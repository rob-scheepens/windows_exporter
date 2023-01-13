module github.com/prometheus-community/windows_exporter

go 1.13

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.6
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d
	github.com/cbwest3-ntnx/win v0.0.0-20230113001944-5585edc28a14
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/dimchansky/utfbom v1.1.1
	github.com/go-kit/log v0.2.1
	github.com/go-ole/go-ole v1.2.6
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/leoluk/perflib_exporter v0.2.0
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/client_model v0.3.0
	github.com/prometheus/common v0.38.0
	github.com/prometheus/exporter-toolkit v0.8.2
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.1 // indirect
	golang.org/x/net v0.0.0-20221014081412-f15817d10f9b // indirect
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783 // indirect
	golang.org/x/sys v0.3.0
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/lxn/win => github.com/cbwest3-ntnx/win v0.0.0-20230113000535-d7d4144b21c0
