// +build windows

package collector

import (
	"fmt"
	"regexp"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/lxn/win"
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
	ReadLatency      *prometheus.Desc
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
	PdhQuery                   string
	PromMetricSuffix string
	PromHelp		string
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


	// BEGIN: golang.org/x/sys/windows APPROACH:
	var handle win.PDH_HQUERY
	var counterHandle win.PDH_HCOUNTER 

	ret := win.PdhOpenQuery(0, 0, &handle)
	if ret != win.PDH_CSTATUS_VALID_DATA {  // Error checking
		fmt.Printf("ERROR: PdhOpenQuery return code is %X\n", ret)
	}

	ret = win.PdhAddEnglishCounter(handle, "\\physicaldisk(*)\\avg. disk sec/read", 0, &counterHandle)
	if ret != win.PDH_CSTATUS_VALID_DATA {  // Error checking
		fmt.Printf("ERROR: PdhAddEnglishCounter return code is %X\n", ret)
	}


	ret = win.PdhCollectQueryData(handle)
	if ret != win.PDH_CSTATUS_VALID_DATA {  // Error checking
		fmt.Printf("ERROR: First PdhCollectQueryData return code is %X\n", ret)
	}


	var derp win.PDH_FMT_COUNTERVALUE_DOUBLE
	var zero uint32 = 0  // TODO (cbwest): Figure out what this argument does.
	ret = win.PdhGetFormattedCounterValueDouble(counterHandle, &zero, &derp)
	if ret != win.PDH_CSTATUS_VALID_DATA {  // Error checking
		fmt.Printf("ERROR: First PdhGetFormattedCounterValueDouble return code is %X\n", ret)
	}
	if derp.CStatus != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: First CStatus is %s (%X)\n", derp.CStatus, derp.CStatus)
	}

	ret = win.PdhCollectQueryData(handle)
	if ret != win.PDH_CSTATUS_VALID_DATA {  // Error checking
		fmt.Printf("ERROR: Second PdhCollectQueryData return code is %X\n", ret)
	}
	fmt.Printf("Collect return code is %X\n", ret) // return code will be ERROR_SUCCESS

	ret = win.PdhGetFormattedCounterValueDouble(counterHandle, &zero, &derp)
	if ret != win.PDH_CSTATUS_VALID_DATA {  // Error checking
		fmt.Printf("ERROR: Second PdhGetFormattedCounterValueDouble return code is %X\n", ret)
	}
	if derp.CStatus != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: Second CStatus is %s (%X)\n", derp.CStatus, derp.CStatus)
	}
	fmt.Printf("derp.DoubleValue=%f\n", derp.DoubleValue)

	// END: golang.org/x/sys/windows APPROACH:


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

	// for _, val := range vals {
	// 	fmt.Println(`I found a value!`)
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

		// ch <- prometheus.MustNewConstMetric(
		// 	c.ReadLatency,
		// 	prometheus.CounterValue,
		// 	val.Value.(float64),
		// 	"disk1-parse-later",
		// )

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
	// }

	return nil, nil
}
