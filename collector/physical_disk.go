//go:build windows
// +build windows

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
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

func init() {
	registerCollector("physical_disk", NewPhysicalDiskCollector, "PhysicalDisk")
}

var (
	nullPtr       *uint16
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
			prometheus.BuildFQName(Namespace, subsystem, "queue_length"),
			"Current Disk Queue Length is the number of requests outstanding on the disk at the time the performance data is collected. It also includes requests in service at the time of the collection. This is a instantaneous snapshot, not an average over the time interval. Multi-spindle disk devices can have multiple requests that are active at one time, but other concurrent requests are awaiting service. This counter might reflect a transitory high or low queue length, but if there is a sustained load on the disk drive, it is likely that this will be consistently high. Requests experience delays proportional to the length of this queue minus the number of spindles on the disks. For good performance, this difference should average less than two.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Device utilization.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\% idle time",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "idle_seconds_total"),
			"Percentage of time during the sample interval that the disk was idle.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Latency.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk sec/read",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_latency_seconds_total"),
			"Average time, in seconds, of a read of data from the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\avg. disk sec/write",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_latency_seconds_total"),
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
			prometheus.BuildFQName(Namespace, subsystem, "reads_total"), // MISNOMER!
			"Rate of read operations on the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk writes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "writes_total"), // MISNOMER!
			"Rate of write operations on the disk.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	// Throughput.
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk read bytes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_bytes_total"), // MISNOMER!
			"Rate at which bytes are transferred from the disk during read operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})
	pdc.PromMetrics = append(pdc.PromMetrics, &PrometheusMetricMap{
		PdhCounterType: win.PDH_FMT_DOUBLE,
		PdhPath:        "\\physicaldisk(*)\\disk write bytes/sec",
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "write_bytes_total"), // MISNOMER!
			"Rate at which bytes are transferred to the disk during write operations.",
			[]string{"disk"},
			nil,
		),
		PromValueType: prometheus.GaugeValue})

	var userData uintptr
	// Append expanded PDH counter to each metric. PDH instances become labels for Prometheus metrics.
	for _, metric := range pdc.PromMetrics {
		paths, diskNumbers, err := localizeAndExpandCounter(queryHandle, metric.PdhPath)
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
			var pdhMetric = PdhMetricMap{CounterHandle: pdhCounterHandle, DiskNumber: diskNumbers[index]}
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

// This function should be reusable by all collectors.
// TODO (cbwest): Do proper error handling.
func localizeAndExpandCounter(pdhQuery win.PDH_HQUERY, path string) (paths []string, diskNumbers []string, err error) {
	var counterHandle win.PDH_HCOUNTER
	var ret = win.PdhAddEnglishCounter(pdhQuery, path, 0, &counterHandle)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: PdhAddEnglishCounter return code is %s (0x%X)\n",
			win.PDHErrors[ret], ret)
	}

	// Call PdhGetCounterInfo twice to get buffer size, per
	// https://learn.microsoft.com/en-us/windows/win32/api/pdh/nf-pdh-pdhgetcounterinfoa#remarks.
	var bufSize uint32 = 0
	var retrieveExplainText uint32 = 0
	ret = win.PdhGetCounterInfo(counterHandle, uintptr(retrieveExplainText), &bufSize, nil)
	if ret != win.PDH_MORE_DATA { // error checking
		fmt.Printf("ERROR: First PdhGetCounterInfo return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	var counterInfo win.PDH_COUNTER_INFO
	ret = win.PdhGetCounterInfo(counterHandle, uintptr(retrieveExplainText), &bufSize, &counterInfo)
	if ret != win.PDH_CSTATUS_VALID_DATA { // error checking
		fmt.Printf("ERROR: Second PdhGetCounterInfo return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	// Call PdhExpandWildCardPath twice, per
	// https://learn.microsoft.com/en-us/windows/win32/api/pdh/nf-pdh-pdhexpandwildcardpathha#remarks.
	var flags uint32 = 0
	var pathListLength uint32 = 0
	ret = win.PdhExpandWildCardPath(nullPtr, counterInfo.SzFullPath, nullPtr, &pathListLength, &flags)
	if ret != win.PDH_MORE_DATA { // error checking
		fmt.Printf("ERROR: First PdhExpandWildCardPath return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}
	if pathListLength < 1 {
		fmt.Printf("ERROR: SOMETHING IS WRONG. pathListLength < 1, is %d.\n", pathListLength)
	}

	expandedPathList := make([]uint16, pathListLength)
	ret = win.PdhExpandWildCardPath(nullPtr, counterInfo.SzFullPath, &expandedPathList[0], &pathListLength, &flags)
	if ret != win.PDH_CSTATUS_VALID_DATA { // error checking
		fmt.Printf("ERROR: Second PdhExpandWildCardPath return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	for i := 0; i < int(pathListLength); i += len(path) + 1 {
		path = win.UTF16PtrToString(&expandedPathList[i])
		if len(path) < 1 { // expandedPathList has two nulls at the end.
			continue
		}

		// Parse PDH instance from the expanded counter path.
		instanceStartIndex := strings.Index(path, "(")
		instanceEndIndex := strings.Index(path, ")")
		if instanceStartIndex < 0 || instanceEndIndex < 0 {
			fmt.Printf("Unable to parse PDH counter instance from '%s'", path)
			continue
		}
		instance := path[instanceStartIndex+1 : instanceEndIndex]

		if instance == "_Total" { // Skip the _Total instance. That is for users to compute.
			continue
		}

		// Parse disk number from the instance.
		diskNumber, _, _ := strings.Cut(instance, " ")
		fmt.Printf("instance='%s', diskNumber='%s'\n", instance, diskNumber)

		paths = append(paths, path)
		diskNumbers = append(diskNumbers, diskNumber)
	}
	return paths, diskNumbers, nil
}

func (c *PhysicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	// TODO (2023-01-19):
	// - Proper error handling.
	// - Windows counters cite "during the sample interval". Is this something we can/should manipulate?
	// - Windows reports rates. Prometheus wants rates to be calculated. Do we use a Gauge or Counter?
	// - In exporter startup:
	//		- Create query.
	//		- Call PdhAddEnglishCounter with the string containing wildcards.
	//		- Use PdhGetCounterInfo on the counter handle.
	//		- Use PdhExpandWildCardPath to expand the wildcards.
	//		- Use PdhAddCounter on each of the resulting paths. Use the returned handle for lookups
	//      - Store returned handles in data structures associated with the Prometheus metric.

	// - In exporter Collect() function:
	// 		- Call PdhCollectData()

	// - In collector Collect() function:
	//		- Iterate through Prometheus metrics, use PDH counter handle to retrieve metrics.
	//		- Perform necessary/minimal parsing to attach labels, etc.
	//		- Add metric to Promethus exporter.

	// Extra credit:
	//		- Allow users to blacklist disks.
	//		- Be smart enough to query disks, and if any were added/removed, re-enumerate.

	ret := win.PdhCollectQueryData(*c.PdhQuery)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: First PdhCollectQueryData return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	for _, metric := range c.PromMetrics {
		fmt.Printf("%s has CounterHandles: %s\n", metric.PromDesc, metric.PdhMetrics)
		for _, pdhMetric := range metric.PdhMetrics {
			var derp win.PDH_FMT_COUNTERVALUE_DOUBLE
			ret = win.PdhGetFormattedCounterValueDouble(pdhMetric.CounterHandle, &metric.PdhCounterType, &derp)
			if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
				fmt.Printf("ERROR: Second PdhGetFormattedCounterValueDouble return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
			}
			if derp.CStatus != win.PDH_CSTATUS_VALID_DATA { // Error checking
				fmt.Printf("ERROR: Second CStatus is %s (0x%X)\n", win.PDHErrors[derp.CStatus], derp.CStatus)
			}
			fmt.Printf("metric.DiskNumber=%s, derp.DoubleValue=%f\n", pdhMetric.DiskNumber, derp.DoubleValue)

			ch <- prometheus.MustNewConstMetric(
				metric.PromDesc,
				metric.PromValueType,
				derp.DoubleValue,
				pdhMetric.DiskNumber,
			)
		}
	}
	return nil, nil
}
