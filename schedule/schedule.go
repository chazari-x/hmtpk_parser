package schedule

import (
	"context"

	"github.com/chazari-x/hmtpk_parser/v2/model"
)

type Adapter interface {
	GetSchedule(ctx context.Context, value, date string) ([]model.Schedule, error)
	GetOptions(ctx context.Context) ([]model.Option, error)
}
