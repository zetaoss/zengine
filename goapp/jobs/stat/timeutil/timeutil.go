package timeutil

import (
	"time"
)

func DailyEndUTC(now time.Time) time.Time {
	current := now.UTC()
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
}

func DailyEndInLocation(now time.Time, loc *time.Location) time.Time {
	current := now.In(loc)
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
}

func HourlyEndUTC(now time.Time, readyMinute int) time.Time {
	current := now.UTC()
	end := time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), 0, 0, 0, time.UTC)
	if current.Minute() < readyMinute {
		end = end.Add(-time.Hour)
	}
	return end
}

func HourlyEndInLocation(now time.Time, loc *time.Location) time.Time {
	current := now.In(loc)
	return time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), 0, 0, 0, loc)
}
