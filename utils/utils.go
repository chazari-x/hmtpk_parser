package utils

import (
	"strings"
	"time"

	"github.com/chazari-x/hmtpk_parser/v2/storage"
)

func GetDate(date string) string {
	d := strings.Fields(date)
	if len(d) != 3 || len(d[1]) < 6 {
		return ""
	}
	switch d[1][:6] {
	case "янв":
		d[1] = "01"
	case "фев":
		d[1] = "02"
	case "мар":
		d[1] = "03"
	case "апр":
		d[1] = "04"
	case "май", "мая":
		d[1] = "05"
	case "июн":
		d[1] = "06"
	case "июл":
		d[1] = "07"
	case "авг":
		d[1] = "08"
	case "сен":
		d[1] = "09"
	case "окт":
		d[1] = "10"
	case "ноя":
		d[1] = "11"
	case "дек":
		d[1] = "12"
	default:
		return ""
	}

	result := strings.Join(d, ".")
	if _, err := time.Parse("2.01.2006", result); err != nil {
		return ""
	}
	return result
}

func RedisIsNil(redis *storage.Redis) bool {
	if redis != nil {
		if redis.Redis != nil {
			return true
		}
	}

	return false
}
