package main

import (
	"path/filepath"

	"github.com/RicheyJang/key_keeper/keeper/example"
	"github.com/RicheyJang/key_keeper/logic"
	"github.com/RicheyJang/key_keeper/utils"
	"github.com/RicheyJang/key_keeper/utils/logger"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	pflag.StringP("host", "h", ":7709", "key service running host")
	pflag.StringP("web", "w", ":7710", "web service running host")
	pflag.StringP("log", "l", "info", "the level of logging")
	configPath := pflag.StringP("config", "c", "./config.toml", "configuration file path")
	pflag.Parse()
	// Host配置
	viper.SetDefault("host", ":7709")
	_ = viper.BindPFlag("host", pflag.Lookup("host"))
	viper.SetDefault("web", ":7710")
	_ = viper.BindPFlag("web", pflag.Lookup("web"))
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

	// 初始化Manager
	manager, err := logic.NewManager(logic.Option{
		KGs: []logic.KeeperGeneratorPair{ // 密钥保管器 及其 对应的生成器 列表 TODO 补充KG列表、使用其它Keeper为默认
			{KeeperName: "Example", Generator: example.NewExampleKeeper, IsDefault: true},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 启动各个服务
	go WebServer(manager, viper.GetString("web"))
	InnerServer(manager, viper.GetString("host"))
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
