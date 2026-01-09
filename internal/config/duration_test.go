package config_test // replace with the actual package name

import (
	"testing"
	"time"

	"github.com/iocalebs/patrolbot/internal/config"
)

func TestDuration_AddTo(t *testing.T) {
	t.Parallel()

	base := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		duration config.Duration
		want     time.Time
	}{
		{
			name: "Hours",
			duration: config.Duration{
				Hours: 5,
			},
			want: base.Add(5 * time.Hour),
		},
		{
			name: "Days",
			duration: config.Duration{
				Days: 3,
			},
			want: base.AddDate(0, 0, 3),
		},
		{
			name: "Weeks",
			duration: config.Duration{
				Weeks: 2,
			},
			want: base.AddDate(0, 0, 14),
		},
		{
			name: "Weeks take precedence over days",
			duration: config.Duration{
				Weeks: 1,
				Days:  5,
			},
			want: base.AddDate(0, 0, 7),
		},
		{
			name: "Days take precedence over hours",
			duration: config.Duration{
				Days:  2,
				Hours: 12,
			},
			want: base.AddDate(0, 0, 2),
		},
		{
			name:     "zero duration returns start time",
			duration: config.Duration{},
			want:     base,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.duration.AddTo(base)
			if !got.Equal(test.want) {
				t.Fatalf("AddTo() = %v, want %v", got, test.want)
			}
		})
	}
}
