package golog

import (
	"fmt"
	"time"
)

var _formatFunc func(level Level, ctime time.Time, line, msg string) string

func SetFormatFunc(f func(level Level, ctime time.Time, line, msg string) string) {
	_formatFunc = f
}

func JsonFormat(level Level, ctime time.Time, line, msg string) string {
	return fmt.Sprintf(`{"createTime": "%s","level": "%s",  "line": "%s", "msg": "%s"}`+"\n", ctime.String()[:23], level, line, msg)
}

func defaultFormat(level Level, ctime time.Time, line, msg string) string {
	return fmt.Sprintf(`%s -- [%s] -- %s -- %s`+"\n", ctime.String()[:23], level, line, msg)
}
