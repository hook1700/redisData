package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"redisData/dao/mysql"
	"redisData/dao/redis"
	"redisData/huobi"
	"redisData/model"
	"redisData/pkg/logger"
	"redisData/pkg/translate"
	"redisData/utils"
	"time"
)

var (
	ErrorUnmarshalFail = errors.New("UnmarshalFail")
)

// StartSetKlineData main.go时，默认缓存16种symbol,1min的数据
func StartSetKlineData() error {

	go huobi.NewSubscribe()
	return nil
}

// StartSetQuotation 自动获取行情数据
func StartSetQuotation() error {

	go huobi.NewQuotation()
	return nil
}

// key 是 kline:xxxx
// GetDataByKey 获取key通过kline

func GetDataByKey(key string) (interface{}, error) {
	//根据key获取值
	kline, err := redis.GetKline(key)
	if err != nil {
		return nil, err
	}
	//将对应key中的value值，将string转化成json后返回
	data := []byte(kline)
	var i interface{}
	//3.解析
	if err := json.Unmarshal(data, &i); err != nil {
		fmt.Println(err)
		return nil, ErrorUnmarshalFail
	}
	return i, nil
}




// SetKlineHistory 开始缓存k线图的历史数据
func SetKlineHistory() error {

	//通过访问mysql获取切片
	var symbol []model.Symbol
	GetAllSymbolErr := mysql.GetAllSymbol(&symbol)
	if GetAllSymbolErr != nil {
		logger.Error(GetAllSymbolErr)
		return GetAllSymbolErr
	}
	ss := make([]string, 0, len(symbol))
	for _, value := range symbol {
		ss = append(ss, value.Name)
	}
	//fmt.Printf("ss is %v", ss)
	//fmt.Printf("ss is %T", ss)

	go func() {
		//传入切片，拼接url参数发起请求，把数据存进redis
		for {
			for i := 0; i < len(ss); i++ {
				url := fmt.Sprintf("https://api.huobi.pro/market/history/kline?period=1min&size=300&symbol=%s", ss[i])
				response, err := http.Get(url)
				if err != nil {
					//log.Fatalf("get api fail err is %v", err)
					logger.Error(err)
					return
				}
				body, _ := ioutil.ReadAll(response.Body)
				//自由币换算
				var kline model.KlineData
				////序列化
				err = json.Unmarshal(body, &kline)
				if err != nil {
					fmt.Println(err)
					return
				}
				scale := TranDecimalScale(ss[i], kline)
				////反序列化
				d, _ := json.Marshal(scale)
				data := string(d)

				//把数据写进redis
				//fmt.Println("redis开始写数据")
				redis.CreateHistoryKline(fmt.Sprintf("\"market.%s.kline.1min\"", ss[i]), data)
				//redis.CreateOrChangeKline(ss[i], data)
				//fmt.Println("redis结束写数据")

			}
			time.Sleep(time.Second * 30)
		}

	}()

	return nil

}

// 新增HistoryKlineKey
func CreateHistoryKline(key string, i interface{}) {
	redis.CreateHistoryKline(key, i)
}


//GetKlineHistory 获取300条k线数据，增加时间参数
func GetKlineHistoryDiy(symbol string, period string) (interface{}, error) {
	if period == "" {
		period = "1min"
	}
	//根据key获取值
	klineHistoryData, err := redis.GetKlineHistory(fmt.Sprintf("\"market.%s.kline.%s\"", symbol, period))
	if err != nil {
		return nil, err
	}
	//将对应key中的value值，将string转化成json后返回
	data := []byte(klineHistoryData)
	var i interface{}
	//3.解析
	if err := json.Unmarshal(data, &i); err != nil {
		fmt.Println(err)
		return nil, ErrorUnmarshalFail
	}
	return i, nil
}

//RequestHuobiKilne 封装火币网请求K线图http请求  "market.btcusdt.kline.5min"
func RequestHuobiKilne(symbol string, period string) ([]byte, error) {
	url := fmt.Sprintf("https://api.huobi.pro/market/history/kline?period=%s&size=300&symbol=%s", period, symbol)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	body, _ := ioutil.ReadAll(response.Body)
	return body, nil
}

// TranDecimalScale 封装自由币换算
func TranDecimalScale(symbol string, data model.KlineData) *model.KlineData {
	//通过数据库得到 自有币位数
	var decimalscale model.DecimalScale
	GetDecimalScaleBySymbolsErr := mysql.GetDecimalScaleBySymbols(symbol, &decimalscale)
	if GetDecimalScaleBySymbolsErr != nil {
		logger.Error(errors.New(fmt.Sprintf("mysql.GetDecimalScaleBySymbols fail err:%v", GetDecimalScaleBySymbolsErr)))
		return nil
	}
	//对数据和自有币位数进行运算，返回修改后的数据
	var multiple float64 // 浮动倍数
	for i := 0; i < len(data.Data); i++ {
		if decimalscale.Value > 0 {
			multiple = float64(decimalscale.Value) + 100
			logger.Info("自由比位数大于零")
			data.Data[i].Amount =translate.Decimal(data.Data[i].Amount * multiple * 0.01)
			data.Data[i].Open = translate.Decimal(data.Data[i].Open * multiple * 0.01)
			data.Data[i].Close = translate.Decimal(data.Data[i].Close * multiple * 0.01)
			data.Data[i].Low = translate.Decimal(data.Data[i].Low * multiple * 0.01)
			data.Data[i].High = translate.Decimal(data.Data[i].High * multiple * 0.01)
			data.Data[i].Vol = translate.Decimal(data.Data[i].Vol * multiple * 0.01)
		}
		if decimalscale.Value < 0 {
			multiple = float64(decimalscale.Value) + 100
			logger.Info("自由比位数xiao于零")
			data.Data[i].Amount =translate.Decimal(data.Data[i].Amount * multiple * 0.01)
			data.Data[i].Open = translate.Decimal(data.Data[i].Open * multiple * 0.01)
			data.Data[i].Close = translate.Decimal(data.Data[i].Close * multiple * 0.01)
			data.Data[i].Low = translate.Decimal(data.Data[i].Low * multiple * 0.01)
			data.Data[i].High = translate.Decimal(data.Data[i].High * multiple * 0.01)
			data.Data[i].Vol = translate.Decimal(data.Data[i].Vol * multiple * 0.01)
		}
		//序列化内部数据
		//json.Marshal(&data.Data)
		//if err != nil {
		//	fmt.Println("是不是内部除了问题")
		//	return nil
		//}

	}
	return &data
}

