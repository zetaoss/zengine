package statmodels

var StatK8sMetricNames = []string{
	"node_cpu_usage",
	"node_cpu_allocatable",
	"node_memory_usage",
	"node_memory_allocatable",
	"pod_cpu_usage",
	"pod_memory_usage",
	"pod_count",
}

type K8sHourly struct {
	Timeslot              string  `gorm:"column:timeslot;primaryKey"`
	NodeCPUUsage          float64 `gorm:"column:node_cpu_usage"`
	NodeCPUAllocatable    float64 `gorm:"column:node_cpu_allocatable"`
	NodeMemoryUsage       float64 `gorm:"column:node_memory_usage"`
	NodeMemoryAllocatable float64 `gorm:"column:node_memory_allocatable"`
	PodCPUUsage           float64 `gorm:"column:pod_cpu_usage"`
	PodMemoryUsage        float64 `gorm:"column:pod_memory_usage"`
	PodCount              int     `gorm:"column:pod_count"`
}

type K8sDaily struct {
	Timeslot              string  `gorm:"column:timeslot;primaryKey"`
	NodeCPUUsage          float64 `gorm:"column:node_cpu_usage"`
	NodeCPUAllocatable    float64 `gorm:"column:node_cpu_allocatable"`
	NodeMemoryUsage       float64 `gorm:"column:node_memory_usage"`
	NodeMemoryAllocatable float64 `gorm:"column:node_memory_allocatable"`
	PodCPUUsage           float64 `gorm:"column:pod_cpu_usage"`
	PodMemoryUsage        float64 `gorm:"column:pod_memory_usage"`
	PodCount              int     `gorm:"column:pod_count"`
}
