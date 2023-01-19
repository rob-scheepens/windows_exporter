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

// Map a single Prometheus metric, e.g. read_latency_seconds_total, to one or
// more Windows PDH counters.
type MetricMap struct {
	PdhPath        string
	PromDesc       *prometheus.Desc
	CounterHandles []win.PDH_HCOUNTER
}

// A PhysicalDiskCollector is a Prometheus collector for perflib PhysicalDisk metrics
type PhysicalDiskCollector struct {
	Metrics []MetricMap
	query   *win.PDH_HQUERY
}

// NewPhysicalDiskCollector ...
func NewPhysicalDiskCollector() (Collector, error) {
	const subsystem = "physical_disk"
	var handle win.PDH_HQUERY
	if ret := win.PdhOpenQuery(0, 0, &handle); ret != 0 {
		fmt.Printf("ERROR: PdhOpenQuery return code is 0x%X\n", ret)
	}
	var pdc = PhysicalDiskCollector{query: &handle}
	pdc.Metrics = append(pdc.Metrics, MetricMap{
		PromDesc: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "read_latency_seconds_total"),
			"Shows the average time, in seconds, of a read operation from the disk (PhysicalDisk.AvgDiskSecPerRead)",
			[]string{"disk"},
			nil,
		),
		PdhPath: "\\physicaldisk(*)\\avg. disk sec/read"})

	var userData uintptr
	// Append expanded PDH counter to each metric. PDH instances become labels for Prometheus metrics.
	for _, metric := range pdc.Metrics {
		paths, err := localizeAndExpandCounter(handle, metric.PdhPath)
		if err != nil {
			fmt.Printf("ERROR: Failed to localize and expand wildcards for: %s", metric.PdhPath)
			continue
		}
		for _, path := range paths {
			var counterHandle win.PDH_HCOUNTER
			ret := win.PdhAddCounter(handle, path, userData, &counterHandle)
			if ret != win.PDH_CSTATUS_VALID_DATA {
				fmt.Printf("ERROR: Failed to add expanded counter: %s", path)
				continue
			}
			metric.CounterHandles = append(metric.CounterHandles, counterHandle)
		}
		fmt.Printf("%s has paths: %s", metric.PromDesc, paths)
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
func localizeAndExpandCounter(query win.PDH_HQUERY, path string) ([]string, error) {
	var counterHandle win.PDH_HCOUNTER
	var ret = win.PdhAddEnglishCounter(query, path, 0, &counterHandle)
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

	var paths []string // avoid another var creation?
	for i := 0; i < int(pathListLength); i += len(path) + 1 {
		path = win.UTF16PtrToString(&expandedPathList[i])
		if len(path) < 1 { // expandedPathList has two nulls at the end.
			continue
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func (c *PhysicalDiskCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {

	// TODO (2023-01-19):
	// - In exporter startup:
	//		- Create query.
	//		- Call PdhAddEnglishCounter with the string containing wildcards.
	//		- Use PdhGetCounterInfo on the counter handle.
	//		- Use PdhExpandWildCardPath to expand the wildcards.
	//		- Use PdhAddCounter on each of the resulting paths. Use the returned handle for lookups
	//      - Store returned handles in data structures associated with the Prometheus metric.

	// - In exporter Collect() function:
	// 		- Call PdhColectData()

	// - In collector Collect() function:
	//		- Iterate through Prometheus metrics, use PDH counter handle to retrieve metrics.
	//		- Perform necessary/minimal parsing to attach labels, etc.
	//		- Add metric to Promethus exporter.

	// Extra credit:
	//		- Allow users to blacklist disks.
	//		- Be smart enough to query disks, and if any were added/removed, re-enumerate.

	ret := win.PdhCollectQueryData(*c.query)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: First PdhCollectQueryData return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	var counterHandle win.PDH_HCOUNTER
	var derp win.PDH_FMT_COUNTERVALUE_DOUBLE
	var format uint32 = win.PDH_FMT_DOUBLE
	ret = win.PdhGetFormattedCounterValueDouble(counterHandle, &format, &derp)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: First PdhGetFormattedCounterValueDouble return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}
	if derp.CStatus != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: First CStatus is %s (0x%X)\n", win.PDHErrors[derp.CStatus], derp.CStatus)
	}

	ret = win.PdhCollectQueryData(*c.query)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: Second PdhCollectQueryData return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}
	fmt.Printf("Collect return code is %s (0x%X)\n", win.PDHErrors[ret], ret) // return code will be ERROR_SUCCESS

	ret = win.PdhGetFormattedCounterValueDouble(counterHandle, &format, &derp)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: Second PdhGetFormattedCounterValueDouble return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}
	if derp.CStatus != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: Second CStatus is %s (0x%X)\n", win.PDHErrors[derp.CStatus], derp.CStatus)
	}
	fmt.Printf("derp.DoubleValue=%f\n", derp.DoubleValue)

	return nil, nil
}
