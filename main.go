package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/RicheyJang/key_keeper/model"

	"github.com/RicheyJang/key_keeper/keeper/example"
	"github.com/RicheyJang/key_keeper/logic"
	"github.com/RicheyJang/key_keeper/utils"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/RicheyJang/key_keeper/utils/logger"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
	// 数据库配置
	viper.SetDefault("db.type", "postgresql")
	viper.SetDefault("db.name", "database")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "username")
	viper.SetDefault("db.password", "password")
	// 证书配置
	viper.SetDefault("cert.ca", "cert/ca.crt")
	viper.SetDefault("cert.self", "cert/server.crt")
	viper.SetDefault("cert.private", "cert/server_rsa_private.pem")
	// 其它设置
	viper.SetDefault("user.maxAge", time.Duration(10*time.Hour))
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
	// 初始化数据库
	db, err := setupDatabase(viper.Sub("db"))
	if err != nil {
		log.Fatal(err)
	}

	// 初始化Manager
	manager, err := logic.NewManager(logic.Option{
		DB:          db,
		UserManager: model.NewUserManger(db),
		KGs: []logic.KeeperGeneratorPair{ // 密钥保管器 及其 对应的生成器 列表 TODO 补充KG列表、使用其它Keeper为默认
			{KeeperName: "Example", Generator: example.NewExampleKeeper},
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

// 初始化gorm数据库
func setupDatabase(config *viper.Viper) (db *gorm.DB, err error) {
	if config == nil {
		return nil, errors.New(-1, "database config is nil")
	}
	// 初始化配置
	gormC := &gorm.Config{ // 1.1 数据库配置
		Logger: logger.NewGormLogger(),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_", // 表名前缀，`User`表为`t_users`
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	}
	// 连接数据库
	switch strings.ToLower(config.GetString("type")) {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetString("user"), config.GetString("password"), config.GetString("host"), config.GetInt("port"), config.GetString("name"))
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN:                       dsn,   // DSN data source name
			DefaultStringSize:         256,   // string 类型字段的默认长度
			SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
		}), gormC)
		if err != nil {
			return nil, err
		}
	case "pg", "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.GetString("host"), config.GetString("user"), config.GetString("password"), config.GetString("name"), config.GetInt("port"))
		db, err = gorm.Open(postgres.Open(dsn), gormC)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(-1, "This type of database is not currently supported")
	}
	return
}
