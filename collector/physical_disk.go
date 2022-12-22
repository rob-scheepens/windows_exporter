//go:build windows
// +build windows

package collector

import (
	"fmt"
	"regexp"

	// "github.com/go-errors/errors"
	"github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("physical_disk", NewPhysicalDiskCollector, "PhysicalDisk")
}

var (
	diskWhitelist = kingpin.Flag(
		"collector.physical_disk.disk-whitelist",
		"Regexp of disks to whitelist. Disk name must both match whitelist and not match blacklist to be included.",
	).Default(".+").String()
	diskBlacklist = kingpin.Flag(
		"collector.physical_disk.disk-blacklist",
		"Regexp of disks to blacklist. Disk name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
)

// A PhysicalDiskCollector is a Prometheus collector for perflib PhysicalDisk metrics
type PhysicalDiskCollector struct {
	Metrics []MetricMap
	// RequestsQueued   *prometheus.Desc
	// ReadBytesTotal   *prometheus.Desc
	// ReadsTotal       *prometheus.Desc
	// WriteBytesTotal  *prometheus.Desc
	// WritesTotal      *prometheus.Desc
	// ReadTime         *prometheus.Desc
	// WriteTime        *prometheus.Desc
	// IdleTime         *prometheus.Desc
	// SplitIOs         *prometheus.Desc
	ReadLatency *prometheus.Desc
	// WriteLatency     *prometheus.Desc
	// ReadWriteLatency *prometheus.Desc

	diskWhitelistPattern *regexp.Regexp
	diskBlacklistPattern *regexp.Regexp
}

// NewPhysicalDiskCollector ...
func NewPhysicalDiskCollector() (Collector, error) {
	const subsystem = "physical_disk"

	// Make a bunch structs?
	// Put them in a list?
	// Submit PDH queries?
	// Iterate through results?

	// Static init, collect in a list:

	// ReadBytesTotal: prometheus.NewDesc(
	// 	prometheus.BuildFQName(Namespace, subsystem, "read_bytes_total"),
	// 	"The number of bytes transferred from the disk during read operations (PhysicalDisk.DiskReadBytesPerSec)",
	// 	[]string{"disk"},
	// 	nil,
	// ),

	// ReadsTotal: prometheus.NewDesc(
	// 	prometheus.BuildFQName(Namespace, subsystem, "reads_total"),
	// 	"The number of read operations on the disk (PhysicalDisk.DiskReadsPerSec)",
	// 	[]string{"disk"},
	// 	nil,
	// ),

	// For statically init'd list: return collectors:

	return &PhysicalDiskCollector{
		// RequestsQueued: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "requests_queued"),
		// 	"The number of requests queued to the disk (PhysicalDisk.CurrentDiskQueueLength)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// ReadBytesTotal: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "read_bytes_total"),
		// 	"The number of bytes transferred from the disk during read operations (PhysicalDisk.DiskReadBytesPerSec)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// ReadsTotal: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "reads_total"),
		// 	"The number of read operations on the disk (PhysicalDisk.DiskReadsPerSec)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// WriteBytesTotal: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "write_bytes_total"),
		// 	"The number of bytes transferred to the disk during write operations (PhysicalDisk.DiskWriteBytesPerSec)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// WritesTotal: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "writes_total"),
		// 	"The number of write operations on the disk (PhysicalDisk.DiskWritesPerSec)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// ReadTime: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "read_seconds_total"),
		// 	"Seconds that the disk was busy servicing read requests (PhysicalDisk.PercentDiskReadTime)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// WriteTime: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "write_seconds_total"),
		// 	"Seconds that the disk was busy servicing write requests (PhysicalDisk.PercentDiskWriteTime)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// IdleTime: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "idle_seconds_total"),
		// 	"Seconds that the disk was idle (PhysicalDisk.PercentIdleTime)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// SplitIOs: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "split_ios_total"),
		// 	"The number of I/Os to the disk were split into multiple I/Os (PhysicalDisk.SplitIOPerSec)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		ReadLatency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_latency_seconds_total"),
			"Shows the average time, in seconds, of a read operation from the disk (PhysicalDisk.AvgDiskSecPerRead)",
			[]string{"disk"},
			nil,
		),

		// WriteLatency: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "write_latency_seconds_total"),
		// 	"Shows the average time, in seconds, of a write operation to the disk (PhysicalDisk.AvgDiskSecPerWrite)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		// ReadWriteLatency: prometheus.NewDesc(
		// 	prometheus.BuildFQName(Namespace, subsystem, "read_write_latency_seconds_total"),
		// 	"Shows the time, in seconds, of the average disk transfer (PhysicalDisk.AvgDiskSecPerTransfer)",
		// 	[]string{"disk"},
		// 	nil,
		// ),

		diskWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *diskWhitelist)),
		diskBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *diskBlacklist)),
	}, nil
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

