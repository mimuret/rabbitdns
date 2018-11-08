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
	"github.com/rabbitdns/rabbitdns/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	LOAD_ERROR int = iota
	OK
)

var (
	ErrReloadError = errors.New("reload error")
)

type Master struct {
	workers []*worker
	config  *config.Config

	monitoringManager *monitoringManager
	serviceManager    *serviceManager
	zoneManager       *zoneManager
	mutex             sync.Mutex
}

func NewMaster(c *config.Config) *Master {
	m := Master{config: c}
	return &m
}

func (m *Master) StartServ(ctx context.Context) error {
	m.monitoringManager = NewMonitoringManager(m.config)
	m.serviceManager = NewServiceManager(m.config, m.monitoringManager)
	m.zoneManager = NewZoneManager(m.config, m.serviceManager)

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load monitoring config")

	if err := m.monitoringManager.LoadMonitors(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load service config")

	if err := m.serviceManager.LoadServices(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"Type": "lib/server/Master",
		"Func": "StartServ",
	}).Info("start to load zone data")

	if err := m.zoneManager.LoadZones(); err != nil {
		return err
	}
	protocols := []string{"tcp", "udp"}
	for _, addr := range m.config.DNSListeners {
		for _, proto := range protocols {
			m.workers = append(m.workers, NewWorker(m.config, m.zoneManager, m.serviceManager, addr, proto))
		}
	}

	for _, worker := range m.workers {
		worker.Run()
	}
	return nil
}
func (m *Master) updateConfig(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mutex.Lock()
			m.monitoringManager.LoadMonitors()
			m.serviceManager.LoadServices()
			m.zoneManager.LoadZones()
			m.zoneManager.DeleteZones()
			m.serviceManager.DeleteServices()
			m.monitoringManager.DeleteMonitors()
			m.mutex.Unlock()
		}
	}
}
