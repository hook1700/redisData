package mysql

import (
	"redisData/model"
	"redisData/pkg/logger"
	"redisData/pkg/mysql"
)

func GetAllSymbol(ss *[]model.Symbol) error {
	sql := "select k_line_code as name from osx_currency"
	if err := mysql.DB.Debug().Raw(sql).Scan(&ss).Error; err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func GetDecimalScaleBySymbols(symbol string, data *model.DecimalScale) error {
	sql := "select decimal_scale as value from osx_currency where k_line_code = ?"
	if err := mysql.DB.Debug().Raw(sql, symbol).Scan(&data).Error; err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
