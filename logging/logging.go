package logging

import (
	"container/list"
	"fmt"
	"github.com/chenxing/cangshan/application"
	"sync"
)

func init() {
	application.RegisterModuleCreater("Logging",
		func() interface{} {
			return new(Logging)
		})
}

var (
	globalLogging *Logging
	caches        = make(map[string]*list.List)
	flushMutex    sync.Mutex
)

type Logging struct {
	Handlers []*Handler
	handlers map[string][]*Handler
}

func (log *Logging) Initialize() error {
	log.handlers = make(map[string][]*Handler)
	for _, handler := range log.Handlers {
		for _, level := range handler.Levels {
			hs := log.handlers[level]
			if hs == nil {
				hs = make([]*Handler, 0, 1)
			}
			log.handlers[level] = append(hs, handler)
		}
	}
	globalLogging = log
	Flush()
	return nil
}

func Log(level string, format string, params ...interface{}) {
	LogSkip(2, level, format, params...)
}

func Debug(format string, params ...interface{}) {
	LogSkip(2, "debug", format, params...)
}

func Info(format string, params ...interface{}) {
	LogSkip(2, "info", format, params...)
}

func Warn(format string, params ...interface{}) {
	LogSkip(2, "warn", format, params...)
}

func Error(format string, params ...interface{}) {
	LogSkip(2, "error", format, params...)
}

func Fatal(format string, params ...interface{}) {
	LogSkip(2, "fatal", format, params...)
}

func Flush() {
	if globalLogging == nil {
		globalLogging = createDefaultLogging()
	}
	flushMutex.Lock()
	defer flushMutex.Unlock()
	for level, cache := range caches {
		for e := cache.Front(); e != nil; e = e.Next() {
			for _, handler := range globalLogging.handlers[level] {
				handler.write(e.Value.(*event))
			}
		}
	}
	caches = make(map[string]*list.List)
}

func LogSkip(skip int, level string, format string, params ...interface{}) {
	LogSkipWithAttr(skip+1, level, nil, format, params...)
}

func LogSkipWithAttr(skip int, level string, attr map[string]string, format string, params ...interface{}) {
	if globalLogging != nil && len(globalLogging.handlers[level]) == 0 {
		fmt.Println("not ready")
		return
	}
	e := newEvent(skip+1, level, attr, format, params...)
	if globalLogging == nil {
		cache := caches[level]
		if cache == nil {
			cache = list.New()
			caches[level] = cache
		}
		cache.PushBack(e)
	} else {
		for _, handler := range globalLogging.handlers[level] {
			handler.write(e)
		}
	}
}

func createDefaultLogging() *Logging {
	log := &Logging{
		Handlers: []*Handler{createDefaultHandler()},
	}
	log.Initialize()
	return log
}