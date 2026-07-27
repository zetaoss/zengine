package main

import (
	"strings"
	"testing"
)

func TestRunMetricsRejectsEndpointFlag(t *testing.T) {
	if err := runMetrics(nil, []string{"--endpoint", "http://metrics.example.test/metrics"}); err == nil {
		t.Fatal("expected --endpoint to be rejected")
	}
}

func TestRunMetricsReturnsErrorOnMissingEnv(t *testing.T) {
	t.Setenv("K8SMETRICS_ENDPOINT", "")
	t.Setenv("K8SMETRICS_NODEPOOL", "")
	t.Setenv("K8SMETRICS_NAMESPACE", "")

	if err := runMetrics(nil, nil); err == nil {
		t.Fatal("expected error when K8SMETRICS env vars are missing")
	}
}

func TestParsePrometheusNodeMetrics(t *testing.T) {
	input := `# HELP k8s_top_node_allocatable_cpu_cores Allocatable CPU of the node in cores.
# TYPE k8s_top_node_allocatable_cpu_cores gauge
k8s_top_node_allocatable_cpu_cores{name="my-cluster-pool-a-node-1"} 1.93
k8s_top_node_allocatable_cpu_cores{name="my-cluster-pool-b-node-2"} 1.93
# HELP k8s_top_node_allocatable_memory_bytes Allocatable memory of the node in bytes.
# TYPE k8s_top_node_allocatable_memory_bytes gauge
k8s_top_node_allocatable_memory_bytes{name="my-cluster-pool-a-node-1"} 1.3918449664e+10
k8s_top_node_allocatable_memory_bytes{name="my-cluster-pool-b-node-2"} 6.31836672e+09
# HELP k8s_top_node_cpu_cores CPU usage of the node in cores.
# TYPE k8s_top_node_cpu_cores gauge
k8s_top_node_cpu_cores{name="my-cluster-pool-a-node-1"} 0.833
k8s_top_node_cpu_cores{name="my-cluster-pool-b-node-2"} 0.187
# HELP k8s_top_node_memory_bytes Memory usage of the node in bytes.
# TYPE k8s_top_node_memory_bytes gauge
k8s_top_node_memory_bytes{name="my-cluster-pool-a-node-1"} 8.217997312e+09
k8s_top_node_memory_bytes{name="my-cluster-pool-b-node-2"} 5.892268032e+09
unrelated_metric{name="not-a-node"} 1
`

	nodes, err := parsePrometheusNodeMetrics(strings.NewReader(input), "pool-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 pool-a node, got %d", len(nodes))
	}

	n1 := nodes[0]
	if n1.Name != "my-cluster-pool-a-node-1" {
		t.Errorf("expected node name my-cluster-pool-a-node-1, got %s", n1.Name)
	}
	if n1.CPUAllocatable != 1.93 {
		t.Errorf("expected CPUAllocatable 1.93, got %f", n1.CPUAllocatable)
	}
	if n1.CPUUsage != 0.833 {
		t.Errorf("expected CPUUsage 0.833, got %f", n1.CPUUsage)
	}
	if n1.MemoryAllocatable != 1.3918449664e+10 {
		t.Errorf("expected MemoryAllocatable 1.3918449664e+10, got %f", n1.MemoryAllocatable)
	}
	if n1.MemoryUsage != 8.217997312e+09 {
		t.Errorf("expected MemoryUsage 8.217997312e+09, got %f", n1.MemoryUsage)
	}
}

func TestParsePrometheusNodeMetricsCustomNodepool(t *testing.T) {
	input := `# HELP k8s_top_node_cpu_cores CPU usage of the node in cores.
# TYPE k8s_top_node_cpu_cores gauge
k8s_top_node_cpu_cores{name="my-cluster-pool-a-node-1"} 0.833
k8s_top_node_cpu_cores{name="my-cluster-pool-b-node-2"} 0.187
`

	nodes, err := parsePrometheusNodeMetrics(strings.NewReader(input), "pool-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 pool-b node, got %d", len(nodes))
	}

	if nodes[0].Name != "my-cluster-pool-b-node-2" {
		t.Errorf("expected node name my-cluster-pool-b-node-2, got %s", nodes[0].Name)
	}
}

func TestParsePrometheusNodeMetricsSupportsNodeLabel(t *testing.T) {
	input := `k8s_top_node_cpu_cores{node="cluster-pool-x-worker-1"} 0.5`

	nodes, err := parsePrometheusNodeMetrics(strings.NewReader(input), "pool-x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Name != "cluster-pool-x-worker-1" || nodes[0].CPUUsage != 0.5 {
		t.Errorf("unexpected node: %+v", nodes[0])
	}
}

func TestParsePrometheusPodMetricsFiltersNamespace(t *testing.T) {
	input := `k8s_top_pod_memory_bytes{name="db-service",namespace="ns-a"} 4.79510528e+08
k8s_top_pod_memory_bytes{name="app-service",namespace="ns-target"} 8.9980928e+08
k8s_top_pod_memory_bytes{name="db-service",namespace="ns-target"} 2.29398528e+09
k8s_top_pod_cpu_cores{name="app-service",namespace="ns-target"} 0.5
k8s_top_pod_cpu_cores{name="db-service",namespace="ns-target"} 0.25
k8s_top_pod_cpu_cores{name="worker-service",namespace="ns-a"} 0.1
`

	pods, err := parsePrometheusPodMetrics(strings.NewReader(input), "ns-target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 2 {
		t.Fatalf("expected 2 ns-target memory metrics, got %d", len(pods))
	}
	if pods[0].Name != "app-service" || pods[0].Namespace != "ns-target" || pods[0].CPUUsage != 0.5 {
		t.Errorf("unexpected first pod: %+v", pods[0])
	}
	if pods[1].Name != "db-service" || pods[1].CPUUsage != 0.25 || pods[1].MemoryUsage != 2.29398528e+09 {
		t.Errorf("unexpected second pod: %+v", pods[1])
	}
}

func TestParsePrometheusPodMetricsCustomNamespace(t *testing.T) {
	input := `k8s_top_pod_memory_bytes{name="db-service",namespace="ns-a"} 4.79510528e+08
k8s_top_pod_memory_bytes{name="app-service",namespace="ns-target"} 8.9980928e+08
`

	pods, err := parsePrometheusPodMetrics(strings.NewReader(input), "ns-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 ns-a pod metric, got %d", len(pods))
	}
	if pods[0].Name != "db-service" || pods[0].Namespace != "ns-a" {
		t.Errorf("unexpected pod: %+v", pods[0])
	}
}

func TestFormatPct(t *testing.T) {
	if got := formatPct(0.724, 1.93); got != "37.51%" {
		t.Errorf("formatPct(0.724, 1.93) = %s, expected 37.51%%", got)
	}
	if got := formatPct(1, 0); got != "0.00%" {
		t.Errorf("formatPct(1, 0) = %s, expected 0.00%%", got)
	}
}

func TestFixedMetricUnits(t *testing.T) {
	if got := formatCPU(0.724); got != "724m" {
		t.Errorf("formatCPU(0.724) = %s, expected 724m", got)
	}
	if got := formatMemoryMi(8.24303616e+09); got != "7861Mi" {
		t.Errorf("formatMemoryMi(8.24303616e+09) = %s, expected 7861Mi", got)
	}
}
