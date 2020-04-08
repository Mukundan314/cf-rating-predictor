package main

import (
	"github.com/cheran-senthil/cf-rating-predictor/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.WithError(err).Fatal()
	}
}
