package hmtpk_schedule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chazari-x/hmtpk_schedule/storage"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	r   *storage.Redis
	log *logrus.Logger
}

//type Config struct {
//	Groups []struct {
//		ID   int    `yaml:"id"`
//		Name string `yaml:"name"`
//	} `yaml:"groups"`
//
//	Teachers []struct {
//		Name string `yaml:"name"`
//	} `yaml:"teachers"`
//}

func NewController(client *redis.Client, logger *logrus.Logger) *Controller {
	s := &Controller{ /*cfg, */ r: &storage.Redis{Redis: client}, log: logger}
	//Groups = make(map[string]int)
	//for _, group := range cfg.Groups {
	//	Groups[group.Name] = group.ID
	//}

	return s
}

//
//var Groups map[string]int
//
//func (s *Controller) GetGroups() []string {
//	var groups []string
//
//	for _, g := range s.cfg.Groups {
//		groups = append(groups, g.Name)
//	}
//
//	return groups
//}
//
//func (s *Controller) GetGroup(groupName string) string {
//	return strconv.Itoa(Groups[groupName])
//}
//
//func (s *Controller) GetTeachers() []string {
//	var teachers []string
//
//	for _, g := range s.cfg.Teachers {
//		teachers = append(teachers, g.Name)
//	}
//
//	return teachers
//}

type Schedule struct {
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lesson"`
	Href    string   `json:"href"`
}

type Lesson struct {
	Num      string `json:"num"`
	Time     string `json:"time"`
	Name     string `json:"name"`
	Room     string `json:"room"`
	Location string `json:"location"`
	Group    string `json:"group"`
	Subgroup string `json:"subgroup"`
	Teacher  string `json:"teacher"`
}

//func (s *Controller) GetScheduleByGroupName(group, date string) ([]Schedule, error) {
//	return s.GetScheduleByGroup(s.GetGroup(group), date)
//}

const (
	numOfColumns = 5
)

// GetScheduleByGroup по идентификатору группы и дате получает расписание на неделю
func (s *Controller) GetScheduleByGroup(group, date string, ctx context.Context) ([]Schedule, error) {
	if group == "0" || group == "" {
		return nil, fmt.Errorf(http.StatusText(http.StatusBadRequest))
	}

	resCh := make(chan []Schedule, 1)
	errCh := make(chan error, 1)
	defer close(resCh)
	defer close(errCh)

	go func() {
		schedule, err := s.getScheduleByGroup(group, date)
		if err != nil {
			select {
			case _, ok := <-errCh:
				if !ok {
					return
				}
			default:
				errCh <- err
			}

			return
		}

		select {
		case _, ok := <-resCh:
			if !ok {
				return
			}
		default:
			resCh <- schedule
		}
	}()

	select {
	case <-ctx.Done():
		return nil, context.Canceled
	case res := <-resCh:
		return res, nil
	case err := <-errCh:
		return nil, err
	}
}

