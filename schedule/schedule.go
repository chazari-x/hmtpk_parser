package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chazari-x/hmtpk_schedule/config"
	"github.com/chazari-x/hmtpk_schedule/model"
	"github.com/chazari-x/hmtpk_schedule/redis"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Schedule struct {
	cfg *config.Schedule
	r   *redis.Redis
}

var Groups map[string]int

func NewSchedule(cfg *config.Schedule, r *redis.Redis) *Schedule {
	s := &Schedule{cfg, r}
	Groups = make(map[string]int)
	for _, group := range cfg.Groups {
		Groups[group.Name] = group.ID
	}

	return s
}

func getDate(date string) string {
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

func (s *Schedule) GetGroups() []string {
	var groups []string

	for _, g := range s.cfg.Groups {
		groups = append(groups, g.Name)
	}

	return groups
}

func (s *Schedule) GetGroup(groupName string) string {
	return strconv.Itoa(Groups[groupName])
}

func (s *Schedule) GetTeachers() []string {
	var teachers []string

	for _, g := range s.cfg.Teachers {
		teachers = append(teachers, g.Name)
	}

	return teachers
}

func (s *Schedule) GetScheduleByGroupName(group, date string) ([]model.Schedule, error) {
	return s.GetScheduleByGroup(s.GetGroup(group), date)
}

func (s *Schedule) GetScheduleByGroup(group, date string) ([]model.Schedule, error) {
	var weeklySchedule []model.Schedule

	log.Trace(group)

	d, err := time.Parse("02.01.2006", date)
	if err != nil {
		return nil, err
	}

	year, week := d.ISOWeek()
	if redisWeeklySchedule, err := s.r.Get(fmt.Sprintf("%d/%d", year, week) + ":" + group); err == nil && redisWeeklySchedule != "" {
		if json.Unmarshal([]byte(redisWeeklySchedule), &weeklySchedule) == nil {
			log.Trace("Данные получены из Redis")
			return weeklySchedule, nil
		}
	}

	href := fmt.Sprintf("https://hmtpk.ru/ru/students/schedule/?group=%s&date_edu1c=%s&send=Показать#current", group, date)
	resp, err := http.Post(href, "", nil)
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

	for scheduleElementNum := 2; scheduleElementNum <= 8; scheduleElementNum++ {
		scheduleDateElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-heading.edu_today > h2", scheduleElementNum))

		date = getDate(strings.Split(scheduleDateElement.Text(), ",")[0])

		weeklySchedule = append(weeklySchedule, model.Schedule{
			Date: scheduleDateElement.Text(),
			Href: fmt.Sprintf("https://hmtpk.ru/ru/students/schedule/?group=%s&date_edu1c=%s&send=Показать#current", group, date),
		})

		lessonsElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-body > #mobile-friendly > tbody:nth-child(2)", scheduleElementNum))
		var lessons []model.Lesson
		for lessonNum := 1; lessonNum < 14; lessonNum++ {
			var lesson model.Lesson
			var exists bool
			lessonElement := lessonsElement.Find(fmt.Sprintf("tr:nth-child(%d)", lessonNum))
			for lessonAttributeNum := 1; lessonAttributeNum <= 5; lessonAttributeNum++ {
				lessonElementAttribute := lessonElement.Find(fmt.Sprintf("td:nth-child(%d)", lessonAttributeNum))
				var value string
				value, exists = lessonElementAttribute.Attr("data-title")
				if exists {
					text := lessonElementAttribute.Text()
					switch value {
					case "Номер урока":
						lesson.Num = text
					case "Время":
						if lesson.Num == "" {
							lesson.Num = lessons[len(lessons)-1].Num
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
				} else if lessonAttributeNum == 5 {
					exists = true
				} else if lessonAttributeNum == 1 {
					break
				}
			}

			if exists {
				lessons = append(lessons, lesson)
			} else {
				break
			}
		}

		weeklySchedule[len(weeklySchedule)-1].Lessons = lessons
	}

	if marshal, err := json.Marshal(weeklySchedule); err == nil {
		if err = s.r.Set(fmt.Sprintf("%d/%d", year, week)+":"+group, string(marshal)); err != nil {
			log.Error(err)
		} else {
			log.Trace("Данные сохранены в Redis")
		}
	}

	return weeklySchedule, nil
}

func (s *Schedule) GetScheduleByTeacher(teacher, date string) ([]model.Schedule, error) {
	var weeklySchedule []model.Schedule

	log.Trace(teacher)

	teacher = strings.ReplaceAll(teacher, " ", "+")
	d, err := time.Parse("02.01.2006", date)
	if err != nil {
		return nil, err
	}

	year, week := d.ISOWeek()
	if redisWeeklySchedule, err := s.r.Get(fmt.Sprintf("%d/%d", year, week) + ":" + teacher); err == nil && redisWeeklySchedule != "" {
		if json.Unmarshal([]byte(redisWeeklySchedule), &weeklySchedule) == nil {
			log.Trace("Данные получены из Redis")
			return weeklySchedule, nil
		}
	}

	href := fmt.Sprintf("https://hmtpk.ru/ru/teachers/schedule/?teacher=%s&date_edu1c=%s&send=Показать#current", teacher, date)
	resp, err := http.Post(href, "", nil)
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

	for scheduleElementNum := 1; scheduleElementNum <= 7; scheduleElementNum++ {
		scheduleDateElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-heading.edu_today > h2", scheduleElementNum))

		date = getDate(strings.Split(scheduleDateElement.Text(), ",")[0])

		weeklySchedule = append(weeklySchedule, model.Schedule{
			Date: scheduleDateElement.Text(),
			Href: fmt.Sprintf("https://hmtpk.ru/ru/teachers/schedule/?teacher=%s&date_edu1c=%s&send=Показать#current", teacher, date),
		})

		lessonsElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-body > table.table > tbody:nth-child(2)", scheduleElementNum))
		var lessons []model.Lesson
		for lessonNum := 1; lessonNum < 14; lessonNum++ {
			var lesson model.Lesson
			lessonElement := lessonsElement.Find(fmt.Sprintf("tr:nth-child(%d)", lessonNum))
			for lessonAttributeNum := 1; lessonAttributeNum <= 5; lessonAttributeNum++ {
				lessonElementAttribute := lessonElement.Find(fmt.Sprintf("td:nth-child(%d)", lessonAttributeNum))
				value := lessonElementAttribute.Text()
				if value == "" {
					break
				}

				value = strings.ReplaceAll(value, "\n", "")
				value = strings.TrimSpace(value)
				switch lessonAttributeNum {
				case 1:
					lesson.Num = value
				case 2:
					lesson.Time = value
				case 3:
					if strings.HasSuffix(value, "(1)") || strings.HasSuffix(value, "(2)") {
						switch value[len(value)-3:] {
						case "(1)":
							lesson.Subgroup = "1"
						case "(2)":
							lesson.Subgroup = "2"
						}
						lesson.Name = strings.TrimRight(strings.TrimRight(value, " (2)"), " (1)")
					} else {
						lesson.Name = value
					}
				case 4:
					lesson.Group = value
				case 5:
					room := strings.TrimSpace(regexp.MustCompile("\\W-[0-9]{1,3}$").FindString(value))
					if room == "" {
						lesson.Room = strings.TrimSpace(value)
					} else {
						lesson.Room = room
						lesson.Location = strings.TrimSpace(strings.TrimRight(value, room))
					}
				}
			}

			if lesson.Num != "" {
				lessons = append(lessons, lesson)
			}
		}

		weeklySchedule[len(weeklySchedule)-1].Lessons = lessons
	}

	if marshal, err := json.Marshal(weeklySchedule); err == nil {
		if err := s.r.Set(fmt.Sprintf("%d/%d", year, week)+":"+teacher, string(marshal)); err != nil {
			log.Error(err)
		} else {
			log.Trace("Данные сохранены в Redis")
		}
	}

	return weeklySchedule, nil
}
