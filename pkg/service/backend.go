package service

import "io"

type ServiceBackendInterface interface {
	ListServices() (*ListServiceResponse, error)
	GetService(string) (io.Reader, error)
}

type ListServiceResponse struct {
	Items []*ServiceInfo
}

type ServiceInfo struct {
	Name         string
	LastModified int64
}
