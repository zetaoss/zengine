package k8s

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchAndParseMetricsPrometheusAPI(t *testing.T) {
	pvcQueryCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("query")
		w.Header().Set("Content-Type", "application/json")

		var res prometheusQueryResponse
		res.Status = "success"
		res.Data.ResultType = "vector"

		switch q {
		case `rate(node_cpu_usage_seconds_total{node=~".*pool-a.*"}[5m])`:
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"node": "my-cluster-pool-a-node-1"},
					Value:  []any{1712345678.0, "0.833"},
				},
			}
		case `kube_node_status_allocatable{node=~".*pool-a.*", resource="cpu", unit="core"}`:
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"node": "my-cluster-pool-a-node-1"},
					Value:  []any{1712345678.0, "1.93"},
				},
			}
		case `node_memory_working_set_bytes{node=~".*pool-a.*"}`:
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"node": "my-cluster-pool-a-node-1"},
					Value:  []any{1712345678.0, "8217997312"},
				},
			}
		case `kube_node_status_allocatable{node=~".*pool-a.*", resource="memory"}`:
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"node": "my-cluster-pool-a-node-1"},
					Value:  []any{1712345678.0, "13918449664"},
				},
			}
		case `rate(pod_cpu_usage_seconds_total{namespace="prod3"}[5m])`:
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"pod": "web-app-1", "namespace": "prod3"},
					Value:  []any{1712345678.0, "0.45"},
				},
			}
		case `pod_memory_working_set_bytes{namespace="prod3"}`:
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"pod": "web-app-1", "namespace": "prod3"},
					Value:  []any{1712345678.0, "524288000"},
				},
			}
		case `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_used_bytes{namespace="prod3", persistentvolumeclaim="db"})`:
			pvcQueryCount++
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"persistentvolumeclaim": "db", "namespace": "prod3"},
					Value:  []any{1712345678.0, "11811160064"},
				},
			}
		case `max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_capacity_bytes{namespace="prod3", persistentvolumeclaim="db"})`:
			pvcQueryCount++
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"persistentvolumeclaim": "db", "namespace": "prod3"},
					Value:  []any{1712345678.0, "106300440576"},
				},
			}
		}

		_ = json.NewEncoder(w).Encode(res)
	}))
	defer ts.Close()

	nodes, pods, err := FetchAndParseMetrics(context.Background(), ts.URL, "pool-a", "prod3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Name != "my-cluster-pool-a-node-1" {
		t.Errorf("expected node name my-cluster-pool-a-node-1, got %s", nodes[0].Name)
	}
	if nodes[0].CPUUsage != 0.833 {
		t.Errorf("expected CPUUsage 0.833, got %f", nodes[0].CPUUsage)
	}
	if nodes[0].CPUAllocatable != 1.93 {
		t.Errorf("expected CPUAllocatable 1.93, got %f", nodes[0].CPUAllocatable)
	}

	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(pods))
	}
	if pods[0].Name != "web-app-1" || pods[0].CPUUsage != 0.45 || pods[0].MemoryUsage != 524288000 {
		t.Errorf("unexpected pod: %+v", pods[0])
	}

	pvcs, err := FetchPVCMetrics(context.Background(), ts.URL, "prod3", "db")
	if err != nil {
		t.Fatalf("unexpected PVC error: %v", err)
	}
	if len(pvcs) != 1 {
		t.Fatalf("expected 1 PVC, got %d", len(pvcs))
	}
	if pvcs[0].Name != "db" || pvcs[0].Namespace != "prod3" || pvcs[0].Usage != 11811160064 || pvcs[0].Capacity != 106300440576 || math.Abs(pvcs[0].UsagePercent-11.11111111111111) > 1e-12 {
		t.Errorf("unexpected PVC: %+v", pvcs[0])
	}
	if pvcQueryCount != 2 {
		t.Errorf("expected 2 PVC queries, got %d", pvcQueryCount)
	}
}

func TestFetchPVCMetricsRejectsMissingCapacity(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := prometheusQueryResponse{Status: "success"}
		if strings.Contains(r.URL.Query().Get("query"), "used_bytes") {
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"persistentvolumeclaim": "db", "namespace": "prod3"},
					Value:  []any{1712345678.0, "0"},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer ts.Close()

	_, err := FetchPVCMetrics(context.Background(), ts.URL, "prod3", "db")
	if err == nil || !strings.Contains(err.Error(), "capacity") {
		t.Fatalf("expected missing capacity error, got %v", err)
	}
}

func TestFetchAndParseMetricsAggregatesContainerMemoryByPod(t *testing.T) {
	const aggregateQuery = `sum by (pod, namespace) (container_memory_working_set_bytes{namespace="prod3", container!=""})`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := prometheusQueryResponse{Status: "success"}
		if r.URL.Query().Get("query") == aggregateQuery {
			res.Data.Result = []prometheusQueryResult{
				{
					Metric: map[string]string{"pod": "multi-container-pod", "namespace": "prod3"},
					Value:  []any{1712345678.0, "314572800"},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer ts.Close()

	_, pods, err := FetchAndParseMetrics(context.Background(), ts.URL, "pool-a", "prod3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 1 || pods[0].Name != "multi-container-pod" || pods[0].MemoryUsage != 314572800 {
		t.Fatalf("unexpected pod metrics: %+v", pods)
	}
}
