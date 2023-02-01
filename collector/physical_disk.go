//go:build windows
// +build windows

// Notes:
// - Windows PDH PhysicalDisk metrics report rates. The Prometheus convention
//   is for rates to be calculated. Gauges were implemented for rate metrics.
// - Windows counters cite "during the sample interval". Is this something we can/should manipulate?

// From https://learn.microsoft.com/en-us/windows/win32/api/pdh/nf-pdh-pdhaddenglishcountera

// Note: If the counter path contains a wildcard character, the non-wildcard
// portions of the path will be localized, but wildcards will not be expanded
// before adding the localized counter path to the query. In this case, you
// will need use the following procedure to add all matching counter names to
// the query.
//     - Make a query
//     - Use PdhAddEnglishCounter with the string containing wildcards
//     - Use PdhGetCounterInfo on the counter handle returned by
//       PdhAddEnglishCounter to get a localized full path (szFullPath.) This
//       string still contains wildcards, but the non-wildcard parts are now
//       localized.
//     - Use PdhExpandWildCardPath to expand the wildcards.
//     - Use PdhAddCounter on each of the resulting paths

package collector

import (
	"fmt"
	"github.com/cbwest3-ntnx/win"
	"github.com/prometheus-community/windows_exporter/headers/pdh"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

func init() {
	registerCollector("physical_disk", NewPhysicalDiskCollector, "PhysicalDisk")
}

var (
	nullPtr       *uint16
)

// A PhysicalDiskCollector is a Prometheus collector for PhysicalDisk metrics gathered with PDH.
type PhysicalDiskCollector struct {
	PromMetrics []*PrometheusMetricMap
	PdhQuery    *win.PDH_HQUERY
}

// Map a single Prometheus metric, e.g. read_latency_seconds_total, to one or
// more Windows PDH counters.
type PrometheusMetricMap struct {
	PdhCounterType uint32
	PdhPath        string // PDH string used to enumerate PDH counters (can include wildcards).
	PdhMetrics     []*PdhMetricMap
	PromDesc       *prometheus.Desc
	PromValueType  prometheus.ValueType
}

type PdhMetricMap struct {
	CounterHandle win.PDH_HCOUNTER
	DiskNumber    string
}

// NewPhysicalDiskCollector ...
func NewPhysicalDiskCollector() (Collector, error) {
	const subsystem = "physical_disk"
	var queryHandle win.PDH_HQUERY
	if ret := win.PdhOpenQuery(0, 0, &queryHandle); ret != 0 {
		fmt.Printf("ERROR: PdhOpenQuery return code is 0x%X\n", ret)
	}
	var pdc = PhysicalDiskCollector{PdhQuery: &queryHandle}

	// Queue length.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\current disk queue length",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_queue_length"),
			"Current Disk Queue Length is the number of requests outstanding on the disk at the time the performance data is collected. It also includes requests in service at the time of the collection. This is a instantaneous snapshot, not an average over the time interval. Multi-spindle disk devices can have multiple requests that are active at one time, but other concurrent requests are awaiting service. This counter might reflect a transitory high or low queue length, but if there is a sustained load on the disk drive, it is likely that this will be consistently high. Requests experience delays proportional to the length of this queue minus the number of spindles on the disks. For good performance, this difference should average less than two.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk queue length",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "average_queue_length"),
			"Average number of both read and write requests that were queued for the selected disk during the sample interval.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk read queue length",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_average_queue_length"),
			"Average number of read requests that were queued for the selected disk during the sample interval.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk write queue length",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_average_queue_length"),
			"Average number of write requests that were queued for the selected disk during the sample interval.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Device utilization.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\% idle time",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_idle_percent"),
			"Percentage of time during the sample interval that the disk was idle.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\% disk read time",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_read_percent"),
			"Percentage of elapsed time that the selected disk drive was busy servicing read requests.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\% disk write time",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_write_percent"),
			"Percentage of elapsed time that the selected disk drive was busy servicing write requests.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\% disk time",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_busy_percent"),
			"Percentage of elapsed time that the selected disk drive was busy servicing read or write requests.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Latency.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk sec/read",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_latency_average_seconds"),
			"Average time, in seconds, of a read of data from the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk sec/transfer",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transfer_latency_average_seconds"),
			"Time, in seconds, of the average disk transfer.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk sec/write",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_latency_average_seconds"),
			"Average time, in seconds, of a write of data to the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Ops.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk reads/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reads_per_second"),
			"Rate of read operations on the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\split io/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "split_io_per_second"),
			"Rate at which I/Os to the disk were split into multiple I/Os. A split I/O may result from requesting data of a size that is too large to fit into a single I/O or that the disk is fragmented.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk transfers/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transfers_per_second"),
			"Rate of read and write operations on the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk writes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "writes_per_second"),
			"Rate of write operations on the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Op sizes.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk bytes/read",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_average_bytes"),
			"Average number of bytes transferred from the disk during read operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk bytes/write",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_average_bytes"),
			"Average number of bytes transferred to the disk during write operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk bytes/transfer",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "transfer_average_bytes"),
			"Average number of bytes transferred to or from the disk during write or read operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Throughput.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk bytes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_per_second"),
			"Rate bytes are transferred to or from the disk during write or read operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk read bytes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_bytes_per_second"),
			"Rate at which bytes are transferred from the disk during read operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk write bytes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_bytes_per_second"),
			"Rate at which bytes are transferred to the disk during write operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	var userData uintptr
	// Append expanded PDH counter to each metric. PDH instances become labels for Prometheus metrics.
	for _, metric := range pdc.PromMetrics {
		paths, instances, err := pdh.LocalizeAndExpandCounter(queryHandle, metric.PdhPath)
		if err != nil {
			fmt.Printf("ERROR: Failed to localize and expand wildcards for: %s", metric.PdhPath)
			continue
		}
		for index, path := range paths {
			var pdhCounterHandle win.PDH_HCOUNTER
			ret := win.PdhAddCounter(queryHandle, path, userData, &pdhCounterHandle)
			if ret != win.PDH_CSTATUS_VALID_DATA {
				fmt.Printf("ERROR: Failed to add expanded counter '%s': %s (0x%X)\n", path, win.PDHErrors[ret], ret)
				continue
			}

			// PhysicalDisk instances include disk number and optionally mounted drives, e.g. '1' or '1 C:'.
			// We only use the disk number as a label.
			diskNumber, _, _:= strings.Cut(instances[index], " ")
			var pdhMetric = PdhMetricMap{CounterHandle: pdhCounterHandle, DiskNumber: diskNumber}
			metric.PdhMetrics = append(metric.PdhMetrics, &pdhMetric)
		}
	}
	fmt.Printf("pdc.PromMetrics: %s\n", pdc.PromMetrics)

	// TODO (cbwest): Figure out where this should live.
	ret := win.PdhCollectQueryData(*pdc.PdhQuery)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: Initial PdhCollectQueryData return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	return &pdc, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *PhysicalDiskCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting physical_disk metrics:", desc, err)
		return err
	}
	return nil
}

