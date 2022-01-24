package glog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const lf = "\n"
const DateFormat = "2006-01-02"
const HourFormat = "2006-01-02-15"
const MinuteFormat = "2006-01-02-15-04"

var _ = DateFormat
var _ = HourFormat

type rotation struct {
	size     int64
	maxCount int

	format string

	maxAge time.Duration

	compress bool
	num      uint
}

type FileLogger struct {
	filepath string
	mu       sync.Mutex
	fd       *os.File
	rotation rotation
}

func NewFileLogger(filePath string, opts ...FileOption) (l *FileLogger, err error) {
	l = &FileLogger{}
	l.filepath, err = filepath.Abs(filePath)
	if err == nil {
		for _, opt := range opts {
			opt(l)
		}

		err = l.openFile()
	}

	return l, err
}

func (l *FileLogger) openFile() (err error) {
	file, err := l.open()
	if err == nil {
		l.set(file)
	}
	return
}

func (l *FileLogger) open() (file *os.File, err error) {
	err = os.MkdirAll(path.Dir(l.filepath), 0755)
	if err == nil {
		file, err = os.OpenFile(l.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	}
	return
}

func (l *FileLogger) set(fd *os.File) {
	prev := l.fd
	if prev == nil {
		// init
	} else { // rotation
		defer func() { _ = prev.Close() }()
	}
	l.fd = fd
	return
}

func (l *FileLogger) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *FileLogger) Fire(entry *logrus.Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.rotate(&entry.Time); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "rotation failure, error=%v", err)
	}

	data, err := entry.Bytes() // performance issue !!!
	if err == nil {
		_, err = l.fd.Write(data)
		if err == nil {
			_, err = l.fd.WriteString(lf)
		}
	}
	return err
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

	stat, err := l.fd.Stat()
	if err == nil {
		timeRotation = !(format == "" || stat.ModTime().Format(format) == t.Format(format))
		sizeRotation = size > 0 && stat.Size() >= size

		if timeRotation || sizeRotation {
			modTime := stat.ModTime()
			rotationPath := l.rotationPath(&modTime)
			err = os.Rename(l.filepath, rotationPath)
			if err == nil { // rename successfully
				err = l.openFile()
				if err == nil {
					go l.afterRotate(rotationPath)
				}
			}
		}
	}

	return err
}

func (l *FileLogger) Close() error {
	return l.fd.Close()
}

func (l *FileLogger) afterRotate(rotationPath string) {
	if l.rotation.compress {
		// todo compress
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
