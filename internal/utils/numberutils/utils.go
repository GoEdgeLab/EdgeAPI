package numberutils

import (
	"fmt"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"strings"
)

func FormatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func FormatInt(value int) string {
	return strconv.Itoa(value)
}

func Max[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](values ...T) T {
	if len(values) == 0 {
		return 0
	}
	var max T
	for index, value := range values {
		if index == 0 {
			max = value
		} else if value > max {
			max = value
		}
	}
	return max
}

func Min[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](values ...T) T {
	if len(values) == 0 {
		return 0
	}
	var min T
	for index, value := range values {
		if index == 0 {
			min = value
		} else if value < min {
			min = value
		}
	}
	return min
}

func FloorFloat64(f float64, decimal int) float64 {
	if decimal <= 0 {
		return f
	}

	var s = fmt.Sprintf("%f", f)
	var index = strings.Index(s, ".")
	if index < 0 || len(s[index:]) <= decimal+1 {
		return f
	}

	return types.Float64(s[:index+decimal+1])
}
