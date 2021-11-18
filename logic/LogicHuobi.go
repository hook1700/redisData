package logic

import (
	"fmt"
	"github.com/iszmxw/huobiapi"
	"github.com/iszmxw/huobiapi/market"
)

func NewMark()  *market.Market {
	market, err := huobiapi.NewMarket()
	if err != nil{
		println(err)
		return nil
	}
	return market
}



// NewSub 订阅主题接口
func NewSub(m *market.Market,symbol string,period string)  {
	// 订阅主题
	m.Subscribe(fmt.Sprintf("market.%s.kline.%s",symbol,period), func(topic string, json *huobiapi.JSON) {
		// 收到数据更新时回调
		fmt.Println(topic, json)
	})
	// 请求数据

	// 进入阻塞等待，这样不会导致进程退出
	m.Loop()
}

func CanelSub(m *market.Market,symbol string,period string)  {
	m.Unsubscribe(fmt.Sprintf("market.%s.kline.%s",symbol,period))
	fmt.Printf("取消market.%s.kline.%s 订阅成功",symbol,period)
}



