package golog

import (
	"fmt"
	"strings"
	"time"
)

var _formatFunc func(ctime time.Time, level, hostname, line, msg string, label map[string]string) string

func SetFormatFunc(f func(ctime time.Time, level, hostname, line, msg string, label map[string]string) string) {
	_formatFunc = f
}

func JsonFormat(ctime time.Time, level, hostname, line, msg string, label map[string]string) string {
	labels := make([]string, 0, len(label))
	if len(label) > 0 {
		for k, v := range label {
			labels = append(labels, fmt.Sprintf(`"%s": "%s"`, k, v))
		}
		return fmt.Sprintf(`{"createTime": "%s", %s, "level": "%s","line": "%s", "msg": "%s"}`+"\n", strings.Join(labels, ","), ctime.String(), level, line, msg)
	}

	return fmt.Sprintf(`{"createTime": "%s","level": "%s",  "line": "%s", "msg": "%s"}`+"\n", ctime.String(), level, line, msg)
}

func defaultFormat(ctime time.Time, level, hostname, line, msg string, label map[string]string) string {
	if len(label) > 0 {
		labels := make([]string, 0, len(label))
		for k, v := range label {
			labels = append(labels, fmt.Sprintf(`"%s": "%s"`, k, v))
		}
		return fmt.Sprintf(`%s -- %s -- [%s] -- %s -- %s`+"\n", ctime.String()[:23], strings.Join(labels, ","), level, line, msg)
	}

	return fmt.Sprintf(`%s -- [%s] -- %s -- %s`+"\n", ctime.String()[:23], level, line, msg)
}
