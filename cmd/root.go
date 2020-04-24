package cmd

import (
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cheran-senthil/cf-rating-predictor/api"
	"github.com/cheran-senthil/cf-rating-predictor/cache"
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

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default $HOME/.cf-rating-predictor.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enables more verbose logging.")
	rootCmd.PersistentFlags().DurationP("update-interval", "i", time.Minute, "")
	rootCmd.PersistentFlags().DurationP("update-rating-before", "r", time.Hour, "Time before contest to update rating")

	viper.BindPFlags(rootCmd.PersistentFlags())
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
	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	c := cache.NewCache(viper.GetDuration("update-rating-before"))

	go func() {
		if err := c.Update(); err != nil {
			logrus.WithError(err).Error()
		}

		for range time.NewTicker(viper.GetDuration("update-interval")).C {
			if err := c.Update(); err != nil {
				logrus.WithError(err).Error()
			}
		}
	}()

	http.Handle("/api/contest.ratingChanges", api.RatingChangesHandler{Cache: c})
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
