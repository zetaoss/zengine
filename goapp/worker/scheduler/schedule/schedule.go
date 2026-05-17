package schedule

import (
	"fmt"
	"time"
)

type Schedule interface {
	Key(time.Time) (string, bool)
	Name() string
}

func DailyAt(hour int, minute int) Schedule { return dailyAt{hour: hour, minute: minute} }

type dailyAt struct {
	hour   int
	minute int
}

func (d dailyAt) Key(t time.Time) (string, bool) {
	if t.Hour() != d.hour || t.Minute() != d.minute {
		return "", false
	}
	return t.Format("2006-01-02"), true
}

func (d dailyAt) Name() string { return "daily:" + twoDigits(d.hour) + ":" + twoDigits(d.minute) }

func HourlyAt(minute int) Schedule { return hourlyAt{minute: minute} }

type hourlyAt struct{ minute int }

func (h hourlyAt) Key(t time.Time) (string, bool) {
	if t.Minute() != h.minute {
		return "", false
	}
	return t.Format("2006-01-02T15"), true
}

func (h hourlyAt) Name() string { return "hourly:" + twoDigits(h.minute) }

func EveryNMinute(minutes int) Schedule {
	if minutes <= 0 {
		minutes = 1
	}
	return everyNMinute{minutes: minutes}
}

type everyNMinute struct{ minutes int }

func (e everyNMinute) Key(t time.Time) (string, bool) {
	if t.Minute()%e.minutes != 0 {
		return "", false
	}
	return t.Format("2006-01-02T15:04"), true
}

func (e everyNMinute) Name() string { return fmt.Sprintf("every-%dm", e.minutes) }

func EveryNSeconds(seconds int) Schedule {
	if seconds <= 0 {
		seconds = 1
	}
	return everyNSeconds{seconds: seconds}
}

type everyNSeconds struct{ seconds int }

func (e everyNSeconds) Key(t time.Time) (string, bool) {
	if t.Second()%e.seconds != 0 {
		return "", false
	}
	return t.Format("2006-01-02T15:04:05"), true
}

func (e everyNSeconds) Name() string { return fmt.Sprintf("every-%ds", e.seconds) }

func twoDigits(v int) string {
	return fmt.Sprintf("%02d", v)
}
