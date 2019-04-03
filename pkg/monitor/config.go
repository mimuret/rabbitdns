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
	"io"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

var (
	ErrReadMonitor       = errors.New("Failed to read monitor config.")
	ErrEmptyMonitor      = errors.New("Empty monitor hash.")
	ErrZeroInterval      = errors.New("Intaval value is zero.")
	ErrZeroUPThreshold   = errors.New("UPThreshold value is zero.")
	ErrZeroOKThreshold   = errors.New("OKThreshold value is zero.")
	ErrZeroNGThreshold   = errors.New("NGThreshold value is zero.")
	ErrTimeoutGTInterval = errors.New("Timeout is grater than Interval.")
)

type Config struct {
	Interval    time.Duration
	Timeout     time.Duration
	UPThreshold uint16
	OKThreshold uint16
	NGThreshold uint16
	Monitor     Monitor
	ModTime     int64
}

func NewConfig() *Config {
	return &Config{}
}
func LoadConfig(r io.Reader, modTime int64) (*Config, error) {
	var err error
	config := NewConfig()

	v := viper.New()
	v.SetConfigType("yml")
	v.SetDefault("Interval", 10)
	v.SetDefault("UPThreshold", 20)
	v.SetDefault("OKThreshold", 10)
	v.SetDefault("NGThreshold", 10)

	config.ModTime = modTime
	if err = v.ReadConfig(r); err != nil {
		return nil, errors.Wrap(ErrReadMonitor, err.Error())
	}
	return config.LoadFromViper(v)
}

func (c *Config) LoadFromViper(v *viper.Viper) (*Config, error) {
	var err error
	c.Interval = time.Duration(v.GetInt("Interval"))
	c.Timeout = time.Duration(v.GetInt("Timeout"))
	c.UPThreshold = uint16(v.GetInt("UPThreshold"))
	c.OKThreshold = uint16(v.GetInt("OKThreshold"))
	c.NGThreshold = uint16(v.GetInt("NGThreshold"))
	if c.Interval == 0 {
		return nil, ErrZeroInterval
	}
	if c.Timeout == 0 {
		c.Timeout = c.Interval / 2
		if c.Timeout == 0 {
			c.Timeout = 1
		}
	}
	if c.Interval < c.Timeout {
		return nil, ErrTimeoutGTInterval
	}
	if c.UPThreshold == 0 {
		return nil, ErrZeroInterval
	}
	if c.OKThreshold == 0 {
		return nil, ErrZeroInterval
	}
	if c.NGThreshold == 0 {
		return nil, ErrZeroInterval
	}
	// RRType
	name := v.Get("monitor")
	if name == nil {
		return nil, ErrEmptyMonitor
	}
	c.Monitor, err = CreateMonitor("monitor", v)
	return c, err
}

func (c *Config) CheckRegister(e *Entry) error {
	return c.Monitor.CheckRegister(e)
}

func (c *Config) Update(current *Config) error {
	return nil
}