// Win32_PerfRawData_PerfDisk_PhysicalDisk docs:
// - https://docs.microsoft.com/en-us/previous-versions/aa394308(v=vs.85) - Win32_PerfRawData_PerfDisk_PhysicalDisk class
type PhysicalDisk struct {
	Name                   string
	CurrentDiskQueueLength float64 `perflib:"Current Disk Queue Length"`
	DiskReadBytesPerSec    float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec        float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec   float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec       float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime    float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime   float64 `perflib:"% Disk Write Time"`
	PercentIdleTime        float64 `perflib:"% Idle Time"`
	SplitIOPerSec          float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead      float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite     float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer  float64 `perflib:"Avg. Disk sec/Transfer"`
}

// Map Prometheus metrics to PDH query strings.
type PDHDiskMap struct {
	Name                   string
	CurrentDiskQueueLength float64 `perflib:"Current Disk Queue Length"`
	DiskReadBytesPerSec    float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec        float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec   float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec       float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime    float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime   float64 `perflib:"% Disk Write Time"`
	PercentIdleTime        float64 `perflib:"% Idle Time"`
	SplitIOPerSec          float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead      float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite     float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer  float64 `perflib:"Avg. Disk sec/Transfer"`
}

type MetricMap struct {
	PdhQuery         string
	PromMetricSuffix string
	PromHelp         string
}

