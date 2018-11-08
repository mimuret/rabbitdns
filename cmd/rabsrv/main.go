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

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/rabbitdns/rabbitdns/internal/server"
	"github.com/rabbitdns/rabbitdns/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ErrExecute = errors.New("execute error")

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	rootCmd := &cobra.Command{
		Use: "rabsrv",
		Run: serv,
	}
	rootCmd.PersistentFlags().BoolP("file", "", true, "use file base backend")
	rootCmd.PersistentFlags().StringP("file.dir", "", ".", "config directory path")
	rootCmd.PersistentFlags().BoolP("etcd", "", false, "use file etcd backend")
	rootCmd.PersistentFlags().StringP("etcd.servers", "", "", "List of etcd servers to connect with (scheme://ip:port), comma separated.")
	rootCmd.PersistentFlags().StringP("etcd.prefix", "", "rabbitdns", "The prefix to prepend to all resource paths in etcd.")
	rootCmd.PersistentFlags().StringP("etcd.cert", "", "", "etcd client certification.")
	rootCmd.PersistentFlags().StringP("etcd.key", "", "", "etcd client key.")
	rootCmd.PersistentFlags().StringP("etcd.ca", "", "", "etcd client CA.")
	rootCmd.PersistentFlags().StringP("log-level", "", "info", "Log level.")
	rootCmd.PersistentFlags().StringP("dns-listen", "", "", "List of dns server listen ip, comma separated.")
	rootCmd.PersistentFlags().StringP("api-listen", "", "", "List of api server listen ip, comma separated.")
	rootCmd.PersistentFlags().StringP("exporter-listen", "", "", "List of exporter server listen ip, comma separated.")
	rootCmd.PersistentFlags().IntP("max-tcp-queries", "", 0, "Number of max tcp query.")

	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"Type":  "rabsrv",
			"Func":  "main",
			"Error": err,
		}).Fatal(ErrExecute)
	}
	os.Exit(0)
}
func serv(cb *cobra.Command, args []string) {
	c := config.NewConfig()
	c.GetFlag(cb)
	if err := config.Check(); err != nil {
		log.WithFields(log.Fields{
			"Type": "rabsrv",
			"Func": "serv",
		}).Fatal("config load error")
	}

	log.WithFields(log.Fields{
		"Type": "rabsrv",
		"Func": "serv",
	}).Info("rabbitdns started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	master := server.NewMaster()
	for {
		select {
		case <-sigCh:
			return
		case newConfig := <-configCh:
			if c == nil {
				c = newConfig
				utils.SetLogLevel(c.LogLevel)
				if err := master.StartServ(context.Background(), c); err != nil {
					log.Fatal(err)
				}
			} else {
				// reload config
				utils.SetLogLevel(newConfig.LogLevel)

			}
		}
	}
}
