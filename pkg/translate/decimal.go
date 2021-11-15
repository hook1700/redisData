package translate

import (
	"fmt"
	"redisData/pkg/logger"
	"strconv"
)

func Decimal(value float64) float64 {
	val, err := strconv.ParseFloat(fmt.Sprintf("%.4f", value), 64)
	if err != nil {
		logger.Info("保留四位小数失败")
		logger.Info(err)
		return 0
	}
	return val
}
