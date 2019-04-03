package zone

import "io"

type ZoneBackendInterface interface {
	ListZones() (*ListZoneResponse, error)
	GetZone(string) (io.Reader, error)
}

type ListZoneResponse struct {
	Items []*ZoneInfo
}

type ZoneInfo struct {
	Name         string
	LastModified int64
}