func (s *Controller) getScheduleByGroup(group, date string) ([]Schedule, error) {
	var weeklySchedule []Schedule

	s.log.Trace(group)

	d, err := time.Parse("02.01.2006", date)
	if err != nil {
		return nil, err
	}

	year, week := d.ISOWeek()
	if s.r != nil {
		if redisWeeklySchedule, err := s.r.Get(fmt.Sprintf("%d/%d", year, week) + ":" + group); err == nil && redisWeeklySchedule != "" {
			if json.Unmarshal([]byte(redisWeeklySchedule), &weeklySchedule) == nil {
				s.log.Trace("Данные получены из redis")
				return weeklySchedule, nil
			}
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

	const (
		firstDayNum = 2
		lastDayNum  = 8
	)

	for scheduleElementNum := firstDayNum; scheduleElementNum <= lastDayNum; scheduleElementNum++ {
		scheduleDateElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-heading.edu_today > h2", scheduleElementNum))

		date = getDate(strings.Split(scheduleDateElement.Text(), ",")[0])

		weeklySchedule = append(weeklySchedule, Schedule{
			Date: scheduleDateElement.Text(),
			Href: fmt.Sprintf("https://hmtpk.ru/ru/students/schedule/?group=%s&date_edu1c=%s&send=Показать#current", group, date),
		})

		lessonsElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-body > #mobile-friendly > tbody:nth-child(2)", scheduleElementNum))
		var lessons []Lesson
		for lessonNum := 1; lessonNum > 0; lessonNum++ {
			var lesson Lesson
			var exists bool
			lessonElement := lessonsElement.Find(fmt.Sprintf("tr:nth-child(%d)", lessonNum))
			for lessonAttributeNum := 1; lessonAttributeNum <= numOfColumns; lessonAttributeNum++ {
				lessonElementAttribute := lessonElement.Find(fmt.Sprintf("td:nth-child(%d)", lessonAttributeNum))
				var value string
				value, exists = lessonElementAttribute.Attr("data-title")
				if !exists {
					if lessonAttributeNum == numOfColumns {
						exists = true
					} else if lessonAttributeNum == 1 {
						break
					}
				}

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
			}

			if exists {
				lessons = append(lessons, lesson)
			} else {
				break
			}
		}

		weeklySchedule[len(weeklySchedule)-1].Lessons = lessons
	}

	if s.r != nil {
		if marshal, err := json.Marshal(weeklySchedule); err == nil {
			if err = s.r.Set(fmt.Sprintf("%d/%d", year, week)+":"+group, string(marshal)); err != nil {
				s.log.Error(err)
			} else {
				s.log.Trace("Данные сохранены в redis")
			}
		}
	}

	return weeklySchedule, nil
}

// GetScheduleByTeacher по ФИО преподавателя и дате получает расписание преподавателя
func (s *Controller) GetScheduleByTeacher(teacher, date string, ctx context.Context) ([]Schedule, error) {
	if teacher == "0" || teacher == "" {
		return nil, fmt.Errorf(http.StatusText(http.StatusBadRequest))
	}

	resCh := make(chan []Schedule, 1)
	errCh := make(chan error, 1)
	defer close(resCh)
	defer close(errCh)

	go func() {
		schedule, err := s.getScheduleByTeacher(teacher, date)
		if err != nil {
			select {
			case _, ok := <-errCh:
				if !ok {
					return
				}
			default:
				errCh <- err
			}

			return
		}

		select {
		case _, ok := <-resCh:
			if !ok {
				return
			}
		default:
			resCh <- schedule
		}
	}()

	select {
	case <-ctx.Done():
		return nil, context.Canceled
	case res := <-resCh:
		return res, nil
	case err := <-errCh:
		return nil, err
	}
}

func (s *Controller) getScheduleByTeacher(teacher, date string) ([]Schedule, error) {
	var weeklySchedule []Schedule

	s.log.Trace(teacher)

	teacher = strings.ReplaceAll(teacher, " ", "+")
	d, err := time.Parse("02.01.2006", date)
	if err != nil {
		return nil, err
	}

	year, week := d.ISOWeek()
	if s.r != nil {
		if redisWeeklySchedule, err := s.r.Get(fmt.Sprintf("%d/%d", year, week) + ":" + teacher); err == nil && redisWeeklySchedule != "" {
			if json.Unmarshal([]byte(redisWeeklySchedule), &weeklySchedule) == nil {
				s.log.Trace("Данные получены из redis")
				return weeklySchedule, nil
			}
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

	const (
		firstDayNum = 1
		lastDayNum  = 7
	)

	for scheduleElementNum := firstDayNum; scheduleElementNum <= lastDayNum; scheduleElementNum++ {
		scheduleDateElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-heading.edu_today > h2", scheduleElementNum))

		date = getDate(strings.Split(scheduleDateElement.Text(), ",")[0])

		weeklySchedule = append(weeklySchedule, Schedule{
			Date: scheduleDateElement.Text(),
			Href: fmt.Sprintf("https://hmtpk.ru/ru/teachers/schedule/?teacher=%s&date_edu1c=%s&send=Показать#current", teacher, date),
		})

		lessonsElement := doc.Children().Find(fmt.Sprintf("div.raspcontent.m5 div:nth-child(%d) div.panel-body > table.table > tbody:nth-child(2)", scheduleElementNum))
		var lessons []Lesson
		for lessonNum := 1; lessonNum > 0; lessonNum++ {
			var lesson Lesson
			var exists bool
			lessonElement := lessonsElement.Find(fmt.Sprintf("tr:nth-child(%d)", lessonNum))
			for lessonAttributeNum := 1; lessonAttributeNum <= numOfColumns; lessonAttributeNum++ {
				lessonElementAttribute := lessonElement.Find(fmt.Sprintf("td:nth-child(%d)", lessonAttributeNum))
				value := lessonElementAttribute.Text()
				if value == "" {
					if lessonAttributeNum == 1 {
						exists = false
					}

					break
				} else {
					exists = true
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

			if exists {
				lessons = append(lessons, lesson)
			} else {
				break
			}
		}

		weeklySchedule[len(weeklySchedule)-1].Lessons = lessons
	}

	if s.r != nil {
		if marshal, err := json.Marshal(weeklySchedule); err == nil {
			if err := s.r.Set(fmt.Sprintf("%d/%d", year, week)+":"+teacher, string(marshal)); err != nil {
				s.log.Error(err)
			} else {
				s.log.Trace("Данные сохранены в redis")
			}
		}
	}

	return weeklySchedule, nil
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
