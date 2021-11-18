package test
//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/iszmxw/huobiapi"
//	"redisData/dao/mysql"
//	"redisData/dao/redis"
//	"redisData/model"
//	"redisData/pkg/logger"
//	"redisData/pkg/translate"
//)
//
//// NewQuotation 新订阅 订阅行情的
//func NewQuotation() {
//	// 创建客户端实例
//	marketNewQuotation, err := huobiapi.NewMarket()
//	if err != nil {
//		logger.Error(err)
//		//err = market.ReConnect()
//		if err != nil {
//			logger.Error(err)
//			return
//		}
//	}
//	//订阅行情信息
//	//"market.btcusdt.depth.step0"
//	// 订阅主题
//	//使用循环一次订阅16条信息
//	var allSymbol []model.Symbol
//	GetAllSymbolErr := mysql.GetAllSymbol(&allSymbol)
//	if GetAllSymbolErr != nil {
//		logger.Error(GetAllSymbolErr)
//		return
//	}
//	count := len(allSymbol)
//	for key, value := range allSymbol {
//		QuotationSubscribe(marketNewQuotation, key, value)
//		if key+1 == count {
//			logger.Info(fmt.Sprintf("数据库共有%v条数据，全部发起订阅完毕，当前订阅的key%v\n", count, key))
//		}
//	}
//	fmt.Println("下一次循环")
//	return
//	//market.Loop()
//
//}
//
//func QuotationSubscribe(markServer *market.Market, key int, value model.Symbol) {
//	SubscribeERR := markServer.Subscribe(fmt.Sprintf("market.%s.depth.step1", value.Name), func(topic string, hjson *huobiapi.JSON) {
//		//logger.Info(fmt.Sprintf("订阅第%v条数据", key))
//		// 收到数据更新时回调
//		//logger.Info(topic)
//		//logger.Info(hjson)
//		//jsonData是订阅后返回的信息 通过MarshalJSON将数据转化成String
//		jsonData, _ := hjson.MarshalJSON()
//		//mapData := utils.JSONToMap(string(jsonData))
//		data := new(Quotation)
//		UnmarshalErr := json.Unmarshal(jsonData, data)
//		if UnmarshalErr != nil {
//			logger.Error(errors.New(fmt.Sprintf("json.Unmarshal subData fail err:%v", UnmarshalErr)))
//			return
//		}
//		logger.Info("输出原来行情的数据")
//		logger.Info(data)
//
//		//自由币换算
//
//		var decimalscale model.DecimalScale
//		GetDecimalScaleBySymbolsErr := mysql.GetDecimalScaleBySymbols(value.Name, &decimalscale)
//		if GetDecimalScaleBySymbolsErr != nil {
//			logger.Error(errors.New(fmt.Sprintf("mysql.GetDecimalScaleBySymbols fail err:%v", GetDecimalScaleBySymbolsErr)))
//			return
//		}
//		fmt.Printf("获取自由币涨幅,键值对%v，自由币位数%v\n", value.Name, decimalscale)
//		//logger.Info("这个是行情接口")
//		//logger.Info( value.Name)
//		//logger.Info(decimalscale)
//		//对数据和自有币位数进行运算，返回修改后的数据
//		var multiple float64
//		if decimalscale.Value > 0 {
//			logger.Info("自由币位数大于零")
//			multiple = float64(decimalscale.Value) + 100
//			for i := 0; i < len(data.Asks); i++ {
//				data.Asks[i][0] = translate.Decimal(data.Asks[i][0] * multiple * 0.01)
//			}
//			for i := 0; i < len(data.Bids); i++ {
//				data.Bids[i][0] = translate.Decimal(data.Bids[i][0] * multiple * 0.01)
//			}
//
//		}
//		if decimalscale.Value < 0 {
//			logger.Info("自由币位数小于零")
//			multiple = float64(decimalscale.Value) + 100
//			for i := 0; i < len(data.Asks); i++ {
//				data.Asks[i][0] = translate.Decimal(data.Asks[i][0] * multiple * 0.01)
//			}
//			for i := 0; i < len(data.Bids); i++ {
//				data.Bids[i][0] = translate.Decimal(data.Bids[i][0] * multiple * 0.01)
//			}
//		}
//		logger.Info("输出自由币换算后行情的数据")
//		logger.Info(data)
//		//取出ch当key使用
//		ch := data.Ch
//		//序列化，存进redis
//		jsonData, MarshalErr := json.Marshal(data)
//		if MarshalErr != nil {
//			logger.Error(MarshalErr)
//		}
//		//存进redis
//		redis.CreateRedisData(fmt.Sprintf("\"%s\"", ch), string(jsonData))
//		//长度小于0不存
//		if len(ch) > 0 && len(string(jsonData)) > 0 {
//			redis.CreateRedisData(fmt.Sprintf("\"%s\"", ch), string(jsonData))
//		} else {
//			logger.Error(errors.New("异常数据"))
//			logger.Info("当前的异常数据的标题" + ch)
//			logger.Info("当前的异常数据的内容" + string(jsonData))
//		}
//
//	})
//	if SubscribeERR != nil {
//		logger.Error(SubscribeERR)
//		return
//	}
//}
