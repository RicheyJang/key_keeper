package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"

	"github.com/RicheyJang/key_keeper/logic"
	"github.com/RicheyJang/key_keeper/utils"
	"github.com/RicheyJang/key_keeper/utils/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/core/router"
	irisRecover "github.com/kataras/iris/v12/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InnerServer(manager *logic.Manager, addr string) {
	// 初始化Gin
	app := iris.New()
	app.Use(irisRecover.New())
	app.Use(logger.Iris("[Inner]")) // 日志

	// 设置路由
	app.PartyFunc("/api/inner", func(inner router.Party) {
		inner.Use(manager.PreRouterOfSetKeeper)
		inner.Post("/key", manager.GetKeyInfo)
		inner.Post("/version", manager.GetLatestVersionKey)
	})

	// 启动
	log.Fatal(app.Run(getRunner(addr),
		iris.WithoutPathCorrectionRedirection,
		iris.WithOptimizations))
}

func getRunner(addr string, hostConfigs ...host.Configurator) iris.Runner {
	// 读取证书
	pool := x509.NewCertPool()
	crt, err := ioutil.ReadFile(viper.GetString("cert.ca"))
	if err != nil {
		log.Fatal("Failed to read CA certificate: ", err.Error())
	}
	pool.AppendCertsFromPEM(crt)
	cert, err := utils.LoadCertificate(viper.GetString("cert.self"), viper.GetString("cert.private"))
	if err != nil {
		log.Fatal("Failed to read Server certificate: ", err.Error())
	}
	// 设置证书认证
	logw, err := logger.GetWriter()
	if err != nil {
		logw = os.Stderr
	}
	s := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert, // 检验客户端证书
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cert, nil
			},
			// NextProtos: []string{"h2", "http/1.1"},
		},
		ErrorLog: stdlog.New(logw, "[kk] ", stdlog.LstdFlags),
	}
	// 生成Runner
	return func(app *iris.Application) error {
		return app.NewHost(s).
			Configure(hostConfigs...).
			ListenAndServeTLS("", "")
	}
}
