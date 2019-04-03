package backend

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"log"
	"time"

	"github.com/rabbitdns/rabbitdns/pkg/monitor"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/rabbitdns/rabbitdns/pkg/service"
	"github.com/rabbitdns/rabbitdns/pkg/zone"
)

var ListOp = clientv3.WithKeysOnly()

func NewEtcdBackend(servers []string, prefix, key, cert, caFile string) BackendInterface {
	var err error
	var tlsConfig *tls.Config
	if key != "" && cert != "" && caFile != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      cert,
			KeyFile:       key,
			TrustedCAFile: caFile,
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	return &EtcdBackend{
		Config: clientv3.Config{
			Endpoints:   servers,
			DialTimeout: 2 * time.Second,
			TLS:         tlsConfig,
		},
		Prefix: prefix,
	}
}

type EtcdBackend struct {
	Config clientv3.Config
	Prefix string
}

func (f EtcdBackend) get(key string) (io.Reader, error) {
	client, err := clientv3.New(f.Config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), key, ListOp)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) < 0 {
		return nil, ErrEmpty
	}
	return bytes.NewReader(resp.Kvs[0].Value), nil
}

func (f *EtcdBackend) ListZones() (*zone.ListZoneResponse, error) {
	client, err := clientv3.New(f.Config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/"+f.Prefix+"/zones", ListOp)
	if err != nil {
		return nil, err
	}
	list := &zone.ListZoneResponse{
		Items: []*zone.ZoneInfo{},
	}
	for _, ev := range resp.Kvs {
		z := &zone.ZoneInfo{
			Name:         string(ev.Key),
			LastModified: ev.Version,
		}
		list.Items = append(list.Items, z)
	}
	return list, nil
}
func (f *EtcdBackend) GetZone(name string) (io.Reader, error) {
	return f.get("/" + f.Prefix + "/services/" + name)
}
func (f *EtcdBackend) ListServices() (*service.ListServiceResponse, error) {
	client, err := clientv3.New(f.Config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/"+f.Prefix+"/services", ListOp)
	if err != nil {
		return nil, err
	}
	list := &service.ListServiceResponse{
		Items: []*service.ServiceInfo{},
	}
	for _, ev := range resp.Kvs {
		s := &service.ServiceInfo{
			Name:         string(ev.Key),
			LastModified: ev.Version,
		}
		list.Items = append(list.Items, s)
	}
	return list, nil
}

func (f *EtcdBackend) GetService(name string) (io.Reader, error) {
	return f.get("/" + f.Prefix + "/services/" + name)
}

func (f *EtcdBackend) ListMonitors() (*monitor.ListMonitorsResponse, error) {

	client, err := clientv3.New(f.Config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/"+f.Prefix+"/monitors", ListOp)
	if err != nil {
		return nil, err
	}
	list := &monitor.ListMonitorsResponse{
		Items: []*monitor.MonitorInfo{},
	}
	for _, ev := range resp.Kvs {
		m := &monitor.MonitorInfo{
			Name:         string(ev.Key),
			LastModified: ev.Version,
		}
		list.Items = append(list.Items, m)
	}
	return list, nil
}

func (f *EtcdBackend) GetMonitor(name string) (io.Reader, error) {
	return f.get("/" + f.Prefix + "/services/" + name)
}
