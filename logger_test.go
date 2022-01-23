package glog

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	name := "test/text.log"
	file, err := os.Create(name)
	if err == nil {
		_ = New(
			WithLevel(logrus.InfoLevel),
			WithWriter(os.Stdout),
			WithWriter(file),
			WithFormatter(new(logrus.JSONFormatter)),
			WithFields(logrus.Fields{"name": "Jack", "project": "auto-math"}),
		).AsStandardLogger()
		logrus.Infof("%s", "hello")
		log.Println("hello world")
		_ = file.Close()
		bytes, err := ioutil.ReadFile(name)
		if err == nil {
			println(string(bytes))
		} else {
			t.Errorf("create %s file error:%v", name, err)
		}
	} else {
		t.Errorf("create %s file error:%v", name, err)
	}

}
