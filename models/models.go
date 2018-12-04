package models

import "cloud.google.com/go/bigquery"

//StoreRecord represents a record in Stores table
type StoreRecord struct {
	StoreCode       string         `json:"store_code"`
	BusinessName    string         `json:"business_name"`
	Address1        string         `json:"address_1"`
	Address2        string         `json:"address_2"`
	City            string         `json:"city"`
	State           string         `json:"state"`
	PostalCode      string         `json:"postal_code"`
	Country         string         `json:"country"`
	PrimaryPhone    string         `json:"primary_phone"`
	Website         string         `json:"website"`
	Description     string         `json:"description"`
	PaymentTypes    string         `json:"payment_types"`
	PrimaryCategory string         `json:"primary_category"`
	Photo           string         `json:"photo"`
	Hours           []*StoreHour   `json:"store_hours"`
	Location        *StoreLocation `json:"location"`
	SapID           string         `json:"sap_id"`
}

//StoreHour represents store hours of operation
type StoreHour struct {
	DayOfWeek string `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
}

//StoreLocation represents store location
type StoreLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

//Save implements ValueSaver interface
func (s *StoreRecord) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"StoreCode":       s.StoreCode,
		"BusinessName":    s.BusinessName,
		"Address1":        s.Address1,
		"Address2":        s.Address2,
		"City":            s.City,
		"State":           s.State,
		"PostalCode":      s.PostalCode,
		"Country":         s.Country,
		"PrimaryPhone":    s.PrimaryPhone,
		"Website":         s.Website,
		"Description":     s.Description,
		"PaymentTypes":    s.PaymentTypes,
		"PrimaryCategory": s.PrimaryCategory,
		"Photo":           s.Photo,
		"Hours":           s.Hours,
		"LOcation":        s.Location,
		"SapID":           s.SapID,
	}, "", nil
}
