package handlers

import (
	"fmt"
	"strings"
	"time"
)

type timeValue struct{ time.Time }

func (t *timeValue) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("فرمت تاریخ نامعتبر است: %s", s)
}

type nullableTimeValue struct{ *time.Time }

func (t *nullableTimeValue) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, s); err == nil {
			parsedUTC := parsed
			t.Time = &parsedUTC
			return nil
		}
	}
	return fmt.Errorf("فرمت تاریخ نامعتبر است: %s", s)
}
