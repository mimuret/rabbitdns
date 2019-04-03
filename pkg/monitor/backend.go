package monitor

import "io"

type MonitorBackendInterface interface {
	ListMonitors() (*ListMonitorsResponse, error)
	GetMonitor(string) (io.Reader, error)
}

type ListMonitorsResponse struct {
	Items []*MonitorInfo
}

type MonitorInfo struct {
	Name         string
	LastModified int64
}
