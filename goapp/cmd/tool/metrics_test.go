package main

import (
	"testing"
)

func TestRunMetricsRejectsEndpointFlag(t *testing.T) {
	if err := runMetrics(nil, []string{"--endpoint", "http://metrics.example.test/metrics"}); err == nil {
		t.Fatal("expected --endpoint to be rejected")
	}
}

func TestRunMetricsReturnsErrorOnMissingEnv(t *testing.T) {
	t.Setenv("BOB_ENDPOINT", "")

	if err := runMetrics(nil, nil); err == nil {
		t.Fatal("expected error when BOB_ENDPOINT is missing")
	}
}

func TestFormatPct(t *testing.T) {
	if got := formatPct(1, 4); got != "25.00%" {
		t.Fatalf("formatPct = %s", got)
	}
	if got := formatPct(1, 0); got != "0.00%" {
		t.Fatalf("formatPct with zero allocatable = %s", got)
	}
}
