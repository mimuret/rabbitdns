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

package config

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	ErrSyntaxNoListen         = errors.New("DNSListens parameter is required")
	ErrSyntaxInvalidListen    = errors.New("DNSListens parameter is invalid format")
	ErrSyntaxNoCtlListen      = errors.New("APIListens parameter is required")
	ErrSyntaxCtlInvalidListen = errors.New("APIListens parameter is invalid format")
	ErrSyntaxMinTCPQueries    = errors.New("MaxTCPQueries parameter must grater than 0")
	ErrSyntaxMaxTCPQryType    = errors.New("MaxTCPQueries parameter must integer")
)

type Config struct {
	FileBackend       bool
	ConfigDir         string
	EtcdBackend       bool
	EtcdServers       []string
	EtcdClientKey     string
	EtcdClientCert    string
	EtcdClientCA      string
	EtcdPrefix        string
	LogLevel          string
	DNSListeners      []string
	APIListeners      []string
	ExporterListeners []string
	MaxTCPQueries     int
	MinimumResponse   bool
}

func NewConfig() (*Config, error) {
	var err error
	c := &Config{
		ConfigDir:         ".",
		EtcdServers:       []string{"http://localhost:2379"},
		EtcdPrefix:        "rabbitdns",
		LogLevel:          "info",
		DNSListeners:      []string{"0.0.0.0:53", "[::]:53"},
		APIListeners:      []string{"0.0.0.0:8053", "[::]:8053"},
		ExporterListeners: []string{"0.0.0.0:9500", "[::]:9500"},
		MaxTCPQueries:     -1,
	}
	if v := os.Getenv("RABBITDNS_FILE_BACKEND"); v != "" {
		switch strings.ToUpper(v) {
		case "ON":
			c.EtcdBackend = true
		case "OFF":
			c.EtcdBackend = true
		}
	}
	if v := os.Getenv("RABBITDNS_CONGIG_DIR"); v != "" {
		c.ConfigDir = v
	}
	if v := os.Getenv("RABBITDNS_ETCD_SERVERS"); v != "" {
		c.EtcdServers = strings.Split(v, ",")
	}
	if v := os.Getenv("RABBITDNS_ETCD_BACKEND"); v != "" {
		switch strings.ToUpper(v) {
		case "ON":
			c.EtcdBackend = true
		case "OFF":
			c.EtcdBackend = true
		}
	}
	if v := os.Getenv("RABBITDNS_ETCD_CLIENT_KEY"); v != "" {
		c.EtcdClientKey = v
	}
	if v := os.Getenv("RABBITDNS_ETCD_CLIENT_CERT"); v != "" {
		c.EtcdClientCert = v
	}
	if v := os.Getenv("RABBITDNS_ETCD_CLIENT_CA"); v != "" {
		c.EtcdClientCA = v
	}
	if v := os.Getenv("RABBITDNS_ETCD_PREFIX"); v != "" {
		c.EtcdPrefix = v
	}
	if v := os.Getenv("RABBITDNS_LOGLEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("RABBITDNS_DNS_LISTEN"); v != "" {
		c.DNSListeners = strings.Split(v, ",")
	}
	if v := os.Getenv("RABBITDNS_API_LISTEN"); v != "" {
		c.APIListeners = strings.Split(v, ",")
	}
	if v := os.Getenv("RABBITDNS_EXPORTER_LISTEN"); v != "" {
		c.ExporterListeners = strings.Split(v, ",")
	}
	if v := os.Getenv("RABBITDNS_MAX_TCP_QUERIES"); v != "" {
		c.MaxTCPQueries, err = strconv.Atoi(v)
		if err != nil {
			log.Warn(ErrSyntaxMaxTCPQryType)
			return nil, ErrSyntaxMaxTCPQryType
		}
	}
	return c, err
}

func (c *Config) GetFlag(cb *cobra.Command) {
	if v, err := cb.PersistentFlags().GetBool("file"); err != nil {
		c.FileBackend = v
	}
	if v, err := cb.PersistentFlags().GetString("file.dir"); err != nil {
		c.ConfigDir = v
	}
	if v, err := cb.PersistentFlags().GetBool("etcd"); err != nil {
		c.EtcdBackend = v
	}
	if v, err := cb.PersistentFlags().GetString("etcd.servers"); err != nil {
		c.EtcdServers = strings.Split(v, ",")
	}
	if v, err := cb.PersistentFlags().GetString("etcd.key"); err != nil {
		c.EtcdClientKey = v
	}
	if v, err := cb.PersistentFlags().GetString("etcd.cert"); err != nil {
		c.EtcdClientCert = v
	}
	if v, err := cb.PersistentFlags().GetString("etcd.ca"); err != nil {
		c.EtcdClientCA = v
	}
	if v, err := cb.PersistentFlags().GetString("etcd.prefix"); err != nil {
		c.EtcdPrefix = v
	}
	if v, err := cb.PersistentFlags().GetString("log-level"); err != nil {
		c.LogLevel = v
	}
	if v, err := cb.PersistentFlags().GetString("dns-listen"); err != nil {
		c.DNSListeners = strings.Split(v, ",")
	}
	if v, err := cb.PersistentFlags().GetString("api-listen"); err != nil {
		c.APIListeners = strings.Split(v, ",")
	}
	if v, err := cb.PersistentFlags().GetString("exporter-listen"); err != nil {
		c.ExporterListeners = strings.Split(v, ",")
	}
	if v, err := cb.PersistentFlags().GetInt("max-tcp-queries"); err != nil {
		c.MaxTCPQueries = v
	}
}

func (c *Config) Check() error {
	syntaxError := &SyntaxError{}
	if c.FileBackend {

	}
	if c.EtcdBackend {
		if len(c.EtcdServers) == 0 {
			syntaxError.Add(ErrSyntaxNoListen)
		}
	}
	if len(c.DNSListeners) == 0 {
		syntaxError.Add(ErrSyntaxNoListen)
	}
	for _, listen := range c.DNSListeners {
		_, err := net.ResolveTCPAddr("tcp", listen)
		if err != nil {
			syntaxError.Add(ErrSyntaxInvalidListen)
		}
	}
	if len(c.APIListeners) == 0 {
		syntaxError.Add(ErrSyntaxNoCtlListen)
	}
	for _, listen := range c.APIListeners {
		_, err := net.ResolveTCPAddr("tcp", listen)
		if err != nil {
			syntaxError.Add(ErrSyntaxCtlInvalidListen)
		}
	}
	if c.MaxTCPQueries == 0 {
		syntaxError.Add(ErrSyntaxMinTCPQueries)
	}
	return syntaxError.Return()
}
