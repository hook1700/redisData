package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"redisData/logic"
	"redisData/pkg/logger"
	"redisData/utils"
	"sync"
	"time"
)

func main() {
	fmt.Println(3294.6371 / (50 * 0.01))
}
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
		//wsConn.Mux.Lock()
		//读取ws中的数据
		mt, message, err := wsConn.Conn.ReadMessage()
		//wsConn.Mux.Unlock()
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