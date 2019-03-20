package main

import (
	"fmt"
	"github.com/choyri/gamyeewai"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kataras/iris"
	"net/http"
)

func main() {
	fmt.Println("Hello world.")
	//dingtalkTest()
	//ginTest()
	//irisTest()
	//logTest()
}

func dingtalkTest() {
	gamyeewai.Dingtalk.WithSilence().Send(map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": "手动设置钉钉消息格式",
		},
	})

	gamyeewai.Dingtalk.SendText("直接发送文本钉钉消息")
	gamyeewai.Dingtalk.WithSilence().SendText("直接发送文本钉钉消息 静默方式")
	gamyeewai.Dingtalk.WithToken("abcdefg").SendText("直接发送文本钉钉消息 自定义 Token")
}

func ginTest() {
	r := gin.Default()
	r.Use(gamyeewai.Gin.CORSHandle())
	r.Use(gamyeewai.Gin.ErrorHandle())

	r.GET("/panic", func(c *gin.Context) {
		panic("手动抛出一个错误")
	})

	r.POST("/panic", func(c *gin.Context) {
		panic("手动抛出一个错误")
	})

	r.GET("/token", func(c *gin.Context) {
		token, err := gamyeewai.Gin.With(c).GetToken()
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.JSON(200, gin.H{"token": token})
		}
	})

	_ = r.Run()
}

func irisTest() {
	app := iris.Default()

	app.Use(gamyeewai.Iris.ErrorHandle())

	app.Get("/panic", func(ctx iris.Context) {
		panic("手动抛出一个错误")
	})

	app.Post("/panic", func(ctx iris.Context) {
		panic("手动抛出一个错误")
	})

	app.Get("/token", func(ctx iris.Context) {
		token, err := gamyeewai.Iris.With(ctx).GetToken()
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			_, _ = ctx.WriteString(err.Error())
			return
		}

		_, _ = ctx.JSON(iris.Map{
			"token": token,
		})
	})

	_ = app.Run(iris.Addr(":8088"))
}

func logTest() {
	gamyeewai.Log.Info().Msg("测试 Info 日志")
	gamyeewai.Log.Error().Msg("测试 Error 日志")
}
