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

package server

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rabbitdns/rabbitdns/pkg/config"
	"github.com/rabbitdns/rabbitdns/pkg/monitor"
	"github.com/rabbitdns/rabbitdns/pkg/service"
	"github.com/rabbitdns/rabbitdns/pkg/worker"
	"github.com/rabbitdns/rabbitdns/pkg/zone"

	log "github.com/sirupsen/logrus"
)

var (
	ErrReloadError = errors.New("reload error")
)

type Controller struct {
	workers []*worker.Worker
	config  *config.Config

	MonitoringManager *monitor.MonitoringManager
	ServiceManager    *service.ServiceManager
	ZoneManager       *zone.ZoneManager
	mutex             sync.Mutex
}

func NewController(c *config.Config) *Master {
	m := Controller{config: c}
	return &m
}

func (m *Controller) StartServ(ctx context.Context) error {
	m.MonitoringManager = manager.NewMonitoringManager(m.config)
	m.ServiceManager = manager.NewServiceManager(m.config, m.MonitoringManager)
	m.ZoneManager = manager.NewZoneManager(m.config, m.ServiceManager)

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load monitoring config")

	if err := m.MonitoringManager.Update(); err != nil {
		return errors.Wrap(err, "can't load monitor config")
	}

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load service config")

	if err := m.ServiceManager.Update(); err != nil {
		return errors.Wrap(err, "can't load service config")
	}

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load zone data")

	if err := m.ZoneManager.Update(); err != nil {
		return errors.Wrap(err, "can't load zone config")
	}
	protocols := []string{"tcp", "udp"}
	for _, addr := range m.config.DNSListeners {
		for _, proto := range protocols {
			m.workers = append(m.workers, NewWorker(m.config, m.ZoneManager, m.ServiceManager, addr, proto))
		}
	}

	for _, worker := range m.workers {
		worker.Run()
	}
	return nil
}
func (m *Controller) updateConfig(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mutex.Lock()
			m.MonitoringManager.Update()
			m.ServiceManager.Update()
			m.ZoneManager.Update()
			m.ZoneManager.Delete()
			m.ServiceManager.Delete()
			m.MonitoringManager.Delete()
			m.mutex.Unlock()
		}
	}
}
