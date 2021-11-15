package model

// 结构体 Symbol接受SQL返回的参数

type Symbol struct {
	Name string `json:"name" db:"k_line_code"`
}

// 结构体 DecimalScale 接受SQL返回的参数

type DecimalScale struct {
	Value int `json:"name" db:"decimal_scale"`
}

type KlineData struct {
	Ch string `json:"ch"`
	Status string `json:"status"`
	Ts int64 `json:"ts"`
	Data []Data  `json:"data"`
}
type Data struct {
	Id int64 `json:"id"`
	Count int64 `json:"count"`
	Open float64 `json:"open"`
	Close float64 `json:"close"`
	Low float64 `json:"low"`
	High float64 `json:"high"`
	Amount float64 `json:"amount"`
	Vol float64 `json:"vol"`

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