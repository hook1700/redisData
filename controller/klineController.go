package controller

import (
	"encoding/json"
	"strings"

	//"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/leizongmin/huobiapi"
	"redisData/huobi"

	//"github.com/leizongmin/huobiapi"
	"go.uber.org/zap"
	"net/http"
	"redisData/logic"
	"redisData/utils"
	"sync"
	"time"
)

//声明一个线程安全的map,存放ws和user
//var users sync.Map

//设置websocket
//CheckOrigin防止跨站点的请求伪造
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsConn 声明并发安全的ws
type WsConn struct {
	*websocket.Conn
	Mux sync.RWMutex
}

// UserInfo 看这个用户订阅了什么
//type UserInfo struct {
//	Uid     string `json:"uid"`
//	Sub    []string `json:"sub_topic"`
//}

func GetRedisData2(c *gin.Context) {
	//每个用户连接,就new一个 ws
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("upGrader.Upgrade fail", zap.Error(err))
		return
	}
	//defer ws.Close()
	//返回前关闭
	//var user UserInfo
	//user.Uid = utils.GetGenerateId()
	wsConn := &WsConn{
		ws,
		sync.RWMutex{},
	}
	defer ws.Close() //返回前关闭

	//每一个ws 对应一个market
	//连接一个market
	//market, err := huobi.NewMarket()
	//if err != nil {
	//	zap.L().Error("huobi.NewMarket() fail", zap.Error(err))
	//}
	//把这个websocket指针和user对应的hash储存起来
	//users.Store(ws,user)

	//读取客户端信息
	for {
		//读取ws中的数据
		wsConn.Mux.Lock()
		mt, message, err := wsConn.Conn.ReadMessage()
		wsConn.Mux.Unlock()
		if err != nil {
			zap.L().Error("wsConn.Conn.ReadMessage fail", zap.Error(err))
			break
		}
		//60秒不发送信息，关闭socket

		//把用户传进来的消息进行处理 msg样式 "market.btcusdt.kline.1min"
		msg := string(message)
		//-------------
		fmt.Println(msg)
		//当请求数据中含有1min或1step这些为已经缓存数据,直接去redis拿
		//if strings.Contains(msg,"1min")||strings.Contains(msg,"step1"){
		go func() {
			for {
				data, err := logic.GetDataByKey(msg)
				if err != nil {
					//如果redis数据获取或者start接口没有被调用，就要重新缓存
					if err == redis.Nil {
						wsConn.Mux.Lock()
						err = wsConn.Conn.WriteMessage(mt, []byte("数据已过期，准备开始缓存"))
						wsConn.Mux.Unlock()
						if err != nil {
							zap.L().Error("wsConn.Conn.WriteMessage fail", zap.Error(err))
							return
						}
						fmt.Printf("logic.GetDataByKey fail %v", err)
						//5s 订阅一次，避免newMarket报错
						//订阅k线图的数据
						err := logic.StartSetKlineData()
						if err != nil {
							zap.L().Error("StartSetKlineData fail", zap.Error(err))
							return
						}
						time.Sleep(2 * time.Second)
						//订阅行情的数据
						err = logic.StartSetQuotation()
						if err != nil {
							zap.L().Error("StartSetQuotation fail", zap.Error(err))
							return
						}
						time.Sleep(10 * time.Second)
					}
				}
				//把读到数据，序列化后返回
				websocketData := utils.Strval(data)
				wsConn.Mux.Lock()
				err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
				wsConn.Mux.Unlock()
				if err != nil {
					zap.L().Error("wsConn.Conn.WriteMessage fail", zap.Error(err))
					ws.Close()
					return
				}
				//每2s推送一次
				time.Sleep(time.Second * 2)
			}
		}()

		//}
		//第二部分逻辑  输入参数为"market.btcusdt.kline.5min" ，等数据库不存在的数据，直接转发60秒后取消订阅，刷新后重新订阅（先不做看性能如何）
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		//msg 用户输入的数据byte转string
		//直接拿msg去订阅
		//连接一个market
		//market, err := huobi.NewMarket()
		fmt.Println("这是第二个流程")
		//fmt.Println(msg)
		//str := "market.btcusdt.kline.5min"
		//newMsg := string([]byte(msg)[1:len([]byte(msg))-1])
		//if newMsg != str{
		//	println(str+"111")
		//	println(newMsg+"222")
		//	println("1111111")
		//}

		//-----------------------
		//newMsg := string([]byte(msg)[1:len([]byte(msg))-1])
		//if true{
		//	err := market.Subscribe(newMsg, func(topic string, hjson *huobiapi.JSON) {
		//		if err != nil{
		//			zap.L().Error("market.Subscrib fail", zap.Error(err))
		//		}
		//		//订阅成功
		//		//fmt.Println("订阅成功")
		//		//120后自动取消订阅
		//		go func() {
		//			time.Sleep(1*time.Minute)
		//			//fmt.Println("取消订阅成功")
		//			market.Unsubscribe(newMsg)
		//			//market.ReceiveTimeout
		//
		//		}()
		//
		//		// 收到数据更新时回调
		//		fmt.Println(topic, hjson)
		//		jsondata, err := hjson.MarshalJSON()
		//		if err != nil {
		//			zap.L().Error("hjson.MarshalJSON  fail", zap.Error(err))
		//			return
		//		}
		//		//把jsondata反序列化后进行，自由币判断运算
		//		klineData := huobi.SubData{}
		//		err = json.Unmarshal(jsondata, &klineData)
		//		if err != nil {
		//			zap.L().Error("json.Unmarshal  fail", zap.Error(err))
		//			return
		//		}
		//		//自由币换算
		//		tranData := logic.TranDecimalScale2(msg, klineData)
		//		//结构体序列化后返回
		//		data, err := json.Marshal(tranData)
		//		if err != nil {
		//			zap.L().Error("json.Marshal(tranData)  fail", zap.Error(err))
		//			return
		//		}
		//		//返回数据给用户
		//		wsConn.Mux.Lock()
		//		err = wsConn.Conn.WriteMessage(mt, data)
		//		wsConn.Mux.Unlock()
		//		//time.Sleep(2*time.Second)
		//		if err != nil {
		//			fmt.Println(err)
		//			ws.Close()
		//
		//		}
		//
		//	})
		//	if err != nil {
		//		return
		//	}
		//}

		//market.Loop()
		//ws.SetCloseHandler(func(code int, text string) error {
		//	market.Close()
		//	fmt.Println(code, text)
		//	return nil
		//})
		//-----------------

	}

}

