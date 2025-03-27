package types

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LeveledLogger struct {
	logger *logrus.Logger
}

type ThreadLogger struct {
	logger *logrus.Logger
	Thread int
}

func (l LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(l.makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(l.makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	if msg == "performing request" {
		return
	}
	l.logger.Debug(l.makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(l.makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Logger() *logrus.Logger {
	return l.logger
}
func (l LeveledLogger) makeMsg(msg string, keysAndValues ...interface{}) string {
	ret := "-> " + msg
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			ret += fmt.Sprintf(" %v", keysAndValues[i+1])
		} else {
			ret += fmt.Sprintf(" %v", keysAndValues[i])
		}
	}
	return ret
}

func NewLeveledLogger(logger *logrus.Logger) LeveledLogger {
	return LeveledLogger{logger: logger}
}

func (l ThreadLogger) Errorf(msg string, keysAndValues ...interface{}) {
	l.logger.Error(l.makeMsgf(msg, keysAndValues...))
}
func (l ThreadLogger) Error(msg string) {
	l.logger.Errorf(l.makeMsg(msg))
}

func (l ThreadLogger) Infof(msg string, keysAndValues ...interface{}) {
	l.logger.Info(l.makeMsgf(msg, keysAndValues...))
}
func (l ThreadLogger) Info(msg string) {
	l.logger.Info(l.makeMsg(msg))
}

func (l ThreadLogger) Debugf(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(l.makeMsgf(msg, keysAndValues...))
}
func (l ThreadLogger) Debug(msg string) {
	l.logger.Debug(l.makeMsg(msg))
}

func (l ThreadLogger) Tracef(msg string, keysAndValues ...interface{}) {
	l.logger.Trace(l.makeMsgf(msg, keysAndValues...))
}
func (l ThreadLogger) Trace(msg string) {
	l.logger.Trace(l.makeMsg(msg))
}

func (l ThreadLogger) Warnf(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(l.makeMsgf(msg, keysAndValues...))
}
func (l ThreadLogger) Warn(msg string) {
	l.logger.Warn(l.makeMsg(msg))
}

func (l ThreadLogger) Logger() *logrus.Logger {
	return l.logger
}
func (l ThreadLogger) makeMsgf(msg string, keysAndValues ...interface{}) string {
	origMsg := fmt.Sprintf(msg, keysAndValues...)
	return fmt.Sprintf("[T%d] %v", l.Thread, origMsg)
}
func (l ThreadLogger) makeMsg(msg string) string {
	return fmt.Sprintf("[T%d] %v", l.Thread, msg)
}

func (l ThreadLogger) GetLogger() *logrus.Logger {
	return l.logger
}

func NewThreadLogger(logger *logrus.Logger, thread int) ThreadLogger {
	return ThreadLogger{logger: logger, Thread: thread}
}
