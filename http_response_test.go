package hmtpk_parser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	parserErrors "github.com/chazari-x/hmtpk_parser/v2/errors"
	"github.com/chazari-x/hmtpk_parser/v2/schedule"
	"github.com/chazari-x/hmtpk_parser/v2/schedule/group"
	"github.com/chazari-x/hmtpk_parser/v2/schedule/teacher"
	"github.com/chazari-x/hmtpk_parser/v2/utils"
	"github.com/sirupsen/logrus"
)

type responseTransport struct {
	status int
	body   string
}

func (r responseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: r.status, Status: fmt.Sprintf("%d %s", r.status, http.StatusText(r.status)), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(r.body)), Request: req}, nil
}

func TestScheduleHTTPResponses(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	for _, kind := range []string{"group", "teacher"} {
		t.Run(kind, func(t *testing.T) {
			var controller schedule.Adapter = group.NewController(nil, logrus.New())
			options := `<select id="group"><option value="123">Test group</option></select>`
			prefix := "<div></div>"
			table := `<table id="mobile-friendly"><thead><tr><th>Lesson</th></tr></thead><tbody><tr><td data-title="Номер урока">1</td><td data-title="Время">08:30</td><td data-title="Название предмета">Math</td><td data-title="Кабинет">101</td><td data-title="Преподаватель">Test</td></tr></tbody></table>`
			if kind == "teacher" {
				controller = teacher.NewController(nil, logrus.New())
				options = `<form id="zstfiltr"><div><div><select><option value="Test">Test teacher</option><option value="">Placeholder</option></select></div></div></form>`
				prefix = ""
				table = `<table class="table"><thead><tr><th>Lesson</th></tr></thead><tbody><tr><td>1</td><td>08:30</td><td>Math</td><td>Test group</td><td>101</td></tr></tbody></table>`
			}
			makeSchedule := func(days int, date, content string) string {
				body := `<div class="raspcontent m5">` + prefix
				for i := 0; i < days; i++ {
					body += `<div class="panel"><div class="panel-heading edu_today"><h2>` + date + `</h2></div><div class="panel-body">` + content + `</div></div>`
				}
				return body + `</div>`
			}
			valid := makeSchedule(7, "23 сентября 2026, среда", table)
			for _, tc := range []struct {
				name    string
				status  int
				body    string
				options bool
				wantErr bool
				lessons int
			}{
				{"options 200", 200, options, true, false, 0},
				{"options 500", 500, options, true, false, 0},
				{"options forbidden", 403, options, true, true, 0},
				{"options unavailable", 503, options, true, true, 0},
				{"error page options", 500, "<h1>Internal Server Error</h1>", true, true, 0},
				{"empty options 200", 200, "<html></html>", true, true, 0},
				{"schedule 200", 200, valid, false, false, 7},
				{"schedule 500", 500, valid, false, false, 7},
				{"empty week", 500, makeSchedule(7, "23 сентября 2026", "Нет занятий"), false, false, 0},
				{"error page schedule", 500, "<h1>Internal Server Error</h1>", false, true, 0},
				{"error page 200", 200, "<h1>Error</h1>", false, true, 0},
				{"truncated week", 500, makeSchedule(6, "23 сентября 2026", table), false, true, 0},
				{"invalid date", 500, makeSchedule(7, "oops", table), false, true, 0},
				{"schedule forbidden", 403, valid, false, true, 0},
				{"schedule unavailable", 503, valid, false, true, 0},
			} {
				t.Run(tc.name, func(t *testing.T) {
					http.DefaultTransport = responseTransport{tc.status, tc.body}
					var err error
					if tc.options {
						result, e := controller.GetOptions(context.Background())
						err = e
						if !tc.wantErr && len(result) != 1 {
							t.Fatalf("options=%v", result)
						}
						if tc.wantErr && len(result) != 0 {
							t.Fatal("invalid response returned options")
						}
					} else {
						result, e := controller.GetSchedule(context.Background(), "123", "23.09.2026")
						err = e
						if !tc.wantErr {
							if len(result) != 7 {
								t.Fatalf("days=%d", len(result))
							}
							lessons := 0
							for _, day := range result {
								lessons += len(day.Lessons)
							}
							if lessons != tc.lessons {
								t.Fatalf("lessons=%d, want %d", lessons, tc.lessons)
							}
						} else if len(result) != 0 {
							t.Fatal("invalid response returned partial schedule")
						}
					}
					if tc.wantErr {
						if !errors.Is(err, parserErrors.ErrorBadResponse) {
							t.Fatalf("error=%v", err)
						}
					} else if err != nil {
						t.Fatal(err)
					}
				})
			}
		})
	}
}

func TestGetDateMalformedInput(t *testing.T) {
	for _, input := range []string{"", "oops", "23 x 2026", "31 февраля 2026", "23 unknown 2026", "23 сентября", "23 сентября 2026 extra"} {
		if got := utils.GetDate(input); got != "" {
			t.Errorf("GetDate(%q)=%q", input, got)
		}
	}
	for input, want := range map[string]string{"2 сентября 2026": "2.09.2026", " 23  сентября 2026 ": "23.09.2026", "02 мая 2026": "02.05.2026"} {
		if got := utils.GetDate(input); got != want {
			t.Errorf("GetDate(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestAnnouncesHTTPResponses(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	valid := `<section class="sf-pagewrap-area overflow-hidden d-flex flex-col justify-content-start"><div><section><main><section><div><div class="row"><div class="iblock-list-item-text p-3"><p class="c-text-secondary">23.09.2026</p><h3><a href="/news/test">Test</a></h3><div class="c-text-secondary">Body</div></div></div></div></section><div class="sf-viewbox position-relative"><div><span>3</span></div></div></main></section></div></section>`
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{"valid 200", 200, valid, false},
		{"valid 500", 500, valid, false},
		{"forbidden", 403, valid, true},
		{"unavailable", 503, valid, true},
		{"error page", 500, "<h1>Internal Server Error</h1>", true},
		{"invalid entry", 500, strings.ReplaceAll(valid, `href="/news/test"`, ""), true},
		{"invalid pagination", 500, strings.ReplaceAll(valid, "<span>3</span>", "<span>oops</span>"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			http.DefaultTransport = responseTransport{tc.status, tc.body}
			result, err := NewController(nil, logrus.New()).GetAnnounces(context.Background(), 1)
			if tc.wantErr {
				if !errors.Is(err, parserErrors.ErrorBadResponse) {
					t.Fatalf("error=%v", err)
				}
				if len(result.Announces) != 0 {
					t.Fatal("invalid response returned announcements")
				}
			} else if err != nil || len(result.Announces) != 1 || result.LastPage != 3 {
				t.Fatalf("result=%v error=%v", result, err)
			}
		})
	}
}
