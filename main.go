package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/RicheyJang/key_keeper/keeper/example"
	"github.com/RicheyJang/key_keeper/logic"
	"github.com/RicheyJang/key_keeper/utils"
	"github.com/RicheyJang/key_keeper/utils/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/core/router"
	irisRecover "github.com/kataras/iris/v12/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
		inner.Post("/key", manager.GetKeyInfo)
		inner.Post("/version", manager.GetLatestVersionKey)
	})
}

func init() {
	pflag.StringP("host", "h", ":7709", "service running host")
	pflag.StringP("log", "l", "info", "the level of logging")
	configPath := pflag.StringP("config", "c", "./config.toml", "configuration file path")
	pflag.Parse()
	// Host配置
	viper.SetDefault("host", ":7709")
	_ = viper.BindPFlag("host", pflag.Lookup("host"))
	_ = viper.BindPFlag("config", pflag.Lookup("config"))
	// 日志配置
	viper.SetDefault("log.level", "info")
	_ = viper.BindPFlag("log.level", pflag.Lookup("log"))
	viper.SetDefault("log.dir", "log")
	viper.SetDefault("log.date", 5)
	// 证书配置
	viper.SetDefault("cert.ca", "cert/ca.crt")
	viper.SetDefault("cert.self", "cert/server.crt")
	viper.SetDefault("cert.private", "cert/server_rsa_private.pem")
	configDir, configFile := filepath.Split(*configPath)
	if err := flushConfig(configDir, configFile); err != nil {
		log.Fatal("setup config error: ", err)
	}
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
	log.Fatal(app.Run(getRunner(viper.GetString("host")),
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
			NextProtos: []string{"h2", "http/1.1"},
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

// 从文件和命令行中刷新所有主配置，若文件不存在将会把配置写入该文件
func flushConfig(configPath string, configFileName string) error {
	// 从文件读取
	viper.AddConfigPath(configPath)
	viper.SetConfigFile(configFileName)
	fullPath := filepath.Join(configPath, configFileName)
	//fileType := filepath.Ext(fullPath)
	//viper.SetConfigType(fileType)
	if utils.FileExists(fullPath) { // 配置文件已存在：合并自配置文件后重新写入
		err := viper.MergeInConfig()
		if err != nil {
			log.Error("flushConfig error in MergeInConfig err: ", err)
			return err
		}
		_ = viper.WriteConfigAs(fullPath)
	} else { // 配置文件不存在：写入配置
		err := viper.SafeWriteConfigAs(fullPath)
		if err != nil {
			log.Error("flushConfig error in SafeWriteConfig err: ", err)
			return err
		}
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) { // 配置文件发生变更之后会调用的回调函数
		_ = logger.SetupLogger()
		log.Infof("reload config from %v", e.Name)
	})
	return nil
}
