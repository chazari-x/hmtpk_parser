# Парсер ХМТПК
[![Go Reference](https://pkg.go.dev/badge/github.com/chazari-x/hmtpk_parser/v2.svg)](https://pkg.go.dev/github.com/chazari-x/hmtpk_parser/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/chazari-x/hmtpk_parser/v2)](https://goreportcard.com/report/github.com/chazari-x/hmtpk_parser/v2)
![License](https://img.shields.io/github/license/chazari-x/hmtpk_parser)
[![Application](https://img.shields.io/badge/VK-Mini-App)](https://vk.com/app51786452)
[![Group](https://img.shields.io/badge/VK-Subscripe-blue)](https://vk.com/hmtpk_app_club)

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
go get github.com/chazari-x/hmtpk_parser/v2
```

## Использование
Пример использования:

```go
package main

import (
  "context"
  "fmt"
  
  hmtpk "github.com/chazari-x/hmtpk_parser/v2"
  "github.com/go-redis/redis/v8"
  "github.com/sirupsen/logrus"
)

func main() {
  // Инициализация клиента Redis
  redisClient := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
  })
  
  // Создание логгера
  logger := logrus.New()
  
  // Создание экземпляра структуры Controller
  controller := hmtpk.NewController(redisClient, logger)
  
  groupScheduleExample(controller)
  
  teacherScheduleExample(controller)
  
  announcementsExample(controller)
}

func groupScheduleExample(controller *hmtpk.Controller) {
  // Получение списка групп
  groups, err := controller.GetGroupOptions(context.Background())
  if err != nil || len(groups) == 0 {
    fmt.Println("Ошибка при получении списка групп:", err)
    return
  }
  
  // Получение расписания для группы
  groupSchedule, err := controller.GetScheduleByGroup(groups[0].Value, "20.03.2024", context.Background())
  if err != nil {
    fmt.Println("Ошибка при получении расписания для группы:", err)
    return
  }
  
  // Вывод расписания на экран
  fmt.Println(groupSchedule)
}

func teacherScheduleExample(controller *hmtpk.Controller) {
  // Получение списка преподавателей
  teachers, err := controller.GetTeacherOptions(context.Background())
  if err != nil || len(teachers) == 0 {
    fmt.Println("Ошибка при получении списка преподавателей:", err)
    return
  }
  
  // Получение расписания для преподавателя
  teacherSchedule, err := controller.GetScheduleByTeacher(teachers[0].Value, "20.03.2024", context.Background())
  if err != nil {
    fmt.Println("Ошибка при получении расписания для преподавателя:", err)
    return
  }
  
  // Вывод расписания на экран
  fmt.Println(teacherSchedule)
}

func announcementsExample(controller *hmtpk.Controller) {
  // Получение списка объявлений
  announcements, err := controller.GetAnnounces(context.Background(), 1)
  if err != nil {
    fmt.Println("Ошибка при получении списка объявлений:", err)
    return
  }
  
  // Вывод объявлений на экран 
  fmt.Println(announcements)
}

```

## Примечание
Данный пакет использует веб-скрейпинг для извлечения данных с сайта Ханты-Мансийского технолого-педагогического колледжа. В случае изменения структуры сайта, пакет может перестать корректно работать. Если вы столкнулись с проблемой, пожалуйста, создайте issue на GitHub.

## Лицензия
Этот проект лицензирован в соответствии с условиями лицензии MIT. См. файл LICENSE для получения дополнительной информации.
