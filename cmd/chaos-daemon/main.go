// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"flag"
	"os"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/fusedev"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
)

var (
	log  = ctrl.Log.WithName("chaos-daemon")
	conf = &chaosdaemon.Config{Host: "0.0.0.0", CrClientConfig: &crclients.CrClientConfig{}}

	printVersion bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.IntVar(&conf.GRPCPort, "grpc-port", 31767, "the port which grpc server listens on")
	flag.IntVar(&conf.HTTPPort, "http-port", 31766, "the port which http server listens on")
	flag.StringVar(&conf.CrClientConfig.Runtime, "runtime", "docker", "current container runtime")
	flag.StringVar(&conf.CrClientConfig.SocketPath,
		"runtime-socket-path",
		"/var/run/docker.sock",
		"current container runtime socket path",
		)
	flag.StringVar(&conf.CaCert, "ca", "", "ca certificate of grpc server")
	flag.StringVar(&conf.Cert, "cert", "", "certificate of grpc server")
	flag.StringVar(&conf.Key, "key", "", "key of grpc server")
	flag.BoolVar(&conf.Profiling, "pprof", false, "enable pprof")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo("Chaos-daemon")

	if printVersion {
		os.Exit(0)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		// Use collectors as prometheus functions deprecated
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	log.Info("grant access to /dev/fuse")
	err := fusedev.GrantAccess()
	if err != nil {
		log.Error(err, "fail to grant access to /dev/fuse")
	}

	if err = chaosdaemon.StartServer(conf, reg); err != nil {
		log.Error(err, "failed to start chaos-daemon server")
		os.Exit(1)
	}
}
