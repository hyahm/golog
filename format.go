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
	// labels := make([]string, 0, len(label))
	// if len(label) > 0 {
	// 	for k, v := range label {
	// 		labels = append(labels, fmt.Sprintf(`"%s": "%s"`, k, v))
	// 	}
	// 	return fmt.Sprintf(`{"createTime": "%s", %s, "level": "%s","line": "%s", "msg": "%s"}`+"\n", strings.Join(labels, ","), ctime.String(), level, line, msg)
	// }

	return fmt.Sprintf(`{"createTime": "%s","level": "%s",  "line": "%s", "msg": "%s"}`+"\n", ctime.String(), level, line, msg)
}

func defaultFormat(level Level, ctime time.Time, line, msg string) string {
	// if len(label) > 0 {

	// 	labels := make([]string, 0, len(label))
	// 	for k, v := range label {
	// 		labels = append(labels, fmt.Sprintf(`"%s": "%s"`, k, v))
	// 	}
	// 	return fmt.Sprintf(`%s -- %s -- [%s] -- %s -- %s`+"\n", ctime.String()[:23], strings.Join(labels, ","), level, line, msg)
	// }

	return fmt.Sprintf(`%s -- [%s] -- %s -- %s`+"\n", ctime.String()[:23], level, line, msg)
}
