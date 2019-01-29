package main

import (
	"errors"
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
		panic(err)
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
		"version", Version,
		"revision", Revision,
		"goVersion", GoVersion,
	)

	client := iaas.NewSakuraCloucClient(c, Version)
	if !client.HasValidAPIKeys() {
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

	// collector info
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(collector.NewExporterCollector(logger, Version, Revision, GoVersion, StartTime))
	r.MustRegister(errors)

	// sakuracloud metrics
	r.MustRegister(collector.NewAutoBackupCollector(logger, errors, client.AutoBackup))
	r.MustRegister(collector.NewCouponCollector(logger, errors, client.Coupon))
	r.MustRegister(collector.NewDatabaseCollector(logger, errors, client.Database))
	r.MustRegister(collector.NewInternetCollector(logger, errors, client.Internet))
	r.MustRegister(collector.NewLoadBalancerCollector(logger, errors, client.LoadBalancer))
	r.MustRegister(collector.NewNFSCollector(logger, errors, client.NFS))
	r.MustRegister(collector.NewMobileGatewayCollector(logger, errors, client.MobileGateway))
	r.MustRegister(collector.NewServerCollector(logger, errors, client.Server))
	r.MustRegister(collector.NewSIMCollector(logger, errors, client.SIM))
	r.MustRegister(collector.NewVPCRouterCollector(logger, errors, client.VPCRouter))
	r.MustRegister(collector.NewZoneCollector(logger, errors, client.Zone))

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
		level.Error(logger).Log("msg", "http listenandserve error", "err", err)
		os.Exit(1)
	}
}
