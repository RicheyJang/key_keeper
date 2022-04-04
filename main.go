package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/RicheyJang/key_keeper/keeper/example"
	"github.com/RicheyJang/key_keeper/logic"
	"github.com/RicheyJang/key_keeper/utils"
	"github.com/RicheyJang/key_keeper/utils/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/core/router"
	irisRecover "github.com/kataras/iris/v12/middleware/recover"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TODO 证书文件路径使用配置
const (
	CACertPath     = "cert/ca.crt"
	ServerCertPath = "cert/server.crt"
	ServerKeyPath  = "cert/server_rsa_private.pem"
)

func setupRoute(group iris.Party) {
	// 初始化Manager
	manager, err := logic.NewManager(logic.Option{
		KGs: []logic.KeeperGeneratorPair{ // 密钥保管器 及其 对应的生成器 列表 TODO 补充KG列表、使用其它Keeper为默认
			{KeeperName: "Example", Generator: example.NewExampleKeeper, IsDefault: true},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 设置路由
	group.PartyFunc("/inner", func(inner router.Party) {
		inner.Use(manager.PreRouterOfSetKeeper)
		inner.Post("/echo", manager.GetKeyInfo)
	})
}

func init() {
	pflag.StringP("host", "h", ":7709", "service running host")
	pflag.StringP("log", "l", "info", "the level of logging")
	pflag.StringP("config", "c", "./config.yaml", "configuration file path")
	pflag.Parse()
	// Host配置
	viper.SetDefault("host", ":7709")
	_ = viper.BindPFlag("host", pflag.Lookup("host"))
	_ = viper.BindPFlag("config", pflag.Lookup("config"))
	// 日志配置
	viper.SetDefault("log.level", "info")
	_ = viper.BindPFlag("log.level", pflag.Lookup("log"))
}

func main() {
	// 初始化日志
	if err := logger.SetupLogger(); err != nil {
		log.Fatal(err)
	}

	// 初始化Gin
	app := iris.New()
	app.Use(irisRecover.New())
	app.Use(logger.Iris()) // 日志

	// 设置路由
	app.PartyFunc("/api", setupRoute)
	// TODO 前端嵌入

	// 启动
	log.Fatal(app.Run(GetRunner(viper.GetString("host")),
		iris.WithoutPathCorrectionRedirection,
		iris.WithOptimizations))
}

func GetRunner(addr string, hostConfigs ...host.Configurator) iris.Runner {
	// 读取证书
	pool := x509.NewCertPool()
	crt, err := ioutil.ReadFile(CACertPath)
	if err != nil {
		log.Fatal("Failed to read CA certificate: ", err.Error())
	}
	pool.AppendCertsFromPEM(crt)
	cert, err := utils.LoadCertificate(ServerCertPath, ServerKeyPath)
	if err != nil {
		log.Fatal("Failed to read Server certificate: ", err.Error())
	}
	// 设置证书认证
	s := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert, // 检验客户端证书
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return cert, nil
			},
			NextProtos: []string{"h2", "http/1.1"},
		},
	}
	// 生成Runner
	return func(app *iris.Application) error {
		return app.NewHost(s).
			Configure(hostConfigs...).
			ListenAndServeTLS("", "")
	}
}
