package glog

import (
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"time"
	"unsafe"
)

const DateFormat = "2006-01-02"
const HourFormat = "2006-01-02-15"
const MinuteFormat = "2006-01-02-15-04"
const datetimeFormat = "2006-01-02 15:04:05"

var _ = DateFormat
var _ = HourFormat
var _ = datetimeFormat

type rotation struct {
	size     int64
	maxCount int

	format string

	maxAge time.Duration

	compress bool
}

type FileLogger struct {
	filepath  string
	mu        sync.Mutex
	rotation  rotation
	file      *os.File
	logger    *logrus.Logger
	extra     logrus.Fields
	formatter logrus.Formatter
}

func NewFileLogger(filePath string, logger *logrus.Logger, opts ...FileOption) (fl *FileLogger, err error) {
	fl = &FileLogger{logger: logger}
	fl.filepath, err = filepath.Abs(filePath)
	if err == nil {
		for _, opt := range opts {
			opt(fl)
		}

		err = fl.open()
	}

	if err == nil {
		if fl.logger.Formatter == nil {
			fl.formatter = &logrus.JSONFormatter{}
		} else {
			fl.formatter = fl.logger.Formatter
		}
		fl.logger.SetFormatter(fl)

		if fl.logger.Hooks == nil {
			fl.logger.Hooks = make(logrus.LevelHooks)
		}
		fl.logger.AddHook(fl)

		fl.logger.SetOutput(fl)
	}

	return fl, err
}

func (l *FileLogger) open() (err error) {
	file, err := l.openFile()
	if err == nil {
		prev := l.file
		l.file = file
		if prev == nil {
			// init
		} else {
			// rotation
			_ = prev.Close()
		}
	}
	return
}

func (l *FileLogger) openFile() (file *os.File, err error) {
	err = os.MkdirAll(path.Dir(l.filepath), 0755)
	if err == nil {
		file, err = os.OpenFile(l.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	}
	return
}

func (l *FileLogger) Levels() []logrus.Level {
	return logrus.AllLevels
}

//func (l *FileLogger) Fire(entry *logrus.Entry) error {
//	l.mu.Lock()
//	defer l.mu.Unlock()
//
//	if err := l.rotate(&entry.Time); err != nil {
//		_, _ = fmt.Fprintf(os.Stderr, "rotation failure, error=%v", err)
//	}
//
//	data, err := entry.Bytes() // performance issue !!!
//	if err == nil {
//		_, err = l.file.Write(data)
//		if err == nil {
//			_, err = l.file.WriteString(lf)
//		}
//	}
//	return err
//}

func (l *FileLogger) Format(entry *logrus.Entry) ([]byte, error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(entry.Time.Unix()))
	entry.Buffer.Write(buf[:])
	return l.formatter.Format(entry)
}

func (l *FileLogger) Fire(entry *logrus.Entry) (e error) {
	if len(l.extra) > 0 {
		entry.Data = entry.WithFields(l.extra).Data
	}
	return
}

func (l *FileLogger) Write(data []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	size := unsafe.Sizeof(uint64(0))
	unix := binary.LittleEndian.Uint64(data)
	date := time.Unix(int64(unix), 0)
	//fmt.Println(date.Format(datetimeFormat))

	if e := l.rotate(&date); e != nil {
		_, _ = fmt.Fprintf(os.Stderr, "rotation failure, error=%v", e)
	}

	return l.file.Write(data[size:])
}

func (l *FileLogger) rotationPath(t *time.Time) (file string) {
	file = l.filepath
	if len(l.rotation.format) > 0 {
		file = file + "." + t.Format(l.rotation.format)
	}
	if l.rotation.size > 0 {
		num := 1

		for {
			name := fmt.Sprintf("%s.%d", file, num)
			if _, err := os.Stat(name); err == nil { // exist
				num++
			} else {
				file = name
				break
			}
		}
	}

	return
}

func (l *FileLogger) rotate(t *time.Time) error {
	var (
		format       = l.rotation.format
		size         = l.rotation.size
		timeRotation = false
		sizeRotation = false
	)

	if format == "" && size == 0 {
		return nil
	}

	stat, err := l.file.Stat()
	if err == nil {
		timeRotation = !(format == "" || stat.ModTime().Format(format) == t.Format(format))
		sizeRotation = size > 0 && stat.Size() >= size

		if timeRotation || sizeRotation {
			modTime := stat.ModTime()
			rotationPath := l.rotationPath(&modTime)
			_ = l.file.Close()
			err = os.Rename(l.filepath, rotationPath)
			if err == nil { // rename successfully
				err = l.open()
				if err == nil {
					go l.afterRotate(rotationPath)
				}
			}
		}
	}

	return err
}

func (l *FileLogger) Close() error {
	return l.file.Close()
}

func (l *FileLogger) afterRotate(rotationPath string) {
	if l.rotation.compress {
		// todo compress
		fmt.Println(rotationPath)
	}
	maxAge := l.rotation.maxAge
	maxCount := l.rotation.maxCount
	if maxAge == 0 && maxCount == 0 {
		return
	}
	matches, err := filepath.Glob(l.filepath + "*")
	if err == nil {
		type file struct {
			name string
			time time.Time
		}
		var files []file
		t := time.Now().Add(-maxAge)
		calcCount := maxCount > 0 && len(matches) > maxCount

		for _, name := range matches {
			if stat, e := os.Lstat(name); e == nil {
				if stat.IsDir() {
					continue
				}
				modTime := stat.ModTime()
				if maxAge > 0 && modTime.Before(t) {
					_ = os.Remove(name)
					continue
				}
				if calcCount {
					files = append(files, file{name: name, time: modTime})
				}
			}

		}

		if len(files) > maxCount {
			sort.Slice(files, func(i, j int) bool { // order by desc
				return files[i].time.After(files[j].time)
			})

			for i := maxCount; i < len(files); i++ {
				_ = os.Remove(files[i].name)
			}
		}
	}
}
