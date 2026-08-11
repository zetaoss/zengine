package k8s

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	statmodels "github.com/zetaoss/zengine/goapp/models/stat"
	"github.com/zetaoss/zengine/goapp/tasks/stat/timeutil"

	"gorm.io/gorm/clause"
)

type NodeMetric struct {
	Name              string
	CPUUsage          float64
	CPUAllocatable    float64
	MemoryUsage       float64
	MemoryAllocatable float64
}

type PodMetric struct {
	Name        string
	Namespace   string
	CPUUsage    float64
	MemoryUsage float64
}

type PVCMetric struct {
	Name         string
	Namespace    string
	Usage        float64
	Capacity     float64
	UsagePercent float64
}

var prometheusHTTPClient = &http.Client{}

type HourlyTask struct{}

func NewHourlyTask() *HourlyTask {
	return &HourlyTask{}
}

type HourlyTaskPayload struct {
	Timeslot string `json:"timeslot"`
}

func (a *HourlyTask) Decode(raw []byte) (any, error) {
	var input HourlyTaskPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, err
		}
	}
	return input, nil
}

func (j *HourlyTask) Execute(ctx context.Context, taskCtx taskctx.Context, input HourlyTaskPayload) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	endpoint := ""
	nodepool := ""
	namespace := ""
	pvcName := ""

	if cfg := taskCtx.Config(); cfg != nil {
		endpoint = cfg.API.MonitoringEndpoint
		nodepool = cfg.API.MonitoringNodepool
		namespace = cfg.API.MonitoringNamespace
		pvcName = cfg.API.MonitoringPVC
	}
	if endpoint == "" {
		endpoint = os.Getenv("MONITORING_ENDPOINT")
	}
	if endpoint == "" {
		return nil, fmt.Errorf("missing MONITORING_ENDPOINT")
	}

	if nodepool == "" {
		nodepool = os.Getenv("MONITORING_NODEPOOL")
	}
	if nodepool == "" {
		return nil, fmt.Errorf("missing MONITORING_NODEPOOL")
	}

	if namespace == "" {
		namespace = os.Getenv("MONITORING_NAMESPACE")
	}
	if namespace == "" {
		return nil, fmt.Errorf("missing MONITORING_NAMESPACE")
	}
	if pvcName == "" {
		pvcName = os.Getenv("MONITORING_PVC")
	}
	if pvcName == "" {
		return nil, fmt.Errorf("missing MONITORING_PVC")
	}

	nodes, pods, err := FetchAndParseMetrics(ctx, endpoint, nodepool, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch k8s metrics: %w", err)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes metrics scraped for nodepool %s", nodepool)
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods metrics scraped for namespace %s", namespace)
	}
	pvcs, err := FetchPVCMetrics(ctx, endpoint, namespace, pvcName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PVC metrics: %w", err)
	}
	if len(pvcs) == 0 {
		return nil, fmt.Errorf("no PVC metrics scraped for %s/%s", namespace, pvcName)
	}

	ts := input.Timeslot
	if ts == "" {
		ts = timeutil.HourlyEndUTC(time.Now().UTC(), 0).Format("2006-01-02 15:04:05")
	}

	var totNodeCPUUsage, totNodeCPUAlloc, totNodeMemUsage, totNodeMemAlloc float64
	for _, n := range nodes {
		totNodeCPUUsage += n.CPUUsage
		totNodeCPUAlloc += n.CPUAllocatable
		totNodeMemUsage += n.MemoryUsage
		totNodeMemAlloc += n.MemoryAllocatable
	}

	var totPodCPUUsage, totPodMemUsage float64
	for _, p := range pods {
		totPodCPUUsage += p.CPUUsage
		totPodMemUsage += p.MemoryUsage
	}

	row := statmodels.K8sHourly{
		Timeslot:              ts,
		NodeCPUUsage:          totNodeCPUUsage,
		NodeCPUAllocatable:    totNodeCPUAlloc,
		NodeMemoryUsage:       totNodeMemUsage,
		NodeMemoryAllocatable: totNodeMemAlloc,
		PodCPUUsage:           totPodCPUUsage,
		PodMemoryUsage:        totPodMemUsage,
		PVCStorageUsage:       pvcs[0].Usage,
		PVCStorageCapacity:    pvcs[0].Capacity,
		PodCount:              len(pods),
	}

	if err := db.Table("stat_k8s_hourly").AutoMigrate(&statmodels.K8sHourly{}); err != nil {
		return nil, err
	}

	if err := db.Table("stat_k8s_hourly").Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&row).Error; err != nil {
		return nil, err
	}

	return app.H{"row": row}, nil
}

