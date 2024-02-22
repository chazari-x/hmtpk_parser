package schedule

import (
	"context"

	"github.com/chazari-x/hmtpk_schedule/model"
)

type Adapter interface {
	GetSchedule(name, date string, ctx context.Context) ([]model.Schedule, error)
}
