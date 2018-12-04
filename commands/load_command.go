package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/storefinder/cli/models"

	log "github.com/sirupsen/logrus"
)

var (
	dataFile, datasetID, tableName, home string

	errMissingParamDatafile  = errors.New("Data file is missing or does not exist")
	errMissingParamDataSet   = errors.New("Dataset is missing")
	errMissingParamTableName = errors.New("Table Name is missing")

	loadCmd = &cobra.Command{
		Use:   "load",
		Short: "load command",
		Long: `Loads store data to a bigquery dataset 
	from an input data file 
	`,
		RunE: loadDataIntoBigQuery,
	}

	exists = func(filename string) (os.FileInfo, error) {
		return os.Stat(filename)
	}
)

func init() {
	log.Info("Initializing load command")
	loadCmd.Flags().StringVarP(&dataFile, "file", "f", "", "Store data file to load from")
	loadCmd.Flags().StringVarP(&datasetID, "dataset", "d", "", "Google BigQuery Dataset ID")
	loadCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table Name in Google BigQuery ")

	home, _ = homedir.Dir()
	rootCmd.AddCommand(loadCmd)
}

func loadDataIntoBigQuery(cmd *cobra.Command, args []string) error {
	var stores []models.StoreRecord

	if len(dataFile) == 0 {
		return errMissingParamDatafile
	}
	if _, err := exists(dataFile); err != nil {
		return errMissingParamDatafile
	}
	if len(datasetID) == 0 {
		return errMissingParamDataSet
	}
	if len(tableName) == 0 {
		return errMissingParamTableName
	}

	projectID := viper.GetString("project")
	credFile := viper.GetString("credentials")
	batchSize := viper.GetInt("batchsize")
	topic := viper.GetString("topic")

	log.Infof("Project ID: %s Credentials: %s Batch Size: %v Topic: %s", projectID, credFile, batchSize, topic)

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(home+"/.storelocator/"+credFile))
	if err != nil {
		return err
	}
	pubSubClient, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(home+"/.storelocator/"+credFile))
	if err != nil {
		return err
	}

	//Create Dataaset
	if err := createDataSet(client); err != nil {
		return err
	}

	//Create table
	if err := createTable(client); err != nil {
		return err
	}

	//insert rows
	var firstLine = true
	buffer, _ := ioutil.ReadFile(dataFile)
	csvPayLoad := string(buffer)
	reader := csv.NewReader(strings.NewReader(csvPayLoad))
	for {
		line, err := reader.Read()
		if firstLine {
			firstLine = false
			continue //skip header
		}
		if err == io.EOF {
			log.Info("End of file detected, breaking the loop...")
			break
		} else {
			lat, _ := strconv.ParseFloat(line[15], 64)
			lon, _ := strconv.ParseFloat(line[16], 64)

			stores = append(stores, models.StoreRecord{
				StoreCode:       line[0],
				BusinessName:    line[1],
				Address1:        line[2],
				Address2:        line[3],
				City:            line[4],
				State:           line[5],
				PostalCode:      line[6],
				Country:         line[7],
				PrimaryPhone:    line[8],
				Website:         line[9],
				Description:     line[10],
				PaymentTypes:    line[11],
				PrimaryCategory: line[12],
				Photo:           line[13],
				Hours:           parseHours(line),
				Location: &models.StoreLocation{
					Latitude:  lat,
					Longitude: lon,
				},
				SapID: line[17],
			})
		}
	}
	var batch []models.StoreRecord
	for i := 0; i < len(stores); i += batchSize {
		end := i + batchSize

		if end > len(stores) {
			end = len(stores)
		}
		batch = stores[i:end]
		log.Infof("Inserting %v stores into Google Big Query", len(batch))
		insertRows(client, batch)
		log.Infof("Publishing message to topic %s for adding the store records to elastic search index", topic)
		publishMessage(pubSubClient, topic, batch)
	}
	log.Info("Done adding stores into Google Big Query")
	return nil
}

func parseHours(record []string) []*models.StoreHour {
	var hours []*models.StoreHour

	dayOfWeek := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	hours = append(hours, parseHour(dayOfWeek[0], record[18]))
	hours = append(hours, parseHour(dayOfWeek[1], record[19]))
	hours = append(hours, parseHour(dayOfWeek[2], record[20]))
	hours = append(hours, parseHour(dayOfWeek[3], record[21]))
	hours = append(hours, parseHour(dayOfWeek[4], record[22]))
	hours = append(hours, parseHour(dayOfWeek[5], record[23]))
	hours = append(hours, parseHour(dayOfWeek[6], record[24]))

	return hours
}

