package hmtpk_parser

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/chazari-x/hmtpk_parser/v2/announce"
	"github.com/chazari-x/hmtpk_parser/v2/model"
	"github.com/chazari-x/hmtpk_parser/v2/schedule/group"
	"github.com/chazari-x/hmtpk_parser/v2/schedule/teacher"
	"github.com/chazari-x/hmtpk_parser/v2/storage"
	"github.com/chazari-x/hmtpk_parser/v2/utils"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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

func TestController_GetGroupValues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	type fields struct {
		r       *storage.Redis
		log     *logrus.Logger
		group   *group.Controller
		teacher *teacher.Controller
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				r:       nil,
				log:     logrus.StandardLogger(),
				group:   group.NewController(nil, logrus.StandardLogger()),
				teacher: teacher.NewController(nil, logrus.StandardLogger()),
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				r:       tt.fields.r,
				log:     tt.fields.log,
				group:   tt.fields.group,
				teacher: tt.fields.teacher,
			}
			got, err := c.GetGroupOptions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroupOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("GetGroupOptions() got = %v, want not empty", got)
			} else {
				t.Log(got)
			}
		})
	}
}

func TestController_GetTeacherValues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	type fields struct {
		r       *storage.Redis
		log     *logrus.Logger
		group   *group.Controller
		teacher *teacher.Controller
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				r:       nil,
				log:     logrus.StandardLogger(),
				group:   group.NewController(nil, logrus.StandardLogger()),
				teacher: teacher.NewController(nil, logrus.StandardLogger()),
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				r:       tt.fields.r,
				log:     tt.fields.log,
				group:   tt.fields.group,
				teacher: tt.fields.teacher,
			}
			got, err := c.GetTeacherOptions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTeacherOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("GetTeacherOptions() got = %v, want not empty", got)
			} else {
				t.Log(got)
			}
		})
	}
}

func TestController_GetAnnounces(t *testing.T) {
	log := logrus.StandardLogger()
	a := announce.NewAnnounce(log)

	tests := []struct {
		name    string
		page    int
		wantErr bool
	}{
		{
			name:    "1",
			page:    1,
			wantErr: false,
		},
		{
			name:    "69",
			page:    69,
			wantErr: false,
		},
		{
			name:    "70",
			page:    70,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				log:      log,
				announce: a,
			}
			got, err := c.GetAnnounces(ctx, tt.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAnnounces() error = %v, wantErr %v", err, tt.wantErr)
				cancel()
				return
			}
			if len(got.Announces) == 0 {
				t.Errorf("GetAnnounces() got = %v, want not empty", got)
			} else {
				t.Log(got)
			}
		})

		cancel()
	}
}
