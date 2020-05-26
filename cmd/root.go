package cmd

import (
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/cors"
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
	rootCmd.PersistentFlags().StringP("addr", "a", ":8080", "Address to serve website.")
	rootCmd.PersistentFlags().DurationP("update-interval", "i", time.Minute, "")
	rootCmd.PersistentFlags().DurationP("update-rating-before", "", time.Hour, "Duration before contest to update rating.")
	rootCmd.PersistentFlags().DurationP("update-rating-changes-after", "", 24*time.Hour, "Duration after contest to update rating changes.")
	rootCmd.PersistentFlags().DurationP("clear-rating-changes-after", "", 24*time.Hour, "Duration after contest to clear rating changes from cache.")

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

		if viper.GetBool("verbose") {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
	})

}

func run(cmd *cobra.Command, args []string) {
	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	c := cache.NewCache()

	go func() {
		for range time.NewTicker(viper.GetDuration("update-interval")).C {
			if err := c.Update(
				viper.GetDuration("update-rating-before"),
				viper.GetDuration("update-rating-changes-after"),
				viper.GetDuration("clear-rating-changes-after"),
			); err != nil {
				logrus.WithError(err).Error()
			}
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/api/contest.ratingChanges", api.RatingChangesHandler{Cache: c})

	handler := cors.Default().Handler(mux)
	logrus.Fatal(http.ListenAndServe(viper.GetString("addr"), handler))
}
