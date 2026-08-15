package k8s

import (
	"context"
	"sort"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	statmodels "github.com/zetaoss/zengine/goapp/models/stat"

	"gorm.io/gorm/clause"
)

type DailyTask struct{}

func NewDailyTask() *DailyTask {
	return &DailyTask{}
}

type dailyGroup struct {
	nodeCPUUsages  []float64
	nodeCPUAllocs  []float64
	nodeMemUsages  []float64
	nodeMemAllocs  []float64
	podCPUUsages   []float64
	podMemUsages   []float64
	pvcStorages    []float64
	pvcCapacities  []float64
	podCounts      []int
	fightingRatios []float64
	maxLevels      []float64
}

func (j *DailyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	since := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
	var hourlyRows []statmodels.K8sHourly
	if err := db.Table("stat_k8s_hourly").Where("timeslot >= ?", since).Find(&hourlyRows).Error; err != nil {
		return nil, err
	}

	if len(hourlyRows) == 0 {
		return app.H{"rows": 0}, nil
	}

	grouped := make(map[string]*dailyGroup)

	for _, h := range hourlyRows {
		if len(h.Timeslot) < 10 {
			continue
		}
		date := h.Timeslot[:10]
		g, exists := grouped[date]
		if !exists {
			g = &dailyGroup{}
			grouped[date] = g
		}
		g.nodeCPUUsages = append(g.nodeCPUUsages, h.NodeCPUUsage)
		g.nodeCPUAllocs = append(g.nodeCPUAllocs, h.NodeCPUAllocatable)
		g.nodeMemUsages = append(g.nodeMemUsages, h.NodeMemoryUsage)
		g.nodeMemAllocs = append(g.nodeMemAllocs, h.NodeMemoryAllocatable)
		g.podCPUUsages = append(g.podCPUUsages, h.PodCPUUsage)
		g.podMemUsages = append(g.podMemUsages, h.PodMemoryUsage)
		// Rows collected before PVC storage tracking was introduced contain the
		// column's zero value. Exclude those missing historical samples.
		if h.PVCStorageUsage > 0 {
			g.pvcStorages = append(g.pvcStorages, h.PVCStorageUsage)
		}
		if h.PVCStorageCapacity > 0 {
			g.pvcCapacities = append(g.pvcCapacities, h.PVCStorageCapacity)
		}
		g.podCounts = append(g.podCounts, h.PodCount)
		g.fightingRatios = append(g.fightingRatios, h.DefenderFightingRatio)
		g.maxLevels = append(g.maxLevels, h.DefenderMaxLevel)
	}

	rows := make([]statmodels.K8sDaily, 0, len(grouped))
	for date, g := range grouped {
		d := statmodels.K8sDaily{
			Timeslot:              date,
			NodeCPUUsage:          medianFloat64(g.nodeCPUUsages),
			NodeCPUAllocatable:    medianFloat64(g.nodeCPUAllocs),
			NodeMemoryUsage:       medianFloat64(g.nodeMemUsages),
			NodeMemoryAllocatable: medianFloat64(g.nodeMemAllocs),
			PodCPUUsage:           medianFloat64(g.podCPUUsages),
			PodMemoryUsage:        medianFloat64(g.podMemUsages),
			PVCStorageUsage:       medianFloat64(g.pvcStorages),
			PVCStorageCapacity:    medianFloat64(g.pvcCapacities),
			PodCount:              medianInt(g.podCounts),
			DefenderFightingRatio: meanFloat64(g.fightingRatios),
			DefenderMaxLevel:      maxFloat64(g.maxLevels),
		}
		rows = append(rows, d)
	}

	if err := db.Table("stat_k8s_daily").AutoMigrate(&statmodels.K8sDaily{}); err != nil {
		return nil, err
	}

	if len(rows) > 0 {
		if err := db.Table("stat_k8s_daily").Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}

	return app.H{"rows": len(rows)}, nil
}

func medianFloat64(vals []float64) float64 {
	n := len(vals)
	if n == 0 {
		return 0
	}
	sorted := make([]float64, n)
	copy(sorted, vals)
	sort.Float64s(sorted)

	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[(n-1)/2] + sorted[n/2]) / 2.0
}

func meanFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}

	var total float64
	for _, val := range vals {
		total += val
	}
	return total / float64(len(vals))
}

func maxFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}

	max := vals[0]
	for _, val := range vals[1:] {
		if val > max {
			max = val
		}
	}
	return max
}

func medianInt(vals []int) int {
	n := len(vals)
	if n == 0 {
		return 0
	}
	sorted := make([]int, n)
	copy(sorted, vals)
	sort.Ints(sorted)

	if n%2 == 1 {
		return sorted[n/2]
	}
	return int((float64(sorted[(n-1)/2])+float64(sorted[n/2]))/2.0 + 0.5)
}
