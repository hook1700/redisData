package routes

import (
	"github.com/gin-gonic/gin"
	"redisData/controller"
	"redisData/middleware"
)

func SetUp() *gin.Engine {
	r := gin.Default()
	//r.Use(middleware.TlsHandler())  // 支持wss
	r.Use(middleware.TraceLogger()) // 日志上下文进行绑定追踪
	//查询，查询redis上的数据，返回给前端
	//请求的 url https://api.huobi.pro/market/history/kline?period=1min&size=1&symbol=btcusdt
	//websocket
	r.GET("/getRedisData", controller.GetRedisData) //
	r.GET("/quotation", controller.QuotationController)
	r.GET("/test3", controller.WsHandle) //
	r.GET("/start", controller.StartController)
	r.GET("/getKlineHistory", controller.GetKlineHistoryController)
	return r

}