type prometheusQueryResult struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

type prometheusQueryData struct {
	ResultType string                  `json:"resultType"`
	Result     []prometheusQueryResult `json:"result"`
}

type prometheusQueryResponse struct {
	Status string              `json:"status"`
	Data   prometheusQueryData `json:"data"`
}

func queryPrometheusURL(endpoint, queryStr string) string {
	baseURL := endpoint
	if !strings.Contains(baseURL, "/api/v1/query") {
		baseURL = strings.TrimRight(baseURL, "/") + "/api/v1/query"
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return endpoint + "?query=" + url.QueryEscape(queryStr)
	}
	q := u.Query()
	q.Set("query", queryStr)
	u.RawQuery = q.Encode()
	return u.String()
}

func queryPrometheusAPI(ctx context.Context, client *http.Client, endpoint, queryStr string) ([]prometheusQueryResult, error) {
	reqURL := queryPrometheusURL(endpoint, queryStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus status code %d", resp.StatusCode)
	}

	var res prometheusQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if res.Status != "success" {
		return nil, fmt.Errorf("prometheus status %s", res.Status)
	}

	return res.Data.Result, nil
}

func queryPrometheusMetric(ctx context.Context, client *http.Client, endpoint string, queries []string) []prometheusQueryResult {
	for _, q := range queries {
		if q == "" {
			continue
		}
		res, err := queryPrometheusAPI(ctx, client, endpoint, q)
		if err == nil && len(res) > 0 {
			return res
		}
	}
	return nil
}

func extractNodeName(m map[string]string) string {
	if n := m["node"]; n != "" {
		return n
	}
	if n := m["name"]; n != "" {
		return n
	}
	if n := m["instance"]; n != "" {
		return n
	}
	if n := m["kubernetes_node"]; n != "" {
		return n
	}
	return ""
}

func extractPodName(m map[string]string) string {
	if p := m["pod"]; p != "" {
		return p
	}
	if p := m["name"]; p != "" {
		return p
	}
	if p := m["pod_name"]; p != "" {
		return p
	}
	return ""
}

func filterNode(nodeName, nodepool string) bool {
	if nodeName == "" {
		return false
	}
	if strings.EqualFold(nodeName, "total") || strings.EqualFold(nodeName, nodepool) {
		return false
	}
	if nodepool == "" {
		return true
	}
	pattern := fmt.Sprintf("-%s-", strings.Trim(nodepool, "-"))
	return strings.Contains(nodeName, pattern) || strings.Contains(nodeName, nodepool)
}

func filterPod(podName, podNS, namespace string) bool {
	if podName == "" {
		return false
	}
	if strings.EqualFold(podName, "total") || strings.EqualFold(podName, namespace) {
		return false
	}
	if namespace != "" && podNS != "" && podNS != namespace {
		return false
	}
	return true
}

