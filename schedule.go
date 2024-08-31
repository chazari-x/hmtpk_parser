package hmtpk_schedule

import (
	"context"
	"errors"

	"github.com/chazari-x/hmtpk_schedule/schedule"
	"github.com/chazari-x/hmtpk_schedule/schedule/group"
	"github.com/chazari-x/hmtpk_schedule/schedule/teacher"
	"github.com/chazari-x/hmtpk_schedule/storage"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/chazari-x/hmtpk_schedule/model"
)

type Controller struct {
	r       *storage.Redis
	log     *logrus.Logger
	group   *group.Controller
	teacher *teacher.Controller
}

func NewController(client *redis.Client, logger *logrus.Logger) *Controller {
	return &Controller{
		r:       &storage.Redis{Redis: client},
		log:     logger,
		group:   group.NewController(client, logger),
		teacher: teacher.NewController(client, logger),
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

// GetGroupValues получает список групп
func (c *Controller) GetGroupValues(ctx context.Context) ([]model.Option, error) {
	return c.group.GetOptions(ctx)
}

// GetTeacherValues получает список преподавателей
func (c *Controller) GetTeacherValues(ctx context.Context) ([]model.Option, error) {
	return c.teacher.GetOptions(ctx)
}

var BadRequest = errors.New("bad request")

func (c *Controller) getSchedule(name, date string, ctx context.Context, adapter schedule.Adapter) ([]model.Schedule, error) {
	if name == "0" || name == "" {
		return nil, BadRequest
	}

	return adapter.GetSchedule(name, date, ctx)
}
