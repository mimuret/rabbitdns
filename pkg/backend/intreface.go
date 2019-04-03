package backend

import (
	"github.com/rabbitdns/rabbitdns/pkg/service"
	"github.com/rabbitdns/rabbitdns/pkg/zone"
)

type BackendInterface interface {
	zone.ZoneBackendInterface
	service.ServiceBackendInterface
}