func parseHour(dayOfWeek string, input string) *models.StoreHour {
	var openTime string
	var closeTime string

	if len(input) == 0 {
		openTime = " "
		closeTime = "CLOSED"
	} else {
		tokens := strings.Split(input, "-")
		if len(tokens) == 2 {
			openTime = tokens[0]
			closeTime = tokens[1]
		}
	}

	hour := models.StoreHour{
		DayOfWeek: dayOfWeek,
		OpenTime:  openTime,
		CloseTime: closeTime,
	}

	return &hour
}

func createDataSet(client *bigquery.Client) error {
	log.Info("Creating dataset")
	ctx := context.Background()
	meta := &bigquery.DatasetMetadata{
		Location: "US", // Create the dataset in the US.
	}
	err := client.Dataset(datasetID).Create(ctx, meta)
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusConflict {
		log.Infof("Dataset %s already exists", datasetID)
	} else {
		log.Infof("Dataset %s is created.", datasetID)
	}
	return nil
}

func createTable(client *bigquery.Client) error {
	ctx := context.Background()
	schema := bigquery.Schema{
		{Name: "StoreCode", Type: bigquery.StringFieldType},
		{Name: "BusinessName", Type: bigquery.StringFieldType},
		{Name: "Address1", Type: bigquery.StringFieldType},
		{Name: "Address2", Type: bigquery.StringFieldType},
		{Name: "City", Type: bigquery.StringFieldType},
		{Name: "State", Type: bigquery.StringFieldType},
		{Name: "PostalCode", Type: bigquery.StringFieldType},
		{Name: "Country", Type: bigquery.StringFieldType},
		{Name: "PrimaryPhone", Type: bigquery.StringFieldType},
		{Name: "Website", Type: bigquery.StringFieldType},
		{Name: "Description", Type: bigquery.StringFieldType},
		{Name: "PaymentTypes", Type: bigquery.StringFieldType},
		{Name: "PrimaryCategory", Type: bigquery.StringFieldType},
		{Name: "Photo", Type: bigquery.StringFieldType},
		{Name: "Hours",
			Type:     bigquery.RecordFieldType,
			Repeated: true,
			Schema: bigquery.Schema{
				{Name: "DayOfWeek", Type: bigquery.StringFieldType},
				{Name: "OpenTime", Type: bigquery.StringFieldType},
				{Name: "CloseTime", Type: bigquery.StringFieldType},
			},
		},
		{Name: "Location",
			Type: bigquery.RecordFieldType,
			Schema: bigquery.Schema{
				{Name: "Latitude", Type: bigquery.FloatFieldType},
				{Name: "Longitude", Type: bigquery.FloatFieldType},
			},
		},
		{Name: "SapID", Type: bigquery.StringFieldType},
	}
	metaData := &bigquery.TableMetadata{
		Schema: schema,
	}
	table := client.Dataset(datasetID).Table(tableName)
	err := table.Create(ctx, metaData)
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusConflict {
		log.Infof("Table already exists %s", tableName)
	} else {
		log.Infof("Table %v created ", tableName)
	}
	return nil
}

func insertRows(client *bigquery.Client, rows []models.StoreRecord) error {
	ctx := context.Background()

	uploader := client.Dataset(datasetID).Table(tableName).Uploader()
	if err := uploader.Put(ctx, rows); err != nil {
		log.Infof("Some errors occured inserting records into Google BigQuery", err)
		return err
	}
	return nil
}

// this function publishes a Google Pub Sub notification message for the Knative service
// to pick up via Eventing to add to Elastic Search Index
func publishMessage(client *pubsub.Client, topic string, rows []models.StoreRecord) error {
	ctx := context.Background()

	//serialize storerecord collection
	payload, err := json.Marshal(rows)
	if err != nil {
		log.Info("Error serializing StoreRecord array")
	}
	t := client.Topic(topic)
	msg := pubsub.Message{
		Data: payload,
	}
	result := t.Publish(ctx, &msg)
	msgID, err := result.Get(ctx)
	if err != nil {
		return err
	}
	log.Infof("Published a message to topic %s Message ID is %s", topic, msgID)
	return nil
}
