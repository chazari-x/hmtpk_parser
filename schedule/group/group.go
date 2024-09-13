package group

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chazari-x/hmtpk_parser/model"
	"github.com/chazari-x/hmtpk_parser/storage"
	"github.com/chazari-x/hmtpk_parser/utils"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	r   *storage.Redis
	log *logrus.Logger
}

func NewController(client *redis.Client, logger *logrus.Logger) *Controller {
	return &Controller{r: &storage.Redis{Redis: client}, log: logger}
}

const (
	firstDayNum  = 2
	lastDayNum   = firstDayNum + 6
	numOfColumns = 5

	href = "https://hmtpk.ru/ru/students/schedule"
)

func (c *Controller) GetSchedule(label, date string, ctx context.Context) ([]model.Schedule, error) {
	var weeklySchedule []model.Schedule

	c.log.Trace(label)

	d, err := time.Parse("02.01.2006", date)
	if err != nil {
		return nil, err
	}

	year, week := d.ISOWeek()
	if utils.RedisIsNil(c.r) {
		if redisWeeklySchedule, err := c.r.Get(fmt.Sprintf("%d/%d", year, week) + ":" + label); err == nil && redisWeeklySchedule != "" {
			if json.Unmarshal([]byte(redisWeeklySchedule), &weeklySchedule) == nil {
				c.log.Trace("Данные получены из redis")
				return weeklySchedule, nil
			}
		}
	}

	href := fmt.Sprintf("%s/?group=%s&date_edu1c=%s&send=Показать#current", href, label, date)
	request, err := http.NewRequestWithContext(ctx, "POST", href, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Ошибка: %s", resp.Status))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	for scheduleElementNum := firstDayNum; scheduleElementNum <= lastDayNum; scheduleElementNum++ {
		weeklySchedule = append(weeklySchedule, c.parseDay(doc, scheduleElementNum, label))
	}

	if utils.RedisIsNil(c.r) {
		if c.r.Redis != nil {
			if marshal, err := json.Marshal(weeklySchedule); err == nil {
				if err = c.r.Set(fmt.Sprintf("%d/%d", year, week)+":"+label, string(marshal)); err != nil {
					c.log.Error(err)
				} else {
					c.log.Trace("Данные сохранены в redis")
				}
			}
		}
	}

	return weeklySchedule, nil
}

const groupsKey = "groups"

func (c *Controller) GetOptions(ctx context.Context) (options []model.Option, err error) {
	if utils.RedisIsNil(c.r) {
		var data string
		if data, err = c.r.Get(groupsKey); err == nil && data != "" {
			if json.Unmarshal([]byte(data), &options) == nil && len(options) != 0 {
				c.log.Trace("Данные получены из redis")
				return
			}
		}
	}

	request, err := http.NewRequestWithContext(ctx, "POST", href, nil)
	if err != nil {
		return
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Ошибка: %s", resp.Status))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	options = c.parseOptions(doc)

	if utils.RedisIsNil(c.r) && len(options) != 0 {
		var marshal []byte
		if marshal, err = json.Marshal(options); err == nil {
			if err = c.r.Set(groupsKey, string(marshal), 60); err != nil {
				c.log.Error(err)
			} else {
				c.log.Trace("Данные сохранены в redis")
			}
		}
	}

	return
}

func (c *Controller) parseOptions(doc *goquery.Document) (options []model.Option) {
	elements := doc.Children().Find("#group > option[value]")
	elements.Each(func(i int, s *goquery.Selection) {
		value, exists := s.Attr("value")
		if exists {
			options = append(options, model.Option{Label: s.Text(), Value: value})
		}
	})

	return
}

func (c *Controller) parseDay(doc *goquery.Document, scheduleElementNum int, name string) model.Schedule {
	scheduleDateElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-heading.edu_today > h2", scheduleElementNum))

	date := utils.GetDate(strings.Split(scheduleDateElement.Text(), ",")[0])
	var schedule = model.Schedule{
		Date: scheduleDateElement.Text(),
		Href: fmt.Sprintf("%s/?group=%s&date_edu1c=%s&send=Показать#current", href, name, date),
	}

	var before string

	lessonsElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-body > #mobile-friendly > tbody:nth-child(2)", scheduleElementNum))
	for lessonNum := 1; lessonNum > 0; lessonNum++ {
		if len(schedule.Lessons) > 0 {
			before = schedule.Lessons[len(schedule.Lessons)-1].Num
		}

		if lesson, exists := c.parseLesson(lessonsElement, lessonNum, before); exists {
			schedule.Lessons = append(schedule.Lessons, lesson)
		} else {
			break
		}
	}

	return schedule
}

func (c *Controller) parseLesson(lessonsElement *goquery.Selection, lessonNum int, before string) (model.Lesson, bool) {
	var lesson model.Lesson
	var exists bool
	lessonElement := lessonsElement.Find(fmt.Sprintf("tr:nth-child(%d)", lessonNum))
	for lessonAttributeNum := 1; lessonAttributeNum <= numOfColumns; lessonAttributeNum++ {
		lesson, exists = c.parseLessonAttribute(lessonElement, lessonAttributeNum, lesson, before)
		if !exists {
			break
		}
	}

	return lesson, exists
}

func (c *Controller) parseLessonAttribute(lessonElement *goquery.Selection, lessonAttributeNum int, lesson model.Lesson, before string) (model.Lesson, bool) {
	lessonElementAttribute := lessonElement.Find(fmt.Sprintf("td:nth-child(%d)", lessonAttributeNum))
	value, exists := lessonElementAttribute.Attr("data-title")
	if !exists {
		if lessonAttributeNum == numOfColumns {
			return lesson, true
		} else if lessonAttributeNum == 1 {
			return lesson, false
		}
	}

	text := lessonElementAttribute.Text()
	switch value {
	case "Номер урока":
		lesson.Num = text
	case "Время":
		if lesson.Num == "" {
			lesson.Num = before
		}
		lesson.Time = text
	case "Название предмета":
		if strings.HasSuffix(text, "(1)") || strings.HasSuffix(text, "(2)") {
			switch text[len(text)-3:] {
			case "(1)":
				lesson.Subgroup = "1"
			case "(2)":
				lesson.Subgroup = "2"
			}
			lesson.Name = strings.TrimSpace(strings.TrimRight(strings.TrimRight(text, " (2)"), " (1)"))
		} else {
			lesson.Name = strings.TrimSpace(text)
		}
	case "Кабинет":
		lesson.Room = text
	case "Преподаватель":
		lesson.Teacher = text
	}

	return lesson, true
}
