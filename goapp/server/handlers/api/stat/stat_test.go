package stat

import (
	"github.com/zetaoss/zengine/goapp/models"
	"testing"
	"time"
)

func TestHourlyEndUTC(t *testing.T) {
	t.Parallel()

	a := hourlyEndUTC(time.Date(2026, 3, 16, 16, 7, 0, 0, time.UTC), 10)
	if got := a.Format(time.RFC3339); got != "2026-03-16T15:00:00Z" {
		t.Fatalf("hourlyEndUTC before cutoff = %s", got)
	}

	b := hourlyEndUTC(time.Date(2026, 3, 16, 16, 10, 0, 0, time.UTC), 10)
	if got := b.Format(time.RFC3339); got != "2026-03-16T16:00:00Z" {
		t.Fatalf("hourlyEndUTC at cutoff = %s", got)
	}
}

func TestBuildHourlyPayload(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	to := from.Add(47 * time.Hour)
	rows := []models.StatCF{
		{Timeslot: from.Add(2 * time.Hour), Name: "sum_requests", Value: "123"},
		{Timeslot: from.Add(2 * time.Hour), Name: "sum_browserMap", Value: "{\"Chrome\":10}"},
	}

	payload := buildCFHourlyPayload(from, to, rows)
	timeslots, ok := payload["timeslots"].([]string)
	if !ok {
		t.Fatalf("timeslots type mismatch: %T", payload["timeslots"])
	}
	if len(timeslots) != 48 {
		t.Fatalf("timeslots len=%d", len(timeslots))
	}

	requests, ok := payload["sum_requests"].([]any)
	if !ok {
		t.Fatalf("sum_requests type mismatch: %T", payload["sum_requests"])
	}
	if requests[2] != float64(123) {
		t.Fatalf("sum_requests[2]=%v", requests[2])
	}
}
