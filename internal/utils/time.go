package utils

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"time"
)

// 计算日期之间的所有日期，格式为YYYYMMDD
func RangeDays(dayFrom string, dayTo string) ([]string, error) {
	ok, err := regexp.MatchString(`^\d{8}$`, dayFrom)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invalid 'dayFrom'")
	}

	ok, err = regexp.MatchString(`^\d{8}$`, dayTo)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invalid 'dayTo'")
	}

	if dayFrom > dayTo {
		dayFrom, dayTo = dayTo, dayFrom
	}

	// 不能超过N天
	maxDays := 100 - 1 // -1 是去掉默认加入的dayFrom
	result := []string{dayFrom}

	year := types.Int(dayFrom[:4])
	month := types.Int(dayFrom[4:6])
	day := types.Int(dayFrom[6:])
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	for {
		t = t.AddDate(0, 0, 1)
		newDay := timeutil.Format("Ymd", t)
		if newDay <= dayTo {
			result = append(result, newDay)
		} else {
			break
		}

		maxDays--
		if maxDays <= 0 {
			break
		}
	}

	return result, nil
}

// 计算小时之间的所有小时，格式为YYYYMMDDHH
func RangeHours(hourFrom string, hourTo string) ([]string, error) {
	ok, err := regexp.MatchString(`^\d{10}$`, hourFrom)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invalid 'hourFrom'")
	}

	ok, err = regexp.MatchString(`^\d{10}$`, hourTo)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invalid 'hourTo'")
	}

	if hourFrom > hourTo {
		hourFrom, hourTo = hourTo, hourFrom
	}

	// 不能超过N天
	maxHours := 100 - 1 // -1 是去掉默认加入的dayFrom
	result := []string{hourFrom}

	year := types.Int(hourFrom[:4])
	month := types.Int(hourFrom[4:6])
	day := types.Int(hourFrom[6:8])
	hour := types.Int(hourFrom[8:])
	t := time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local)
	for {
		t = t.Add(1 * time.Hour)
		newHour := timeutil.Format("YmdH", t)
		if newHour <= hourTo {
			result = append(result, newHour)
		} else {
			break
		}

		maxHours--
		if maxHours <= 0 {
			break
		}
	}

	return result, nil
}
