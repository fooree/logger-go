package glog

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewFileLogger(t *testing.T) {
	filepath := "test/test.log"
	fileLogger, err := NewFileLogger(filepath,
		WithRotationFormat(MinuteFormat),
		WithRotationSize(20<<1),
		WithRotationCount(5),
		WithRotationMaxAge(3*time.Minute),
	)
	if err == nil {

		defer func() { _ = fileLogger.Close() }()
		logger := New(
			WithWriter(os.Stderr),
			WithHook(fileLogger),
		)
		for i := 0; i < 10; i++ {
			logger.InfoF(context.Background(), "%s file logger", "hello")
			time.Sleep(time.Minute)
		}

	} else {
		t.Errorf("new file logger error: %w", err)
	}

}
