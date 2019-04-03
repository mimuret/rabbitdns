package backend

import (
	"io"
	"strings"
	"time"

	"github.com/rabbitdns/rabbitdns/pkg/service"
	"github.com/rabbitdns/rabbitdns/pkg/zone"
)

func NewTestBackend(baseDir string) BackendInterface {
	return &TestBackend{}
}

type TestBackend struct {
}

func (t *TestBackend) ListZones() (*zone.ListZoneResponse, error) {
	return &zone.ListZoneResponse{
		Items: []*zone.ZoneInfo{
			{
				Name:         "example.jp",
				LastModified: time.Now().Unix(),
			},
		},
	}, nil
}
func (t *TestBackend) GetZone(name string) (io.Reader, error) {
	return strings.NewReader(`$ORIGIN example.jp.
$TTL 3600
@ IN SOA ns.example.jp. root.example.jp. 1 3600 900 86400 900
@ IN NS ns.example.jp.
ns IN A 127.0.0.1
www IN DYNA service1
`), nil
}
func (t *TestBackend) ListServices() (*service.ListServiceResponse, error) {
	return &service.ListServiceResponse{
		Items: []*service.ServiceInfo{
			{
				Name:         "service1",
				LastModified: time.Now().Unix(),
			},
		},
	}, nil
}
func (t *TestBackend) GetService(name string) (io.Reader, error) {
	return strings.NewReader(`apiVersion: rabbitdns.dev/v1
kind: ServiceEndpoint
metadata:
	name: service1
spec:
	rrtype: A
	rdatas:
		- name: host1
			value: 192.168.0.1
		  livenessProbe:
			  httpGet:
					scheme: HTTP
					path: /health
					port: 8080
		- name: host1
		  value: 192.168.80.1
`), nil
}
