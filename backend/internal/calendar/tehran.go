package calendar

import "time"

var tehran = loadTehran()

func loadTehran() *time.Location {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return time.FixedZone("Tehran", 3*3600+1800)
	}
	return loc
}

func Location() *time.Location {
	return tehran
}

func TehranToday() time.Time {
	now := time.Now().In(tehran)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tehran)
}

func DayRangeUTC(day time.Time) (time.Time, time.Time) {
	local := day.In(tehran)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, tehran)
	return start.UTC(), start.Add(24 * time.Hour).UTC()
}
