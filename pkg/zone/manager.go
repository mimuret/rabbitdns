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

package zone

import (
	"io"
	"net"

	"github.com/pkg/errors"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/pkg/rdns"
	"github.com/rabbitdns/rabbitdns/pkg/resource_interface"
	log "github.com/sirupsen/logrus"
)

var (
	ErrParseRR         = errors.New("parse error")
	ErrParseZone       = errors.New("parse zone")
	ErrServiceNotFount = errors.New("service not found")
	ErrLoad            = errors.New("load error")
)

type ZoneManager struct {
	zoneSet         *Tree
	loading         map[string]bool
	backends        []ZoneBackendInterface
	serviceResource resource_interface.ServiceResourceInterface
}

var responseServfail = &GetResourceResponse{Rcode: dns.RcodeServerFailure}
var responseRefused = &GetResourceResponse{Rcode: dns.RcodeRefused}
var responseNXDOMAIN = &GetResourceResponse{Rcode: dns.RcodeNameError}

func NewZoneManager(backends []ZoneBackendInterface) *ZoneManager {
	return &ZoneManager{
		zoneSet:  NewTree(),
		loading:  map[string]bool{},
		backends: backends,
	}
}

func (m *ZoneManager) Update() error {
	for k, _ := range m.loading {
		m.loading[k] = false
	}
	for _, b := range m.backends {
		if list, err := b.ListZones(); err != nil {
			for _, i := range list.Items {
				var modTime int64
				m.loading[i.Name] = true

				// get mod time
				originLabels := rdns.Labels(i.Name)
				zoneNode := m.zoneSet.AddNode(originLabels)
				if v, ok := zoneNode.Get("ModTime"); ok == true {
					switch mt := v.(type) {
					case int64:
						modTime = mt
					}
				}
				if modTime != i.LastModified {
					reader, err := b.GetZone(i.Name)
					if err != nil {
						return errors.Wrap(err, "can't read zone")
					}
					if err := m.addZone(i, reader); err != nil {
						return errors.Wrap(err, "failed to read zone")
					}
				}
			}
		}
	}
	return nil
}

func (m *ZoneManager) DeleteZones() {
	for k, v := range m.loading {
		if v == false {
			origin := dns.Fqdn(k)
			originLabels := rdns.Labels(origin)

			if node := m.zoneSet.SearchNode(originLabels, true); node != nil {
				if s, ok := node.Get("services"); ok {
					if services, ok := s.([]string); ok {
						for _, serviceName := range services {
							m.serviceResource.UnRegisterService(serviceName, origin)
						}
					}
				}
				node.DeleteAll()
			}
			m.zoneSet.DeleteNode(originLabels, false)
		}
	}
}

func (m *ZoneManager) addZone(i *ZoneInfo, r io.Reader) error {
	origin := dns.Fqdn(i.Name)
	originLabels := rdns.Labels(origin)
	services := []string{}
	zoneTree := NewTree()
	zoneTree.Auth = true
	var RRs []dns.RR
	success := true
	for x := range dns.ParseZone(r, i.Name, "") {
		if x.Error != nil {
			log.WithFields(log.Fields{
				"Type":  "lib/server/ZoneManager",
				"Func":  "LoadZones",
				"Error": x.Error,
			}).Warn(ErrParseRR)
			success = false

			return ErrParseZone
		} else {
			if dyn, ok := x.RR.(*dns.PrivateRR); ok {
				if rdata, ok := dyn.Data.(*rdns.DYNRR); ok {
					if _, err := m.serviceResource.GetResources(&resource_interface.GetResourcesRequest{
						Name: rdata.Resource,
					}); err == resource_interface.ErrNotDefineService {
						return errors.Wrap(ErrServiceNotFount, "name:"+dyn.Header().Name+",ServiceName:"+rdata.Resource)
					}
					services = append(services, rdata.Resource)
				}
			}
			node := zoneTree.AddRR(x.RR)
			if x.RR.Header().Rrtype == dns.TypeNS && dns.Fqdn(x.RR.Header().Name) != i.Name {
				node.Auth = false
			}
		}
	}
	if err := zoneTree.verifyZone(originLabels); err != nil {
		m.zoneSet.Set("state", false)
		return err
	}
	if success != true {
		m.zoneSet.Set("state", false)
		return ErrParseZone
	}
	m.zoneSet.Set("ModTime", i.LastModified)
	m.zoneSet.Set("Records", RRs)
	m.zoneSet.Set("state", true)
	m.zoneSet.Set("ZoneTree", zoneTree)
	m.zoneSet.Set("services", services)
	for _, service_name := range services {
		m.serviceResource.RegisterService(service_name, origin)
	}

	log.WithFields(log.Fields{
		"Type":     "lib/server/ZoneManager",
		"Func":     "readZone",
		"zonename": origin,
	}).Info("load zone")

	return nil
}

func (m *ZoneManager) searchZone(labels []string) *Tree {
	tree := m.zoneSet.SearchNode(labels, false)
	for tree.Label != "" {
		if _, ok := tree.Get("provide"); ok == true {
			return tree
		}
		tree = tree.Parent
	}
	return nil
}

type GetResourceRequest struct {
	Qname        string
	Qtype        uint16
	Labels       []string
	SourceIP     net.IP
	ClientSubnet []net.IPNet
}

type GetResourceResponse struct {
	Rcode         int
	Authoritative bool
	Answer        []dns.RR
	Authority     []dns.RR
	Additional    []dns.RR
}

func (m *ZoneManager) GetResource(req *GetResourceRequest) (*GetResourceResponse, error) {
	zoneNode := m.searchZone(req.Labels)
	if zoneNode == nil {
		return responseRefused, nil
	}
	if v, ok := zoneNode.Get("ZoneTree"); ok == true {
		if zoneTree, ok := v.(*Tree); ok {
			response, err := zoneTree.GetAnswer(req, m.serviceResource)
			if err != nil {
				return responseServfail, errors.Wrap(err, "can't create response")
			}
			return response, nil
		}
	}
	return responseServfail, errors.New("Not ZoneTree exist")
}
