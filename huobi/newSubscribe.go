package huobi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leizongmin/huobiapi"
	"github.com/leizongmin/huobiapi/market"
	"redisData/dao/mysql"
	"redisData/dao/redis"
	"redisData/model"
	"redisData/pkg/logger"
	"redisData/pkg/translate"
	//"redisData/pkg/translate"
	//"time"
)

type SubData struct {
	Ch    string `json:"ch"`
	Ts    int64  `json:"ts"`
	*Tick `json:"tick"`
}

type Tick struct {
	ID     int64   `json:"id"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Amount float64 `json:"amount"`
	Vol    float64 `json:"vol"`
	Count  int64   `json:"count"`
}

type Quotation struct {
	Ch     string `json:"ch"`
	Ts     int64  `json:"ts"`
	*Ticks `json:"tick"`
}
type Ticks struct {
	Bids    [][]float64 `json:"bids"`
	Asks    [][]float64 `json:"asks"`
	Version int64       `json:"version"`
	Ts      int64       `json:"ts"`
}

type Data struct {
	Ch       string `json:"ch"`
	Ts       int64  `json:"ts"`
	TickData string `json:"tick"`
}

// NewSubscribe 新订阅 订阅K线图的
func NewSubscribe() {
	//fmt.Printf("market.%s.kline.%s", symbol, period)
	//参数校验

	// 创建客户端实例
	marketService, err := huobiapi.NewMarket()
	if err != nil {
		logger.Error(err)
		return
	}
	// 订阅主题
	//使用循环一次订阅16条信息
	var allSymbol []model.Symbol
	GetAllSymbolErr := mysql.GetAllSymbol(&allSymbol)
	if err != nil {
		logger.Error(GetAllSymbolErr)
		return
	}
	logger.Info(len(allSymbol))
	logger.Info(allSymbol)

	// 获取数据总条数
	count := len(allSymbol)
	for key, value := range allSymbol {
		marketSubscribe(marketService, key, value)
		if key+1 == count {
			logger.Info(fmt.Sprintf("数据库共有%v条数据，全部发起订阅完毕，当前订阅的key%v", count, key))
		}
	}

	fmt.Println("下一次循环")
	return
	//market.Loop()

}

// 从火币网发送socket请求获取数据
func marketSubscribe(markSer *market.Market, key int, value model.Symbol) {
	SubscribeERR := markSer.Subscribe(fmt.Sprintf("market.%s.kline.1min", value.Name), func(topic string, hjson *huobiapi.JSON) {
		logger.Info(fmt.Sprintf("订阅第%v条数据", key))
		// 收到数据更新时回调
		logger.Info(topic)
		logger.Info(hjson)
		//jsonData是订阅后返回的信息 通过MarshalJSON将数据转化成String
		jsonData, _ := hjson.MarshalJSON()
		//mapData := utils.JSONToMap(string(jsonData))
		subData := &SubData{}
		if UnmarshalErr := json.Unmarshal(jsonData, subData); UnmarshalErr != nil {
			logger.Error(errors.New(fmt.Sprintf("json.Unmarshal subData fail err:%v", UnmarshalErr)))

		}
		//通过数据库得到 自有币位数
		var decimalscale model.DecimalScale
		GetDecimalScaleBySymbolsErr := mysql.GetDecimalScaleBySymbols(value.Name, &decimalscale)
		if GetDecimalScaleBySymbolsErr != nil {
			logger.Error(errors.New(fmt.Sprintf("mysql.GetDecimalScaleBySymbols fail err:%v", GetDecimalScaleBySymbolsErr)))
			return
		}
		logger.Info(subData)
		logger.Info(fmt.Sprintf("%v:%v", value.Name, decimalscale.Value))
		//对数据和自有币位数进行运算，返回修改后的数据
		var multiple float64 // 浮动倍数
		if decimalscale.Value > 0 {
			multiple = float64(decimalscale.Value) + 100
			logger.Info("自由比位数大于零")
			subData.Amount = translate.Decimal(subData.Amount * multiple * 0.01)
			subData.Open = translate.Decimal(subData.Open * multiple * 0.01)
			subData.Close = translate.Decimal(subData.Close * multiple * 0.01)
			subData.Low = translate.Decimal(subData.Low * multiple * 0.01)
			subData.High = translate.Decimal(subData.High * multiple * 0.01)
			subData.Vol = translate.Decimal(subData.Vol * multiple * 0.01)
		}
		if decimalscale.Value < 0 {
			multiple = float64(decimalscale.Value) + 100
			logger.Info("自由比位数xiao于零")
			subData.Amount = translate.Decimal(subData.Amount * multiple * 0.01)
			subData.Open = translate.Decimal(subData.Open * multiple * 0.01)
			subData.Close = translate.Decimal(subData.Close * multiple * 0.01)
			subData.Low = translate.Decimal(subData.Low * multiple * 0.01)
			subData.High = translate.Decimal(subData.High * multiple * 0.01)
			subData.Vol = translate.Decimal(subData.Vol * multiple * 0.01)
		}
		logger.Info(subData)
		//取出ch当key
		ch := subData.Ch
		//把修改后的对象反序列化，存进redis
		redisData, MarshalErr := json.Marshal(subData)
		if MarshalErr != nil {
			logger.Error(errors.New(fmt.Sprintf("json.Marshal(subData) fail err:%v", MarshalErr)))
		}
		//根据推送返回的数据，以字符串的形式存入reids
		//redis.CreateOrChangeKline(topic, string(redisData))
		if len(ch) > 0 && len(string(redisData)) > 0 {
			redis.CreateOrChangeKline(fmt.Sprintf("\"%s\"", ch), string(redisData))
		} else {
			logger.Error(errors.New("异常数据"))
			logger.Info("当前的异常数据的标题" + ch)
			logger.Info("当前的异常数据的内容" + string(redisData))
		}
		//fmt.Println(string(redisData))
	})
	if SubscribeERR != nil {
		logger.Error(SubscribeERR)
		return
	}
}

// NewQuotation 新订阅 订阅行情的
func NewQuotation() {
	// 创建客户端实例
	marketNewQuotation, err := huobiapi.NewMarket()
	if err != nil {
		fmt.Println(err)
		//err = market.ReConnect()
		if err != nil {
			logger.Error(err)
			return
		}
	}
	//订阅行情信息
	//"market.btcusdt.depth.step0"
	// 订阅主题
	//使用循环一次订阅16条信息

	var allSymbol []model.Symbol
	GetAllSymbolErr := mysql.GetAllSymbol(&allSymbol)
	if GetAllSymbolErr != nil {
		logger.Error(GetAllSymbolErr)
		return
	}
	for _, value := range allSymbol {
		SubscribeErr := marketNewQuotation.Subscribe(fmt.Sprintf("market.%s.depth.step1", value.Name), func(topic string, hjson *huobiapi.JSON) {
			// 收到数据更新时回调
			//fmt.Println(topic, hjson)
			jsonData, _ := hjson.MarshalJSON()
			//println(string(jsonData))

			data := new(Quotation)
			UnmarshalErr := json.Unmarshal(jsonData, data)
			if UnmarshalErr != nil {
				logger.Error(UnmarshalErr)
				return
			}
			//redis.CreateRedisData(fmt.Sprintf("\"%s\"", topic), string(jsonData))
			//fmt.Printf("%#v", data.Ticks.Bids[1][0])
			//根据自由币变量修改
			//通过数据库得到 自有币位数
			var decimalscale model.DecimalScale
			GetDecimalScaleBySymbolsErr := mysql.GetDecimalScaleBySymbols(value.Name, &decimalscale)
			if GetDecimalScaleBySymbolsErr != nil {
				logger.Error(errors.New(fmt.Sprintf("mysql.GetDecimalScaleBySymbols fail err:%v", GetDecimalScaleBySymbolsErr)))
				return
			}
			//对数据和自有币位数进行运算，返回修改后的数据
			if decimalscale.Value > 0 {
				for i := 0; i < len(data.Asks); i++ {
					data.Asks[i][0] = data.Asks[i][0] * float64(decimalscale.Value) * 0.01
				}
				for i := 0; i < len(data.Bids); i++ {
					data.Bids[i][0] = data.Bids[i][0] * float64(decimalscale.Value) * 0.01
				}

			}
			if decimalscale.Value < 0 {
				decimalscale.Value = decimalscale.Value * -1
				for i := 0; i < len(data.Asks); i++ {
					data.Asks[i][0] = data.Asks[i][0] / float64(decimalscale.Value) * 0.01
				}
				for i := 0; i < len(data.Bids); i++ {
					data.Bids[i][0] = data.Bids[i][0] / float64(decimalscale.Value) * 0.01
				}
			}
			//取出ch当key使用
			ch := data.Ch
			//序列化，存进redis
			jsonData, err = json.Marshal(data)
			if err != nil {
				fmt.Printf("json.Marshal(data) fail,err%s", err)
			}
			redis.CreateRedisData(fmt.Sprintf("\"%s\"", ch), string(jsonData))

		})
		if SubscribeErr != nil {
			logger.Error(SubscribeErr)
			return
		}
	}
	marketNewQuotation.Loop()
}

func NewSubscribeParam(symbol string, period string) {
	// 创建客户端实例
	marketNewSubscribeParam, err := huobiapi.NewMarket()
	if err != nil {
		logger.Error(err)
		err = marketNewSubscribeParam.ReConnect()
		if err != nil {
			logger.Error(err)
			return
		}
	}
	// 订阅主题
	//使用循环一次订阅16条信息
	//allSymbol, err := mysql.GetAllSymbol()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	SubscribeErr := marketNewSubscribeParam.Subscribe(fmt.Sprintf("market.%s.kline.%s", symbol, period), func(topic string, hjson *huobiapi.JSON) {
		// 收到数据更新时回调
		//fmt.Println(topic, json)

		//redis.CreateOrChangeKline(topic, *json)
		//fmt.Println(json)
		//utils.JSONToMap(string(json.MarshalJSON()))
		//jsonData是订阅后返回的信息 通过MarshalJSON将数据转化成String
		jsonData, _ := hjson.MarshalJSON()
		//mapData := utils.JSONToMap(string(jsonData))
		subData := &SubData{}
		if err := json.Unmarshal(jsonData, subData); err != nil {
			fmt.Printf("json.Unmarshal subData fail err:%v", err)
		}
		//通过数据库得到 自有币位数
		var decimalscale model.DecimalScale
		GetDecimalScaleBySymbolsErr := mysql.GetDecimalScaleBySymbols(symbol, &decimalscale)
		if GetDecimalScaleBySymbolsErr != nil {
			logger.Error(errors.New(fmt.Sprintf("mysql.GetDecimalScaleBySymbols fail err:%v", GetDecimalScaleBySymbolsErr)))
			return
		}
		//对数据和自有币位数进行运算，返回修改后的数据
		if decimalscale.Value > 0 {
			subData.Amount = subData.Amount * float64(decimalscale.Value) * 0.01
			subData.Open = subData.Open * float64(decimalscale.Value) * 0.01
			subData.Close = subData.Close * float64(decimalscale.Value) * 0.01
			subData.Low = subData.Low * float64(decimalscale.Value) * 0.01
			subData.High = subData.High * float64(decimalscale.Value) * 0.01
			subData.Vol = subData.Vol * float64(decimalscale.Value) * 0.01
		}
		if decimalscale.Value < 0 {
			decimalscale.Value = decimalscale.Value * -1
			subData.Amount = subData.Amount / float64(decimalscale.Value) * 0.01
			subData.Open = subData.Open / float64(decimalscale.Value) * 0.01
			subData.Close = subData.Close / float64(decimalscale.Value) * 0.01
			subData.Low = subData.Low / float64(decimalscale.Value) * 0.01
			subData.High = subData.High / float64(decimalscale.Value) * 0.01
			subData.Vol = subData.Vol / float64(decimalscale.Value) * 0.01
		}
		//取出ch当key
		ch := subData.Ch
		//把修改后的对象反序列化，存进redis
		redisData, err := json.Marshal(subData)
		if err != nil {
			fmt.Printf("json.Marshal(subData) fail err:%v", err)
		}
		//根据推送返回的数据，以字符串的形式存入reids
		//redis.CreateOrChangeKline(topic, string(redisData))
		redis.CreateOrChangeKline(fmt.Sprintf("\"%s\"", ch), string(redisData))
		//fmt.Println(string(redisData))

	})
	if SubscribeErr != nil {
		logger.Error(SubscribeErr)
		return
	}

	marketNewSubscribeParam.Loop()
}
