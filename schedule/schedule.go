package schedule

import (
	"context"

	"github.com/chazari-x/hmtpk_parser/v2/model"
)

type Adapter interface {
	GetSchedule(valueLabel, date string, ctx context.Context) ([]model.Schedule, error)
	GetOptions(ctx context.Context) ([]model.Option, error)
}
