package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/pdh"
)

var (
	diskReadBytesDesc  = prometheus.NewDesc("disk_read_bytes", "Disk read bytes", []string{"drive"}, nil)
	diskWriteBytesDesc = prometheus.NewDesc("disk_write_bytes", "Disk write bytes", []string{"drive"}, nil)
	diskReadsDesc      = prometheus.NewDesc("disk_reads", "Disk reads", []string{"drive"}, nil)
	diskWritesDesc     = prometheus.NewDesc("disk_writes", "Disk writes", []string{"drive"}, nil)
)

type pdhCounter struct {
	query   windows.Handle
	counter windows.Handle
	path    string
}

func (c *pdhCounter) collect() (float64, error) {
	var value windows.PDH_FMT_COUNTERVALUE_DOUBLE
	if err := windows.PdhCollectQueryData(c.query); err != nil {
		return 0, err
	}
	if err := windows.PdhGetFormattedCounterValue(c.counter, windows.PDH_FMT_DOUBLE, nil, &value); err != nil {
		return 0, err
	}
	return value.DoubleValue, nil
}

type diskCollector struct {
	counters map[string]*pdhCounter
}

func (c *diskCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- diskReadBytesDesc
	ch <- diskWriteBytesDesc
	ch <- diskReadsDesc
	ch <- diskWritesDesc
}

func (c *diskCollector) Collect(ch chan<- prometheus.Metric) {
	for drive, counter := range c.counters {
		value, err := counter.collect()
		if err != nil {
			fmt.Printf("Error collecting counter %s: %s", counter.path, err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(diskReadBytesDesc, prometheus.GaugeValue, value, drive)
		ch <- prometheus.MustNewConstMetric(diskWriteBytesDesc, prometheus.GaugeValue, value, drive)
		ch <- prometheus.MustNewConstMetric(diskReadsDesc, prometheus.GaugeValue, value, drive)
		ch <- prometheus.MustNewConstMetric(diskWritesDesc, prometheus.GaugeValue, value, drive)
	}
}

func newDiskCollector() (*diskCollector, error) {
	var counter *pdhCounter
	var err error
	counters := make(map[string]*pdhCounter)

	for _, drive := range []string{"C:", "D:", "E:", "F:"} {
		counter, err = newPDHCounter(fmt.Sprintf("\\PhysicalDisk(%s)\\Disk Read Bytes/sec", drive))
		if err != nil {
			return nil, err
		}
		counters[drive] = counter

		counter, err = newPDHCounter(fmt.Sprintf("\\PhysicalDisk(%s)\\Disk Write Bytes/sec", drive))
		if err != nil {
			return nil, err
		}
		counters[drive] = counter

		counter, err = newPDHCounter(fmt.Sprintf("\\PhysicalDisk(%s)\\Disk Reads/sec", drive))
		if err != nil {
			return nil, err
		}
		counters[drive] = counter

		counter, err = newPDHCounter(fmt.Sprintf("\\PhysicalDisk(%s)\\Disk Writes/sec", drive))
		if err != nil {
			return nil, err
		}
		counters[drive] = counter
	}

	return &diskCollector{counters: counters}, nil
}

func newPDHCounter(path string) (*pdhCounter, error) {
	var err error
	query, err := windows.PdhOpenQuery(nil, 0)
	if err != nil {
		return nil, err
	}
	counter, err := windows.PdhAddCounter(query, windows.StringToUTF16Ptr(path), 0)
	if err != nil {
		windows.PdhCloseQuery(query)
		return nil, err
	}
	return &pdhCounter{query: query, counter: counter, path: path}, nil
}
