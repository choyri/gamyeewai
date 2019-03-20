package gamyeewai

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

func disposeErrMsg(err interface{}) (ret string) {
	ret = fmt.Sprintf("%v", err)

	if ret == "" {
		ret = "发生了未知的错误"
	}

	return
}

func disposeKibanaLink(timestamp int64) (ret string) {
	domain := getStringEnv(EnvKeyGamyeewaiKibanaDomain)
	index := getStringEnv(EnvKeyGamyeewaiKibanaIndex)

	if domain == "" || index == "" {
		return
	}

	domain = strings.TrimRight(strings.Trim(domain, ""), "/")

	url := fmt.Sprintf("%s/app/kibana#/discover?_g=(time:(from:now-3d,mode:quick,to:now))&_a=(filters:!((query:(match:(timestamp:(query:'%v'))))),index:'%s')", domain, timestamp, index)

	// 链接里有括号 得转义一下 否则钉钉的阉割版 Markdown 不识别
	url = strings.Replace(url, "(", "%28", -1)
	url = strings.Replace(url, ")", "%29", -1)

	return fmt.Sprintf("[查看详情](%s)", url)
}

func getTrace() string {
	trace := make([]byte, 2048)
	trace = trace[:runtime.Stack(trace, true)]

	return string(trace)
}

func sendDingtalkAlarm(data *requestData) {
	kibanaLink := disposeKibanaLink(data.Timestamp.(int64))

	data.Timestamp = time.Unix(data.Timestamp.(int64), 0).Format("2006-01-02 15:04:05")

	key := reflect.TypeOf(*data)
	value := reflect.ValueOf(*data)

	var content string

	for i := 0; i < key.NumField(); i++ {
		k := key.Field(i).Name
		v := value.Field(i).Interface().(string)

		if k == "RawData" || v == "" {
			continue
		}

		if k == "Trace" {
			tmp := []rune(v)
			v = string(tmp[:len(tmp)/3]) + "…"
		}

		item := []string{"### " + k, "```", v, "```", "---", ""}
		content += strings.Join(item, "\n")
	}

	content += kibanaLink
	println(content)

	Dingtalk.WithSilence().Send(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": data.Message,
			"text":  content,
		},
	})
}
