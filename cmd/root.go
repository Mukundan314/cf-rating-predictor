package cmd

import (
	"time"

	"github.com/fsnotify/fsnotify"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cheran-senthil/cf-rating-predictor/db"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "cf-rating-predictor",
		Short: "Server for cf-rating-predictor",
		Run:   run,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.cf-rating-predictor.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			logrus.WithError(err).Fatal()
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".cf-rating-predictor")
	}

	viper.SetEnvPrefix("cf_rating_predictor")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		logrus.WithError(err).Info("No config file found")
	} else {
		if err == nil {
			logrus.WithField("configFile", viper.ConfigFileUsed()).Info("Using config file")
		} else {
			logrus.WithError(err).Error("Error when reading config")
		}

		viper.WatchConfig()
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithField("configFile", e.Name).Info("Config file changed")
	})

}

func run(cmd *cobra.Command, args []string) {
	logrus.SetLevel(logrus.DebugLevel)

	d := db.NewDB(time.Hour * 1)

	for _ = range time.NewTicker(time.Minute).C {
		d.Update()
	}
}