func (c *PhysicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	// Old stuff:
	// var dst []PhysicalDisk
	// if err := unmarshalObject(ctx.perfObjects["PhysicalDisk"], &dst); err != nil {
	// 	return nil, err
	// }

	// Goals
	// Query PDH for specified counters for ALL disks in a system.
	// Extra credit: allow users to blacklist disks.

	// BEGIN: Imported test case to drive PDH query.
	pc, err := newPerfCounter(`\physicaldisk(*)\avg. disk sec/read`, true) // TODO (cbwest): check what 'true' does.
	if err != nil {
		log.Fatal(fmt.Sprintf("0x%X %s %s", err.(*win_perf_counters.PdhError).ErrorCode,
			win_perf_counters.PDHErrors[err.(*win_perf_counters.PdhError).ErrorCode], err))
	}

	// OLD HARD-CODED VALUE.
	// var vals [1]win_perf_counters.CounterValue
	// vals[0] = win_perf_counters.CounterValue{InstanceName: `\physicaldisk(1)\avg. disk sec/read`, Value: 42.0}

	var vals []win_perf_counters.CounterValue
	vals, err = pc.query.GetFormattedCounterArrayDouble(pc.handle)
	if err != nil {
		fmt.Println(vals)
		vals, err = pc.query.GetFormattedCounterArrayDouble(pc.handle)
		if err != nil {
			fmt.Println(vals)
			log.Fatal(fmt.Sprintf("0x%X %s %s", err.(*win_perf_counters.PdhError).ErrorCode,
				win_perf_counters.PDHErrors[err.(*win_perf_counters.PdhError).ErrorCode], err))
		}
	}

	err = pc.query.Close()
	if err != nil {
		log.Fatal(fmt.Sprintf("0x%X %s %s", err.(*win_perf_counters.PdhError).ErrorCode,
			win_perf_counters.PDHErrors[err.(*win_perf_counters.PdhError).ErrorCode], err))
	}
	// END: Imported test case to drive PDH query.

	// Rework this to allow disk blacklisting.
	// if disk.Name == "_Total" ||
	// 	c.diskBlacklistPattern.MatchString(disk.Name) ||
	// 	!c.diskWhitelistPattern.MatchString(disk.Name) {
	// 	continue
	// }

	// BAD! (nested loops)
	// for val in range vals {
	// 	for metric in metrics {
	// 		if val.InstanceName == metric.Name {
	// 		}
	// 	}
	// }

	for _, val := range vals {
		fmt.Println(`I found a value!`)
		// ch <- prometheus.MustNewConstMetric(
		// 	c.RequestsQueued,
		// 	prometheus.GaugeValue,
		// 	disk.CurrentDiskQueueLength,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.ReadBytesTotal,
		// 	prometheus.CounterValue,
		// 	disk.DiskReadBytesPerSec,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.ReadsTotal,
		// 	prometheus.CounterValue,
		// 	disk.DiskReadsPerSec,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.WriteBytesTotal,
		// 	prometheus.CounterValue,
		// 	disk.DiskWriteBytesPerSec,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.WritesTotal,
		// 	prometheus.CounterValue,
		// 	disk.DiskWritesPerSec,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.ReadTime,
		// 	prometheus.CounterValue,
		// 	disk.PercentDiskReadTime,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.WriteTime,
		// 	prometheus.CounterValue,
		// 	disk.PercentDiskWriteTime,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.IdleTime,
		// 	prometheus.CounterValue,
		// 	disk.PercentIdleTime,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.SplitIOs,
		// 	prometheus.CounterValue,
		// 	disk.SplitIOPerSec,
		// 	disk.Name,
		// )

		ch <- prometheus.MustNewConstMetric(
			c.ReadLatency,
			prometheus.CounterValue,
			val.Value.(float64),
			"disk1-parse-later",
		)

		// ch <- prometheus.MustNewConstMetric(
		// 	c.WriteLatency,
		// 	prometheus.CounterValue,
		// 	disk.AvgDiskSecPerWrite*ticksToSecondsScaleFactor,
		// 	disk.Name,
		// )

		// ch <- prometheus.MustNewConstMetric(
		// 	c.ReadWriteLatency,
		// 	prometheus.CounterValue,
		// 	disk.AvgDiskSecPerTransfer*ticksToSecondsScaleFactor,
		// 	disk.Name,
		// )
	}

	return nil, nil
}

// Taken from: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/0eff6a45c4cac90f93fad640ff0c5d63561a2a34/pkg/winperfcounters/watcher.go#L43
type perfCounter struct {
	path   string
	query  win_perf_counters.PerformanceQuery
	handle win_perf_counters.PDH_HCOUNTER
}

// Taken from: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/0eff6a45c4cac90f93fad640ff0c5d63561a2a34/pkg/winperfcounters/watcher.go#L68
// newPerfCounter returns a new performance counter for the specified descriptor.
func newPerfCounter(counterPath string, collectOnStartup bool) (*perfCounter, error) {
	query := &win_perf_counters.PerformanceQueryImpl{}
	err := query.Open()
	if err != nil {
		return nil, err
	}

	var handle win_perf_counters.PDH_HCOUNTER
	handle, err = query.AddEnglishCounterToQuery(counterPath)
	if err != nil {
		return nil, err
	}

	// Some perf counters (e.g. cpu) return the usage stats since the last measure.
	// We collect data on startup to avoid an invalid initial reading
	if collectOnStartup {
		err = query.CollectData()
		if err != nil {
			return nil, err
		}
	}

	counter := &perfCounter{
		path:   counterPath,
		query:  query,
		handle: handle,
	}

	return counter, nil
}
