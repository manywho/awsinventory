package main

import "github.com/sirupsen/logrus"

var logger *logrus.Logger

func initLogger() {
	logger = logrus.New()

	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
		break
	case "info":
		logger.SetLevel(logrus.InfoLevel)
		break
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
		break
	case "warning":
		logger.SetLevel(logrus.WarnLevel)
	default:
		logger.SetLevel(logrus.WarnLevel)
		logger.Warning("unknown log level, using default (warning)")
	}
}
