package glog

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

type Extension interface {
	logrus.Hook
	io.Writer
}

type Logger struct {
	logger  *logrus.Logger
	writers []io.Writer
	fields  logrus.Fields
}

func New(opts ...Option) *Logger {
	l := &Logger{
		logger:  logrus.New(),
		writers: make([]io.Writer, 0, 8),
	}

	l.logger.Hooks.Add(l)

	for _, opt := range opts {
		opt(l)
	}

	l.logger.Out = l

	return l
}

func (l *Logger) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *Logger) Fire(entry *logrus.Entry) error {
	if len(l.fields) > 0 {
		// local variable
		newEntry := entry.WithFields(l.fields)
		entry.Data = newEntry.Data
	}
	return nil
}

func (l *Logger) Write(message []byte) (n int, err error) {
	for i := 0; i < len(l.writers); i++ {
		if _, e := l.writers[i].Write(message); e == nil {
			continue
		}
		_, _ = fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
	return
}

func (l *Logger) AsStandardLogger() *Logger {
	logrus.SetOutput(l)
	logrus.SetLevel(l.logger.Level)
	logrus.SetFormatter(l.logger.Formatter)
	logrus.SetReportCaller(l.logger.ReportCaller)
	logrus.StandardLogger().Hooks = l.logger.Hooks
	logrus.StandardLogger().ExitFunc = l.logger.ExitFunc
	return l
}

func (l *Logger) InfoF(_ context.Context, format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}
