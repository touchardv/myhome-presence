package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

var preferredLogFilenames = [...]string{
	"/var/log/myhome/presence.log",
	"./presence.log",
	"/tmp/presence.log",
}

// SetupLogging initalizes the logger.
func SetupLogging(logLevel string, daemonized bool) func() error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatal(err)
	}

	if daemonized {
		logFile := openLogFile()

		logrus.SetOutput(logFile)
		logrus.SetLevel(level)
		return logFile.Close
	}
	logrus.SetFormatter(&logrus.TextFormatter{DisableLevelTruncation: true, FullTimestamp: true})
	logrus.SetLevel(level)
	return func() error {
		return nil
	}
}

func openLogFile() *os.File {
	for _, filename := range preferredLogFilenames {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err == nil {
			return file
		}
	}
	panic("Failed to create a log file")
}
