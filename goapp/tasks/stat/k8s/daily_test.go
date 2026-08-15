package k8s

import (
	"testing"
)

func TestMedianFloat64(t *testing.T) {
	odd := []float64{10.0, 1.0, 5.0}
	if got := medianFloat64(odd); got != 5.0 {
		t.Errorf("medianFloat64(odd) = %f, expected 5.0", got)
	}

	even := []float64{10.0, 1.0, 5.0, 20.0}
	if got := medianFloat64(even); got != 7.5 {
		t.Errorf("medianFloat64(even) = %f, expected 7.5", got)
	}

	empty := []float64{}
	if got := medianFloat64(empty); got != 0 {
		t.Errorf("medianFloat64(empty) = %f, expected 0", got)
	}
}

func TestDefenderDailyAggregations(t *testing.T) {
	if got := meanFloat64([]float64{0, 0.25, 0.75}); got != 1.0/3.0 {
		t.Errorf("meanFloat64() = %f, expected %f", got, 1.0/3.0)
	}
	if got := meanFloat64(nil); got != 0 {
		t.Errorf("meanFloat64(nil) = %f, expected 0", got)
	}
	if got := maxFloat64([]float64{2, 9, 4}); got != 9 {
		t.Errorf("maxFloat64() = %f, expected 9", got)
	}
	if got := maxFloat64(nil); got != 0 {
		t.Errorf("maxFloat64(nil) = %f, expected 0", got)
	}
}

func TestMedianInt(t *testing.T) {
	odd := []int{10, 1, 5}
	if got := medianInt(odd); got != 5 {
		t.Errorf("medianInt(odd) = %d, expected 5", got)
	}

	even := []int{10, 1, 5, 20}
	if got := medianInt(even); got != 8 {
		t.Errorf("medianInt(even) = %d, expected 8", got)
	}
}
