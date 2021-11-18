package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"net/http"
	"redisData/logic"
	"redisData/model"
	"redisData/pkg/logger"
	"redisData/utils"
	"sync"
	"time"
)

//设置websocket
//CheckOrigin防止跨站点的请求伪造
//var upGrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//}
//
//type WsConn struct {
//	*websocket.Conn
//	Mux sync.RWMutex
//}

//wsConn.Mux.Lock() //加锁
//err=wsConn.Conn.WriteMessage(websocket.TextMessage,msgByte)
//wsConn.Mux.Unlock()

// StartController 每次先手动启动一下
func StartController(c *gin.Context) {

	//启动获取k线图数据
	if err := logic.StartSetKlineData(); err != nil {
		fmt.Printf("logic.StartSetKlineData() fail err:%v", err)
	}
	//启动获取行情数据
	if err := logic.StartSetQuotation(); err != nil {
		fmt.Printf("logic.StartSetQuotation() fail err:%v", err)
	}
	go func() {
		err := logic.SetKlineHistory()
		if err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Println("huobiService Start success")
	c.JSON(http.StatusOK, gin.H{

		"msg": "Start success",
	})
}

// GetRedisData websocket请求,根据发送的内容返回键值对
func GetRedisData(c *gin.Context) {

	//升级get请求为webSocket协议
	ws, _ := upGrader.Upgrade(c.Writer, c.Request, nil)
	wsConn := &WsConn{
		ws,
		sync.RWMutex{},
	}
	for {
		//读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		if err != nil {
			logger.Info(err)
			continue
		}
		//对数据进行切割，读取参数
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		strMsg := string(message)
		//打印请求参数
		logger.Info(strMsg)

		//写入ws数据
		go func(wsConn *WsConn) {
			for {
				data, GetDataByKeyErr := logic.GetDataByKey(strMsg)
				//修改，当拿不到key重新订阅，10秒订阅一次
				if GetDataByKeyErr == redis.Nil {
					//err = wsConn.Conn.WriteMessage(mt, []byte("key不存在，准备开始缓存"))
					//if err != nil {
					//	return
					//}
					StartSetKlineDataErr := logic.StartSetKlineData()
					if StartSetKlineDataErr != nil {
						logger.Error(StartSetKlineDataErr)
						continue
					}
					time.Sleep(10 * time.Second)
				}
				websocketData := utils.Strval(data)
				err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
				if err != nil {
					logger.Info(mt)
					logger.Info(websocketData)
					logger.Info(data)
					logger.Error(err)
					CloseErr := wsConn.Close()
					if CloseErr != nil {
						logger.Error(CloseErr)
						continue
					}
					continue
				}
				time.Sleep(time.Second * 2)
			}

		}(wsConn)

	}

}

// QuotationController 请求行情数据接口
func QuotationController(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, wsErr := upGrader.Upgrade(c.Writer, c.Request, nil)
	if wsErr != nil {
		logger.Error(wsErr)
		return
	}
	wsConn := &WsConn{
		ws,
		sync.RWMutex{},
	}
	for {
		//读取ws中的数据，数据是"market.btcusdt.depth.step1"类型
		mt, message, messageErr := wsConn.Conn.ReadMessage()
		if messageErr != nil {
			logger.Error(messageErr)
			return
		}
		strMsg := string(message)
		//打印请求参数
		fmt.Println(strMsg)
		//分割
		//resultList := utils.Split(strMsg, ".")

		go func() {
			for {
				data, err := logic.GetDataByKey(strMsg)
				//修改，当拿不到key重新订阅，10秒订阅一次
				if err == redis.Nil {
					StartSetQuotationErr := logic.StartSetQuotation()
					if StartSetQuotationErr != nil {
						logger.Error(StartSetQuotationErr)
						return
					}
					time.Sleep(10 * time.Second)
				}
				websocketData := utils.Strval(data)
				if len(websocketData) > 0 {
					wsConn.Mux.Lock()
					WriteMessageErr := wsConn.Conn.WriteMessage(mt, []byte(websocketData))
					wsConn.Mux.Unlock()
					if WriteMessageErr != nil {
						logger.Error(WriteMessageErr)
						CloseErr := wsConn.Close()
						if CloseErr != nil {
							logger.Error(CloseErr)
							return
						}
						return
					}
				}
				time.Sleep(time.Second * 2)
			}
		}()
	}
}

//https://api.huobi.pro/market/history/kline?period=1day&size=200&symbol=btcusdt

// GetKlineHistoryController 通过key获取历史300条数据
func GetKlineHistoryController(c *gin.Context) {
	//参数检验
	//校验参数，不写直接给默认值
	symbol := c.Query("symbol")
	period := c.Query("period")
	fmt.Println(symbol)
	if symbol == "" {
		c.JSON(http.StatusOK, gin.H{
			"msg": "symbol param is require",
		})
		return
	}
	if period == "" {
		c.JSON(http.StatusOK, gin.H{
			"msg": "period param is require",
		})
		return
	}
	//逻辑
	//判断key是否存在，存在直接拿
	key := fmt.Sprintf("\"market.%s.kline.%s\"", symbol, period)
	res := logic.ExistKey(key)
	if res == true {
		fmt.Println("key已经存在")
		//直接从reids查询返回
		diy, err := logic.GetKlineHistoryDiy(symbol, period)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsondata := utils.Strval(diy)

		c.JSON(http.StatusOK, gin.H{
			"data": jsondata,
		})
		return
	}
	//period != 1min,请求时再缓存
	if period != "1min" {
		fmt.Println("period != 1min")
		//请求火币网，拿到数据换算，存进redis ,取redis
		kilneData, err := logic.RequestHuobiKilne(symbol, period)
		if err != nil {
			fmt.Println(err)
			return
		}
		//反序列化
		var data model.KlineData
		err = json.Unmarshal(kilneData, &data)
		if err != nil {
			fmt.Println(err)
			return
		}
		//自有币换算
		tranData := logic.TranDecimalScale(symbol, data)
		//序列化
		jsonData, err := json.Marshal(tranData)
		if err != nil {
			fmt.Println(jsonData)
		}
		//存进redis
		logic.CreateHistoryKline(fmt.Sprintf("\"market.%s.kline.%s\"", symbol, period), string(jsonData))
		//logic.GetKlineHistory(fmt.Sprintf("\"market.%s.kline.%s\"",symbol,period),string(tranData))
		c.JSON(http.StatusOK, gin.H{
			//返回数据
			"data": tranData,
		})
		return

	}

	//period = 1min自动缓存
	historyData, err := logic.GetKlineHistoryDiy(symbol, "")
	//historyData, err := logic.GetKlineHistory(symbol)
	if err != nil {
		if err == redis.Nil {
			SetKlineHistoryErr := logic.SetKlineHistory()
			//c.JSON(http.StatusOK, gin.H{
			//	"msg": "正在缓存数据,请2s后继续访问",
			//})
			if SetKlineHistoryErr != nil {
				logger.Error(SetKlineHistoryErr)
			}
			time.Sleep(10 * time.Second)
			return
		}
		fmt.Println(err)
		return
	}
	//返回数据
	c.JSON(http.StatusOK, gin.H{
		"data": historyData,
	})
	return
}
