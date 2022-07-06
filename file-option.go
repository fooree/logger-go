package glog

import (
	"github.com/sirupsen/logrus"
	"time"
)

const (
	KB uint64 = 1024
	MB        = KB * KB
	GB        = MB * KB
	TB        = GB * KB
)

var _ = TB
var _ = WithRotationSize
var _ = WithRotationCount
var _ = WithRotationMaxAge
var _ = WithRotationFormat
var _ = WithRotationSize
var _ = WithFormatter

type FileOption func(*FileLogger)

func WithRotationSize(size uint64) FileOption {
	return func(l *FileLogger) {
		s := int64(size)
		if s > 0 {
			l.rotation.size = s
		}
	}
}

func WithRotationCount(maxCount uint) FileOption {
	return func(l *FileLogger) {
		l.rotation.maxCount = int(maxCount)
	}
}

func WithRotationMaxAge(maxAge time.Duration) FileOption {
	return func(l *FileLogger) {
		l.rotation.maxAge = maxAge
	}
}

func WithRotationFormat(format string) FileOption {
	return func(l *FileLogger) {
		l.rotation.format = format
	}
}

func WithFormatter(formatter logrus.Formatter) FileOption {
	return func(l *FileLogger) {
		l.logger.Formatter = formatter
	}
}

func WithFields(fields logrus.Fields) FileOption {
	return func(l *FileLogger) {
		l.extra = fields
	}
}
