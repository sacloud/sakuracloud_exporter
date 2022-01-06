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
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

// TODO test certificates(common name and expire date)

type dummyProxyLBClient struct {
	find       []*sacloud.ProxyLB
	findErr    error
	cert       *sacloud.ProxyLBCertificates
	certErr    error
	monitor    *sacloud.MonitorConnectionValue
	monitorErr error
}

func (d *dummyProxyLBClient) Find(ctx context.Context) ([]*sacloud.ProxyLB, error) {
	return d.find, d.findErr
}
func (d *dummyProxyLBClient) GetCertificate(ctx context.Context, id types.ID) (*sacloud.ProxyLBCertificates, error) {
	return d.cert, d.certErr
}
func (d *dummyProxyLBClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorConnectionValue, error) {
	return d.monitor, d.monitorErr
}

func TestProxyLBCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewProxyLBCollector(context.Background(), testLogger, testErrors, &dummyProxyLBClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.ProxyLBInfo,
		c.BindPortInfo,
		c.ServerInfo,
		c.CertificateInfo,
		c.CertificateExpireDate,
		c.ActiveConnections,
		c.ConnectionPerSec,
	}))
}

func TestProxyLBCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewProxyLBCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             iaas.ProxyLBClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyProxyLBClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list proxyLBs" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyProxyLBClient{},
			wantMetrics: nil,
		},
		{
			name: "a proxyLB",
			in: &dummyProxyLBClient{
				find: []*sacloud.ProxyLB{
					{
						ID:           101,
						Name:         "proxylb",
						Description:  "desc",
						Tags:         types.Tags{"tag1", "tag2"},
						Availability: types.Availabilities.Available,
						Plan:         types.ProxyLBPlans.CPS100,
						HealthCheck: &sacloud.ProxyLBHealthCheck{
							Protocol:  types.ProxyLBProtocols.HTTP,
							Path:      "/",
							DelayLoop: 10,
						},
						SorryServer: &sacloud.ProxyLBSorryServer{
							IPAddress: "192.168.0.21",
							Port:      80,
						},
						BindPorts: []*sacloud.ProxyLBBindPort{
							{
								ProxyMode: types.ProxyLBProxyModes.HTTP,
								Port:      80,
							},
							{
								ProxyMode: types.ProxyLBProxyModes.HTTPS,
								Port:      443,
							},
						},
						Servers: []*sacloud.ProxyLBServer{
							{
								IPAddress: "192.168.0.101",
								Port:      80,
								Enabled:   true,
							},
						},
						UseVIPFailover:   true,
						Region:           types.ProxyLBRegions.TK1,
						ProxyNetworks:    []string{"133.242.0.0/24"},
						FQDN:             "site-xxx.proxylb.sakura.ne.jp",
						VirtualIPAddress: "192.0.2.1",
					},
				},
				cert: &sacloud.ProxyLBCertificates{
					PrimaryCert: &sacloud.ProxyLBPrimaryCert{
						ServerCertificate:       "",
						IntermediateCertificate: "",
						PrivateKey:              "",
						CertificateEndDate:      time.Time{},
						CertificateCommonName:   "",
					},
				},
				monitor: &sacloud.MonitorConnectionValue{
					Time:              monitorTime,
					ActiveConnections: 100,
					ConnectionsPerSec: 200,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "proxylb",
					}),
				},
				{
					desc: c.BindPortInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "proxylb",
						"bind_port_index": "0",
						"proxy_mode":      "http",
						"port":            "80",
					}),
				},
				{
					desc: c.BindPortInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "proxylb",
						"bind_port_index": "1",
						"proxy_mode":      "https",
						"port":            "443",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "proxylb",
						"server_index": "0",
						"ipaddress":    "192.168.0.101",
						"port":         "80",
						"enabled":      "1",
					}),
				},
				{
					desc: c.ProxyLBInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                     "101",
						"name":                   "proxylb",
						"plan":                   "100",
						"vip":                    "192.0.2.1",
						"fqdn":                   "site-xxx.proxylb.sakura.ne.jp",
						"proxy_networks":         ",133.242.0.0/24,",
						"sorry_server_ipaddress": "192.168.0.21",
						"sorry_server_port":      "80",
						"tags":                   ",tag1,tag2,",
						"description":            "desc",
					}),
				},
				{
					desc: c.ActiveConnections,
					metric: createGaugeWithTimestamp(100, map[string]string{
						"id":   "101",
						"name": "proxylb",
					}, monitorTime),
				},
				{
					desc: c.ConnectionPerSec,
					metric: createGaugeWithTimestamp(200, map[string]string{
						"id":   "101",
						"name": "proxylb",
					}, monitorTime),
				},
			},
		},
		{
			name: "activity monitor APIs return error",
			in: &dummyProxyLBClient{
				find: []*sacloud.ProxyLB{
					{
						ID:           101,
						Name:         "proxylb",
						Description:  "desc",
						Tags:         types.Tags{"tag1", "tag2"},
						Availability: types.Availabilities.Available,
						Plan:         types.ProxyLBPlans.CPS100,
						HealthCheck: &sacloud.ProxyLBHealthCheck{
							Protocol:  types.ProxyLBProtocols.HTTP,
							Path:      "/",
							DelayLoop: 10,
						},
						SorryServer: &sacloud.ProxyLBSorryServer{
							IPAddress: "192.168.0.21",
							Port:      80,
						},
						BindPorts: []*sacloud.ProxyLBBindPort{
							{
								ProxyMode: types.ProxyLBProxyModes.HTTP,
								Port:      80,
							},
							{
								ProxyMode: types.ProxyLBProxyModes.HTTPS,
								Port:      443,
							},
						},
						Servers: []*sacloud.ProxyLBServer{
							{
								IPAddress: "192.168.0.101",
								Port:      80,
								Enabled:   true,
							},
						},
						UseVIPFailover:   true,
						Region:           types.ProxyLBRegions.TK1,
						ProxyNetworks:    []string{"133.242.0.0/24"},
						FQDN:             "site-xxx.proxylb.sakura.ne.jp",
						VirtualIPAddress: "192.0.2.1",
					},
				},
				certErr:    errors.New("dummy2"),
				monitorErr: errors.New("dummy3"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "proxylb",
					}),
				},
				{
					desc: c.BindPortInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "proxylb",
						"bind_port_index": "0",
						"proxy_mode":      "http",
						"port":            "80",
					}),
				},
				{
					desc: c.BindPortInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "proxylb",
						"bind_port_index": "1",
						"proxy_mode":      "https",
						"port":            "443",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "proxylb",
						"server_index": "0",
						"ipaddress":    "192.168.0.101",
						"port":         "80",
						"enabled":      "1",
					}),
				},
				{
					desc: c.ProxyLBInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                     "101",
						"name":                   "proxylb",
						"plan":                   "100",
						"vip":                    "192.0.2.1",
						"fqdn":                   "site-xxx.proxylb.sakura.ne.jp",
						"proxy_networks":         ",133.242.0.0/24,",
						"sorry_server_ipaddress": "192.168.0.21",
						"sorry_server_port":      "80",
						"tags":                   ",tag1,tag2,",
						"description":            "desc",
					}),
				},
			},
			wantErrCounter: 2,
			wantLogs: []string{
				`level=warn msg="can't get certificate: proxyLB=101" err=dummy2`,
				`level=warn msg="can't get proxyLB's metrics: ProxyLBID=101" err=dummy3`,
			},
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "proxylb")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
