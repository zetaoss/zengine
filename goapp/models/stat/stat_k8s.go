package statmodels

var StatK8sMetricNames = []string{
	"node_cpu_usage",
	"node_cpu_allocatable",
	"node_memory_usage",
	"node_memory_allocatable",
	"pod_cpu_usage",
	"pod_memory_usage",
	"pvc_storage_usage",
	"pvc_storage_capacity",
	"pod_count",
	"defender_fighting_ratio",
	"defender_max_level",
}

type K8sHourly struct {
	Timeslot              string  `gorm:"column:timeslot;primaryKey"`
	NodeCPUUsage          float64 `gorm:"column:node_cpu_usage"`
	NodeCPUAllocatable    float64 `gorm:"column:node_cpu_allocatable"`
	NodeMemoryUsage       float64 `gorm:"column:node_memory_usage"`
	NodeMemoryAllocatable float64 `gorm:"column:node_memory_allocatable"`
	PodCPUUsage           float64 `gorm:"column:pod_cpu_usage"`
	PodMemoryUsage        float64 `gorm:"column:pod_memory_usage"`
	PVCStorageUsage       float64 `gorm:"column:pvc_storage_usage"`
	PVCStorageCapacity    float64 `gorm:"column:pvc_storage_capacity"`
	PodCount              int     `gorm:"column:pod_count"`
	DefenderFightingRatio float64 `gorm:"column:defender_fighting_ratio"`
	DefenderMaxLevel      float64 `gorm:"column:defender_max_level"`
}

type K8sDaily struct {
	Timeslot              string  `gorm:"column:timeslot;primaryKey"`
	NodeCPUUsage          float64 `gorm:"column:node_cpu_usage"`
	NodeCPUAllocatable    float64 `gorm:"column:node_cpu_allocatable"`
	NodeMemoryUsage       float64 `gorm:"column:node_memory_usage"`
	NodeMemoryAllocatable float64 `gorm:"column:node_memory_allocatable"`
	PodCPUUsage           float64 `gorm:"column:pod_cpu_usage"`
	PodMemoryUsage        float64 `gorm:"column:pod_memory_usage"`
	PVCStorageUsage       float64 `gorm:"column:pvc_storage_usage"`
	PVCStorageCapacity    float64 `gorm:"column:pvc_storage_capacity"`
	PodCount              int     `gorm:"column:pod_count"`
	DefenderFightingRatio float64 `gorm:"column:defender_fighting_ratio"`
	DefenderMaxLevel      float64 `gorm:"column:defender_max_level"`
}
