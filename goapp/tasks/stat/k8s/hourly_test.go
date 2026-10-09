package k8s

import (
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const bobBody = `{"status":"ok","time":"2026-08-15T12:00:00Z","result":{
"node_cpu_usage":[{"labels":{"node":"n1"},"value":0.833}],
"node_cpu_allocatable":[{"labels":{"node":"n1"},"value":1.93}],
"node_memory_usage":[{"labels":{"node":"n1"},"value":8217997312}],
"node_memory_allocatable":[{"labels":{"node":"n1"},"value":13918449664}],
"pod_cpu_usage":[{"labels":{"namespace":"prod3","pod":"web-1"},"value":0.45},{"labels":{"namespace":"prod3","pod":"web-2"},"value":0.05}],
"pod_memory_usage":[{"labels":{"namespace":"prod3","pod":"web-1"},"value":524288000}],
"pod_count":[{"labels":{},"value":2}],
"pvc_storage_usage":[{"labels":{"namespace":"prod3","persistentvolumeclaim":"db"},"value":11811160064}],
"pvc_storage_capacity":[{"labels":{"namespace":"prod3","persistentvolumeclaim":"db"},"value":106300440576}],
"defender_fighting_ratio":[{"labels":{},"value":0.25}],
"defender_max_level":[{"labels":{},"value":7}]}}`

func fakeBob(t *testing.T, status int, body string, gotURL *string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotURL != nil {
			*gotURL = r.URL.String()
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestFetchMetrics(t *testing.T) {
	var gotURL string
	at := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	m, err := FetchMetrics(context.Background(), fakeBob(t, http.StatusOK, bobBody, &gotURL), &at)
	if err != nil {
		t.Fatal(err)
	}
	if gotURL != "/metrics/?time=2026-08-15T12%3A00%3A00Z" {
		t.Errorf("request URL = %s", gotURL)
	}
	if m.Total("pod_cpu_usage") != 0.5 || m.Total("pod_count") != 2 || m.Total("defender_max_level") != 7 {
		t.Errorf("totals: pod_cpu=%v pods=%v level=%v", m.Total("pod_cpu_usage"), m.Total("pod_count"), m.Total("defender_max_level"))
	}

	nodes := m.Nodes()
	if len(nodes) != 1 || nodes[0] != (NodeMetric{Name: "n1", CPUUsage: 0.833, CPUAllocatable: 1.93, MemoryUsage: 8217997312, MemoryAllocatable: 13918449664}) {
		t.Errorf("nodes=%+v", nodes)
	}
	pods := m.Pods()
	if len(pods) != 2 || pods[0] != (PodMetric{Name: "web-1", Namespace: "prod3", CPUUsage: 0.45, MemoryUsage: 524288000}) || pods[1].MemoryUsage != 0 {
		t.Errorf("pods=%+v", pods)
	}
	pvcs := m.PVCs()
	if len(pvcs) != 1 || pvcs[0].Name != "db" || pvcs[0].Namespace != "prod3" || math.Abs(pvcs[0].UsagePercent-11.11111111111111) > 1e-12 {
		t.Errorf("pvcs=%+v", pvcs)
	}
}

func TestFetchMetricsErrors(t *testing.T) {
	if _, err := FetchMetrics(context.Background(), "", nil); err == nil || !strings.Contains(err.Error(), "BOB_ENDPOINT") {
		t.Errorf("empty endpoint: %v", err)
	}
	_, err := FetchMetrics(context.Background(), fakeBob(t, http.StatusBadGateway, `{"status":"error","error":"pod_count: prometheus status code 400"}`, nil), nil)
	if err == nil || !strings.Contains(err.Error(), "pod_count") {
		t.Errorf("bob error: %v", err)
	}
}

func TestFetchMetricsByName(t *testing.T) {
	var gotURL string
	if _, err := FetchMetrics(context.Background(), fakeBob(t, http.StatusOK, bobBody, &gotURL), nil, "pod_count", "node_cpu_usage"); err != nil {
		t.Fatal(err)
	}
	if gotURL != "/metrics/?name=pod_count&name=node_cpu_usage" {
		t.Errorf("request URL = %s", gotURL)
	}

	_, err := FetchMetrics(context.Background(), fakeBob(t, http.StatusNotFound, `{"status":"error","error":"unknown metric: pvc_storage_usage"}`, nil), nil, hourlyMetrics...)
	if err == nil || !strings.Contains(err.Error(), "unknown metric: pvc_storage_usage") {
		t.Errorf("unknown metric: %v", err)
	}
}

func TestValidate(t *testing.T) {
	m, err := FetchMetrics(context.Background(), fakeBob(t, http.StatusOK, bobBody, nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.validate(); err != nil {
		t.Fatalf("complete metrics: %v", err)
	}

	with := func(name string, samples []Sample) Metrics {
		out := Metrics{}
		for k, v := range m {
			out[k] = v
		}
		out[name] = samples
		return out
	}
	if err := with("pvc_storage_usage", []Sample{}).validate(); err == nil || !strings.Contains(err.Error(), "pvc_storage_usage has no data") {
		t.Errorf("empty PVC usage: %v", err)
	}
	if err := with("pvc_storage_usage", []Sample{{Value: 0}}).validate(); err != nil {
		t.Errorf("zero PVC usage should pass: %v", err)
	}
	if err := with("defender_max_level", []Sample{}).validate(); err != nil {
		t.Errorf("no defender data should pass: %v", err)
	}
	if err := with("pvc_storage_capacity", []Sample{{Value: 0}}).validate(); err == nil || !strings.Contains(err.Error(), "capacity") {
		t.Errorf("zero capacity: %v", err)
	}
}

func TestParseHourlyTimeslot(t *testing.T) {
	got, err := parseHourlyTimeslot("2026-08-15 12:00:00")
	if err != nil || !got.Equal(time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("got %v, %v", got, err)
	}
}
