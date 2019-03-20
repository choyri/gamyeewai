package gamyeewai

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ginEntry struct {
	sync.Mutex
	ctx *gin.Context
}

var Gin ginEntry

func (e *ginEntry) With(ctx *gin.Context) *ginEntry {
	e.Lock()
	e.ctx = ctx
	e.Unlock()
	return e
}

func (e *ginEntry) CORSHandle() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization")

	return cors.New(config)
}

func (e *ginEntry) DisposeError(err string) {
	requestData := e.parseRequest()
	msg := disposeErrMsg(err)

	// 打印日志
	Log.WithRequestData(requestData).Error().Msg(msg)

	// 钉钉报警
	requestData.Message = msg
	sendDingtalkAlarm(requestData)
}

func (e *ginEntry) ErrorHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		e.With(c)

		defer func() {
			err := recover()
			if err == nil {
				return
			}

			e.DisposeError(disposeErrMsg(err))

			c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		}()

		c.Next()
	}
}

func (e *ginEntry) GetToken() (string, error) {
	ctx := e.getCtx()

	token := ctx.Query("token")

	if token != "" {
		return token, nil
	}

	header := ctx.GetHeader("Authorization")

	if header != "" && strings.HasPrefix(header, "Bearer ") {
		return header[7:], nil
	}

	token, err := ctx.Cookie("token")

	if err == nil {
		return token, nil
	}

	return "", errors.New("token 不存在")
}

func (e *ginEntry) getBody() string {
	body, err := e.getCtx().GetRawData()

	if err != nil {
		return ""
	}

	e.getCtx().Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	r := regexp.MustCompile("\\s+")
	ret := r.ReplaceAllString(string(body), "")

	return ret
}

func (e *ginEntry) getCtx() *gin.Context {
	e.Lock()
	if e.ctx != nil {
		e.Unlock()
		return e.ctx
	}

	panic("Gin Context 不存在")
}

func (e *ginEntry) getForm() string {
	_ = e.getCtx().Request.ParseForm()

	form := make(map[string][]string)

	if f := e.getCtx().Request.Form; len(f) > 0 {
		for k, v := range f {
			form[k] = append(form[k], v...)
		}
	}

	if f := e.getCtx().Request.PostForm; len(f) > 0 {
		for k, v := range f {
			form[k] = append(form[k], v...)
		}
	}

	if len(form) == 0 {
		return ""
	}

	ret, err := json.Marshal(form)
	if err != nil {
		return ""
	}

	return string(ret)
}

func (e *ginEntry) parseRequest() *requestData {
	ctx := e.getCtx()

	token, _ := e.GetToken()
	rawData, _ := httputil.DumpRequest(ctx.Request, false)

	ret := requestData{
		URL:       ctx.Request.Method + " " + ctx.Request.Host + ctx.Request.RequestURI,
		Query:     ctx.Request.URL.RawQuery,
		Form:      e.getForm(),
		Body:      e.getBody(),
		Token:     token,
		ClientIP:  ctx.ClientIP(),
		RawData:   string(bytes.TrimSpace(rawData)),
		Timestamp: time.Now().Unix(),
		Trace:     getTrace(),
	}

	return &ret
}
