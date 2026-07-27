package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/cmd/util/tablewriter"
	"github.com/zetaoss/zengine/goapp/tasks/stat/k8s"
)

type NodeMetric = k8s.NodeMetric
type PodMetric = k8s.PodMetric

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

	endpoint := ""
	nodepool := ""
	namespace := ""
	if cfg != nil {
		endpoint = cfg.API.K8sMetricsEndpoint
		nodepool = cfg.API.K8sMetricsNodepool
		namespace = cfg.API.K8sMetricsNamespace
	}
	if endpoint == "" {
		endpoint = os.Getenv("K8SMETRICS_ENDPOINT")
	}
	if endpoint == "" {
		return fmt.Errorf("missing K8SMETRICS_ENDPOINT")
	}
	if nodepool == "" {
		nodepool = os.Getenv("K8SMETRICS_NODEPOOL")
	}
	if nodepool == "" {
		return fmt.Errorf("missing K8SMETRICS_NODEPOOL")
	}
	if namespace == "" {
		namespace = os.Getenv("K8SMETRICS_NAMESPACE")
	}
	if namespace == "" {
		return fmt.Errorf("missing K8SMETRICS_NAMESPACE")
	}

	show := func() error {
		if *watch || *watchShort {
			fmt.Print("\033[H\033[J")
			_, _ = fmt.Printf("%s\n\n", time.Now().Format(time.RFC3339))
		}
		nodes, pods, err := fetchAndParseMetrics(endpoint, nodepool, namespace)
		if err != nil {
			return fmt.Errorf("failed to fetch metrics from %s: %w", endpoint, err)
		}
		if err := printNodeMetrics(nodes); err != nil {
			return err
		}
		_, _ = fmt.Println()
		return printPodMetrics(pods, namespace)
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

func fetchAndParseMetrics(endpoint, nodepool, namespace string) ([]NodeMetric, []PodMetric, error) {
	return k8s.FetchAndParseMetrics(context.Background(), endpoint, nodepool, namespace)
}

func parsePrometheusNodeMetrics(r io.Reader, nodepool string) ([]NodeMetric, error) {
	return k8s.ParsePrometheusNodeMetrics(r, nodepool)
}

func parsePrometheusPodMetrics(r io.Reader, namespace string) ([]PodMetric, error) {
	return k8s.ParsePrometheusPodMetrics(r, namespace)
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

func printPodMetrics(pods []PodMetric, namespace string) error {
	if len(pods) == 0 {
		if namespace == "" {
			namespace = "prod3"
		}
		_, _ = fmt.Printf("No %s pod metrics found.\n", namespace)
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

func formatCPU(cores float64) string {
	return fmt.Sprintf("%.0fm", cores*1000)
}

func formatMemoryMi(bytes float64) string {
	const MiB = 1024 * 1024
	return fmt.Sprintf("%.0fMi", bytes/MiB)
}

func formatPct(used, alloc float64) string {
	if alloc <= 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", (used/alloc)*100)
}
