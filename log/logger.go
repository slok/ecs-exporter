package log

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type customLogger struct {
	*logrus.Logger
	sync.Mutex
}

// global logger
var logger = customLogger{
	Logger: logrus.New(),
	Mutex:  sync.Mutex{},
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Level are the logging levels
type Level logrus.Level

// log levels
const (
	DebugLevel = Level(logrus.DebugLevel)
	InfoLevel  = Level(logrus.InfoLevel)
	WarnLevel  = Level(logrus.WarnLevel)
	ErrorLevel = Level(logrus.ErrorLevel)
	FatalLevel = Level(logrus.FatalLevel)
	PanicLevel = Level(logrus.PanicLevel)
)

// SetLevel sets log level
func SetLevel(level Level) {
	logger.Lock()
	defer logger.Unlock()
	logger.Level = logrus.Level(level)
}
