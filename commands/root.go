package commands

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var rootCmd = &cobra.Command{
	Use:   "storelocator",
	Short: "Simple CLI for storelocator service",
	Long: `A simple command line application
		for storelocator service
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
var cfgFile, project string

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("project", "p", "", "Google CLoud Project")
	rootCmd.PersistentFlags().StringP("credentials", "c", "", "Google Cloud Credentials")
	rootCmd.PersistentFlags().StringP("topic", "q", "", "Google pubsub topic")
	rootCmd.PersistentFlags().Int32P("batchSize", "b", 10, "Batch size for Google Bigquery Inserts")

	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.BindPFlag("credentials", rootCmd.PersistentFlags().Lookup("credentials"))
	viper.BindPFlag("topic", rootCmd.PersistentFlags().Lookup("topic"))
	viper.BindPFlag("batchSize", rootCmd.PersistentFlags().Lookup("batchSize"))
}

//Execute executes the root command
func Execute() {
	rootCmd.SilenceErrors = false
	rootCmd.SilenceUsage = false

	if err := rootCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetConfigType("yaml")
	if len(cfgFile) > 0 {
		log.Info("Config file :", cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		if home, err := homedir.Dir(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			log.Info("Home directory: ", home)
			viper.AddConfigPath(home + "/.storelocator")
			viper.SetConfigName("config")
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}
