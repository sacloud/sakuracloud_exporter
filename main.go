// Copyright 2019-2022 The sakuracloud_exporter Authors
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
	"github.com/sacloud/sakuracloud_exporter/platform"
)

var (
	// Version of sakuracloud_exporter.
	Version = "0.17.1"
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

	level.Info(logger).Log( // nolint
		"msg", "starting sakuracloud_exporter",
		"rate-limit", c.RateLimit,
		"version", Version,
		"revision", Revision,
		"goVersion", GoVersion,
	)

	client := platform.NewSakuraCloudClient(c, Version)
	ctx := context.Background()

	if !client.HasValidAPIKeys(ctx) {
		panic(errors.New("unauthorized: invalid API key is applied"))
	}
	if !c.NoCollectorWebAccel && !client.HasWebAccelPermission(ctx) {
		logger.Log("warn", "API key doesn't have webaccel permission") // nolint
	}

	errs := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sakuracloud_exporter_errors_total",
		Help: "The total number of errors per collector",
	}, []string{"collector"})

	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		PidFn: func() (int, error) { return os.Getpid(), nil },
	}))

	ctx, cancel := context.WithCancel(ctx)

	// collector info
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(collector.NewExporterCollector(ctx, logger, Version, Revision, GoVersion, StartTime))
	r.MustRegister(errs)

	// sakuracloud metrics
	if !c.NoCollectorAutoBackup {
		r.MustRegister(collector.NewAutoBackupCollector(ctx, logger, errs, client.AutoBackup))
	}
	if !c.NoCollectorCoupon {
		r.MustRegister(collector.NewCouponCollector(ctx, logger, errs, client.Coupon))
	}
	if !c.NoCollectorDatabase {
		r.MustRegister(collector.NewDatabaseCollector(ctx, logger, errs, client.Database))
	}
	if !c.NoCollectorESME {
		r.MustRegister(collector.NewESMECollector(ctx, logger, errs, client.ESME))
	}
	if !c.NoCollectorInternet {
		r.MustRegister(collector.NewInternetCollector(ctx, logger, errs, client.Internet))
	}
	if !c.NoCollectorLoadBalancer {
		r.MustRegister(collector.NewLoadBalancerCollector(ctx, logger, errs, client.LoadBalancer))
	}
	if !c.NoCollectorLoadBalancer {
		r.MustRegister(collector.NewLocalRouterCollector(ctx, logger, errs, client.LocalRouter))
	}
	if !c.NoCollectorNFS {
		r.MustRegister(collector.NewNFSCollector(ctx, logger, errs, client.NFS))
	}
	if !c.NoCollectorMobileGateway {
		r.MustRegister(collector.NewMobileGatewayCollector(ctx, logger, errs, client.MobileGateway))
	}
	if !c.NoCollectorProxyLB {
		r.MustRegister(collector.NewProxyLBCollector(ctx, logger, errs, client.ProxyLB))
	}
	if !c.NoCollectorServer {
		r.MustRegister(collector.NewServerCollector(ctx, logger, errs, client.Server, c.NoCollectorServerExceptMaintenance))
	}
	if !c.NoCollectorSIM {
		r.MustRegister(collector.NewSIMCollector(ctx, logger, errs, client.SIM))
	}
	if !c.NoCollectorVPCRouter {
		r.MustRegister(collector.NewVPCRouterCollector(ctx, logger, errs, client.VPCRouter))
	}
	if !c.NoCollectorZone {
		r.MustRegister(collector.NewZoneCollector(ctx, logger, errs, client.Zone))
	}
	if !c.NoCollectorWebAccel {
		r.MustRegister(collector.NewWebAccelCollector(ctx, logger, errs, client.WebAccel))
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

	level.Info(logger).Log("msg", "listening", "addr", c.WebAddr) // nolint
	if err := http.ListenAndServe(c.WebAddr, nil); err != nil {
		cancel()
		level.Error(logger).Log("msg", "http listenandserve error", "err", err) // nolint
		os.Exit(2)
	}
}
