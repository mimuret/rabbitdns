// Copyright (C) 2018 Manabu Sonoda.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitor

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rabbitdns/rabbitdns/pkg/resource_interface"

	log "github.com/sirupsen/logrus"
)

var (
	ErrNotDefineMonitor = errors.New("Not define monitor")
	ErrEmptyState       = errors.New("empty state")
)

type MonitoringManager struct {
	loading  map[string]bool
	monitors map[string]*Config
	entries  map[string]map[string]map[string]*Entry
	states   map[string]map[string]map[string]bool
	backends []MonitorBackendInterface
}

func NewMonitoringManager(backends []MonitorBackendInterface) *MonitoringManager {
	return &MonitoringManager{
		loading:  map[string]bool{},
		monitors: map[string]*Config{},
		entries:  map[string]map[string]map[string]*Entry{},
		states:   map[string]map[string]map[string]bool{},
		backends: backends,
	}
}

func (m *MonitoringManager) GetMonitors() []map[string]string {
	results := []map[string]string{}
	for name, _ := range m.monitors {
		result := map[string]string{
			"name": name,
		}
		results = append(results, result)
	}
	return results
}

func (m *MonitoringManager) CheckRegisterMonitor(monitorName, service string, endpoint resource_interface.Endpoint) error {
	mon, ok := m.monitors[monitorName]
	if ok == false {
		return ErrNotDefineMonitor
	}
	entry := &Entry{Value: endpoint.GetValue(), RRtype: endpoint.GetRRtype()}

	err := mon.CheckRegister(entry)
	if err != nil {
		return err
	}
	return nil
}
func (m *MonitoringManager) RegisterMonitor(monitorName, service string, e resource_interface.Endpoint) error {
	mon, ok := m.monitors[monitorName]
	path := e.GetPath()
	if ok == false {
		return ErrNotDefineMonitor
	}
	if m.entries[monitorName][service] == nil {
		m.entries[monitorName][service] = make(map[string]*Entry)
	}

	entry := NewEntry(service, mon, e)
	var status bool
	var err error
	if current, ok := m.entries[monitorName][service][path]; ok {
		status = current.getStatus()
		current.CancelFunc()
		log.WithFields(log.Fields{
			"Type":          "lib/server/MonitoringManager",
			"Func":          "RegisterMonitor",
			"monitor_name":  monitorName,
			"service_name":  service,
			"endpoint_path": path,
			"status":        status,
		}).Debug("stop current monitor")

	} else {
		status, err = m.getState(monitorName, service, path)
		if err != nil {
			status = entry.Monitor(context.Background())
		}
	}
	entry.initStatus(status)
	m.entries[monitorName][service][path] = entry
	log.WithFields(log.Fields{
		"Type":            "lib/server/MonitoringManager",
		"Func":            "RegisterMonitor",
		"monitor_name":    monitorName,
		"service_name":    service,
		"endpoint_path":   e.GetPath(),
		"endpoint_value":  e.GetValue(),
		"endpoint_rrtype": e.GetRRtype(),
	}).Debug("start monitoring")

	go entry.Start(context.Background())
	return nil
}
func (m *MonitoringManager) UnRegisterMonitor(monitorName, service string, e resource_interface.Endpoint) {
	if m.entries[monitorName][service] != nil {
		if entry, exist := m.entries[monitorName][service][e.GetPath()]; exist {
			entry.Stop()
			delete(m.entries[monitorName][service], e.GetPath())
		}
		if len(m.entries[monitorName][service]) == 0 {
			delete(m.entries[monitorName], service)
		}
	}
}

func (m *MonitoringManager) addMonitor(bmon *MonitorInfo, r io.Reader) error {
	mon, err := LoadConfig(r, bmon.LastModified)
	if err != nil {
		return err
	}
	m.monitors[bmon.Name] = mon
	m.entries[bmon.Name] = make(map[string]map[string]*Entry)

	return nil
}
func (m *MonitoringManager) getState(monitorName, serviceName, path string) (bool, error) {
	if mmap, ok := m.states[serviceName]; ok {
		if smap, ok := mmap[serviceName]; ok {
			if state, ok := smap[path]; ok {
				return state, nil
			}
		}
	}
	return false, ErrEmptyState
}
func (m *MonitoringManager) saveStates() {
	for monitorName, _ := range m.entries {
		m.states[monitorName] = make(map[string]map[string]bool)
		for serviceName, _ := range m.entries[monitorName] {
			m.states[monitorName][serviceName] = make(map[string]bool)
			for path, entry := range m.entries[monitorName][serviceName] {
				m.states[monitorName][serviceName][path] = entry.getStatus()
			}
		}
	}
}
func (m *MonitoringManager) loadMonitors() error {
	for k := range m.loading {
		m.loading[k] = false
	}
	for _, b := range m.backends {
		if list, err := b.ListMonitors(); err != nil {
			for _, bmon := range list.Items {
				m.loading[bmon.Name] = true

				if _, ok := m.monitors[bmon.Name]; !(ok && bmon.LastModified == m.monitors[bmon.Name].ModTime) {
					r, err := b.GetMonitor(bmon.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"Type":  "lib/server/MonitoringManager",
							"Func":  "LoadMonitor",
							"Error": err,
							"name":  bmon.Name,
						}).Warn(ErrReadMonitor)
						return ErrReadMonitor
					}
					if err := m.addMonitor(bmon, r); err != nil {
						log.WithFields(log.Fields{
							"Type":  "lib/server/MonitoringManager",
							"Func":  "LoadMonitor",
							"Error": err,
							"name":  bmon.Name,
						}).Warn(ErrReadMonitor)
						return ErrReadMonitor
					}
				}
			}
		}
	}
	return nil
}

func (m *MonitoringManager) deleteMonitors() {
	for k, v := range m.loading {
		if v == false {
			name := strings.TrimSuffix(filepath.Base(k), ".yml")
			if len(m.entries[name]) > 0 {
				log.WithFields(log.Fields{
					"Type":         "lib/server/ServiceManager",
					"Func":         "LoadServices",
					"service_name": name,
				}).Warn("faild to delete this monitor. because zone use this service.")
				continue
			}
			delete(m.entries, name)
			delete(m.monitors, name)
		}
	}
}
