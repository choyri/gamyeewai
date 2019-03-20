package gamyeewai

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin/json"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"io/ioutil"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"
)

type irisEntry struct {
	ctx iris.Context
}

var Iris irisEntry

func (e *irisEntry) With(ctx iris.Context) *irisEntry {
	e.ctx = ctx
	return e
}

func (e *irisEntry) DisposeError(err string) {
	requestData := e.parseRequest()
	msg := disposeErrMsg(err)

	// 打印日志
	Log.WithRequestData(requestData).Error().Msg(msg)

	// 钉钉报警
	requestData.Message = msg
	sendDingtalkAlarm(requestData)
}

func (e *irisEntry) ErrorHandle() context.Handler {
	return func(ctx context.Context) {
		e.With(ctx)

		defer func() {
			err := recover()
			if err == nil {
				return
			}

			e.DisposeError(disposeErrMsg(err))
		}()

		ctx.Next()
	}
}

func (e *irisEntry) GetToken() (string, error) {
	ctx := e.getCtx()

	token := ctx.URLParam("token")

	if token != "" {
		return token, nil
	}

	header := ctx.GetHeader("Authorization")

	if header != "" && strings.HasPrefix(header, "Bearer ") {
		return header[7:], nil
	}

	token = ctx.GetCookie("token")

	if token != "" {
		return token, nil
	}

	return "", errors.New("token 不存在")
}

func (e *irisEntry) getBody() string {
	body := e.getCtx().Request().Body
	if body == nil {
		return ""
	}

	rawData, err := ioutil.ReadAll(body)
	if err != nil {
		return ""
	}

	e.getCtx().Request().Body = ioutil.NopCloser(bytes.NewBuffer(rawData))

	r := regexp.MustCompile("\\s+")
	ret := r.ReplaceAllString(string(rawData), "")

	return ret
}

func (e *irisEntry) getCtx() iris.Context {
	if e.ctx != nil {
		return e.ctx
	}

	panic("Iris Context 不存在")
}

func (e *irisEntry) getForm() string {
	values := e.getCtx().FormValues()
	if len(values) == 0 {
		return ""
	}

	ret, err := json.Marshal(values)
	if err != nil {
		return ""
	}

	return string(ret)
}

func (e *irisEntry) parseRequest() *requestData {
	ctx := e.getCtx()

	token, _ := e.GetToken()
	rawData, _ := httputil.DumpRequest(ctx.Request(), false)

	ret := requestData{
		URL:       ctx.Request().Method + " " + ctx.Request().Host + ctx.Request().RequestURI,
		Query:     ctx.Request().URL.RawQuery,
		Form:      e.getForm(),
		Body:      e.getBody(),
		Token:     token,
		ClientIP:  ctx.RemoteAddr(),
		RawData:   string(bytes.TrimSpace(rawData)),
		Timestamp: time.Now().Unix(),
		Trace:     getTrace(),
	}

	return &ret
}
