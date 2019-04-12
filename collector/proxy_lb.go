package collector

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// ProxyLBCollector collects metrics about all proxyLBs.
type ProxyLBCollector struct {
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.ProxyLBClient

	Up          *prometheus.Desc
	ProxyLBInfo *prometheus.Desc

	BindPortInfo *prometheus.Desc

	ServerInfo *prometheus.Desc

	CertificateInfo       *prometheus.Desc
	CertificateExpireDate *prometheus.Desc

	ActiveConnections *prometheus.Desc
	ConnectionPerSec  *prometheus.Desc
}

// NewProxyLBCollector returns a new ProxyLBCollector.
func NewProxyLBCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.ProxyLBClient) *ProxyLBCollector {
	errors.WithLabelValues("proxyLB").Add(0)

	proxyLBLabels := []string{"id", "name"}
	proxyLBInfoLabels := append(proxyLBLabels, "plan", "vip", "fqdn",
		"proxy_networks", "sorry_server_ipaddress", "sorry_server_port", "tags", "description")

	proxyLBBindPortLabels := append(proxyLBLabels, "bind_port_index", "proxy_mode", "port")
	proxyLBServerLabels := append(proxyLBLabels, "server_index", "ipaddress", "port", "enabled")
	proxyLBCertificateLabels := append(proxyLBLabels, "cert_index")
	proxyLBCertificateInfoLabels := append(proxyLBCertificateLabels, "common_name", "issuer_name")

	return &ProxyLBCollector{
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_proxylb_up",
			"If 1 the ProxyLB is available, 0 otherwise",
			proxyLBLabels, nil,
		),
		ProxyLBInfo: prometheus.NewDesc(
			"sakuracloud_proxylb_info",
			"A metric with a constant '1' value labeled by proxyLB information",
			proxyLBInfoLabels, nil,
		),
		BindPortInfo: prometheus.NewDesc(
			"sakuracloud_proxylb_bind_port_info",
			"A metric with a constant '1' value labeled by BindPort information",
			proxyLBBindPortLabels, nil,
		),
		ServerInfo: prometheus.NewDesc(
			"sakuracloud_proxylb_server_info",
			"A metric with a constant '1' value labeled by real-server information",
			proxyLBServerLabels, nil,
		),
		CertificateInfo: prometheus.NewDesc(
			"sakuracloud_proxylb_cert_info",
			"A metric with a constant '1' value labeled by certificate information",
			proxyLBCertificateInfoLabels, nil,
		),
		CertificateExpireDate: prometheus.NewDesc(
			"sakuracloud_proxylb_cert_expire",
			"Certificate expiration date in seconds since epoch (1970)",
			proxyLBCertificateLabels, nil,
		),
		ActiveConnections: prometheus.NewDesc(
			"sakuracloud_proxylb_active_connections",
			"Active connection count",
			proxyLBLabels, nil,
		),
		ConnectionPerSec: prometheus.NewDesc(
			"sakuracloud_proxylb_connection_per_sec",
			"Connection count per second",
			proxyLBLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ProxyLBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.ProxyLBInfo
	ch <- c.BindPortInfo
	ch <- c.ServerInfo
	ch <- c.CertificateInfo
	ch <- c.CertificateExpireDate
	ch <- c.ActiveConnections
	ch <- c.ConnectionPerSec
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ProxyLBCollector) Collect(ch chan<- prometheus.Metric) {
	proxyLBs, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("proxylb").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list proxyLBs",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(proxyLBs))

	for i := range proxyLBs {
		go func(proxyLB *sacloud.ProxyLB) {
			defer wg.Done()

			proxyLBLabels := c.proxyLBLabels(proxyLB)

			var up float64
			if proxyLB.IsAvailable() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				proxyLBLabels...,
			)

			for i := range proxyLB.Settings.ProxyLB.BindPorts {
				wg.Add(1)
				go func(index int) {
					c.collectProxyLBBindPortInfo(ch, proxyLB, index)
					wg.Done()
				}(i)
			}

			for i := range proxyLB.Settings.ProxyLB.Servers {
				wg.Add(1)
				go func(index int) {
					c.collectProxyLBServerInfo(ch, proxyLB, index)
					wg.Done()
				}(i)
			}

			wg.Add(1)
			go func() {
				c.collectProxyLBInfo(ch, proxyLB)
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				c.collectProxyLBCertInfo(ch, proxyLB)
				wg.Done()
			}()

			if proxyLB.IsAvailable() {
				now := time.Now()

				wg.Add(1)
				go func() {
					c.collectProxyLBMetrics(ch, proxyLB, now)
					wg.Done()
				}()
			}

		}(proxyLBs[i])
	}

	wg.Wait()
}

func (c *ProxyLBCollector) proxyLBLabels(proxyLB *sacloud.ProxyLB) []string {
	return []string{
		proxyLB.GetStrID(),
		proxyLB.Name,
	}
}

