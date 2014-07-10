package logger

import (
	"github.com/cihub/seelog"
)

var (
	globalLog ILogger
)

func init() {
	globalLog, _ = seelog.LoggerFromConfigAsFile("logger.xml")
}

func Tracef(format string, params ...interface{}) {
	if globalLog != nil {
		globalLog.Tracef(format, params)
	}
}

func Debugf(format string, params ...interface{}) {
	if globalLog != nil {
		globalLog.Debugf(format, params)
	}
}

func Infof(format string, params ...interface{}) {
	if globalLog != nil {
		globalLog.Infof(format, params)
	}
}

func Warnf(format string, params ...interface{}) error {
	if globalLog != nil {
		return globalLog.Warnf(format, params)
	}
	return nil
}

func Errorf(format string, params ...interface{}) error {
	if globalLog != nil {
		return globalLog.Errorf(format, params)
	}
	return nil
}

func Criticalf(format string, params ...interface{}) error {
	if globalLog != nil {
		return globalLog.Criticalf(format, params)
	}
	return nil
}

func Trace(v ...interface{}) {
	if globalLog != nil {
		globalLog.Trace(v)
	}
}

func Debug(v ...interface{}) {
	if globalLog != nil {
		globalLog.Debug(v)
	}
}

func Info(v ...interface{}) {
	if globalLog != nil {
		globalLog.Info(v)
	}
}

func Warn(v ...interface{}) error {
	if globalLog != nil {
		return globalLog.Warn(v)
	}
	return nil
}

func Error(v ...interface{}) error {
	if globalLog != nil {
		return globalLog.Error(v)
	}
	return nil
}

func Critical(v ...interface{}) error {
	if globalLog != nil {
		return globalLog.Critical(v)
	}
	return nil
}
