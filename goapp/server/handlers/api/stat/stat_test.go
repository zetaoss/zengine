package stat

import (
	"testing"
	"time"

	"github.com/zetaoss/zengine/goapp/models/stat"
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
	rows := []statmodels.StatCF{
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

func TestBuildK8sHourlyPayloadIncludesPVCStorage(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	rows := []statmodels.K8sHourly{
		{Timeslot: from.Format("2006-01-02 15:04:05"), PVCStorageUsage: 11 * 1024 * 1024 * 1024, PVCStorageCapacity: 100 * 1024 * 1024 * 1024, DefenderFightingRatio: 0.25, DefenderMaxLevel: 7},
		{Timeslot: to.Format("2006-01-02 15:04:05"), PVCStorageUsage: 1},
	}

	payload := buildK8sHourlyPayload(from, to, rows)
	storage, ok := payload["pvc_storage_usage"].([]any)
	if !ok {
		t.Fatalf("pvc_storage_usage type mismatch: %T", payload["pvc_storage_usage"])
	}
	if storage[0] != float64(11*1024*1024*1024) {
		t.Fatalf("pvc_storage_usage[0]=%v", storage[0])
	}
	if storage[1] != nil {
		t.Fatalf("pvc_storage_usage[1]=%v, expected nil", storage[1])
	}
	capacity, ok := payload["pvc_storage_capacity"].([]any)
	if !ok || capacity[0] != float64(100*1024*1024*1024) {
		t.Fatalf("pvc_storage_capacity=%v", payload["pvc_storage_capacity"])
	}
	if capacity[1] != nil {
		t.Fatalf("pvc_storage_capacity[1]=%v, expected nil", capacity[1])
	}
	fightingRatio, ok := payload["defender_fighting_ratio"].([]any)
	if !ok || fightingRatio[0] != 0.25 {
		t.Fatalf("defender_fighting_ratio=%v", payload["defender_fighting_ratio"])
	}
	maxLevel, ok := payload["defender_max_level"].([]any)
	if !ok || maxLevel[0] != float64(7) {
		t.Fatalf("defender_max_level=%v", payload["defender_max_level"])
	}
}

func TestBuildK8sDailyPayloadIncludesPVCStorage(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 1)
	rows := []statmodels.K8sDaily{
		{Timeslot: from.Format("2006-01-02"), PVCStorageUsage: 12 * 1024 * 1024 * 1024, PVCStorageCapacity: 100 * 1024 * 1024 * 1024, DefenderFightingRatio: 0.5, DefenderMaxLevel: 9},
		{Timeslot: to.Format("2006-01-02"), PVCStorageUsage: 1},
	}

	payload := buildK8sDailyPayload(from, to, rows)
	storage, ok := payload["pvc_storage_usage"].([]any)
	if !ok {
		t.Fatalf("pvc_storage_usage type mismatch: %T", payload["pvc_storage_usage"])
	}
	if storage[0] != float64(12*1024*1024*1024) {
		t.Fatalf("pvc_storage_usage[0]=%v", storage[0])
	}
	if storage[1] != nil {
		t.Fatalf("pvc_storage_usage[1]=%v, expected nil", storage[1])
	}
	capacity, ok := payload["pvc_storage_capacity"].([]any)
	if !ok || capacity[0] != float64(100*1024*1024*1024) {
		t.Fatalf("pvc_storage_capacity=%v", payload["pvc_storage_capacity"])
	}
	if capacity[1] != nil {
		t.Fatalf("pvc_storage_capacity[1]=%v, expected nil", capacity[1])
	}
	fightingRatio, ok := payload["defender_fighting_ratio"].([]any)
	if !ok || fightingRatio[0] != 0.5 {
		t.Fatalf("defender_fighting_ratio=%v", payload["defender_fighting_ratio"])
	}
	maxLevel, ok := payload["defender_max_level"].([]any)
	if !ok || maxLevel[0] != float64(9) {
		t.Fatalf("defender_max_level=%v", payload["defender_max_level"])
	}
}
