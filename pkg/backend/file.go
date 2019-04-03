package backend

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rabbitdns/rabbitdns/pkg/monitor"
	"github.com/rabbitdns/rabbitdns/pkg/service"
	"github.com/rabbitdns/rabbitdns/pkg/zone"
)

func NewFileBackend(baseDir string) BackendInterface {
	return &FileBackend{
		BaseDir: baseDir,
	}
}

type FileBackend struct {
	BaseDir string
}

func (f *FileBackend) get(path string) (io.Reader, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil

}
func (f *FileBackend) ListZones() (*zone.ListZoneResponse, error) {
	list := &zone.ListZoneResponse{
		Items: []*zone.ZoneInfo{},
	}
	matches, err := filepath.Glob(f.BaseDir + "/zones/*")
	if err != nil {
		return nil, err
	}
	for _, f := range matches {
		stat, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		origin := filepath.Base(f)

		z := &zone.ZoneInfo{
			Name:         origin,
			LastModified: stat.ModTime().Unix(),
		}
		list.Items = append(list.Items, z)
	}
	return list, nil
}
func (f *FileBackend) GetZone(name string) (io.Reader, error) {
	return f.get(f.BaseDir + "/zones/" + name)
}
func (f *FileBackend) ListServices() (*service.ListServiceResponse, error) {
	list := &service.ListServiceResponse{
		Items: []*service.ServiceInfo{},
	}
	matches, err := filepath.Glob(f.BaseDir + "/services/*.yaml")
	if err != nil {
		return nil, err
	}
	for _, f := range matches {
		stat, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		origin := filepath.Base(f)

		s := &service.ServiceInfo{
			Name:         origin,
			LastModified: stat.ModTime().Unix(),
		}
		list.Items = append(list.Items, s)

	}
	return list, nil
}
func (f *FileBackend) GetService(name string) (io.Reader, error) {
	return f.get(f.BaseDir + "/services/" + name + ".yaml")
}
func (f *FileBackend) ListMonitors() (*monitor.ListMonitorsResponse, error) {
	list := &monitor.ListMonitorsResponse{
		Items: []*monitor.MonitorInfo{},
	}
	matches, err := filepath.Glob(f.BaseDir + "/monitors/*")
	if err != nil {
		return nil, err
	}
	for _, f := range matches {
		stat, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		origin := filepath.Base(f)

		m := &monitor.MonitorInfo{
			Name:         origin,
			LastModified: stat.ModTime().Unix(),
		}
		list.Items = append(list.Items, m)

	}
	return list, nil

}
func (f *FileBackend) GetMonitor(name string) (io.Reader, error) {
	return f.get(f.BaseDir + "/monitors/" + name + ".yaml")
}
