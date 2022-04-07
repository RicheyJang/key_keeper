package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kataras/iris/v12"
	requestLogger "github.com/kataras/iris/v12/middleware/logger"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// SetupLogger 设置日志
func SetupLogger() error {
	// 配置日志等级
	log.SetLevel(log.InfoLevel)
	if l, ok := flagLToLevel[strings.ToLower(viper.GetString("log.level"))]; ok {
		log.SetLevel(l)
	}
	// 日志格式
	log.SetFormatter(&SimpleFormatter{})
	// 日志滚动切割
	logW, err := GetWriter()
	if err != nil {
		log.Error("Get rotate logs file err: ", err)
		return err
	}
	// 日志输出
	log.SetOutput(logW)
	return nil
}

// GetWriter 获取日志输出writer
func GetWriter() (io.Writer, error) {
	writerLock.Lock()
	defer writerLock.Unlock()
	if writer == nil {
		dir := viper.GetString("log.dir")
		logf, err := rotatelogs.New(
			filepath.Join(dir, "kk-%Y-%m-%d.log"),
			rotatelogs.WithLinkName(filepath.Join(dir, "kk.log")),
			rotatelogs.WithMaxAge(time.Duration(viper.GetInt("log.date"))*24*time.Hour),
			rotatelogs.WithRotationTime(24*time.Hour),
		)
		if err != nil {
			return nil, err
		}
		writer = io.MultiWriter(os.Stdout, logf)
	}
	return writer, nil
}

var writer io.Writer
var writerLock sync.Mutex

var flagLToLevel = map[string]log.Level{
	"debug":   log.DebugLevel,
	"info":    log.InfoLevel,
	"warn":    log.WarnLevel,
	"warning": log.WarnLevel,
	"error":   log.ErrorLevel,
}

// logrus 日志格式化

type SimpleFormatter struct{}

const stringOfSymbol = "[kk]"
const stringOfStarter = ": "

func (f SimpleFormatter) Format(entry *log.Entry) ([]byte, error) {
	var output bytes.Buffer
	// 标识
	output.WriteString(stringOfSymbol)
	// 时间
	output.WriteString(entry.Time.Format("[2006-01-02 15:04:05.000ms]"))
	// 等级
	output.WriteRune('[')
	output.WriteString(entry.Level.String())
	output.WriteRune(']')
	// 消息
	output.WriteString(stringOfStarter)
	output.WriteString(entry.Message)
	// 键值对
	output.WriteRune(' ')
	for k, val := range entry.Data {
		output.WriteString(k)
		output.WriteRune(':')
		output.WriteString(cast.ToString(val))
		output.WriteRune(' ')
	}
	output.WriteRune('\n')
	return output.Bytes(), nil
}

// iris 日志相关

func Iris() iris.Handler {
	return requestLogger.New(requestLogger.Config{
		Status:     true,
		IP:         true,
		Method:     true,
		Path:       true,
		Query:      true,
		LogFunc:    irisLoggerFunc,
		LogFuncCtx: nil,
		Skippers:   nil,
	})
}

func irisLoggerFunc(endTime time.Time, latency time.Duration, status, ip, method, path string,
	message interface{}, headerMessage interface{}) {
	log.Infof("[IRIS] %3v | %13v | %15v | %-7v  %#v | msg: %v",
		status,
		latency,
		ip,
		method,
		path,
		message,
	)
}