// TranDecimalScale2 封装自由币换算,修改下参数,从symbol改成sub, 这个方法是订阅5分钟数据时调用
func TranDecimalScale2(sub string, subData huobi.SubData) *huobi.SubData {
	//字符串切割 去双引号
	sub = string([]byte(sub)[1 : len(sub)-1])
	//通过“.”分割,取出symbol
	res := utils.Split(sub, ".")
	fmt.Println(res[1])
	//去双引号
	//通过数据库得到 自有币位数
	var decimalscale model.DecimalScale
	GetDecimalScaleBySymbolsErr := mysql.GetDecimalScaleBySymbols(res[1], &decimalscale)
	if GetDecimalScaleBySymbolsErr != nil {
		logger.Error(errors.New(fmt.Sprintf("mysql.GetDecimalScaleBySymbols fail err:%v", GetDecimalScaleBySymbolsErr)))
		return nil
	}
	//对数据和自有币位数进行运算，返回修改后的数据
	var multiple float64 // 浮动倍数
	//for i := 0;i < len(data.Data);i++{
	if decimalscale.Value > 0 {

		multiple = float64(decimalscale.Value) + 100
		logger.Info("自由比位数大于零")
		subData.Amount = translate.Decimal(subData.Amount * multiple * 0.01)
		subData.Open = translate.Decimal(subData.Open * multiple * 0.01)
		subData.Close = translate.Decimal(subData.Close * multiple * 0.01)
		subData.Low = translate.Decimal(subData.Low * multiple * 0.01)
		subData.High = translate.Decimal(subData.High * multiple * 0.01)
		subData.Vol = translate.Decimal(subData.Vol * multiple * 0.01)
		//subData.Amount = translate.Decimal(subData.Amount * float64(decimalscale.Value) * 0.01)
		//subData.Open = translate.Decimal(subData.Open * float64(decimalscale.Value) * 0.01)
		//subData.Close = translate.Decimal(subData.Close * float64(decimalscale.Value) * 0.01)
		//subData.Low = translate.Decimal(subData.Low * float64(decimalscale.Value) * 0.01)
		//subData.High = translate.Decimal(subData.High * float64(decimalscale.Value) * 0.01)
		//subData.Vol = translate.Decimal(subData.Vol * float64(decimalscale.Value) * 0.01)
	}
	if decimalscale.Value < 0 {
		multiple = float64(decimalscale.Value) + 100
		logger.Info("自由比位数大于零")
		subData.Amount = translate.Decimal(subData.Amount * multiple * 0.01)
		subData.Open = translate.Decimal(subData.Open * multiple * 0.01)
		subData.Close = translate.Decimal(subData.Close * multiple * 0.01)
		subData.Low = translate.Decimal(subData.Low * multiple * 0.01)
		subData.High = translate.Decimal(subData.High * multiple * 0.01)
		subData.Vol = translate.Decimal(subData.Vol * multiple * 0.01)
		//decimalscale.Value = decimalscale.Value * -1
		//subData.Amount = translate.Decimal(subData.Amount / float64(decimalscale.Value) * 0.01)
		//subData.Open = translate.Decimal(subData.Open / float64(decimalscale.Value) * 0.01)
		//subData.Close = translate.Decimal(subData.Close / float64(decimalscale.Value) * 0.01)
		//subData.Low = translate.Decimal(subData.Low / float64(decimalscale.Value) * 0.01)
		//subData.High = translate.Decimal(subData.High / float64(decimalscale.Value) * 0.01)
		//subData.Vol = translate.Decimal(subData.Vol / float64(decimalscale.Value) * 0.01)
	}



	return &subData
}

//判断key是否已经缓存
func ExistKey(key string) bool {
	existKey := redis.ExistKey(key)
	return existKey
}

func InitStartData() {
	//启动获取k线图数据
	if err := StartSetKlineData(); err != nil {
		fmt.Printf("logic.StartSetKlineData() fail err:%v", err)
	}
	//启动获取行情数据
	if err := StartSetQuotation(); err != nil {
		fmt.Printf("logic.StartSetQuotation() fail err:%v", err)
	}
	go func() {
		err := SetKlineHistory()
		if err != nil {
			fmt.Println(err)
		}
	}()
	zap.L().Debug("初始化数据执行完毕")
}
