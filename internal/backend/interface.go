package backend

import (
	"io"
)

type Backend interface {
	ListDomains() (*ZoneList, error)
	GetZone(string) (io.Reader, error)
	ListServices() (*ServiceList, error)
	GetService(string) (io.Reader, error)
	ListMonitors() (*MonitorList, error)
	GetMonitor(string) (io.Reader, error)
}

func NewZoneList() *ZoneList {
	return &ZoneList{
		Items: []Zone{},
	}
}
func NewServiceList() *ServiceList {
	return &ServiceList{
		Items: []Service{},
	}
}

func NewMonitorList() *MonitorList {
	return &MonitorList{
		Items: []Monitor{},
	}
}

type ZoneList struct {
	Items []Zone
}

type Zone struct {
	Name         string
	LastModified int64
}

type ServiceList struct {
	Items []Service
}

type Service struct {
	Name         string
	LastModified int64
}

type MonitorList struct {
	Items []Monitor
}

type Monitor struct {
	Name         string
	LastModified int64
}
