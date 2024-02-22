package hmtpk_schedule

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/chazari-x/hmtpk_schedule/model"
	"github.com/chazari-x/hmtpk_schedule/storage"
	"github.com/chazari-x/hmtpk_schedule/utils"
	"github.com/sirupsen/logrus"
)

func Test_getDate(t *testing.T) {
	tests := []struct {
		name string
		date string
		want string
	}{
		{
			name: "1",
			date: "02 февраля 2025",
			want: "02.02.2025",
		},
		{
			name: "2",
			date: "02 марта 2025",
			want: "02.03.2025",
		},
		{
			name: "3",
			date: "02 декабря 2025",
			want: "02.12.2025",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.GetDate(tt.date); got != tt.want {
				t.Errorf("getDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_GetScheduleByGroup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	type args struct {
		group string
		date  string
		ctx   context.Context
	}
	tests := []struct {
		name    string
		r       *storage.Redis
		log     *logrus.Logger
		args    args
		noWant  []model.Schedule
		wantErr bool
	}{
		{
			name: "",
			r:    nil,
			log:  logrus.StandardLogger(),
			args: args{
				group: "114808",
				date:  "22.02.2024",
				ctx:   ctx,
			},
			noWant:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewController(nil, tt.log)
			got, err := c.GetScheduleByGroup(tt.args.group, tt.args.date, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScheduleByGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.noWant) {
				t.Errorf("GetScheduleByGroup() got = %v, noWant %v", got, tt.noWant)
			} else {
				t.Log(got[3].Lessons)
			}
		})
	}
}

func TestController_GetScheduleByTeacher(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	type args struct {
		teacher string
		date    string
		ctx     context.Context
	}
	tests := []struct {
		name    string
		r       *storage.Redis
		log     *logrus.Logger
		args    args
		noWant  []model.Schedule
		wantErr bool
	}{
		{
			name: "",
			r:    nil,
			log:  logrus.StandardLogger(),
			args: args{
				teacher: "<>",
				date:    "21.02.2024",
				ctx:     ctx,
			},
			noWant:  nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewController(nil, tt.log)
			got, err := c.GetScheduleByTeacher(tt.args.teacher, tt.args.date, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScheduleByTeacher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.noWant) {
				t.Errorf("GetScheduleByTeacher() got = %v, noWant %v", got, tt.noWant)
			} else {
				t.Log(len(got[2].Lessons))
				t.Log(got[2].Lessons)
			}
		})
	}
}
