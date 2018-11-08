package backend

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/rabbitdns/rabbitdns/internal/config"
)

var ListOp = clientv3.OpOption.WithKeysOnly()

func NewEtcdBackend(c *config.Config) *EtcdBackend {
	var err error
	var tlsConfig *tls.Config
	if c.EtcdClientKey != "" && c.EtcdClientCert != "" && c.EtcdClientCA {
		tlsInfo := transport.TLSInfo{
			CertFile:      "/tmp/test-certs/test-name-1.pem",
			KeyFile:       "/tmp/test-certs/test-name-1-key.pem",
			TrustedCAFile: "/tmp/test-certs/trusted-ca.pem",
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	return &EtcdBackend{
		Config: &clientv3.Config{
			Endpoints:   c.EtcdServers,
			DialTimeout: 2 * time.Second,
		},
		Prefix: c.EtcdPrefix,
		TLS:    tlsConfig,
	}
}

type EtcdBackend struct {
	Config *clientv3.Config
}

func (f *EtcdBackend) ListZones() (*ZoneList, error) {
	client, err := clientv3.New(f.Config)
	if err != nil {
		return nil, err
	}
	kv := clientv3.NewKV(client)
	res, err := kv.Get(context.TODO(), "/"+c.Prefix+"/zones", ListOp)
	if err != nil {
		return nil, err
	}
	list := NewZoneList()
	for _, ev := range resp.Kvs {
		ev.Version
		z := Zone{
			Name:         ev.Key,
			LastModified: ev.Version,
		}
		list.Items = append(list.Items, z)
	}
	return list, nil
}
func (f *EtcdBackend) GetZone(name string) (io.Reader, error) {
	return nil, nil
}
func (f *EtcdBackend) ListServices() (*ServiceList, error) {
	return nil, nil
}
func (f *EtcdBackend) GetService(name string) (io.Reader, error) {
	return nil, nil
}
func (f *EtcdBackend) ListMonitors() (*MonitorList, error) {
	return nil, nil
}
func (f *EtcdBackend) GetMonitor(name string) (io.Reader, error) {
	return nil, nil
}
