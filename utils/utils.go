package utils

import (
	"strings"

	"github.com/chazari-x/hmtpk_schedule/storage"
)

func GetDate(date string) string {
	d := strings.Split(date, " ")
	switch d[1][:6] {
	case "янв":
		d[1] = "01"
	case "фев":
		d[1] = "02"
	case "мар":
		d[1] = "03"
	case "апр":
		d[1] = "04"
	case "май":
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
	}

	return strings.Join(d, ".")
}

func RedisIsNil(redis *storage.Redis) bool {
	if redis != nil {
		if redis.Redis != nil {
			return true
		}
	}

	return false
}
