package testutils

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

type TestLogger struct {
	log.Logger
	Test *testing.T

	DebugLogs []*string
	InfoLogs  []*string
	ErrorLogs []*string
	WarnLogs  []*string

	DoLog bool
	mutex sync.Mutex
}

func (t *TestLogger) Panicf(msg string, args ...any) {
	t.Test.Errorf("PANIC: %s", fmt.Sprintf(msg, args...))
	return
}
func (l *TestLogger) Warnf(format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.WarnLogs == nil {
		l.WarnLogs = []*string{}
	}
	o := fmt.Sprintf(format, args...)
	l.WarnLogs = append(l.WarnLogs, &o)
	if l.DoLog {
		l.Test.Logf("ERROR: %s", o)
	}
	return
}
func (t *TestLogger) Printf(msg string, args ...any) {
	t.Test.Logf(msg, args...)
	return
}
func (l *TestLogger) Errorf(format string, v ...any) {
	l.mutex.Lock()
	if l.ErrorLogs == nil {
		l.ErrorLogs = []*string{}
	}
	o := fmt.Sprintf(format, v...)
	l.ErrorLogs = append(l.ErrorLogs, &o)
	l.mutex.Unlock()
	if l.DoLog {
		l.Test.Logf("ERROR: %s", o)
	}
	return
}
func (l *TestLogger) Infof(format string, v ...any) {
	l.mutex.Lock()
	if l.DebugLogs == nil {
		l.InfoLogs = []*string{}
	}
	o := fmt.Sprintf(format, v...)
	l.InfoLogs = append(l.InfoLogs, &o)
	l.mutex.Unlock()
	if l.DoLog {
		l.Test.Logf("INFO: %s", o)
	}
	return
}

func (l *TestLogger) Debugf(format string, v ...interface{}) {
	l.mutex.Lock()
	if l.DebugLogs == nil {
		l.DebugLogs = []*string{}
	}
	o := fmt.Sprintf(format, v...)
	l.DebugLogs = append(l.DebugLogs, &o)
	l.mutex.Unlock()
	if l.DoLog {
		l.Test.Logf("DEBUG: %s", o)
	}
}
