# ХМТПК Расписание

Этот пакет предоставляет простой способ получения расписания занятий с сайта [Ханты-Мансийский технолого-педагогический колледж (ХМТПК)](https://hmtpk.ru/ru/).

## Установка
Для установки пакета, выполните следующую команду:

```bash
go get github.com/chazari-x/hmtpk_schedule
```

## Использование
Пример использования:

```go
package main

import (
	"context"
	"fmt"

	"github.com/chazari-x/hmtpk_schedule"
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