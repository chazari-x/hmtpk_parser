package hmtpk_schedule

import "testing"

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
			if got := getDate(tt.date); got != tt.want {
				t.Errorf("getDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
