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

package service

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/rabbitdns/rabbitdns/pkg/resource_interface"

	"github.com/pkg/errors"

	"github.com/miekg/dns"

	log "github.com/sirupsen/logrus"
)

type ServiceBackend interface {
}

type ServiceManager struct {
	services        map[string]*Config
	loading         map[string]bool
	using           map[string]map[string]bool
	monitorResource resource_interface.MonitorResourceInterface
	backends        []ServiceBackendInterface
}

func NewServiceManager(backends []ServiceBackendInterface, m resource_interface.MonitorResourceInterface) *ServiceManager {
	return &ServiceManager{
		services:        map[string]*Config{},
		loading:         map[string]bool{},
		using:           make(map[string]map[string]bool),
		monitorResource: m,
		backends:        backends,
	}
}

func (s *ServiceManager) GetResources(req *resource_interface.GetResourcesRequest) ([]dns.RR, error) {
	svc, exist := s.services[req.Name]
	if !exist {
		return nil, resource_interface.ErrNotDefineService
	}
	if svc.RRType != req.RRtype {
		return nil, resource_interface.ErrMismatchServiceRRtype
	}

	response, err := svc.Service.GetEndpoints(&GetEndpointsRequest{
		SourceIP: req.SourceIP,
		ECS:      req.ECS,
	})
	return response, err
}
func (s *ServiceManager) GetService(name string) (*Config, bool) {
	service, ok := s.services[name]
	return service, ok
}
func (s *ServiceManager) RegisterService(service string, zonename string) {
	s.using[service][zonename] = true
}
func (s *ServiceManager) UnRegisterService(service string, zonename string) {
	delete(s.using[service], zonename)
}
func (s *ServiceManager) addService(srv *ServiceInfo, r io.Reader) error {
	svc, err := LoadConfig(r, srv.LastModified)
	if err != nil {
		return errors.Wrap(err, "can't load service config")
	}

	for endpoint, monitor := range svc.Monitors {
		log.WithFields(log.Fields{
			"Type":    "lib/server/ServiceManager",
			"Func":    "addService",
			"monitor": monitor,
			"name":    srv.Name,
			"value":   endpoint.Value,
			"RRType":  endpoint.RRType,
		}).Debug("CheckRegisterMonitor")

		err = s.monitorResource.CheckRegisterMonitor(monitor, srv.Name, endpoint)
		if err != nil {
			return errors.Wrap(err, "can't register monitor")
		}
	}
	for endpoint, monitor := range svc.Monitors {
		log.WithFields(log.Fields{
			"Type":    "lib/server/ServiceManager",
			"Func":    "addService",
			"monitor": monitor,
			"name":    srv.Name,
			"value":   endpoint.Value,
			"RRType":  endpoint.RRType,
		}).Debug("RegisterMonitor")

		s.monitorResource.RegisterMonitor(monitor, srv.Name, endpoint)
	}

	s.services[srv.Name] = svc
	s.using[srv.Name] = make(map[string]bool)
	log.WithFields(log.Fields{
		"Type": "lib/server/ServiceManager",
		"Func": "addService",
		"name": srv.Name,
	}).Debug("Done create service")

	return nil
}
func (s *ServiceManager) Update() error {
	for k := range s.loading {
		s.loading[k] = false
	}
	for _, b := range s.backends {
		if list, err := b.ListServices(); err != nil {
			for _, r := range list.Items {
				s.loading[r.Name] = true
				if _, ok := s.services[r.Name]; !(ok && r.LastModified == s.services[r.Name].ModTime) {
					reader, err := b.GetService(r.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"Type":  "lib/server/LoadService",
							"Func":  "LoadService",
							"Error": err,
							"name":  r.Name,
						}).Warn("failed to read service config.")
						return errors.Wrap(err, "failed to read service config.")
					}

					if err := s.addService(r, reader); err != nil {
						log.WithFields(log.Fields{
							"Type":  "lib/server/MonitoringManager",
							"Func":  "LoadMonitor",
							"Error": err,
							"name":  r.Name,
						}).Warn("failed to read service config.")
						return errors.Wrap(err, "failed to read service config.")
					}
				}
			}
		}
	}
	return nil
}

func (s *ServiceManager) deleteServices() {
	for k, v := range s.loading {
		if v == false {
			name := strings.TrimSuffix(filepath.Base(k), ".yml")
			if len(s.using[name]) > 0 {
				log.WithFields(log.Fields{
					"Type":         "lib/server/ServiceManager",
					"Func":         "LoadServices",
					"service_name": name,
				}).Warn("faild to delete service. because zone use this service.")
				continue
			}
			for endpoint, monitor := range s.services[k].Monitors {
				s.monitorResource.UnRegisterMonitor(monitor, name, endpoint)
			}
			delete(s.using, name)
		}
	}
}
