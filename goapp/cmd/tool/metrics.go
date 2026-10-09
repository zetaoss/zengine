package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/cmd/util/tablewriter"
	"github.com/zetaoss/zengine/goapp/tasks/stat/k8s"
)

type NodeMetric = k8s.NodeMetric
type PodMetric = k8s.PodMetric
type PVCMetric = k8s.PVCMetric

func runMetrics(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("metrics", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	watch := fs.Bool("watch", false, "watch and refresh")
	watchShort := fs.Bool("w", false, "watch and refresh")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("metrics does not accept positional arguments")
	}

	bobEndpoint := ""
	if cfg != nil {
		bobEndpoint = cfg.API.BobEndpoint
	}
	if bobEndpoint == "" {
		bobEndpoint = strings.TrimRight(os.Getenv("BOB_ENDPOINT"), "/")
	}
	if bobEndpoint == "" {
		return fmt.Errorf("missing BOB_ENDPOINT")
	}

	show := func() error {
		if *watch || *watchShort {
			fmt.Print("\033[H\033[J")
			_, _ = fmt.Printf("%s\n\n", time.Now().Format(time.RFC3339))
		}
		m, err := k8s.FetchMetrics(context.Background(), bobEndpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch metrics from %s: %w", bobEndpoint, err)
		}
		if err := printNodeMetrics(m.Nodes()); err != nil {
			return err
		}
		_, _ = fmt.Println()
		if err := printPodMetrics(m.Pods()); err != nil {
			return err
		}
		_, _ = fmt.Println()
		if err := printPVCMetrics(m.PVCs()); err != nil {
			return err
		}
		_, _ = fmt.Println()
		return printDefenderMetrics(m.Total("defender_fighting_ratio"), m.Total("defender_max_level"))
	}

	if err := show(); err != nil {
		return err
	}

	if !*watch && !*watchShort {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := show(); err != nil {
			return err
		}
	}
	return nil
}

func printNodeMetrics(nodes []NodeMetric) error {
	if len(nodes) == 0 {
		_, _ = fmt.Println("No node metrics found.")
		return nil
	}

	tw := tablewriter.New(os.Stdout, "NAME", "CPU(cores)", "CPU(%)", "MEMORY(bytes)", "MEMORY(%)")
	if err := tw.Header(); err != nil {
		return err
	}

	var totalCPUUsage, totalCPUAlloc, totalMemUsage, totalMemAlloc float64

	for _, n := range nodes {
		totalCPUUsage += n.CPUUsage
		totalCPUAlloc += n.CPUAllocatable
		totalMemUsage += n.MemoryUsage
		totalMemAlloc += n.MemoryAllocatable

		cpuUsage := formatCPU(n.CPUUsage)
		cpuPct := formatPct(n.CPUUsage, n.CPUAllocatable)

		memUsage := formatMemoryMi(n.MemoryUsage)
		memPct := formatPct(n.MemoryUsage, n.MemoryAllocatable)

		if err := tw.Row(n.Name, cpuUsage, cpuPct, memUsage, memPct); err != nil {
			return err
		}
	}

	totCPUPct := formatPct(totalCPUUsage, totalCPUAlloc)
	totMemPct := formatPct(totalMemUsage, totalMemAlloc)
	if err := tw.Row("TOTAL", formatCPU(totalCPUUsage), totCPUPct, formatMemoryMi(totalMemUsage), totMemPct); err != nil {
		return err
	}

	return tw.Flush()
}

func printPodMetrics(pods []PodMetric) error {
	if len(pods) == 0 {
		_, _ = fmt.Println("No pod metrics found.")
		return nil
	}

	tw := tablewriter.New(os.Stdout, "NAMESPACE", "POD", "CPU(cores)", "MEMORY(bytes)")
	if err := tw.Header(); err != nil {
		return err
	}

	var totalCPUUsage, totalMemUsage float64

	for _, pod := range pods {
		totalCPUUsage += pod.CPUUsage
		totalMemUsage += pod.MemoryUsage

		if err := tw.Row(pod.Namespace, pod.Name, formatCPU(pod.CPUUsage), formatMemoryMi(pod.MemoryUsage)); err != nil {
			return err
		}
	}

	if err := tw.Row("TOTAL", "", formatCPU(totalCPUUsage), formatMemoryMi(totalMemUsage)); err != nil {
		return err
	}

	return tw.Flush()
}

func printPVCMetrics(pvcs []PVCMetric) error {
	if len(pvcs) == 0 {
		_, _ = fmt.Println("No PVC metrics found.")
		return nil
	}

	tw := tablewriter.New(os.Stdout, "NAMESPACE", "PVC", "STORAGE(bytes)", "STORAGE(%)")
	if err := tw.Header(); err != nil {
		return err
	}

	for _, pvc := range pvcs {
		if err := tw.Row(
			pvc.Namespace,
			pvc.Name,
			formatStorageGi(pvc.Usage),
			fmt.Sprintf("%.2f%%", pvc.UsagePercent),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printDefenderMetrics(fightingRatio, maxLevel float64) error {
	tw := tablewriter.New(os.Stdout, "METRIC", "VALUE")
	if err := tw.Header(); err != nil {
		return err
	}
	if err := tw.Row("Fighting Ratio (1h)", formatDefenderFightingRatio(fightingRatio)); err != nil {
		return err
	}
	if err := tw.Row("Max Level (1h)", formatDefenderMaxLevel(maxLevel)); err != nil {
		return err
	}
	return tw.Flush()
}

func formatCPU(cores float64) string {
	return fmt.Sprintf("%.0fm", cores*1000)
}

func formatMemoryMi(bytes float64) string {
	const MiB = 1024 * 1024
	return fmt.Sprintf("%.0fMi", bytes/MiB)
}

func formatStorageGi(bytes float64) string {
	const GiB = 1024 * 1024 * 1024
	return fmt.Sprintf("%.0fGi", bytes/GiB)
}

func formatPct(used, alloc float64) string {
	if alloc <= 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", (used/alloc)*100)
}

func formatDefenderFightingRatio(ratio float64) string {
	return fmt.Sprintf("%.2f%%", ratio*100)
}

func formatDefenderMaxLevel(level float64) string {
	return fmt.Sprintf("%.0f", level)
}
