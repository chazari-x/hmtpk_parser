# Парсер ХМТПК
[![Go Reference](https://pkg.go.dev/badge/github.com/chazari-x/hmtpk_schedule/v2.svg)](https://pkg.go.dev/github.com/chazari-x/hmtpk_parser/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/chazari-x/hmtpk_schedule)](https://goreportcard.com/report/github.com/chazari-x/hmtpk_parser)
![License](https://img.shields.io/github/license/chazari-x/hmtpk_parser)
[![Application](https://img.shields.io/badge/VK-Mini-App)](https://vk.com/hmtpk_schedule)
[![Group](https://img.shields.io/badge/VK-Subscripe-blue)](https://vk.com/hmtpk_schedule_club)

Этот пакет предоставляет простой способ получения данных с сайта [Ханты-Мансийский технолого-педагогический колледж (ХМТПК)](https://hmtpk.ru/ru/).

## Получаемые данные
Пакет позволяет получить следующие данные:
- Расписание занятий для группы
- Расписание занятий для преподавателя
- Список групп
- Список преподавателей
- Объявления

## Установка
Для установки пакета, выполните следующую команду:

```bash
go get github.com/chazari-x/hmtpk_schedule/v2
```

## Использование
Пример использования:

```go
package main

import (
	"context"
	"fmt"

	"github.com/chazari-x/hmtpk_parser/v2"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func main() {
	// Инициализация клиента Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Создание логгера
	logger := logrus.New()

	// Создание экземпляра структуры Controller
	controller := NewController(redisClient, logger)

	// Получение расписания для группы
	groupSchedule, err := controller.GetScheduleByGroup("Ваша_группа", "20.03.2024", context.Background())
	if err != nil {
		fmt.Println("Ошибка при получении расписания для группы:", err)
		return
	}

	// Вывод расписания на экран
	fmt.Println(groupSchedule)

	// Получение расписания для преподавателя
	teacherSchedule, err := controller.GetScheduleByTeacher("ФИО_преподавателя", "20.03.2024", context.Background())
	if err != nil {
		fmt.Println("Ошибка при получении расписания для преподавателя:", err)
		return
	}

	// Вывод расписания на экран
	fmt.Println(teacherSchedule)
}
```

## Примечание
Данный пакет использует веб-скрейпинг для извлечения данных с сайта Ханты-Мансийского технолого-педагогического колледжа. В случае изменения структуры сайта, пакет может перестать корректно работать. Если вы столкнулись с проблемой, пожалуйста, создайте issue на GitHub.

## Лицензия
Этот проект лицензирован в соответствии с условиями лицензии MIT. См. файл LICENSE для получения дополнительной информации.