package gajob

import (
	"testing"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
)

func TestParseGARows(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Seoul")
	payload := app.H{
		"rows": []any{
			app.H{
				"dimensionValues": []any{
					app.H{"value": "20260521"},
				},
				"metricValues": []any{
					app.H{"value": "10"},
					app.H{"value": "20"},
					app.H{"value": "30"},
					app.H{"value": "5"},
				},
			},
		},
	}

	t.Run("Daily", func(t *testing.T) {
		layout := "20060102"
		outLayout := "2006-01-02"
		rows := parseGARows(payload, layout, outLayout, loc, false)
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		expected := "2026-05-21"
		if rows[0].Timeslot != expected {
			t.Errorf("expected timeslot %s, got %s", expected, rows[0].Timeslot)
		}
	})

	t.Run("Hourly", func(t *testing.T) {
		payloadHourly := app.H{
			"rows": []any{
				app.H{
					"dimensionValues": []any{
						app.H{"value": "20260521"},
						app.H{"value": "00"},
					},
					"metricValues": []any{
						app.H{"value": "10"},
						app.H{"value": "20"},
						app.H{"value": "30"},
						app.H{"value": "5"},
					},
				},
			},
		}
		layout := "20060102 15"
		outLayout := "2006-01-02 15:04:05"
		rows := parseGARows(payloadHourly, layout, outLayout, loc, true)
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		// 2026-05-21 00:00:00 KST is 2026-05-20 15:00:00 UTC
		expected := "2026-05-20 15:00:00"
		if rows[0].Timeslot != expected {
			t.Errorf("expected timeslot %s, got %s", expected, rows[0].Timeslot)
		}
	})
}
