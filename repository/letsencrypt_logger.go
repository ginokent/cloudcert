package repository

import (
	"fmt"
	"io"
	"strings"

	"github.com/rec-logger/rec.go"
)

// legoStdLogger is an alternative logger that satisfies the log.StdLogger interface.
type legoStdLogger struct {
	logger *rec.Logger
}

// DefaultLegoStdLogger returns LegoLogger that used as default in letsencrypt package.
func newLegoStdLogger(writer io.Writer) *legoStdLogger {
	const callerSkipForLegoPackage = 3

	return &legoStdLogger{
		logger: rec.Must(rec.New(writer, rec.WithCallerSkip(callerSkipForLegoPackage))),
	}
}

// GetInternalLogger returns internal logger in LegoStdLogger.
func (l *legoStdLogger) GetInternalLogger() *rec.Logger {
	return l.logger
}

// SetInternalLogger sets internal logger in LegoStdLogger.
func (l *legoStdLogger) SetInternalLogger(logger *rec.Logger) *legoStdLogger {
	l.logger = logger

	return l
}

// Fatal writes a log entry.
func (l *legoStdLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	panic(args)
}

// Fatalln writes a log entry.
func (l *legoStdLogger) Fatalln(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	panic(args)
}

// Fatalf writes a log entry.
func (l *legoStdLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
	panic(args)
}

// Print writes a log entry.
func (l *legoStdLogger) Print(args ...interface{}) {
	message := fmt.Sprint(args...)

	l.getLogger(message)(message)
}

// Println writes a log entry.
func (l *legoStdLogger) Println(args ...interface{}) {
	message := fmt.Sprint(args...)

	l.getLogger(message)(message)
}

// Printf writes a log entry.
func (l *legoStdLogger) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	l.getLogger(message)(message)
}

func (l *legoStdLogger) getLogger(message string) (logger func(msg string, fields ...rec.Field)) {
	const (
		legoLogInfoPrefix = "[INFO]"
		legoLogWarnPrefix = "[WARN]"
	)

	switch {
	case strings.HasPrefix(message, legoLogInfoPrefix):
		return l.logger.Info
	case strings.HasPrefix(message, legoLogWarnPrefix):
		return l.logger.Warning
	default:
		return l.logger.Info
	}
}