func WsHandle(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, CloseErr := upGrader.Upgrade(c.Writer, c.Request, nil)
	wsConn := &WsConn{
		ws,
		sync.RWMutex{},
	}
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	defer ws.Close() //返回前关闭
	market, err := huobi.NewMarket()
	if err != nil {
		fmt.Println(111)
		fmt.Println(err)
		fmt.Println(666)
	}
	for {
		if CloseErr != nil {
			fmt.Println(123456)
			fmt.Println(CloseErr)
			fmt.Println(err)
			fmt.Println(123456)
		}
		//读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		if err != nil {
			fmt.Println(666)
			marketErr := market.Close()
			if marketErr != nil {
				fmt.Println("关闭连接失败1")
				fmt.Println(marketErr.Error())
				fmt.Println("关闭连接失败2")
				return
			} else {
				fmt.Println("关闭成功")
			}
			fmt.Println(err)
			fmt.Println(666)
			break
		}
		//对数据进行切割，读取参数
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		msg := string(message)
		newMsg := string([]byte(msg)[1 : len([]byte(msg))-1])
		//打印请求参数
		fmt.Println(newMsg)

		if strings.Contains(msg, "1min") || strings.Contains(msg, "step1") {
			go func() {
				for {
					data, err := logic.GetDataByKey(msg)
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
		} else {
			//写入ws数据
			go func() {
				for {

					go func() {
						err = market.Subscribe(newMsg, func(topic string, hjson *huobiapi.JSON) {
							fmt.Println(msg)
							if err != nil {
								zap.L().Error("market.Subscrib fail", zap.Error(err))
							}
							//订阅成功
							//fmt.Println("订阅成功")
							//120后自动取消订阅
							go func() {
								time.Sleep(1 * time.Minute)
								//fmt.Println("取消订阅成功")
								market.Unsubscribe(newMsg)
								//market.ReceiveTimeout

							}()

							// 收到数据更新时回调
							fmt.Println(topic, hjson)
							jsondata, err := hjson.MarshalJSON()
							if err != nil {
								zap.L().Error("hjson.MarshalJSON  fail", zap.Error(err))
								return
							}
							//把jsondata反序列化后进行，自由币判断运算
							klineData := huobi.SubData{}
							err = json.Unmarshal(jsondata, &klineData)
							if err != nil {
								zap.L().Error("json.Unmarshal  fail", zap.Error(err))
								return
							}
							//自由币换算
							tranData := logic.TranDecimalScale2(msg, klineData)
							//结构体序列化后返回
							data, err := json.Marshal(tranData)
							if err != nil {
								zap.L().Error("json.Marshal(tranData)  fail", zap.Error(err))
								return
							}
							//返回数据给用户
							wsConn.Mux.Lock()
							err = wsConn.Conn.WriteMessage(mt, data)
							wsConn.Mux.Unlock()
							//time.Sleep(2*time.Second)
							if err != nil {
								fmt.Println(err)
								ws.Close()

							}

						})
						go func() {
							time.Sleep(60 * time.Second)
							market.Unsubscribe(newMsg)
						}()
					}()
					market.Loop()

				}

			}()
		}

	}

}

// WsHandle2 仅有K线图可用
func WsHandle2(c *gin.Context) {
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
	market, err := huobi.NewMarket()
	if err != nil {
		fmt.Println(err)
	}
	for {
		//读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		//对数据进行切割，读取参数
		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
		msg := string(message)
		newMsg := string([]byte(msg)[1 : len([]byte(msg))-1])
		//打印请求参数
		fmt.Println(newMsg)

		//写入ws数据
		go func() {
			for {

				go func() {
					err = market.Subscribe(newMsg, func(topic string, hjson *huobiapi.JSON) {
						fmt.Println(msg)
						if err != nil {
							zap.L().Error("market.Subscrib fail", zap.Error(err))
						}
						//订阅成功
						//fmt.Println("订阅成功")
						//120后自动取消订阅
						go func() {
							time.Sleep(1 * time.Minute)
							//fmt.Println("取消订阅成功")
							market.Unsubscribe(newMsg)
							//market.ReceiveTimeout

						}()

						// 收到数据更新时回调
						fmt.Println(topic, hjson)
						jsondata, err := hjson.MarshalJSON()
						if err != nil {
							zap.L().Error("hjson.MarshalJSON  fail", zap.Error(err))
							return
						}
						//把jsondata反序列化后进行，自由币判断运算
						klineData := huobi.SubData{}
						err = json.Unmarshal(jsondata, &klineData)
						if err != nil {
							zap.L().Error("json.Unmarshal  fail", zap.Error(err))
							return
						}
						//自由币换算
						tranData := logic.TranDecimalScale2(msg, klineData)
						//结构体序列化后返回
						data, err := json.Marshal(tranData)
						if err != nil {
							zap.L().Error("json.Marshal(tranData)  fail", zap.Error(err))
							return
						}
						//返回数据给用户
						wsConn.Mux.Lock()
						err = wsConn.Conn.WriteMessage(mt, data)
						wsConn.Mux.Unlock()
						//time.Sleep(2*time.Second)
						if err != nil {
							fmt.Println(err)
							ws.Close()

						}

					})
					go func() {
						time.Sleep(60 * time.Second)
						market.Unsubscribe(newMsg)
					}()
				}()
				market.Loop()

			}

		}()

	}

}
