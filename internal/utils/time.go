package utils

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"time"
)

// 分钟时间点
type timeMinute struct {
	Day    string
	Minute string
}

// 分钟时间范围
type timeMinuteRange struct {
	Day        string
	MinuteFrom string
	MinuteTo   string
}

// RangeDays 计算日期之间的所有日期，格式为YYYYMMDD
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

// RangeMonths 计算日期之间的所有月份，格式为YYYYMM
func RangeMonths(dayFrom string, dayTo string) ([]string, error) {
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

	var result = []string{dayFrom[:6]}

	var year = types.Int(dayFrom[:4])
	var month = types.Int(dayFrom[4:6])
	var day = types.Int(dayFrom[6:])
	var t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	for {
		t = t.AddDate(0, 0, 20)
		var newDay = timeutil.Format("Ymd", t)
		if newDay <= dayTo {
			var monthString = newDay[:6]
			if !lists.ContainsString(result, monthString) {
				result = append(result, monthString)
			}
		} else {
			break
		}
	}

	var endMonth = dayTo[:6]
	if !lists.ContainsString(result, endMonth) {
		result = append(result, endMonth)
	}

	return result, nil
}

// RangeHours 计算小时之间的所有小时，格式为YYYYMMDDHH
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
	var maxHours = 100 - 1 // -1 是去掉默认加入的dayFrom
	var result = []string{hourFrom}

	var year = types.Int(hourFrom[:4])
	var month = types.Int(hourFrom[4:6])
	var day = types.Int(hourFrom[6:8])
	var hour = types.Int(hourFrom[8:])
	var t = time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local)
	for {
		t = t.Add(1 * time.Hour)
		var newHour = timeutil.Format("YmdH", t)
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

// RangeMinutes 计算若干个时间点，返回结果为 [ [day1, minute1], [day2, minute2] ... ]
func RangeMinutes(toTime time.Time, count int, everyMinutes int64) []timeMinute {
	var everySeconds = everyMinutes * 60
	if everySeconds <= 0 {
		everySeconds = 300
	}
	var result = []timeMinute{}
	var fromTime = time.Unix(toTime.Unix()-everySeconds*int64(count-1), 0)
	for {
		var timestamp = fromTime.Unix() / everySeconds * everySeconds
		result = append(result, timeMinute{
			Day:    timeutil.FormatTime("Ymd", timestamp),
			Minute: timeutil.FormatTime("Hi", timestamp),
		})
		fromTime = time.Unix(fromTime.Unix()+everySeconds, 0)

		count--
		if count <= 0 {
			break
		}
	}

	return result
}

// GroupMinuteRanges 将时间点分组
func GroupMinuteRanges(minutes []timeMinute) []timeMinuteRange {
	var result = []*timeMinuteRange{}
	var lastDay = ""
	var lastRange *timeMinuteRange
	for _, minute := range minutes {
		if minute.Day != lastDay {
			lastDay = minute.Day
			lastRange = &timeMinuteRange{
				Day:        minute.Day,
				MinuteFrom: minute.Minute,
				MinuteTo:   minute.Minute,
			}
			result = append(result, lastRange)
		} else {
			if lastRange != nil {
				lastRange.MinuteTo = minute.Minute
			}
		}
	}

	var finalResult = []timeMinuteRange{}
	for _, minutePtr := range result {
		finalResult = append(finalResult, *minutePtr)
	}
	return finalResult
}
