package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/leizongmin/huobiapi"
	"net/http"
	"redisData/huobi"
	"redisData/logic"
	"redisData/model"
	"redisData/pkg/logger"
	"redisData/utils"
	"strings"
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
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			logger.Error(err)
		}
	}(ws) //返回前关闭
	for {
		wsConn.Mux.Lock()
		//读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		wsConn.Mux.Unlock()
		if err != nil {
			fmt.Println(err)
			break
		}
		//对数据进行切割，读取参数
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		strMsg := string(message)
		//打印请求参数
		fmt.Println(strMsg)

		//写入ws数据
		go func() {
			for {
				data, err := logic.GetDataByKey(strMsg)
				//修改，当拿不到key重新订阅，10秒订阅一次
				if err == redis.Nil {
					//err = wsConn.Conn.WriteMessage(mt, []byte("key不存在，准备开始缓存"))
					if err != nil {
						return
					}
					StartSetKlineDataErr := logic.StartSetKlineData()
					if StartSetKlineDataErr != nil {
						logger.Error(StartSetKlineDataErr)
						return
					}
					time.Sleep(10 * time.Second)
				}
				websocketData := utils.Strval(data)
				wsConn.Mux.Lock()
				err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
				wsConn.Mux.Unlock()
				if err != nil {
					logger.Error(err)
					CloseErr := ws.Close()
					if CloseErr != nil {
						logger.Error(CloseErr)
						return
					}
					return
				}
				time.Sleep(time.Second * 2)
			}

		}()

	}

}

// GetRedisData4 自定义订阅symbol和时间
func GetRedisData4(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, _ := upGrader.Upgrade(c.Writer, c.Request, nil)
	wsConn := &WsConn{
		ws,
		sync.RWMutex{},
	}
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	defer ws.Close() //返回前关闭
	for {
		//读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		//打印参数
		fmt.Println(string(message))
		if err != nil {
			fmt.Println(err)
			break
		}
		//拿到参数进行校验
		//如果含有1min中直接请求redis
		b := strings.Contains(string(message), "1min")
		if b == true {
			//直接查询redis
			go func() {
				for {
					data, err := logic.GetDataByKey(string(message))
					//修改，当拿不到key重新订阅，10秒订阅一次
					if err == redis.Nil {
						err = wsConn.Conn.WriteMessage(mt, []byte("key不存在，准备开始缓存"))
						if err != nil {
							return
						}
						logic.StartSetKlineData()
						time.Sleep(10 * time.Second)
					}
					websocketData := utils.Strval(data)
					wsConn.Mux.Lock()
					err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
					wsConn.Mux.Unlock()
					if err != nil {
						fmt.Println(err)
						ws.Close()
						return
					}
					time.Sleep(time.Second * 2)
				}
			}()
		}

		//如果不含有1min，缓存到redis在redis里面拿

		//fmt.Println(string(message))
		//对数据进行切割，读取参数
		newMessage := message[1 : len(message)-1]
		//fmt.Println(string(newMessage))
		res := utils.Split(string(newMessage), ".")
		//截取2和4作为参数
		//传入参数直接请求 存取redis
		go huobi.NewSubscribeParam(res[1], res[3])
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		//打印请求参数

		go func() {
			for {
				data, err := logic.GetDataByKey(string(message))
				if err != nil {
					fmt.Println(err)
					return
				}
				websocketData := utils.Strval(data)
				//fmt.Println(websocketData)
				wsConn.Mux.Lock()
				err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
				wsConn.Mux.Unlock()
				time.Sleep(time.Second * 2)
			}

			//}()
			//通过redis的key取值

			//写入ws数据
			//go func() {
			//	for {
			//		data, err := logic.GetDataByKey(strMsg)
			//		//修改，当拿不到key重新订阅，10秒订阅一次
			//		if err == redis.Nil {
			//			err = wsConn.Conn.WriteMessage(mt, []byte("key不存在，准备开始缓存"))
			//			if err != nil {
			//				return
			//			}
			//			logic.StartSetKlineData()
			//			time.Sleep(10 * time.Second)
			//		}
			//		websocketData := utils.Strval(data)
			//		wsConn.Mux.Lock()
			//		err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
			//		wsConn.Mux.Unlock()
			//		if err != nil {
			//			fmt.Println(err)
			//			ws.Close()
			//			return
			//		}
			//		time.Sleep(time.Second * 2)
			//	}
			//
		}()

	}

}

