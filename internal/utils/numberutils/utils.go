package numberutils

import (
	"fmt"
	"github.com/iwind/TeaGo/types"
	"strconv"
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

func FloorFloat32(f float32, decimal int) float32 {
	if decimal <= 0 {
		return f
	}

	return types.Float32(fmt.Sprintf("%.2f", f))
}
