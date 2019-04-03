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
	"errors"
	"io"

	"github.com/miekg/dns"
	"github.com/rabbitdns/rabbitdns/pkg/rdns"
	"github.com/spf13/viper"
)

var (
	ErrReadService        = errors.New("because can't read service file")
	ErrRRtypeEmpty        = errors.New("because RRType is empty")
	ErrRRTypeNotSupported = errors.New("because RRType is not supported")
	ErrServiceEmpty       = errors.New("because service is empty")
)

type Config struct {
	Name     string
	RRType   uint16
	Service  Service
	Monitors map[*Endpoint]string
	ModTime  int64
}

func NewConfig() *Config {
	return &Config{Monitors: map[*Endpoint]string{}}
}

func LoadConfig(r io.Reader, modTime int64) (*Config, error) {
	var err error
	config := NewConfig()

	v := viper.New()
	v.SetConfigType("yml")

	config.ModTime = modTime

	if err = v.ReadConfig(r); err != nil {
		return nil, ErrReadService
	}
	return config.LoadFromViper(v)
}

func (c *Config) LoadFromViper(v *viper.Viper) (*Config, error) {
	var err error
	// RRType
	rrStr, ok := v.Get("RRtype").(string)
	if ok == false {
		return nil, ErrRRtypeEmpty
	}

	rrType, ok := dns.StringToType[rrStr]
	if ok == false {
		return nil, ErrRRTypeNotSupported
	}
	if _, ok := rdns.StaticDynamicMap[rrType]; ok == false {
		return nil, ErrRRTypeNotSupported

	}
	c.RRType = rrType

	// Rule
	if ok := v.Get("service"); ok == nil {
		return nil, ErrServiceEmpty
	}
	c.Service, err = CreateService(c, "service", v)
	return c, err
}
func (c *Config) Delete() {
}
func (c *Config) Update(current *Config) error {
	return nil
}
