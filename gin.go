package gamyeewai

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type ginEntry struct {
	Context *gin.Context
}

var Gin = new(ginEntry)

type requestData struct {
	Message   string      `json:"message"`
	URL       string      `json:"url"`
	Body      string      `json:"body"`
	Token     string      `json:"token"`
	ClientIP  string      `json:"client_ip"`
	RawData   string      `json:"raw_data"`
	Timestamp interface{} `json:"timestamp"`
}

func (e *ginEntry) CORSHandle() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization")

	return cors.New(config)
}

func (e *ginEntry) ErrorHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		e.Context = c

		defer func() {
			err := recover()
			if err == nil {
				return
			}

			msg := disposeErrMsg(err)

			Log.WithGin().Error().Msg(msg)

			data := e.parseRequest()
			data.Message = msg
			sendDingtalkAlarm(data)

			c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		}()

		c.Next()
	}
}

func (e *ginEntry) GetToken() (string, error) {
	token := e.Context.Query("token")

	if token != "" {
		return token, nil
	}

	header := e.Context.GetHeader("Authorization")

	if header != "" && strings.HasPrefix(header, "Bearer ") {
		return header[7:], nil
	}

	token, err := e.Context.Cookie("token")

	if err == nil {
		return token, nil
	}

	return "", errors.New("token 不存在")
}

func (e *ginEntry) getBody() string {
	body, err := e.Context.GetRawData()

	if err != nil {
		return "n/a"
	}

	e.Context.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	r := regexp.MustCompile("\\s+")
	return r.ReplaceAllString(string(body), "")
}

func (e *ginEntry) parseRequest() (ret requestData) {
	token, err := e.GetToken()

	if err != nil {
		token = "n/a"
	}

	rawData, _ := httputil.DumpRequest(e.Context.Request, false)

	ret = requestData{
		URL:       e.Context.Request.Method + " " + e.Context.Request.Host + e.Context.Request.RequestURI,
		Body:      e.getBody(),
		Token:     token,
		ClientIP:  e.Context.ClientIP(),
		RawData:   string(bytes.TrimSpace(rawData)),
		Timestamp: time.Now().Unix(),
	}

	return
}

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

func sendDingtalkAlarm(data requestData) {
	kibanaLink := disposeKibanaLink(data.Timestamp.(int64))

	data.Timestamp = time.Unix(data.Timestamp.(int64), 0).Format("2006-01-02 15:04:05")

	key := reflect.TypeOf(data)
	value := reflect.ValueOf(data)

	var content string

	for i := 0; i < key.NumField(); i++ {
		if key.Field(i).Name == "RawData" {
			continue
		}
		item := []string{"### " + key.Field(i).Name, "```", value.Field(i).Interface().(string), "```", "---", ""}
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
