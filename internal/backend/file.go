package backend

import (
	"io"
	"os"
	"path/filepath"
)

func NewFileBackend(baseDir string) *FileBackend {
	return &FileBackend{
		BaseDir: baseDir,
	}
}

type FileBackend struct {
	BaseDir string
}

func (f *FileBackend) ListZones() (*ZoneList, error) {
	list := NewZoneList()
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

		z := Zone{
			Name:         origin,
			LastModified: stat.ModTime().Unix(),
		}
		list.Items = append(list.Items, z)
	}
	return list, nil
}
func (f *FileBackend) GetZone(name string) (io.Reader, error) {
	path := f.BaseDir + "/zones/" + name
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (f *FileBackend) ListServices() (*ServiceList, error) {
	list := NewServiceList()
	matches, err := filepath.Glob(f.BaseDir + "/services/*.yml")
	if err != nil {
		return nil, err
	}
	for _, f := range matches {
		stat, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		origin := filepath.Base(f)

		s := Service{
			Name:         origin,
			LastModified: stat.ModTime().Unix(),
		}
		list.Items = append(list.Items, s)

	}
	return list, nil
}
func (f *FileBackend) GetService(name string) (io.Reader, error) {
	path := f.BaseDir + "/services/" + name + ".yml"
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (f *FileBackend) ListMonitors() (*MonitorList, error) {
	list := NewMonitorList()
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

		m := Monitor{
			Name:         origin,
			LastModified: stat.ModTime().Unix(),
		}
		list.Items = append(list.Items, m)

	}
	return list, nil

}
func (f *FileBackend) GetMonitor(name string) (io.Reader, error) {
	path := f.BaseDir + "/monitors/" + name + ".yml"
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}
