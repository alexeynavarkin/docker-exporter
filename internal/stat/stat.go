package stat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/alexeynavarkin/docker_exporter/internal/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
)

type response struct {
	NetworkStats map[string]types.NetworkStats `json:"networks"`
	MemStats     types.MemoryStats             `json:"memory_stats"`
	CpuStats     types.CPUStats                `json:"cpu_stats"`
}

var (
	labelNames = []string{"containerName", "serviceName", "serviceID", "type"}

	descCPU = prometheus.NewDesc(
		"docker_cpu_usage_ns_total",
		"Cumulative container CPU usage in nanoseconds.",
		labelNames, nil,
	)
	descMem = prometheus.NewDesc(
		"docker_mem_usage_bytes",
		"Container memory usage in bytes.",
		labelNames, nil,
	)
	descNet = prometheus.NewDesc(
		"docker_net_bytes_total",
		"Cumulative container network I/O in bytes.",
		labelNames, nil,
	)
)

type Gatherer struct {
	cli *client.Client
}

func NewGatherer(cli *client.Client) *Gatherer {
	return &Gatherer{cli: cli}
}

func (g *Gatherer) Describe(ch chan<- *prometheus.Desc) {
	ch <- descCPU
	ch <- descMem
	ch <- descNet
}

func (g *Gatherer) Collect(ch chan<- prometheus.Metric) {
	containers, err := g.cli.ContainerList(
		context.Background(),
		container.ListOptions{
			Filters: filters.NewArgs(filters.Arg("status", "running")),
		},
	)
	if err != nil {
		fmt.Printf("error get container list %v\n", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, c := range containers {
		go func(c types.Container) {
			defer wg.Done()
			g.collectContainer(c, ch)
		}(c)
	}
	wg.Wait()
}

func (g *Gatherer) collectContainer(c types.Container, ch chan<- prometheus.Metric) {
	stats, _ := g.cli.ContainerStats(context.Background(), c.ID, false)
	data, _ := io.ReadAll(stats.Body)
	var resp response
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	containerName := strings.Join(c.Names, "")
	svcName := util.GetMapValue(c.Labels, util.LabelNameServiceName, util.LabelDefaultValue)
	svcID := util.GetMapValue(c.Labels, util.LabelNameServiceID, util.LabelDefaultValue)

	label := func(t string) []string { return []string{containerName, svcName, svcID, t} }

	// Memory — gauge, current usage
	ch <- prometheus.MustNewConstMetric(descMem, prometheus.GaugeValue,
		float64(resp.MemStats.Usage), label("used")...)

	// CPU — counter, absolute cumulative nanoseconds from Docker
	ch <- prometheus.MustNewConstMetric(descCPU, prometheus.CounterValue,
		float64(resp.CpuStats.CPUUsage.UsageInUsermode), label("usermode")...)
	ch <- prometheus.MustNewConstMetric(descCPU, prometheus.CounterValue,
		float64(resp.CpuStats.CPUUsage.UsageInKernelmode), label("kernelmode")...)
	ch <- prometheus.MustNewConstMetric(descCPU, prometheus.CounterValue,
		float64(resp.CpuStats.ThrottlingData.ThrottledTime), label("throttled")...)

	// Network — counter, sum all interfaces, absolute cumulative bytes from Docker
	var rx, tx, drop, errs uint64
	for _, iface := range resp.NetworkStats {
		rx += iface.RxBytes
		tx += iface.TxBytes
		drop += iface.RxDropped + iface.TxDropped
		errs += iface.RxErrors + iface.TxErrors
	}
	ch <- prometheus.MustNewConstMetric(descNet, prometheus.CounterValue, float64(rx), label("rx")...)
	ch <- prometheus.MustNewConstMetric(descNet, prometheus.CounterValue, float64(tx), label("tx")...)
	ch <- prometheus.MustNewConstMetric(descNet, prometheus.CounterValue, float64(drop), label("drop")...)
	ch <- prometheus.MustNewConstMetric(descNet, prometheus.CounterValue, float64(errs), label("error")...)
}