// QuotationController 请求行情数据接口
func QuotationController(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close() //返回前关闭
	for {
		//读取ws中的数据，数据是"market.btcusdt.depth.step1"类型
		mt, message, err := ws.ReadMessage()
		if err != nil {
			break
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
					logic.StartSetQuotation()
					time.Sleep(10 * time.Second)
				}
				websocketData := utils.Strval(data)
				err = ws.WriteMessage(mt, []byte(websocketData))
				if err != nil {
					return
				}
				time.Sleep(time.Second * 1)
			}
		}()
	}
}

//https://api.huobi.pro/market/history/kline?period=1day&size=200&symbol=btcusdt

// KlineHistoryController 每10秒缓存300条数据  已经移入start里面了
func KlineHistoryController(c *gin.Context) {
	//参数校验-无
	//逻辑处理
	go func() {
		err := logic.SetKlineHistory()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	//返回参数
	c.JSON(http.StatusOK, gin.H{
		"msg": "保存K线图历史数据成功",
	})

}

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
	if symbol == "" {
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

	}

	//period = 1min自动缓存
	historyData, err := logic.GetKlineHistoryDiy(symbol, "")
	//historyData, err := logic.GetKlineHistory(symbol)
	if err != nil {
		if err == redis.Nil {
			err := logic.SetKlineHistory()
			c.JSON(http.StatusOK, gin.H{
				"msg": "正在缓存数据,请2s后继续访问",
			})
			fmt.Println(err)
			return

			time.Sleep(10 * time.Second)
		}
		fmt.Println(err)
		return
	}
	//返回数据
	c.JSON(http.StatusOK, gin.H{
		"data": historyData,
	})
}

//1.启动一个websocket 客户端
func GetRedisData3(c *gin.Context) {
	market, err := huobiapi.NewMarket()
	if err != nil {
		fmt.Printf("huobiapi.NewMarket() %v", err)
		return
	}
	//升级get请求为webSocket协议
	ws, _ := upGrader.Upgrade(c.Writer, c.Request, nil)
	wsConn := &WsConn{
		ws,
		sync.RWMutex{},
	}
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	defer ws.Close() //返回前关闭
	for {
		//2.读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		err = market.Subscribe(string(message), func(topic string, hjson *huobiapi.JSON) {
			// 收到数据更新时回调,收到信息返回给前端
			jsonData, _ := hjson.MarshalJSON()
			wsConn.Mux.Lock()
			//3.写数据给
			err = wsConn.Conn.WriteMessage(mt, jsonData)
			if err != nil {
				fmt.Printf("webSocket Write Data fail err%v", err)
			}
			wsConn.Mux.Unlock()
			//fmt.Println(topic, hjson)
		})
		if err != nil {
			fmt.Printf("market.Subscribe fail %v", err)
			return
		}

		//对数据进行切割，读取参数
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		//strMsg := string(message)
		//打印请求参数
		//fmt.Println(strMsg)

		//写入ws数据
		//go func() {
		//	for {
		//		data, err := logic.GetDataByKey(strMsg)
		//		//修改，当拿不到key重新订阅，10秒订阅一次
		//		if err == redis.Nil {
		//			err = wsConn.Conn.WriteMessage(mt, []byte("key不存在，准备开始缓存"))
		//			if err != nil {
		//				return
		//			}
		//			logic.StartSetKlineData()
		//			time.Sleep(10 * time.Second)
		//		}
		//		websocketData := utils.Strval(data)
		//		wsConn.Mux.Lock()
		//		err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
		//		wsConn.Mux.Unlock()
		//		if err != nil {
		//			fmt.Println(err)
		//			ws.Close()
		//			return
		//		}
		//		time.Sleep(time.Second * 2)
		//	}
		//
		//}()

	}
	market.Loop()

}
