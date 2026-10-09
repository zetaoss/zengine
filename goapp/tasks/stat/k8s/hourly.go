package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
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

// Sample is one series of a bob metric.
type Sample struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
}

// Metrics is bob's /metrics/ result. Metric names match the stat_k8s_hourly columns; the PromQL
// behind them (node pool, namespace, PVC) is configured in bob.
type Metrics map[string][]Sample

var bobHTTPClient = &http.Client{Timeout: 15 * time.Second}

// FetchMetrics gets the named metrics (all when names is empty) from bob at the given time
// (nil: now). bob answers 404 when a name is not configured.
func FetchMetrics(ctx context.Context, bobEndpoint string, at *time.Time, names ...string) (Metrics, error) {
	if bobEndpoint == "" {
		return nil, fmt.Errorf("BOB_ENDPOINT is required")
	}
	params := url.Values{"name": names}
	if at != nil {
		params.Set("time", at.UTC().Format(time.RFC3339))
	}
	u := bobEndpoint + "/metrics/"
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := bobHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Status string  `json:"status"`
		Error  string  `json:"error"`
		Result Metrics `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("bob metrics: status %d: %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || body.Status != "ok" {
		return nil, fmt.Errorf("bob metrics: status %d: %s", resp.StatusCode, body.Error)
	}
	return body.Result, nil
}

// Total sums every series of a metric.
func (m Metrics) Total(name string) float64 {
	var total float64
	for _, s := range m[name] {
		total += s.Value
	}
	return total
}

// byLabels groups a metric's values by the given label values.
func (m Metrics) byLabels(name string, labels ...string) map[[2]string]float64 {
	out := map[[2]string]float64{}
	for _, s := range m[name] {
		var key [2]string
		for i, l := range labels {
			key[i] = s.Labels[l]
		}
		out[key] += s.Value
	}
	return out
}

// Nodes returns per-node values (label node), sorted by name.
func (m Metrics) Nodes() []NodeMetric {
	nodes := map[string]*NodeMetric{}
	set := func(metric string, apply func(*NodeMetric, float64)) {
		for key, v := range m.byLabels(metric, "node") {
			n, ok := nodes[key[0]]
			if !ok {
				n = &NodeMetric{Name: key[0]}
				nodes[key[0]] = n
			}
			apply(n, v)
		}
	}
	set("node_cpu_usage", func(n *NodeMetric, v float64) { n.CPUUsage = v })
	set("node_cpu_allocatable", func(n *NodeMetric, v float64) { n.CPUAllocatable = v })
	set("node_memory_usage", func(n *NodeMetric, v float64) { n.MemoryUsage = v })
	set("node_memory_allocatable", func(n *NodeMetric, v float64) { n.MemoryAllocatable = v })

	out := make([]NodeMetric, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Pods returns per-pod values (labels namespace, pod), sorted by name.
func (m Metrics) Pods() []PodMetric {
	pods := map[[2]string]*PodMetric{}
	set := func(metric string, apply func(*PodMetric, float64)) {
		for key, v := range m.byLabels(metric, "namespace", "pod") {
			p, ok := pods[key]
			if !ok {
				p = &PodMetric{Namespace: key[0], Name: key[1]}
				pods[key] = p
			}
			apply(p, v)
		}
	}
	set("pod_cpu_usage", func(p *PodMetric, v float64) { p.CPUUsage = v })
	set("pod_memory_usage", func(p *PodMetric, v float64) { p.MemoryUsage = v })

	out := make([]PodMetric, 0, len(pods))
	for _, p := range pods {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// PVCs returns per-PVC values (labels namespace, persistentvolumeclaim).
func (m Metrics) PVCs() []PVCMetric {
	capacity := m.byLabels("pvc_storage_capacity", "namespace", "persistentvolumeclaim")
	out := []PVCMetric{}
	for key, used := range m.byLabels("pvc_storage_usage", "namespace", "persistentvolumeclaim") {
		pvc := PVCMetric{Namespace: key[0], Name: key[1], Usage: used, Capacity: capacity[key]}
		if pvc.Capacity > 0 {
			pvc.UsagePercent = pvc.Usage / pvc.Capacity * 100
		}
		out = append(out, pvc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Namespace+out[i].Name < out[j].Namespace+out[j].Name })
	return out
}

// hourlyMetrics are the stat_k8s_hourly columns, requested from bob by name.
var hourlyMetrics = []string{
	"node_cpu_usage", "node_cpu_allocatable", "node_memory_usage", "node_memory_allocatable",
	"pod_cpu_usage", "pod_memory_usage", "pod_count",
	"pvc_storage_usage", "pvc_storage_capacity",
	"defender_fighting_ratio", "defender_max_level",
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

	bobEndpoint := ""
	if cfg := taskCtx.Config(); cfg != nil {
		bobEndpoint = cfg.API.BobEndpoint
	}

	ts := input.Timeslot
	if ts == "" {
		ts = timeutil.HourlyEndUTC(time.Now().UTC(), 0).Format("2006-01-02 15:04:05")
	}
	evaluationTime, err := parseHourlyTimeslot(ts)
	if err != nil {
		return nil, fmt.Errorf("invalid timeslot %q: %w", ts, err)
	}

	m, err := FetchMetrics(ctx, bobEndpoint, &evaluationTime, hourlyMetrics...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch k8s metrics: %w", err)
	}

	row := statmodels.K8sHourly{
		Timeslot:              ts,
		NodeCPUUsage:          m.Total("node_cpu_usage"),
		NodeCPUAllocatable:    m.Total("node_cpu_allocatable"),
		NodeMemoryUsage:       m.Total("node_memory_usage"),
		NodeMemoryAllocatable: m.Total("node_memory_allocatable"),
		PodCPUUsage:           m.Total("pod_cpu_usage"),
		PodMemoryUsage:        m.Total("pod_memory_usage"),
		PVCStorageUsage:       m.Total("pvc_storage_usage"),
		PVCStorageCapacity:    m.Total("pvc_storage_capacity"),
		PodCount:              int(m.Total("pod_count")),
		DefenderFightingRatio: m.Total("defender_fighting_ratio"),
		DefenderMaxLevel:      m.Total("defender_max_level"),
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

func parseHourlyTimeslot(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("expected RFC3339 or 2006-01-02 15:04:05")
}
