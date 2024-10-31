package process

import "github.com/sirupsen/logrus"

type LeveledLogger struct {
	logger *logrus.Logger
}

func (l LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorf(msg, keysAndValues...)
}
func (l LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infof(msg, keysAndValues...)
}
func (l LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debugf(msg, keysAndValues...)
}
func (l LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warnf(msg, keysAndValues...)
}
func (l LeveledLogger) Logger() *logrus.Logger {
	return l.logger
}

func NewLeveledLogger(logger *logrus.Logger) LeveledLogger {
	return LeveledLogger{logger: logger}
}