func parsePrometheusValue(val []any) (float64, bool) {
	if len(val) < 2 {
		return 0, false
	}
	switch v := val[1].(type) {
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
	case float64:
		return v, true
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func FetchPVCMetrics(ctx context.Context, endpoint, namespace, pvcName string) ([]PVCMetric, error) {
	if namespace == "" {
		return nil, fmt.Errorf("missing namespace parameter")
	}
	if pvcName == "" {
		return nil, fmt.Errorf("missing PVC parameter")
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := prometheusHTTPClient
	pvcMap := make(map[string]*PVCMetric)
	usageFound := make(map[string]bool)
	capacityFound := make(map[string]bool)
	getPVC := func(name, ns string) *PVCMetric {
		key := ns + "\x00" + name
		pvc, ok := pvcMap[key]
		if !ok {
			pvc = &PVCMetric{Name: name, Namespace: ns}
			pvcMap[key] = pvc
		}
		return pvc
	}

	used, err := queryPrometheusAPI(
		reqCtx,
		client,
		endpoint,
		fmt.Sprintf(
			`max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_used_bytes{namespace=%q, persistentvolumeclaim=%q})`,
			namespace,
			pvcName,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("query PVC usage: %w", err)
	}
	for _, result := range used {
		name := result.Metric["persistentvolumeclaim"]
		ns := result.Metric["namespace"]
		if name != pvcName || ns != namespace {
			continue
		}
		if value, ok := parsePrometheusValue(result.Value); ok {
			getPVC(name, ns).Usage = value
			usageFound[ns+"\x00"+name] = true
		}
	}

	capacity, err := queryPrometheusAPI(
		reqCtx,
		client,
		endpoint,
		fmt.Sprintf(
			`max by (namespace, persistentvolumeclaim) (kubelet_volume_stats_capacity_bytes{namespace=%q, persistentvolumeclaim=%q})`,
			namespace,
			pvcName,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("query PVC capacity: %w", err)
	}
	for _, result := range capacity {
		name := result.Metric["persistentvolumeclaim"]
		ns := result.Metric["namespace"]
		if name != pvcName || ns != namespace {
			continue
		}
		if value, ok := parsePrometheusValue(result.Value); ok {
			getPVC(name, ns).Capacity = value
			capacityFound[ns+"\x00"+name] = true
		}
	}

	for key, pvc := range pvcMap {
		if !usageFound[key] {
			return nil, fmt.Errorf("missing PVC usage metric for %s/%s", pvc.Namespace, pvc.Name)
		}
		if !capacityFound[key] || pvc.Capacity <= 0 {
			return nil, fmt.Errorf("missing or invalid PVC capacity metric for %s/%s", pvc.Namespace, pvc.Name)
		}
		pvc.UsagePercent = (pvc.Usage / pvc.Capacity) * 100
	}

	pvcs := make([]PVCMetric, 0, len(pvcMap))
	for _, pvc := range pvcMap {
		pvcs = append(pvcs, *pvc)
	}
	sort.Slice(pvcs, func(i, j int) bool {
		if pvcs[i].Namespace == pvcs[j].Namespace {
			return pvcs[i].Name < pvcs[j].Name
		}
		return pvcs[i].Namespace < pvcs[j].Namespace
	})
	return pvcs, nil
}

func FetchAndParseMetrics(ctx context.Context, endpoint, nodepool, namespace string) ([]NodeMetric, []PodMetric, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := prometheusHTTPClient

	nodeMap := make(map[string]*NodeMetric)
	podMap := make(map[string]*PodMetric)

	getNode := func(name string) *NodeMetric {
		n, ok := nodeMap[name]
		if !ok {
			n = &NodeMetric{Name: name}
			nodeMap[name] = n
		}
		return n
	}

	getPod := func(name, ns string) *PodMetric {
		p, ok := podMap[name]
		if !ok {
			p = &PodMetric{Name: name, Namespace: ns}
			podMap[name] = p
		}
		return p
	}

	nodepoolFilter := nodepool
	if nodepoolFilter != "" {
		nodepoolFilter = strings.Trim(nodepool, "-")
	}

	// PromQL Queries based on MONITORING_ENDPOINT Prometheus API
	// Node CPU Usage (cores)
	cpuUsageRes := queryPrometheusMetric(reqCtx, client, endpoint, []string{
		fmt.Sprintf(`rate(node_cpu_usage_seconds_total{node=~".*%s.*"}[5m])`, nodepoolFilter),
		`rate(node_cpu_usage_seconds_total[5m])`,
		`sum by (node) (rate(node_cpu_usage_seconds_total[5m]))`,
		`k8s_top_node_cpu_cores`,
	})
	for _, r := range cpuUsageRes {
		nodeName := extractNodeName(r.Metric)
		if filterNode(nodeName, nodepool) {
			if val, ok := parsePrometheusValue(r.Value); ok {
				getNode(nodeName).CPUUsage = val
			}
		}
	}

	// Node CPU Allocatable (cores)
	cpuAllocRes := queryPrometheusMetric(reqCtx, client, endpoint, []string{
		fmt.Sprintf(`kube_node_status_allocatable{node=~".*%s.*", resource="cpu", unit="core"}`, nodepoolFilter),
		fmt.Sprintf(`kube_node_status_allocatable{node=~".*%s.*", resource="cpu"}`, nodepoolFilter),
		`kube_node_status_allocatable{resource="cpu", unit="core"}`,
		`kube_node_status_allocatable{resource="cpu"}`,
		`k8s_top_node_allocatable_cpu_cores`,
	})
	for _, r := range cpuAllocRes {
		nodeName := extractNodeName(r.Metric)
		if filterNode(nodeName, nodepool) {
			if val, ok := parsePrometheusValue(r.Value); ok {
				getNode(nodeName).CPUAllocatable = val
			}
		}
	}

	// Node Memory Usage (bytes)
	memUsageRes := queryPrometheusMetric(reqCtx, client, endpoint, []string{
		fmt.Sprintf(`node_memory_working_set_bytes{node=~".*%s.*"}`, nodepoolFilter),
		`node_memory_working_set_bytes`,
		`node_memory_usage_bytes`,
		`k8s_top_node_memory_bytes`,
	})
	for _, r := range memUsageRes {
		nodeName := extractNodeName(r.Metric)
		if filterNode(nodeName, nodepool) {
			if val, ok := parsePrometheusValue(r.Value); ok {
				getNode(nodeName).MemoryUsage = val
			}
		}
	}

	// Node Memory Allocatable (bytes)
	memAllocRes := queryPrometheusMetric(reqCtx, client, endpoint, []string{
		fmt.Sprintf(`kube_node_status_allocatable{node=~".*%s.*", resource="memory"}`, nodepoolFilter),
		`kube_node_status_allocatable{resource="memory"}`,
		`k8s_top_node_allocatable_memory_bytes`,
	})
	for _, r := range memAllocRes {
		nodeName := extractNodeName(r.Metric)
		if filterNode(nodeName, nodepool) {
			if val, ok := parsePrometheusValue(r.Value); ok {
				getNode(nodeName).MemoryAllocatable = val
			}
		}
	}

	// Pod CPU Usage (cores)
	podCPURes := queryPrometheusMetric(reqCtx, client, endpoint, []string{
		fmt.Sprintf(`rate(pod_cpu_usage_seconds_total{namespace="%s"}[5m])`, namespace),
		fmt.Sprintf(`sum by (pod, namespace) (rate(pod_cpu_usage_seconds_total{namespace="%s"}[5m]))`, namespace),
		fmt.Sprintf(`sum by (pod, namespace) (rate(container_cpu_usage_seconds_total{namespace="%s", container!=""}[5m]))`, namespace),
		fmt.Sprintf(`k8s_top_pod_cpu_cores{namespace="%s"}`, namespace),
	})
	for _, r := range podCPURes {
		podName := extractPodName(r.Metric)
		podNS := r.Metric["namespace"]
		if podNS == "" {
			podNS = namespace
		}
		if filterPod(podName, podNS, namespace) {
			if val, ok := parsePrometheusValue(r.Value); ok {
				getPod(podName, podNS).CPUUsage = val
			}
		}
	}

	// Pod Memory Usage (bytes)
	podMemRes := queryPrometheusMetric(reqCtx, client, endpoint, []string{
		fmt.Sprintf(`pod_memory_working_set_bytes{namespace="%s"}`, namespace),
		fmt.Sprintf(`container_memory_working_set_bytes{namespace="%s", container!=""}`, namespace),
		fmt.Sprintf(`pod_memory_usage_bytes{namespace="%s"}`, namespace),
		fmt.Sprintf(`k8s_top_pod_memory_bytes{namespace="%s"}`, namespace),
	})
	for _, r := range podMemRes {
		podName := extractPodName(r.Metric)
		podNS := r.Metric["namespace"]
		if podNS == "" {
			podNS = namespace
		}
		if filterPod(podName, podNS, namespace) {
			if val, ok := parsePrometheusValue(r.Value); ok {
				getPod(podName, podNS).MemoryUsage = val
			}
		}
	}

	if len(nodeMap) > 0 || len(podMap) > 0 {
		nodes := make([]NodeMetric, 0, len(nodeMap))
		for _, n := range nodeMap {
			nodes = append(nodes, *n)
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})

		pods := make([]PodMetric, 0, len(podMap))
		for _, p := range podMap {
			pods = append(pods, *p)
		}
		sort.Slice(pods, func(i, j int) bool {
			return pods[i].Name < pods[j].Name
		})

		return nodes, pods, nil
	}

	// Fallback to direct HTTP GET on endpoint for raw exporter text input
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	nodes, err := ParsePrometheusNodeMetrics(strings.NewReader(string(body)), nodepool)
	if err != nil {
		return nil, nil, err
	}
	pods, err := ParsePrometheusPodMetrics(strings.NewReader(string(body)), namespace)
	if err != nil {
		return nil, nil, err
	}
	return nodes, pods, nil
}

func ParsePrometheusNodeMetrics(r io.Reader, nodepool string) ([]NodeMetric, error) {
	if nodepool == "" {
		return nil, fmt.Errorf("missing nodepool parameter")
	}

	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(bodyBytes))
	if strings.HasPrefix(trimmed, "{") {
		var res prometheusQueryResponse
		if err := json.Unmarshal([]byte(trimmed), &res); err == nil && res.Status == "success" {
			nodeMap := make(map[string]*NodeMetric)
			for _, r := range res.Data.Result {
				nodeName := extractNodeName(r.Metric)
				if !filterNode(nodeName, nodepool) {
					continue
				}
				val, ok := parsePrometheusValue(r.Value)
				if !ok {
					continue
				}
				node, exists := nodeMap[nodeName]
				if !exists {
					node = &NodeMetric{Name: nodeName}
					nodeMap[nodeName] = node
				}
				metricName := r.Metric["__name__"]
				switch metricName {
				case "k8s_top_node_cpu_cores", "node_cpu_usage_seconds_total":
					node.CPUUsage = val
				case "k8s_top_node_allocatable_cpu_cores", "kube_node_status_allocatable":
					if r.Metric["resource"] == "cpu" || r.Metric["resource"] == "" {
						node.CPUAllocatable = val
					}
				case "k8s_top_node_memory_bytes", "node_memory_working_set_bytes":
					node.MemoryUsage = val
				case "k8s_top_node_allocatable_memory_bytes":
					node.MemoryAllocatable = val
				}
			}
			nodes := make([]NodeMetric, 0, len(nodeMap))
			for _, n := range nodeMap {
				nodes = append(nodes, *n)
			}
			sort.Slice(nodes, func(i, j int) bool {
				return nodes[i].Name < nodes[j].Name
			})
			return nodes, nil
		}
	}

	return parseTextNodeMetrics(strings.NewReader(trimmed), nodepool)
}

func parseTextNodeMetrics(r io.Reader, nodepool string) ([]NodeMetric, error) {
	nodeMap := make(map[string]*NodeMetric)
	scanner := bufio.NewScanner(r)

	pattern := fmt.Sprintf("-%s-", strings.Trim(nodepool, "-"))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		metricName, labels, val, err := parsePrometheusLine(line)
		if err != nil {
			continue
		}
		if !isNodeMetric(metricName) {
			continue
		}

		nodeName := labels["name"]
		if nodeName == "" {
			nodeName = labels["node"]
		}
		if strings.EqualFold(nodeName, "total") || strings.EqualFold(nodeName, nodepool) {
			continue
		}
		if !strings.Contains(nodeName, pattern) && !strings.Contains(nodeName, nodepool) {
			continue
		}

		node, exists := nodeMap[nodeName]
		if !exists {
			node = &NodeMetric{Name: nodeName}
			nodeMap[nodeName] = node
		}

		switch metricName {
		case "k8s_top_node_cpu_cores":
			node.CPUUsage = val
		case "k8s_top_node_allocatable_cpu_cores":
			node.CPUAllocatable = val
		case "k8s_top_node_memory_bytes":
			node.MemoryUsage = val
		case "k8s_top_node_allocatable_memory_bytes":
			node.MemoryAllocatable = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	nodes := make([]NodeMetric, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, *n)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return nodes, nil
}

func ParsePrometheusPodMetrics(r io.Reader, namespace string) ([]PodMetric, error) {
	if namespace == "" {
		return nil, fmt.Errorf("missing namespace parameter")
	}

	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(bodyBytes))
	if strings.HasPrefix(trimmed, "{") {
		var res prometheusQueryResponse
		if err := json.Unmarshal([]byte(trimmed), &res); err == nil && res.Status == "success" {
			podMap := make(map[string]*PodMetric)
			for _, r := range res.Data.Result {
				podName := extractPodName(r.Metric)
				podNS := r.Metric["namespace"]
				if podNS == "" {
					podNS = namespace
				}
				if !filterPod(podName, podNS, namespace) {
					continue
				}
				val, ok := parsePrometheusValue(r.Value)
				if !ok {
					continue
				}
				pod, exists := podMap[podName]
				if !exists {
					pod = &PodMetric{Name: podName, Namespace: podNS}
					podMap[podName] = pod
				}
				metricName := r.Metric["__name__"]
				switch metricName {
				case "k8s_top_pod_cpu_cores", "pod_cpu_usage_seconds_total", "container_cpu_usage_seconds_total":
					pod.CPUUsage = val
				case "k8s_top_pod_memory_bytes", "pod_memory_working_set_bytes", "container_memory_working_set_bytes":
					pod.MemoryUsage = val
				}
			}
			pods := make([]PodMetric, 0, len(podMap))
			for _, pod := range podMap {
				pods = append(pods, *pod)
			}
			sort.Slice(pods, func(i, j int) bool {
				return pods[i].Name < pods[j].Name
			})
			return pods, nil
		}
	}

	return parseTextPodMetrics(strings.NewReader(trimmed), namespace)
}

func parseTextPodMetrics(r io.Reader, namespace string) ([]PodMetric, error) {
	podMap := make(map[string]*PodMetric)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		metricName, labels, val, err := parsePrometheusLine(line)
		if err != nil || (metricName != "k8s_top_pod_cpu_cores" && metricName != "k8s_top_pod_memory_bytes") {
			continue
		}
		podName := labels["name"]
		if strings.EqualFold(podName, "total") || strings.EqualFold(podName, namespace) {
			continue
		}
		if labels["namespace"] != namespace || podName == "" {
			continue
		}

		pod, exists := podMap[podName]
		if !exists {
			pod = &PodMetric{Name: podName, Namespace: labels["namespace"]}
			podMap[pod.Name] = pod
		}
		switch metricName {
		case "k8s_top_pod_cpu_cores":
			pod.CPUUsage = val
		case "k8s_top_pod_memory_bytes":
			pod.MemoryUsage = val
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	pods := make([]PodMetric, 0, len(podMap))
	for _, pod := range podMap {
		pods = append(pods, *pod)
	}
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})
	return pods, nil
}

func isNodeMetric(name string) bool {
	switch name {
	case "k8s_top_node_cpu_cores",
		"k8s_top_node_allocatable_cpu_cores",
		"k8s_top_node_memory_bytes",
		"k8s_top_node_allocatable_memory_bytes":
		return true
	default:
		return false
	}
}

func parsePrometheusLine(line string) (string, map[string]string, float64, error) {
	labels := make(map[string]string)
	braceStart := strings.IndexByte(line, '{')
	braceEnd := strings.LastIndexByte(line, '}')

	var metricName, valStr string

	if braceStart != -1 && braceEnd != -1 && braceEnd > braceStart {
		metricName = strings.TrimSpace(line[:braceStart])
		labelStr := line[braceStart+1 : braceEnd]
		valStr = strings.TrimSpace(line[braceEnd+1:])

		for _, pair := range strings.Split(labelStr, ",") {
			pair = strings.TrimSpace(pair)
			eq := strings.IndexByte(pair, '=')
			if eq > 0 {
				k := strings.TrimSpace(pair[:eq])
				v := strings.Trim(strings.TrimSpace(pair[eq+1:]), `"`)
				labels[k] = v
			}
		}
	} else {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return "", nil, 0, fmt.Errorf("invalid line format")
		}
		metricName = fields[0]
		valStr = fields[1]
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return "", nil, 0, err
	}

	return metricName, labels, val, nil
}
