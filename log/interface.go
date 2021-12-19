package log

type Logger interface {
	Error(interface{})
	Errorf(string, ...interface{})
	Warn(interface{})
	Warnf(string, ...interface{})
	Info(interface{})
	Infof(string, ...interface{})
	Debug(interface{})
	Debugf(string, ...interface{})
	With(string) Logger
}
