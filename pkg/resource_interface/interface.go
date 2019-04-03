package resource_interface

import (
	"errors"
	"net"

	"github.com/miekg/dns"
)

type MonitorResourceInterface interface {
	CheckRegisterMonitor(string, string, Endpoint) error
	RegisterMonitor(string, string, Endpoint) error
	UnRegisterMonitor(string, string, Endpoint)
}
type Endpoint interface {
	GetPath() string
	GetValue() string
	GetRRtype() uint16
	SetStatus(bool)
}
type ServiceResourceInterface interface {
	GetResources(*GetResourcesRequest) ([]dns.RR, error)
	RegisterService(string, string)
	UnRegisterService(string, string)
}

var (
	ErrNotDefineService      = errors.New("Not define service.")
	ErrMismatchServiceRRtype = errors.New("rrtype mismatch.")
)

type GetResourcesRequest struct {
	Name     string
	RRtype   uint16
	SourceIP net.IP
	ECS      []net.IPNet
}
