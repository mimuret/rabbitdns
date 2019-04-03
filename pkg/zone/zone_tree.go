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
	"strings"

	"github.com/pkg/errors"

	"github.com/rabbitdns/rabbitdns/pkg/resource_interface"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/pkg/rdns"
)

var (
	ErrVerifyZoneEmptyApex           = errors.New("Apex node is empty.")
	ErrVerifyZoneEmptySOA            = errors.New("SOA RR is empty.")
	ErrVerifyZoneDupulicateSOA       = errors.New("More than 1 SOA RR found.")
	ErrVerifyZoneEmptyApexNS         = errors.New("Apex NS not found.")
	ErrVerifyNodeDupulicateCNAME     = errors.New("More than 1 CNAME RR found in same name.")
	ErrVerifyNodeOtherRRInCNAMENode  = errors.New("Found other RR in CNAME node.")
	ErrVerifyNodeDupulicateDNAME     = errors.New("More than 1 DNAME RR found in same name.")
	ErrVerifyNodeFoundChildNodeDNAME = errors.New("Found child node in DNAME node.")
)

// for RR
func (t *Tree) AddRR(rr dns.RR) *Tree {
	labels := rdns.Labels(rr.Header().Name)
	rrNode := t.AddNode(labels)
	rrNode.SetRR(rr)
	return rrNode
}

func (t *Tree) SetRR(rr dns.RR) {
	t.Resources[rr.Header().Rrtype] = append(t.Resources[rr.Header().Rrtype], rr)
}

func (t *Tree) GetRR(rrType uint16) ([]dns.RR, bool) {
	v, ok := t.Resources[rrType]
	return v, ok
}

func (t *Tree) DeleteRR(rrType uint16, rr dns.RR) {
	delete(t.Resources, rrType)
}

func (t *Tree) GetRRWithName(name string, rrType uint16) ([]dns.RR, bool) {
	labels := rdns.Labels(name)
	node := t.SearchNode(labels, true)
	if node == nil {
		return nil, false
	}
	return node.GetRR(rrType)
}

func (t *Tree) verifyZone(originLabels []string) error {
	// APEX CHECK
	apex := t.SearchNode(originLabels, true)
	if apex == nil {
		return ErrVerifyZoneEmptyApex
	}
	// SOA CHECK
	soa, exist := apex.GetRR(dns.TypeSOA)
	if !exist {
		return ErrVerifyZoneEmptySOA
	}
	if len(soa) > 1 {
		return ErrVerifyZoneDupulicateSOA
	}
	// Zone APEX NS CHECK
	if _, exist := apex.GetRR(dns.TypeNS); !exist {
		return ErrVerifyZoneEmptyApexNS
	}

	return apex.verifyNode()
}