func (c *ProxyLBCollector) collectProxyLBInfo(ch chan<- prometheus.Metric, proxyLB *sacloud.ProxyLB) {
	sorryServerPort := ""
	if proxyLB.Settings.ProxyLB.SorryServer.Port != nil {
		sorryServerPort = fmt.Sprintf("%d", *proxyLB.Settings.ProxyLB.SorryServer.Port)
	}

	labels := append(c.proxyLBLabels(proxyLB),
		fmt.Sprintf("%d", int(proxyLB.GetPlan())),
		proxyLB.Status.VirtualIPAddress,
		proxyLB.Status.FQDN,
		flattenStringSlice(proxyLB.Status.ProxyNetworks),
		proxyLB.Settings.ProxyLB.SorryServer.IPAddress,
		sorryServerPort,
		flattenStringSlice(proxyLB.Tags),
		proxyLB.Description,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProxyLBInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *ProxyLBCollector) collectProxyLBBindPortInfo(ch chan<- prometheus.Metric, proxyLB *sacloud.ProxyLB, index int) {
	bindPort := proxyLB.Settings.ProxyLB.BindPorts[index]
	labels := append(c.proxyLBLabels(proxyLB),
		fmt.Sprintf("%d", index),
		bindPort.ProxyMode,
		fmt.Sprintf("%d", bindPort.Port),
	)

	ch <- prometheus.MustNewConstMetric(
		c.BindPortInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *ProxyLBCollector) collectProxyLBServerInfo(ch chan<- prometheus.Metric, proxyLB *sacloud.ProxyLB, index int) {
	server := proxyLB.Settings.ProxyLB.Servers[index]
	var enabled = "0"
	if server.Enabled {
		enabled = "1"
	}
	labels := append(c.proxyLBLabels(proxyLB),
		fmt.Sprintf("%d", index),
		server.IPAddress,
		fmt.Sprintf("%d", server.Port),
		enabled,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServerInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *ProxyLBCollector) collectProxyLBCertInfo(ch chan<- prometheus.Metric, proxyLB *sacloud.ProxyLB) {
	cert, err := c.client.GetCertificate(proxyLB.ID)
	if err != nil {
		c.errors.WithLabelValues("proxylb").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get certificate: proxyLB=%d", proxyLB.ID),
			"err", err,
		)
		return
	}
	if cert.PrivateKey == "" || cert.ServerCertificate == "" {
		// cert is not registered
		return
	}

	var commonName, issuerName string
	block, _ := pem.Decode([]byte(cert.ServerCertificate))
	if block != nil {
		c, err := x509.ParseCertificate(block.Bytes) // ignore err
		if err == nil {
			commonName = c.Subject.CommonName
			issuerName = c.Issuer.CommonName
		}
	}

	certLabels := append(c.proxyLBLabels(proxyLB), "0")
	infoLabels := append(certLabels, commonName, issuerName)

	ch <- prometheus.MustNewConstMetric(
		c.CertificateInfo,
		prometheus.GaugeValue,
		float64(1.0),
		infoLabels...,
	)

	// expire date
	ch <- prometheus.MustNewConstMetric(
		c.CertificateExpireDate,
		prometheus.GaugeValue,
		float64(cert.CertificateEndDate.Unix())*1000,
		certLabels...,
	)

	for i, cert := range cert.AdditionalCerts {
		var commonName, issuerName string
		block, _ := pem.Decode([]byte(cert.ServerCertificate))
		if block != nil {
			c, err := x509.ParseCertificate(block.Bytes) // ignore err
			if err == nil {
				commonName = c.Subject.CommonName
				issuerName = c.Issuer.CommonName
			}
		}

		certLabels := append(c.proxyLBLabels(proxyLB), fmt.Sprintf("%d", i+1))
		infoLabels := append(certLabels, commonName, issuerName)

		ch <- prometheus.MustNewConstMetric(
			c.CertificateInfo,
			prometheus.GaugeValue,
			float64(1.0),
			infoLabels...,
		)

		// expire date
		ch <- prometheus.MustNewConstMetric(
			c.CertificateExpireDate,
			prometheus.GaugeValue,
			float64(cert.CertificateEndDate.Unix())*1000,
			certLabels...,
		)
	}
}

func (c *ProxyLBCollector) collectProxyLBMetrics(ch chan<- prometheus.Metric, proxyLB *sacloud.ProxyLB, now time.Time) {

	values, err := c.client.Monitor(proxyLB.ID, now)
	if err != nil {
		c.errors.WithLabelValues("proxylb").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get proxyLB's metrics: ProxyLBID=%d", proxyLB.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	if values.ActiveConnections != nil {
		m := prometheus.MustNewConstMetric(
			c.ActiveConnections,
			prometheus.GaugeValue,
			values.ActiveConnections.Value,
			c.proxyLBLabels(proxyLB)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.ActiveConnections.Time, m)
	}
	if values.ConnectionsPerSec != nil {
		m := prometheus.MustNewConstMetric(
			c.ConnectionPerSec,
			prometheus.GaugeValue,
			values.ConnectionsPerSec.Value,
			c.proxyLBLabels(proxyLB)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.ConnectionsPerSec.Time, m)
	}
}
