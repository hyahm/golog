package golog

import (
	"sync"

	"github.com/fatih/color"
)

var logColor *levelColors

type levelColors struct {
	attrs map[level][]color.Attribute
	mu    *sync.RWMutex
}

func init() {
	logColor = &levelColors{
		mu:    &sync.RWMutex{},
		attrs: make(map[level][]color.Attribute),
	}
	SetColor(ERROR, []color.Attribute{color.FgRed})
	SetColor(WARN, []color.Attribute{color.FgYellow})
	SetColor(DEBUG, []color.Attribute{color.FgGreen})
}

// 设置某级别的颜色
func SetColor(lv level, attrs []color.Attribute) {
	logColor.mu.Lock()
	logColor.attrs[lv] = attrs
	logColor.mu.Unlock()
}

func GetColor(lv level) []color.Attribute {
	logColor.mu.RLock()
	defer logColor.mu.RUnlock()
	if attrs, ok := logColor.attrs[lv]; ok {
		return attrs
	}
	return nil
}

func CleanColor(lv level, attrs []color.Attribute) {
	logColor.mu.Lock()
	delete(logColor.attrs, lv)
	logColor.mu.Unlock()
}
