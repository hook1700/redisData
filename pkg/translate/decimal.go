package translate

import (
	"fmt"
	"strconv"
)

func Decimal(value float64) float64 {
	val, err := strconv.ParseFloat(fmt.Sprintf("%.4f", value), 64)
	if err != nil {
		return 0
	}
	return val
}
