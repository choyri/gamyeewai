package gamyeewai

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"runtime"
	"time"
)

type logstashHook struct{}

type logEntry struct {
	*zerolog.Logger
	withGin bool
}

var l = zerolog.New(os.Stderr).With().Timestamp().Logger().Hook(logstashHook{})

var Log = logEntry{&l, false}

func (h logstashHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("@version", "1").
		Int64("@timestamp", time.Now().Unix()).
		Str("type", "golang").
		Str("channel", getAppName()).
		Str("env", getAppEnv())

	if level < zerolog.ErrorLevel {
		return
	}

	trace := make([]byte, 2048)
	trace = trace[:runtime.Stack(trace, true)]
	e.Str("trace", string(trace))

	if !Log.withGin || Gin.Context == nil {
		return
	}

	// 用完后改回默认值
	Log.withGin = false

	requestData := Gin.parseRequest()

	key := reflect.TypeOf(requestData)
	value := reflect.ValueOf(requestData)

	for i := 0; i < key.NumField(); i++ {
		if key.Field(i).Name == "Message" {
			continue
		}
		e.Str(key.Field(i).Tag.Get("json"), fmt.Sprintf("%v", value.Field(i).Interface()))
	}
}

func (e *logEntry) WithGin() *logEntry {
	e.withGin = true
	return e
}
