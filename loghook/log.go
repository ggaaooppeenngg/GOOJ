package loghook

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

type CallerHook struct {
}

func NewCallerHook() *CallerHook {
	return &CallerHook{}
}

func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	entry.Data["caller"] = hook.caller()
	return nil
}

func (hook *CallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *CallerHook) caller() string {
	if _, file, line, ok := runtime.Caller(7); ok {
		return strings.Join([]string{file, strconv.Itoa(line)}, ":")
	}
	if _, file, line, ok := runtime.Caller(5); ok {
		return strings.Join([]string{file, strconv.Itoa(line)}, ":")
	}

	return "???"
}
