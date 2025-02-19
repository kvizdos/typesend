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
	mutex     sync.Mutex
}

func (t *TestLogger) Panicf(msg string, args ...any) {
	t.Test.Errorf("PANIC: %s", fmt.Sprintf(msg, args...))
	return
}
func (t *TestLogger) Warnf(msg string, args ...any) {
	t.Test.Errorf("Warn: %s", fmt.Sprintf(msg, args...))
	return
}
func (l *TestLogger) Errorf(format string, v ...any) {
	l.mutex.Lock()
	if l.ErrorLogs == nil {
		l.ErrorLogs = []*string{}
	}
	o := fmt.Sprintf(format, v...)
	l.ErrorLogs = append(l.DebugLogs, &o)
	l.mutex.Unlock()
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
}
