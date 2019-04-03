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

package rdns

import (
	"strings"

	"github.com/miekg/dns"
)

const (
	TypeANAME = 0xFF17
)

type ANAME struct {
	Resource string `dns:"octet"`
}

func init() {
	dns.PrivateHandle("ANAME", TypeANAME, NewANAME)
}

func NewANAME() dns.PrivateRdata { return &ANAME{} }

func (rd *ANAME) Len() int       { return len([]byte(rd.Resource)) }
func (rd *ANAME) String() string { return rd.Resource }
func (rd *ANAME) Parse(txt []string) error {
	rd.Resource = strings.TrimSpace(strings.Join(txt, " "))
	return nil
}

func (rd *ANAME) Pack(buf []byte) (int, error) {
	b := []byte(rd.Resource)
	n := copy(buf, b)
	if n != len(b) {
		return n, dns.ErrBuf
	}
	return n, nil
}

func (rd *ANAME) Unpack(buf []byte) (int, error) {
	rd.Resource = string(buf)
	return len(buf), nil
}

func (rd *ANAME) Copy(dest dns.PrivateRdata) error {
	d, ok := dest.(*ANAME)
	if !ok {
		return dns.ErrRdata
	}
	d.Resource = rd.Resource
	return nil
}
