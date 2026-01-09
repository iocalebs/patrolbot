package config

import "time"

// Duration represents a span of time expressed in hours, days, or weeks.
type Duration struct {
	Hours int `json:"hours,omitempty" jsonschema:"oneof,minimum=1"`
	Days  int `json:"days,omitempty"  jsonschema:"oneof,minimum=1"`
	Weeks int `json:"weeks,omitempty" jsonschema:"oneof,minimum=1"`
}

// AddTo returns the time obtained by adding the duration to `start`.
//
// The calculation uses calendar arithmetic for weeks and days, and a fixed
// duration for hours. If multiple fields are set, the largest unit takes
// precedence: weeks, then days, then hours.
func (d Duration) AddTo(start time.Time) time.Time {
	if d.Weeks != 0 {
		return start.AddDate(0, 0, 7*d.Weeks) //nolint:mnd
	}

	if d.Days != 0 {
		return start.AddDate(0, 0, d.Days)
	}

	return start.Add(time.Duration(d.Hours) * time.Hour)
}
