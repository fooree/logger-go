package glog

import (
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestNewFileLogger(t *testing.T) {
	filepath := "test/test.log"
	logg := &logrus.Logger{
		Level: logrus.InfoLevel,
	}
	fileLogger, err := NewFileLogger(filepath,
		logg,
		WithRotationFormat(MinuteFormat),
		WithRotationSize(100),
		WithRotationCount(5),
		WithRotationMaxAge(3*time.Minute),
		WithFields(logrus.Fields{"额外信息": "信息数据"}),
	)
	if err == nil {

		defer func() { _ = fileLogger.Close() }()

		for i := 0; i < 10; i++ {
			logg.Infof("%s file logger", "hello")
			//time.Sleep(time.Second * 3)
		}

	} else {
		t.Errorf("new file logger error: %v", err)
	}

}
