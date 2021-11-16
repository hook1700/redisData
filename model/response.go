package model

import (
	"gorm.io/gorm"
	"time"
)

// 结构体 Symbol接受SQL返回的参数

type Symbol struct {
	Name string `json:"name" db:"k_line_code"`
}

// 结构体 DecimalScale 接受SQL返回的参数

type DecimalScale struct {
	Value int `json:"name" db:"decimal_scale"`
}

type KlineData struct {
	Ch     string `json:"ch"`
	Status string `json:"status"`
	Ts     int64  `json:"ts"`
	Data   []Data `json:"data"`
}
type Data struct {
	Id     int64   `json:"id"`
	Count  int64   `json:"count"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Amount float64 `json:"amount"`
	Vol    float64 `json:"vol"`
}

//type SubData struct {
//	Ch    string `json:"ch"`
//	Ts    int64  `json:"ts"`
//	*Tick `json:"tick"`
//}
//
//type Tick struct {
//	ID     int64   `json:"id"`
//	Open   float64 `json:"open"`
//	Close  float64 `json:"close"`
//	Low    float64 `json:"low"`
//	High   float64 `json:"high"`
//	Amount float64 `json:"amount"`
//	Vol    float64 `json:"vol"`
//	Count  int64   `json:"count"`
//}

// gorm 数据库结构体

type Currency struct {
	Id                   int            `json:"id"`                     //主键id
	Name                 string         `json:"name"`                   //名称
	TradingPairId        int            `json:"trading_pair_id"`        //交易对
	KLineCode            string         `json:"k_line_code"`            //K线图代码
	DecimalScale         int            `json:"decimal_scale"`          //自有币位数
	Type                 string         `json:"type"`                   //交易显示：（币币交易，永续合约，期权合约）
	Sort                 int8           `json:"sort"`                   //排序
	IsHidden             int8           `json:"is_hidden"`              //是否展示：0-否，1-展示
	FluctuationMin       float64        `json:"fluctuation_min"`        //行情波动值（小）
	FluctuationMax       float64        `json:"fluctuation_max"`        //行情波动值（大）
	FeeCurrencyCurrency  string         `json:"fee_currency_currency"`  //币币交易手续费%
	FeePerpetualContract string         `json:"fee_perpetual_contract"` //永续合约手续费%
	FeeOptionContract    string         `json:"fee_option_contract"`    //期权合约手续费%
	CreatedAt            time.Time      `json:"created_at"`             //创建时间
	UpdatedAt            time.Time      `json:"updated_at"`             //更新时间
	DeletedAt            gorm.DeletedAt `json:"deleted_at"`             //删除时间，为 null 则是没删除
}
