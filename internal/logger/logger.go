package logger

import (
	"time"

	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	})
	l.SetLevel(logrus.TraceLevel)
	l.Info("Test")
	return l
}
