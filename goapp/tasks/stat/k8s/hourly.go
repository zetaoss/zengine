package k8s

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	if cfg := taskCtx.Config(); cfg != nil {
		endpoint = cfg.API.K8sMetricsEndpoint
		nodepool = cfg.API.K8sMetricsNodepool
		namespace = cfg.API.K8sMetricsNamespace
	}
	if endpoint == "" {
		endpoint = os.Getenv("K8SMETRICS_ENDPOINT")
	}
	if endpoint == "" {
		return nil, fmt.Errorf("missing K8SMETRICS_ENDPOINT")
	}

	if nodepool == "" {
		nodepool = os.Getenv("K8SMETRICS_NODEPOOL")
	}
	if nodepool == "" {
		return nil, fmt.Errorf("missing K8SMETRICS_NODEPOOL")
	}

	if namespace == "" {
		namespace = os.Getenv("K8SMETRICS_NAMESPACE")
	}
	if namespace == "" {
		return nil, fmt.Errorf("missing K8SMETRICS_NAMESPACE")
	}

	nodes, pods, err := FetchAndParseMetrics(ctx, endpoint, nodepool, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch k8s metrics: %w", err)
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

func FetchAndParseMetrics(ctx context.Context, endpoint, nodepool, namespace string) ([]NodeMetric, []PodMetric, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
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
