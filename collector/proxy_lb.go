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

package collector

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// ProxyLBCollector collects metrics about all proxyLBs.
type ProxyLBCollector struct {
	ctx    context.Context
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
func NewProxyLBCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.ProxyLBClient) *ProxyLBCollector {
	errors.WithLabelValues("proxylb").Add(0)

	proxyLBLabels := []string{"id", "name"}
	proxyLBInfoLabels := append(proxyLBLabels, "plan", "vip", "fqdn",
		"proxy_networks", "sorry_server_ipaddress", "sorry_server_port", "tags", "description")

	proxyLBBindPortLabels := append(proxyLBLabels, "bind_port_index", "proxy_mode", "port")
	proxyLBServerLabels := append(proxyLBLabels, "server_index", "ipaddress", "port", "enabled")
	proxyLBCertificateLabels := append(proxyLBLabels, "cert_index")
	proxyLBCertificateInfoLabels := append(proxyLBCertificateLabels, "common_name", "issuer_name")

	return &ProxyLBCollector{
		ctx:    ctx,
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
	proxyLBs, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("proxylb").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", "can't list proxyLBs",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(proxyLBs))

	for i := range proxyLBs {
		func(proxyLB *sacloud.ProxyLB) {
			defer wg.Done()

			proxyLBLabels := c.proxyLBLabels(proxyLB)

			var up float64
			if proxyLB.Availability.IsAvailable() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				proxyLBLabels...,
			)

			for i := range proxyLB.BindPorts {
				wg.Add(1)
				go func(index int) {
					c.collectProxyLBBindPortInfo(ch, proxyLB, index)
					wg.Done()
				}(i)
			}

			for i := range proxyLB.Servers {
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

			if proxyLB.Availability.IsAvailable() {
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
		proxyLB.ID.String(),
		proxyLB.Name,
	}
}

func (c *ProxyLBCollector) collectProxyLBInfo(ch chan<- prometheus.Metric, proxyLB *sacloud.ProxyLB) {
	sorryServerPort := ""
	if proxyLB.SorryServer.Port > 0 {
		sorryServerPort = fmt.Sprintf("%d", proxyLB.SorryServer.Port)
	}

	labels := append(c.proxyLBLabels(proxyLB),
		fmt.Sprintf("%d", int(proxyLB.GetPlan())),
		proxyLB.VirtualIPAddress,
		proxyLB.FQDN,
		flattenStringSlice(proxyLB.ProxyNetworks),
		proxyLB.SorryServer.IPAddress,
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
	bindPort := proxyLB.BindPorts[index]
	labels := append(c.proxyLBLabels(proxyLB),
		fmt.Sprintf("%d", index),
		string(bindPort.ProxyMode),
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
	server := proxyLB.Servers[index]
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
	cert, err := c.client.GetCertificate(c.ctx, proxyLB.ID)
	if err != nil {
		c.errors.WithLabelValues("proxylb").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get certificate: proxyLB=%d", proxyLB.ID),
			"err", err,
		)
		return
	}
	if cert == nil {
		return
	}
	if cert.PrimaryCert.PrivateKey == "" || cert.PrimaryCert.ServerCertificate == "" {
		// cert is not registered
		return
	}

	var commonName, issuerName string
	block, _ := pem.Decode([]byte(cert.PrimaryCert.ServerCertificate))
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
		float64(cert.PrimaryCert.CertificateEndDate.Unix())*1000,
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
	values, err := c.client.Monitor(c.ctx, proxyLB.ID, now)
	if err != nil {
		c.errors.WithLabelValues("proxylb").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get proxyLB's metrics: ProxyLBID=%d", proxyLB.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	m := prometheus.MustNewConstMetric(
		c.ActiveConnections,
		prometheus.GaugeValue,
		values.ActiveConnections,
		c.proxyLBLabels(proxyLB)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
	m = prometheus.MustNewConstMetric(
		c.ConnectionPerSec,
		prometheus.GaugeValue,
		values.ConnectionsPerSec,
		c.proxyLBLabels(proxyLB)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}
