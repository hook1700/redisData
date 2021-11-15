package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	abnormal "redisData/pkg"
	"redisData/pkg/logger"
)

var lg *zap.Logger

// InitLogger 初始化Logger
func InitLogger(mode string) (err error) {

	writeSyncer := getLogWriter(
		viper.GetString("log.filename"),
		viper.GetInt("log.max_size"),
		viper.GetInt("log.max_backups"),
		viper.GetInt("log.max_age"))
	encoder := getEncoder()
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(viper.GetString("log.level")))
	if err != nil {
		return
	}
	//--------------------------------------------------------------------------------------
	//设置日志输出位置，在控制台和文本中选择
	var core zapcore.Core
	if mode == "dev" {
		//进入开发模式，日志输出到终端
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, l),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		)
	} else {
		core = zapcore.NewCore(encoder, writeSyncer, l)
	}
	//-------------------------------------------------------------------------------------
	lg = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(lg) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}

const DefaultHeader = "Tracking-Id"

func TraceLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer abnormal.Stack("服务器抛出异常")
		// 每个请求生成的请求RequestId具有全局唯一性
		RequestId := ctx.GetHeader(DefaultHeader)
		// 如果不存在，则生成TrackingID
		if RequestId == "" {
			RequestId = uuid.New().String()
			ctx.Header(DefaultHeader, RequestId)
		}
		fmt.Printf("当前请求ID为：%v\n", RequestId)
		ctx.Set(DefaultHeader, RequestId)
		logger.RequestId = RequestId
		logger.NewContext(ctx, zap.String("RequestId", RequestId))
		// 为日志添加请求的地址以及请求参数等信息
		logger.NewContext(ctx, zap.String("request.method", ctx.Request.Method))
		logger.NewContext(ctx, zap.String("request.url", ctx.Request.URL.String()))
		headers, _ := json.Marshal(ctx.Request.Header)
		logger.NewContext(ctx, zap.String("request.headers", string(headers)))
		// 将请求参数json序列化后添加进日志上下文
		data, err := ctx.GetRawData()
		if err != nil {
			logger.Error(err)
		}
		// 很关键,把读过的字节流重新放到body
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		logger.NewContext(ctx, zap.Any("request.params", string(data)))
		logger.WithContext(ctx).Info("请求信息："+RequestId, zap.Skip())

		path := ctx.FullPath()
		Connection := ctx.Request.Header.Get("Connection")
		logger.Info(path)
		logger.Info(Connection)
		// 继续往下面执行
		if Connection != "Upgrade" {
			switch path {
			case "/start":
			case "/klineHistory":
			case "/getKlineHistory":
				ctx.Next()
				break
			default:
				ctx.String(200, "hello world!")
				ctx.Abort()
				break
			}
		}
		ctx.Next()
	}
}
