package routes

import (
	"github.com/gin-gonic/gin"
	"redisData/controller"
	"redisData/logger"
)

func SetUp() *gin.Engine {

	r := gin.New()
	r.Use(logger.TraceLogger())
	//查询，查询redis上的数据，返回给前端
	//请求的 url https://api.huobi.pro/market/history/kline?period=1min&size=1&symbol=btcusdt
	//websocket
	r.GET("/getRedisData", controller.GetRedisData)
	r.GET("/quotation", controller.QuotationController)
	r.GET("/ws", controller.GetRedisData2)
	r.GET("/websocketData", controller.GetRedisData3)
	r.GET("/test", controller.GetRedisData4)
	r.GET("/test2", controller.WsHandle2)
	r.GET("/test3", controller.WsHandle)
	//http
	r.GET("/start", controller.StartController)
	r.GET("/klineHistory", controller.KlineHistoryController)
	r.GET("/getKlineHistory", controller.GetKlineHistoryController)
	return r

}
