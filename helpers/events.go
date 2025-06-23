package H

import (
	"runtime"
	"time"

	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

type EventArgs map[string]interface{}
type EventFunc func(event_uuid string, args EventArgs)
type EventData struct {
	File  string
	Line  int
	Event EventFunc
}

type ListenerData struct {
	db     *gorm.DB
	events map[string][]EventData
	logger *echo.Logger
}

func (l *ListenerData) Load(logger *echo.Logger) (err error) {
	l.db = db
	l.logger = logger
	l.events = make(map[string][]EventData)
	l.AddListener("mail.send", func(event_uuid string, args EventArgs) {

	})
	return nil
}

func (l *ListenerData) AddListener(name string, fn EventFunc) {
	_, file, line, _ := runtime.Caller(1)
	if _, ok := l.events[name]; !ok {
		l.events[name] = make([]EventData, 0)
	}
	l.events[name] = append(l.events[name], EventData{file, line, fn})
	(*l.logger).Debugf("Added Listener: %s-%d File: %s[%d]", name, len(l.events[name])-1, file, line)
}

func (l *ListenerData) Fire(event_name string, args EventArgs) {
	_, file_fire, line_fire, _ := runtime.Caller(1)
	if list_fn, ok := l.events[event_name]; ok {
		for idx, data := range list_fn {
			go func(idx int, data EventData, args EventArgs) {
				random_uuid := NewUUID()
				(*l.logger).Debugf("Init Fire Event: %s[%d] - uuid: %s - caller in: %s[%d] - caller from: %s[%d]\n", event_name, idx, random_uuid, data.File, data.Line, file_fire, line_fire)
				start := time.Now().UTC()
				go data.Event(random_uuid, args)
				elapsed := time.Since(start)
				(*l.logger).Debugf("End Fire Event %s[%d] - uuid: [%s] in %s\n", event_name, idx, random_uuid, elapsed)
			}(idx, data, args)
		}
	} else {
		(*l.logger).Fatalf("Event '%s' not found: caller from: %s[%d]", event_name, file_fire, line_fire)
	}
}

var Listener ListenerData
