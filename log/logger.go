package log

import (
	"fmt"
	glog "log"
	"time"
)

type Log struct {
	*glog.Logger
}

func (l Log) printLogf(level string, format string, v ...interface{}) {
	l.Println(fmt.Sprintf("%s %s", level, fmt.Sprintf(format, v...)))
}

func (l Log) printLog(level string, v interface{}) {
	l.Print(fmt.Sprintf("%s %s %v", time.Now().String(), level, v))
}

func (l Log) Error(v interface{}) {
	l.printLog("ERR", v)
}

func (l Log) Errorf(format string, v ...interface{}) {
	l.printLogf("ERR", format, v...)
}

func (l Log) Warn(v interface{}) {
	l.printLog("WARN", v)
}

func (l Log) Warnf(format string, v ...interface{}) {
	l.printLogf("WARN", format, v...)
}

func (l Log) Info(v interface{}) {
	l.printLog("INFO", v)
}

func (l Log) Infof(format string, v ...interface{}) {
	l.printLogf("INFO", format, v...)
}

func (l Log) Debug(v interface{}) {
	l.printLog("DEBUG", v)
}

func (l Log) Debugf(format string, v ...interface{}) {
	l.printLogf("DEBUG", format, v...)
}

var _ Logger = Log{}

func NewDefaultLogger() Logger {
	return Log{
		Logger: glog.Default(),
	}
}
