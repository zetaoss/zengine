package main

import (
	"testing"
	"time"
)

func TestInterpolateDailySamples(t *testing.T) {
	day := func(raw string) time.Time {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	samples := []dailySample{
		{timeslot: day("2026-01-01"), values: map[string]float64{"count": 10, "rate": 1.5}},
		{timeslot: day("2026-01-04"), values: map[string]float64{"count": 19, "rate": 4.5}},
	}
	rows := interpolateDailySamples(samples, []dailyMetric{{name: "count", integer: true}, {name: "rate"}})
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	if rows[0]["timeslot"] != "2026-01-02" || rows[0]["count"] != float64(13) || rows[0]["rate"] != 2.5 {
		t.Fatalf("first row = %#v", rows[0])
	}
	if rows[1]["timeslot"] != "2026-01-03" || rows[1]["count"] != float64(16) || rows[1]["rate"] != 3.5 {
		t.Fatalf("second row = %#v", rows[1])
	}
}

func TestInterpolateDailySamplesLeavesEdgesUntouched(t *testing.T) {
	day, _ := time.Parse("2006-01-02", "2026-01-02")
	rows := interpolateDailySamples([]dailySample{{timeslot: day, values: map[string]float64{"count": 10}}}, []dailyMetric{{name: "count", integer: true}})
	if len(rows) != 0 {
		t.Fatalf("rows = %#v, want none", rows)
	}
}

func TestParseDailyTimeslotAcceptsDatetime(t *testing.T) {
	got, ok := parseDailyTimeslot("2026-05-16 00:00:00")
	if !ok || got.Format("2006-01-02") != "2026-05-16" {
		t.Fatalf("parseDailyTimeslot() = %v, %t", got, ok)
	}
}

func TestParseNumberAcceptsMySQLUnsignedIntegers(t *testing.T) {
	got, ok := parseNumber(uint64(265292))
	if !ok || got != 265292 {
		t.Fatalf("parseNumber() = %v, %t", got, ok)
	}
}
