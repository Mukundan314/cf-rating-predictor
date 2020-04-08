package cmd

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
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
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".cf-rating-predictor")
	}

	viper.SetEnvPrefix("cf_rating_predictor")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logrus.WithField("configFile", viper.ConfigFileUsed()).Info("Using config file")
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithField("configFile", e.Name).Info("Config file changed")
	})
}

func run(cmd *cobra.Command, args []string) {
}
