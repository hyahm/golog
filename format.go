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

func defaultFormat(ctime time.Time, level, hostname, line, msg string, label map[string]string) string {
	labels := make([]string, 0, len(label))
	if len(label) > 0 {
		for k, v := range label {
			labels = append(labels, fmt.Sprintf(`"%s": "%s"`, k, v))
		}
		return fmt.Sprintf(`{"createTime": "%s", %s, "level": "%s","hostname": "%s", "line": "%s", "msg": "%s"}`+"\n", strings.Join(labels, ","), ctime.String(), level, hostname, line, msg)
	}

	return fmt.Sprintf(`{"createTime": "%s","level": "%s", "hostname": "%s", "line": "%s", "msg": "%s"}`+"\n", ctime.String(), level, hostname, line, msg)
}
