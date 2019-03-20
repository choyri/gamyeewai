package gamyeewai

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"sync"
	"time"
)

type logstashHook struct{}

type logEntry struct {
	*zerolog.Logger
	sync.Mutex
	requestData *requestData
}

var (
	l   = zerolog.New(os.Stderr).With().Timestamp().Logger().Hook(logstashHook{})
	Log = logEntry{Logger: &l}
)

func (e *logEntry) WithRequestData(data *requestData) *logEntry {
	e.Lock()
	e.requestData = data
	e.Unlock()
	return e
}

func (h logstashHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("@version", "1").
		Int64("@timestamp", time.Now().Unix()).
		Str("type", "golang").
		Str("channel", getAppName()).
		Str("env", getAppEnv())

	if level < zerolog.ErrorLevel {
		return
	}

	Log.Lock()

	if Log.requestData == nil {
		Log.Unlock()
		return
	}

	key := reflect.TypeOf(*Log.requestData)
	value := reflect.ValueOf(*Log.requestData)

	for i := 0; i < key.NumField(); i++ {
		if key.Field(i).Name == "Message" {
			continue
		}
		e.Str(key.Field(i).Tag.Get("json"), fmt.Sprintf("%v", value.Field(i).Interface()))
	}

	Log.Unlock()
}
