package mysql

import (
	"fmt"
	"redisData/model"
	"redisData/pkg/logger"
)

func GetAllSymbol() (Symbols *[]model.Symbol, err error) {
	ss := new([]model.Symbol)
	sql := `select k_line_code from osx_currency`
	if err = db.Select(ss, sql); err != nil {
		fmt.Errorf("查询数据失败")
		return nil, err
	}
	//fmt.Printf("ss is %v", *ss)
	return ss, nil
}

func GetDecimalScaleBySymbols(symbol string) (*model.DecimalScale, error) {
	var d model.DecimalScale
	sql := `select decimal_scale from osx_currency where k_line_code = ?`
	logger.Info(sql)
	logger.Info(symbol)
	if err := db.Get(&d, sql, symbol); err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return nil, err
	}
	return &d, nil
}