func (t *Tree) verifyNode() error {
	if len(t.Resources) > 0 {
		// CNAME CHECK
		if cname, exist := t.GetRR(dns.TypeCNAME); exist {
			if len(cname) > 1 {
				return ErrVerifyNodeDupulicateCNAME
			}
			if len(t.Resources) > 2 {
				return ErrVerifyNodeOtherRRInCNAMENode
			}
			for rrtype := range t.Resources {
				if rrtype != dns.TypeCNAME && rrtype != dns.TypeDNAME {
					return ErrVerifyNodeOtherRRInCNAMENode
				}
			}
		}
		// DNAME CHECK
		if dname, exist := t.GetRR(dns.TypeDNAME); exist {
			if len(dname) > 1 {
				return ErrVerifyNodeDupulicateDNAME
			}
			if len(t.Children) > 0 {
				return ErrVerifyNodeFoundChildNodeDNAME
			}
		}
	}
	for _, child := range t.Children {
		if err := child.verifyNode(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) findZoneCut() *Tree {
	if t.Auth {
		return nil
	}
	if t.Parent.Auth {
		return t
	}
	return t.Parent.findZoneCut()
}

var (
	ErrZoneCutFind = errors.New("not founf zone cut")
)

func (t *Tree) GetAnswer(req *GetResourceRequest, sr resource_interface.ServiceResourceInterface) (*GetResourceResponse, error) {
	response := &GetResourceResponse{
		Rcode:         dns.RcodeSuccess,
		Authoritative: false,
		Answer:        []dns.RR{},
		Authority:     []dns.RR{},
		Additional:    []dns.RR{},
	}
	err := t.getAnswer(response, req, sr, req.Qname, 10, false)
	return response, err
}

func isAuth(t *Tree, zoneName string, rrtype uint16) bool {
	if rrtype == dns.TypeDS {
		return zoneName != t.Label
	}
	return t.Auth
}

func (t *Tree) getAnswer(res *GetResourceResponse, req *GetResourceRequest, sr resource_interface.ServiceResourceInterface, sname string, count int, isWildcard bool) error {
	labels := rdns.Labels(sname)
	if count <= 0 {
		return nil
	}

	node := t.SearchNode(labels, isWildcard)
	if node == nil {
		return nil
	}
	if isAuth(node, sname, req.Qtype) {
		res.Authoritative = true
		if node.Label == sname {
			// found name
			res.Rcode = dns.RcodeSuccess
			if rrs, exist := node.GetRR(req.Qtype); exist {
				// found RR
				for _, rr := range rrs {
					if isWildcard {
						rr.Header().Name = req.Qname
					}
					res.Answer = append(res.Answer, rr)
				}
			} else if rrs, exist := node.GetRR(dns.TypeCNAME); exist {
				// found CNAME
				res.Answer = append(res.Answer, rrs[0])
				if cname, ok := rrs[0].(*dns.CNAME); ok {
					creq := &GetResourceRequest{
						Qname:        cname.Target,
						Qtype:        req.Qtype,
						Labels:       rdns.Labels(cname.Target),
						SourceIP:     req.SourceIP,
						ClientSubnet: req.ClientSubnet,
					}
					return t.getAnswer(res, creq, sr, creq.Qname, count-1, false)
				} else if dynamicRR, exist := rdns.StaticDynamicMap[req.Qtype]; exist {
					if rrs, ok := node.GetRR(dynamicRR); ok {
						if dyn, ok := rrs[0].(*dns.PrivateRR); ok {
							if rdata, ok := dyn.Data.(*rdns.DYNRR); ok {
								if resources, err := sr.GetResources(&resource_interface.GetResourcesRequest{
									Name:     rdata.Resource,
									RRtype:   req.Qtype,
									SourceIP: req.SourceIP,
									ECS:      req.ClientSubnet,
								}); err != nil {
									return errors.Wrap(err, "can't get service resource.")
								} else {
									for _, rr := range resources {
										rr.Header().Name = req.Qname
										rr.Header().Rrtype = req.Qtype
										rr.Header().Class = dyn.Header().Class
										rr.Header().Ttl = dyn.Header().Ttl
										res.Answer = append(res.Answer, rr)
									}
								}
							}
						}
					}
				}
			}
		} else {
			if rrs, exist := node.GetRR(dns.TypeDNAME); exist {
				dname := rrs[0]
				dname.Header().Name = req.Qname
				dname.Header().Ttl = 0
				res.Authority = append(res.Authority, dname)
			} else if !isWildcard {
				wildcard := dns.Fqdn("*." + strings.Join(labels[1:], "."))
				return t.getAnswer(res, req, sr, wildcard, count-1, true)
			}
		}
	} else {
		// found Delegation
		zoneCut := node.findZoneCut()
		if zoneCut == nil {
			return ErrZoneCutFind
		}
		if rrs, ok := zoneCut.GetRR(dns.TypeNS); ok == true {
			res.Authority = append(res.Authority, rrs...)
		}
		if rrs, ok := zoneCut.GetRR(dns.TypeDS); ok == true {
			res.Authority = append(res.Authority, rrs...)
		}
		res.Authoritative = false
	}
	return nil
}
