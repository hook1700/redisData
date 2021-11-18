package controller

//import (
//	"encoding/json"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"github.com/go-redis/redis"
//	"github.com/gorilla/websocket"
//	"github.com/iszmxw/huobiapi"
//	"net/http"
//	"redisData/huobi"
//	"redisData/logic"
//	"redisData/pkg/logger"
//	"redisData/utils"
//	"sync"
//	"time"
//)
//
//
//// TransmitController 转发功能开发
//// 逻辑1.通过ws获取到信息 2.新建一个结构体存放用户订阅列表，socketID
//
//func TransmitController(c *gin.Context)  {
//
//}
//
//func TestMap(c *gin.Context)  {
//	for i:=0;i<10;i++{
//		go logic.InitStartData()
//		fmt.Printf("第%d个线程",i)
//	}
//	c.JSON(http.StatusOK,"ok")
//}
//
//
//
//
//
//
//func GetRedisData2(c *gin.Context) {
//	//每个用户连接,就new一个 ws
//	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		logger.Error(err)
//		return
//	}
//	//defer ws.Close()
//	//返回前关闭
//	//var user UserInfo
//	//user.Uid = utils.GetGenerateId()
//	wsConn := &WsConn{
//		ws,
//		sync.RWMutex{},
//	}
//	defer func(ws *websocket.Conn) {
//		err := ws.Close()
//		if err != nil {
//			logger.Error(err)
//		}
//	}(ws) //返回前关闭
//
//	//读取客户端信息
//	for {
//		//读取ws中的数据
//		wsConn.Mux.Lock()
//		mt, message, ReadMessageErr := wsConn.Conn.ReadMessage()
//		wsConn.Mux.Unlock()
//		if err != nil {
//			logger.Error(ReadMessageErr)
//			break
//		}
//		//60秒不发送信息，关闭socket
//
//		//把用户传进来的消息进行处理 msg样式 "market.btcusdt.kline.1min"
//		msg := string(message)
//		//-------------
//		logger.Info(msg)
//		//当请求数据中含有1min或1step这些为已经缓存数据,直接去redis拿
//		//if strings.Contains(msg,"1min")||strings.Contains(msg,"step1"){
//		go func() {
//			for {
//				data, GetDataByKeyErr := logic.GetDataByKey(msg)
//				if GetDataByKeyErr != nil {
//					logger.Error(GetDataByKeyErr)
//					//如果redis数据获取或者start接口没有被调用，就要重新缓存
//					if GetDataByKeyErr == redis.Nil {
//						wsConn.Mux.Lock()
//						WriteMessageErr := wsConn.Conn.WriteMessage(mt, []byte("数据已过期，准备开始缓存"))
//						wsConn.Mux.Unlock()
//						if WriteMessageErr != nil {
//							logger.Error(WriteMessageErr)
//							return
//						}
//						fmt.Printf("logic.GetDataByKey fail %v", GetDataByKeyErr)
//						//5s 订阅一次，避免newMarket报错
//						//订阅k线图的数据
//						StartSetKlineDataErr := logic.StartSetKlineData()
//						if StartSetKlineDataErr != nil {
//							logger.Error(StartSetKlineDataErr)
//							return
//						}
//						time.Sleep(2 * time.Second)
//						//订阅行情的数据
//						StartSetQuotationErr := logic.StartSetQuotation()
//						if StartSetQuotationErr != nil {
//							logger.Error(StartSetQuotationErr)
//							return
//						}
//						time.Sleep(10 * time.Second)
//					}
//				}
//				//把读到数据，序列化后返回
//				websocketData := utils.Strval(data)
//				wsConn.Mux.Lock()
//				err = wsConn.Conn.WriteMessage(mt, []byte(websocketData))
//				wsConn.Mux.Unlock()
//				if err != nil {
//					logger.Error(err)
//					err := ws.Close()
//					if err != nil {
//						logger.Error(err)
//						return
//					}
//					return
//				}
//				//每2s推送一次
//				time.Sleep(time.Second * 2)
//			}
//		}()
//		logger.Info("这是第二个流程")
//
//	}
//
//}
//
//// WsHandle2 仅有K线图可用
//func WsHandle2(c *gin.Context) {
//	//升级get请求为webSocket协议
//	ws, _ := upGrader.Upgrade(c.Writer, c.Request, nil)
//	wsConn := &WsConn{
//		ws,
//		sync.RWMutex{},
//	}
//	defer func(ws *websocket.Conn) {
//		err := ws.Close()
//		if err != nil {
//			logger.Error(err)
//		}
//	}(ws) //返回前关闭
//	for {
//		market, err := huobi.NewMarket()
//		if err != nil {
//			logger.Info(err)
//		}
//		//读取ws中的数据
//		mt, message, err := wsConn.Conn.ReadMessage()
//		if err != nil {
//			logger.Info(err)
//			break
//		}
//		//对数据进行切割，读取参数
//		//如果请求的是market.ethbtc.kline.5min,订阅这条信息，然后再返回
//		msg := string(message)
//		newMsg := string([]byte(msg)[1 : len([]byte(msg))-1])
//		//打印请求参数
//		logger.Info(newMsg)
//
//		//写入ws数据
//		go func() {
//			for {
//
//				go func() {
//					err = market.Subscribe(newMsg, func(topic string, hjson *huobiapi.JSON) {
//						logger.Info(msg)
//						if err != nil {
//							logger.Error(err)
//						}
//						//订阅成功
//						//logger.Info("订阅成功")
//						//120后自动取消订阅
//						go func() {
//							time.Sleep(1 * time.Minute)
//							//logger.Info("取消订阅成功")
//							market.Unsubscribe(newMsg)
//							//market.ReceiveTimeout
//
//						}()
//						// 收到数据更新时回调
//						logger.Info(topic)
//						logger.Info(hjson)
//						jsondata, MarshalJSONErr := hjson.MarshalJSON()
//						if err != nil {
//							logger.Error(MarshalJSONErr)
//							return
//						}
//						//把jsondata反序列化后进行，自由币判断运算
//						klineData := huobi.SubData{}
//						err = json.Unmarshal(jsondata, &klineData)
//						if err != nil {
//							logger.Error(err)
//							return
//						}
//						//自由币换算
//						tranData := logic.TranDecimalScale2(msg, klineData)
//						//结构体序列化后返回
//						data, dataErr := json.Marshal(tranData)
//						if dataErr != nil {
//							logger.Error(dataErr)
//							return
//						}
//						//返回数据给用户
//						wsConn.Mux.Lock()
//						err = wsConn.Conn.WriteMessage(mt, data)
//						wsConn.Mux.Unlock()
//						//time.Sleep(2*time.Second)
//						if err != nil {
//							logger.Info(err)
//							err := ws.Close()
//							if err != nil {
//								logger.Info(err)
//								return
//							}
//
//						}
//
//					})
//					go func() {
//						time.Sleep(60 * time.Second)
//						market.Unsubscribe(newMsg)
//					}()
//				}()
//				market.Loop()
//
//			}
//
//		}()
//
//	}
//
//}