package gamyeewai

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const dingtalkHost = "https://oapi.dingtalk.com/robot/send?access_token="

type dingtalkEntry struct {
	token   string
	silence bool
}

var Dingtalk = dingtalkEntry{silence: false}

func init() {
	token := getStringEnv(EnvKeyGamyeewaiDingtalkToken)
	if token != "" {
		Dingtalk.token = token
	}
}

func (e dingtalkEntry) WithSilence() dingtalkEntry {
	e.silence = true
	return e
}

func (e dingtalkEntry) WithToken(token string) dingtalkEntry {
	e.token = token
	return e
}

// 发送原始格式消息
func (e dingtalkEntry) Send(data map[string]interface{}) {
	if e.token == "" {
		msg := "未设置钉钉 Token，消息停止发送"

		if e.silence {
			Log.Warn().Msg(msg)
			return
		}

		panic(msg)
	}

	encode, _ := json.Marshal(data)
	resp, _ := http.Post(dingtalkHost+e.token, "application/json", bytes.NewBuffer(encode))
	_ = resp.Body.Close()
}

// 发送文本消息
func (e dingtalkEntry) SendText(content string) {
	data := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}

	e.Send(data)
}
