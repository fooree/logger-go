package glog

import (
	"github.com/sirupsen/logrus"
	"io"
)

var _ = WithCaller
var _ = WithLevel
var _ = WithHook
var _ = WithWriter
var _ = WithFormatter
var _ = WithFields

type Option func(*Logger)

func WithLevel(level logrus.Level) Option {
	return func(l *Logger) {
		l.logger.Level = level
	}
}

func WithHook(hook logrus.Hook) Option {
	return func(l *Logger) {
		l.logger.Hooks.Add(hook)
	}
}

func WithWriter(w io.Writer) Option {
	return func(l *Logger) {
		l.writers = append(l.writers, w)
	}
}

func WithFormatter(formatter logrus.Formatter) Option {
	return func(l *Logger) {
		l.logger.Formatter = formatter
	}
}

func WithCaller(Caller bool) Option {
	return func(l *Logger) {
		l.logger.ReportCaller = Caller
	}
}

func WithFields(fields logrus.Fields) Option {
	return func(l *Logger) {
		l.fields = fields
	}
}
