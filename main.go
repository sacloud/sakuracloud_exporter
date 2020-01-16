// Copyright 2019-2020 The sakuracloud_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sacloud/sakuracloud_exporter/collector"
	"github.com/sacloud/sakuracloud_exporter/config"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

var (
	// Version of sakuracloud_exporter.
	Version string
	// Revision or Commit this binary was built from.
	Revision string
	// GoVersion running this binary.
	GoVersion = runtime.Version()
	// StartTime has the time this was started.
	StartTime = time.Now()
)

func main() {

	c, err := config.InitConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filterOption := level.AllowInfo()
	if c.Debug {
		filterOption = level.AllowDebug()
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, filterOption)
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	level.Info(logger).Log(
		"msg", "starting sakuracloud_exporter",
		"rate-limit", c.RateLimit,
		"version", Version,
		"revision", Revision,
		"goVersion", GoVersion,
	)

	client := iaas.NewSakuraCloucClient(c, Version)
	if !client.HasValidAPIKeys(context.TODO()) {
		panic(errors.New("unauthorized: invalid API key is applied"))
	}

	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sakuracloud_exporter_errors_total",
		Help: "The total number of errors per collector",
	}, []string{"collector"})

	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		PidFn: func() (int, error) { return os.Getpid(), nil },
	}))

	ctx, cancel := context.WithCancel(context.Background())

	// collector info
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(collector.NewExporterCollector(ctx, logger, Version, Revision, GoVersion, StartTime))
	r.MustRegister(errors)

	// sakuracloud metrics
	if !c.NoCollectorAutoBackup {
		r.MustRegister(collector.NewAutoBackupCollector(ctx, logger, errors, client.AutoBackup))
	}
	if !c.NoCollectorCoupon {
		r.MustRegister(collector.NewCouponCollector(ctx, logger, errors, client.Coupon))
	}
	if !c.NoCollectorDatabase {
		r.MustRegister(collector.NewDatabaseCollector(ctx, logger, errors, client.Database))
	}
	if !c.NoCollectorInternet {
		r.MustRegister(collector.NewInternetCollector(ctx, logger, errors, client.Internet))
	}
	if !c.NoCollectorLoadBalancer {
		r.MustRegister(collector.NewLoadBalancerCollector(ctx, logger, errors, client.LoadBalancer))
	}
	if !c.NoCollectorNFS {
		r.MustRegister(collector.NewNFSCollector(ctx, logger, errors, client.NFS))
	}
	if !c.NoCollectorMobileGateway {
		r.MustRegister(collector.NewMobileGatewayCollector(ctx, logger, errors, client.MobileGateway))
	}
	if !c.NoCollectorProxyLB {
		r.MustRegister(collector.NewProxyLBCollector(ctx, logger, errors, client.ProxyLB))
	}
	if !c.NoCollectorServer {
		r.MustRegister(collector.NewServerCollector(ctx, logger, errors, client.Server))
	}
	if !c.NoCollectorSIM {
		r.MustRegister(collector.NewSIMCollector(ctx, logger, errors, client.SIM))
	}
	if !c.NoCollectorVPCRouter {
		r.MustRegister(collector.NewVPCRouterCollector(ctx, logger, errors, client.VPCRouter))
	}
	if !c.NoCollectorZone {
		r.MustRegister(collector.NewZoneCollector(ctx, logger, errors, client.Zone))
	}

	http.Handle(c.WebPath,
		promhttp.HandlerFor(r, promhttp.HandlerOpts{}),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>SakuraCloud Exporter</title></head>
			<body>
			<h1>SakuraCloud Exporter</h1>
			<p><a href="` + c.WebPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "listening", "addr", c.WebAddr)
	if err := http.ListenAndServe(c.WebAddr, nil); err != nil {
		cancel()
		level.Error(logger).Log("msg", "http listenandserve error", "err", err)
		os.Exit(2)
	}
}
