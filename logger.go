package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

var logLevelMap = map[string]logrus.Level{
	"trace": logrus.TraceLevel,
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
}

func setupLogger(logLevel string) error {
	level, ok := logLevelMap[logLevel]
	if !ok {
		return fmt.Errorf("Invalid log level: %s", logLevel)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return nil
}