func (c *PhysicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	// TODO (2023-01-19):
	// - Proper error handling.
	// - In exporter startup:
	//		- Create query.
	//		- Call PdhAddEnglishCounter with the string containing wildcards.
	//		- Use PdhGetCounterInfo on the counter handle.
	//		- Use PdhExpandWildCardPath to expand the wildcards.
	//		- Use PdhAddCounter on each of the resulting paths. Use the returned handle for lookups
	//      - Store returned handles in data structures associated with the Prometheus metric.

	// - In exporter Collect() function:
	// 		- Call PdhCollectData()

	// Extra credit:
	//		- Allow users to blacklist disks.
	//		- Be smart enough to query disks, and if any were added/removed, re-enumerate.

	ret := win.PdhCollectQueryData(*c.PdhQuery)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: First PdhCollectQueryData return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	for _, metric := range c.PromMetrics {
		//fmt.Printf("%s has CounterHandles: %s\n", metric.PromDesc, metric.PdhMetrics)
		for _, pdhMetric := range metric.PdhMetrics {
			var counter win.PDH_FMT_COUNTERVALUE_DOUBLE
			ret = win.PdhGetFormattedCounterValueDouble(pdhMetric.CounterHandle, &metric.PdhCounterType, &counter)
			if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
				fmt.Printf("ERROR: Second PdhGetFormattedCounterValueDouble return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
			}
			if counter.CStatus != win.PDH_CSTATUS_VALID_DATA { // Error checking
				fmt.Printf("ERROR: Second CStatus is %s (0x%X)\n", win.PDHErrors[counter.CStatus], counter.CStatus)
			}
			//fmt.Printf("metric.DiskNumber=%s, counter.DoubleValue=%f\n", pdhMetric.DiskNumber, counter.DoubleValue)
			ch <- prometheus.MustNewConstMetric(
				metric.PromDesc,
				metric.PromValueType,
				counter.DoubleValue,
				pdhMetric.DiskNumber,
			)
		}
	}
	return nil, nil
}
