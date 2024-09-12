package hmtpk_parser

import (
	"context"
	"errors"

	"github.com/chazari-x/hmtpk_parser/announce"
	"github.com/chazari-x/hmtpk_parser/schedule"
	"github.com/chazari-x/hmtpk_parser/schedule/group"
	"github.com/chazari-x/hmtpk_parser/schedule/teacher"
	"github.com/chazari-x/hmtpk_parser/storage"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/chazari-x/hmtpk_parser/model"
)

type Controller struct {
	r        *storage.Redis
	log      *logrus.Logger
	group    *group.Controller
	teacher  *teacher.Controller
	announce *announce.Announce
}

func NewController(client *redis.Client, logger *logrus.Logger) *Controller {
	return &Controller{
		r:        &storage.Redis{Redis: client},
		log:      logger,
		group:    group.NewController(client, logger),
		teacher:  teacher.NewController(client, logger),
		announce: announce.NewAnnounce(logger),
	}
}

// GetScheduleByGroup по идентификатору группы и дате получает расписание на неделю
func (c *Controller) GetScheduleByGroup(group, date string, ctx context.Context) ([]model.Schedule, error) {
	return c.getSchedule(group, date, ctx, c.group)
}

// GetScheduleByTeacher по ФИО преподавателя и дате получает расписание преподавателя
func (c *Controller) GetScheduleByTeacher(teacher, date string, ctx context.Context) ([]model.Schedule, error) {
	return c.getSchedule(teacher, date, ctx, c.teacher)
}

// GetGroupOptions получает список групп
func (c *Controller) GetGroupOptions(ctx context.Context) ([]model.Option, error) {
	return c.group.GetOptions(ctx)
}

// GetTeacherOptions получает список преподавателей
func (c *Controller) GetTeacherOptions(ctx context.Context) ([]model.Option, error) {
	return c.teacher.GetOptions(ctx)
}

var BadRequest = errors.New("bad request")

func (c *Controller) getSchedule(name, date string, ctx context.Context, adapter schedule.Adapter) ([]model.Schedule, error) {
	if name == "0" || name == "" {
		return nil, BadRequest
	}

	return adapter.GetSchedule(name, date, ctx)
}

// GetAnnounces получает html блок с объявлениями
func (c *Controller) GetAnnounces(ctx context.Context, page int) ([]model.Announce, error) {
	if page < 1 {
		return nil, BadRequest
	}

	return c.announce.GetAnnounces(ctx, page)
}
