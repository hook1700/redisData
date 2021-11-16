package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/unrolled/secure"
	"go.uber.org/zap"
	"io/ioutil"
	abnormal "redisData/pkg"
	"redisData/pkg/logger"
)

const DefaultHeader = "Tracking-Id"

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     fmt.Sprintf("localhost:%d", viper.GetInt("port")),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

// 日志追踪

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
