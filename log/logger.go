package log

import (
	"fmt"
	glog "log"
	"os"
)

const (
	defaultLogName = ""
	defaultLogFmt  = " [%s] %v\n"
)

type defaultLog struct {
	name string
	*glog.Logger
}

func (l defaultLog) printLogf(level string, format string, v ...interface{}) {
	l.Print(fmt.Sprintf(defaultLogFmt, level, fmt.Sprintf(format, v...)))
}

func (l defaultLog) printLog(level string, v interface{}) {
	l.Print(fmt.Sprintf(defaultLogFmt, level, v))
}

func (l defaultLog) Error(v interface{}) {
	l.printLog("ERROR", v)
}

func (l defaultLog) Errorf(format string, v ...interface{}) {
	l.printLogf("ERROR", format, v...)
}

func (l defaultLog) Warn(v interface{}) {
	l.printLog("WARN", v)
}

func (l defaultLog) Warnf(format string, v ...interface{}) {
	l.printLogf("WARN", format, v...)
}

func (l defaultLog) Info(v interface{}) {
	l.printLog("INFO", v)
}

func (l defaultLog) Infof(format string, v ...interface{}) {
	l.printLogf("INFO", format, v...)
}

func (l defaultLog) Debug(v interface{}) {
	l.printLog("DEBUG", v)
}

func (l defaultLog) Debugf(format string, v ...interface{}) {
	l.printLogf("DEBUG", format, v...)
}

func (l defaultLog) With(name string) Logger {
	if l.name != "" {
		name = fmt.Sprintf("%s.%s", l.name, name)
	}
	return defaultLog{
		name:   name,
		Logger: glog.New(os.Stdout, name+" - ", glog.LstdFlags),
	}
}

func buildDefaultLogger() defaultLog {
	return defaultLog{
		name:   defaultLogName,
		Logger: glog.New(os.Stdout, defaultLogName, glog.LstdFlags),
	}
}

var root Logger = buildDefaultLogger()

func NewLogger(name string) Logger {
	return root.With(name)
}

func SetLogger(logger Logger) {
	root = logger
}
