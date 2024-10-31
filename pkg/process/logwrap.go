package process

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LeveledLogger struct {
	logger *logrus.Logger
}

func (l LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Trace(makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(makeMsg(msg, keysAndValues...))
}
func (l LeveledLogger) Logger() *logrus.Logger {
	return l.logger
}

func NewLeveledLogger(logger *logrus.Logger) LeveledLogger {
	return LeveledLogger{logger: logger}
}

func makeMsg(msg string, keysAndValues ...interface{}) string {
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
