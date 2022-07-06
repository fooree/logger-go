package glog

import (
	"github.com/sirupsen/logrus"
	"io"
)

var _ = WithCaller
var _ = WithLevel
var _ = WithHook
var _ = WithWriter

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

func WithCaller(Caller bool) Option {
	return func(l *Logger) {
		l.logger.ReportCaller = Caller
	}
}
