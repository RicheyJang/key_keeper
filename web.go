package main

import (
	"embed"
	"net/http"

	"github.com/RicheyJang/key_keeper/logic"
	"github.com/RicheyJang/key_keeper/utils/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	irisRecover "github.com/kataras/iris/v12/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func WebServer(manager *logic.Manager, addr string) {
	// 初始化Gin
	app := iris.New()
	app.Use(irisRecover.New())

	// 设置路由
	app.PartyFunc("/api", func(api router.Party) {
		api.Use(logger.Iris("[ Web ]")) // 日志
		// TODO 注册后端接口
	})
	// 前端页面
	setupStatic(app)

	// 启动
	log.Fatal(app.Run(iris.TLS(addr, viper.GetString("cert.self"), viper.GetString("cert.private")),
		iris.WithoutPathCorrectionRedirection,
		iris.WithOptimizations))
}

//go:embed dist/*
var Static embed.FS

func setupStatic(engine *iris.Application) {
	fsys := iris.PrefixDir("dist", http.FS(Static)) // 去除前缀

	// 配置
	option := router.DefaultDirOptions
	option.SPA = true // 指定为 单页面SPA

	engine.HandleDir("/", fsys, option)
}
