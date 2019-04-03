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

package worker

import (
	"net"

	"github.com/pkg/errors"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/pkg/config"
	"github.com/rabbitdns/rabbitdns/pkg/rdns"
	"github.com/rabbitdns/rabbitdns/pkg/zone"
	log "github.com/sirupsen/logrus"
)

var (
	ErrServ             = errors.New("failed to start dns server.")
	ErrZoneCutFind      = errors.New("failed to find zone cut.")
	ErrNotFoundZoneData = errors.New("faild to find zone data in zone node.")
)

const (
	Answer int = iota
	Authoritative
	Additional
)
const (
	NXDOMAIN int = iota
	FOUND
	NOTFOUND
)

type Worker struct {
	listener    *dns.Server
	mux         *dns.ServeMux
	config      *config.Config
	zoneManager *zone.ZoneManager
}

func NewWorker(minimumResponse bool, maxTCPQueries int, addr string, proto string, zoneManager *zone.ZoneManager) *Worker {
	worker := &Worker{
		listener: &dns.Server{
			Addr:          addr,
			Net:           proto,
			MaxTCPQueries: maxTCPQueries,
		},
		zoneManager: zoneManager,
	}
	mux := dns.NewServeMux()
	mux.Handle(".", worker)
	worker.mux = mux

	return worker
}

func (s *Worker) Run() {
	go func(l *dns.Server) {
		if err := l.ListenAndServe(); err != nil {
			log.WithFields(log.Fields{
				"Type":   "lib/server/Worker",
				"Func":   "Run",
				"Error":  err,
				"server": l,
			}).Fatal(ErrServ)
		}
	}(s.listener)
}

func (s *Worker) serverDNSCAHOS(m *dns.Msg, req *dns.Msg) {
	qname := req.Question[0].Name
	m.MsgHdr.AuthenticatedData = false
	m.MsgHdr.CheckingDisabled = false
	m.MsgHdr.Authoritative = true

	if req.Question[0].Qtype != dns.TypeTXT {
		m.Rcode = dns.RcodeNXRrset
		return
	}
	switch qname {
	case "version.bind.", "version.server":
		hdr := dns.RR_Header{Name: qname, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
		m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"1.0.0"}}}
		m.Rcode = dns.RcodeSuccess
	case "hostname.bind.", "id.server.":
		hdr := dns.RR_Header{Name: qname, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
		m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"localhost."}}}
		m.Rcode = dns.RcodeSuccess
	default:
		m.Rcode = dns.RcodeNXRrset
	}
}

func (s *Worker) serveDNSINET(w dns.ResponseWriter, m *dns.Msg, req *dns.Msg) error {
	var sourceIP net.IP
	switch addr := w.LocalAddr().(type) {
	case *net.UDPAddr:
		sourceIP = addr.IP
	case *net.TCPAddr:
		sourceIP = addr.IP
	}
	var ecs = []net.IPNet{}
	for _, rr := range m.Extra {
		if edns, ok := rr.(*dns.OPT); ok {
			for _, opt := range edns.Option {
				if e, ok := opt.(*dns.EDNS0_SUBNET); ok {
					len := 32
					if e.Family == 2 {
						len = 128
					}
					subnet := net.IPNet{
						IP:   e.Address,
						Mask: net.CIDRMask(int(e.SourceNetmask), len),
					}
					ecs = append(ecs, subnet)
				}
			}
		}
	}
	resourceReq := &zone.GetResourceRequest{
		Qname:        req.Question[0].Name,
		Qtype:        req.Question[0].Qtype,
		Labels:       rdns.Labels(req.Question[0].Name),
		SourceIP:     sourceIP,
		ClientSubnet: ecs,
	}
	// todo ecs

	res, err := s.zoneManager.GetResource(resourceReq)
	if err != nil {
		return err
	}
	m.Rcode = res.Rcode
	m.Authoritative = res.Authoritative
	m.Answer = res.Answer
	m.Ns = res.Authority
	m.Extra = res.Additional
	return nil
}

func (s *Worker) servfail(m *dns.Msg) {
	m.Rcode = dns.RcodeServerFailure
}
func (s *Worker) refused(m *dns.Msg) {
	m.Rcode = dns.RcodeRefused
}
func (s *Worker) notImplemented(m *dns.Msg) {
	m.Rcode = dns.RcodeNotImplemented
}

func (s *Worker) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	switch req.Question[0].Qclass {
	case dns.ClassCHAOS:
		s.serverDNSCAHOS(m, req)
	case dns.ClassINET:
		err := s.serveDNSINET(w, m, req)
		if err != nil {
			s.servfail(m)
		}
	default:
		s.notImplemented(m)
	}
	w.WriteMsg(m)
}
