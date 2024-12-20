// Copyright 2019-2023 The sakuracloud_exporter Authors
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
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyProxyLBClient struct {
	find       []*iaas.ProxyLB
	findErr    error
	cert       *iaas.ProxyLBCertificates
	certErr    error
	monitor    *iaas.MonitorConnectionValue
	monitorErr error
}

func (d *dummyProxyLBClient) Find(ctx context.Context) ([]*iaas.ProxyLB, error) {
	return d.find, d.findErr
}
func (d *dummyProxyLBClient) GetCertificate(ctx context.Context, id types.ID) (*iaas.ProxyLBCertificates, error) {
	return d.cert, d.certErr
}
func (d *dummyProxyLBClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorConnectionValue, error) {
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
		in             platform.ProxyLBClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyProxyLBClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=WARN msg="can't list proxyLBs" err=dummy`},
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
				find: []*iaas.ProxyLB{
					{
						ID:           101,
						Name:         "proxylb",
						Description:  "desc",
						Tags:         types.Tags{"tag1", "tag2"},
						Availability: types.Availabilities.Available,
						Plan:         types.ProxyLBPlans.CPS100,
						HealthCheck: &iaas.ProxyLBHealthCheck{
							Protocol:  types.ProxyLBProtocols.HTTP,
							Path:      "/",
							DelayLoop: 10,
						},
						SorryServer: &iaas.ProxyLBSorryServer{
							IPAddress: "192.168.0.21",
							Port:      80,
						},
						BindPorts: []*iaas.ProxyLBBindPort{
							{
								ProxyMode: types.ProxyLBProxyModes.HTTP,
								Port:      80,
							},
							{
								ProxyMode: types.ProxyLBProxyModes.HTTPS,
								Port:      443,
							},
						},
						Servers: []*iaas.ProxyLBServer{
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
				cert: &iaas.ProxyLBCertificates{
					PrimaryCert: &iaas.ProxyLBPrimaryCert{
						ServerCertificate: `-----BEGIN CERTIFICATE-----
MIIDzTCCArWgAwIBAgIUZllLmMvTzLYyCQPqmnmt8zfOdVowDQYJKoZIhvcNAQEL
BQAwdjELMAkGA1UEBhMCSlAxDjAMBgNVBAgMBVRva3lvMREwDwYDVQQHDAhTaGlu
anVrdTEaMBgGA1UECgwRVGVzdCBPcmdhbml6YXRpb24xEjAQBgNVBAsMCVRlc3Qg
VW5pdDEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjQxMjIwMDcwNTUyWhcNMjUx
MjIwMDcwNTUyWjB2MQswCQYDVQQGEwJKUDEOMAwGA1UECAwFVG9reW8xETAPBgNV
BAcMCFNoaW5qdWt1MRowGAYDVQQKDBFUZXN0IE9yZ2FuaXphdGlvbjESMBAGA1UE
CwwJVGVzdCBVbml0MRQwEgYDVQQDDAtleGFtcGxlLmNvbTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBANhpoUrynlFZDXVVlr7XYYUYRVRnPDNsHGKopF81
6V63WosAJpIz+8biFFA+OfwX2b/VX2VsE4Nakg5TGnxtEe+LFi5bphrbGmLFsxoT
8IMFu4qEKrybI+61jdkvhDWd5D82dohkE4poOvGePqrEhECREWQ17d5Oqc9cj39d
rerBfY2j9k+w0PxYtdQo7+FrBfQBOxMmDVqY1umTZTswfTn8sXsugqn4UrHrBtYd
O1/MeFsx4c63n48D5DepquvBmwnTa9ccnHbrdIItWs7BwgJKbDt7NJ1rtTED/1G9
xnk/pld2iPySqGLlPRyqETtMNcdyx3KfkOnH7Q5H17Wi1kMCAwEAAaNTMFEwHQYD
VR0OBBYEFO0w5+4Hp1fkxLAThWyLF5v4sC61MB8GA1UdIwQYMBaAFO0w5+4Hp1fk
xLAThWyLF5v4sC61MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AKMEfy6bA0/d7yNTEXssPpEhC7/XolkAqKntl741TQ0mgJAkgeUGIfFkFNioCeQc
m3Aqam6IsMyHcZwo9gJR4KnE02N+jQpLJbDw8ym2BwCpF9g43x5K9qzvFEml4Idg
nq9UP0T4Yz1eKvCmVCm8cApVqr02TYnYMg9Oo3QE0giPIEHdG0mDuWM46eDAzoLL
8ib9EPnmyswhfNzSZyoH5nNV8137VOwPGtcBAg8fmdO+hmOVgEU5OGxz3U26toi5
yfHUC+O5jhCLSTAJwd2RWeCMEcN9FVI1IaGZV2WxrbXC+/5qZTjSvdvrmVbVAAd2
ybZBwFTVijAdTHYmC1VNSxQ=
-----END CERTIFICATE-----`,
						IntermediateCertificate: `-----BEGIN CERTIFICATE-----
MIID0jCCArqgAwIBAgIUcCPU6qCiTPDVQ1LW9bePo+PMQKEwDQYJKoZIhvcNAQEL
BQAwdjELMAkGA1UEBhMCSlAxDjAMBgNVBAgMBVRva3lvMREwDwYDVQQHDAhTaGlu
anVrdTEaMBgGA1UECgwRVGVzdCBPcmdhbml6YXRpb24xEjAQBgNVBAsMCVRlc3Qg
VW5pdDEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjQxMjIwMDcwNjE0WhcNMjUx
MjIwMDcwNjE0WjCBizELMAkGA1UEBhMCSlAxDjAMBgNVBAgMBVRva3lvMREwDwYD
VQQHDAhTaGluanVrdTEaMBgGA1UECgwRVGVzdCBPcmdhbml6YXRpb24xGjAYBgNV
BAsMEUludGVybWVkaWF0ZSBVbml0MSEwHwYDVQQDDBhpbnRlcm1lZGlhdGUuZXhh
bXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDoW7sb/ahy
PdoC+duRXjGoNp2caCTS02JcxMjFzE3yKj/p+SFNr7ufNTxMRIGcFLzmgYHRo0C/
MWYXPF5aKwhZYtln2ur87NErrZfPT/8xBdY/H5fJOpKyBB/ByfnIeYgFBBkZRCfT
Dytu/WOZENTsd1JAiirzM3xXvlopwiICsQ3JyMNfcbvYPQqLIY6Aynj1S5+aJDpg
x/F+n1r7Ji1egpfblaIMeX0Q0goDLNEfGESzFbbqFzs5OTBpexknbST9yNH6Fb9u
Tv4MEhjDsjYkDvjIV1QN0PH8R63toclcp2P3bMOxczkds5EKSVMkS5Bp2+rYNbK+
Hgpdt6wR5U/rAgMBAAGjQjBAMB0GA1UdDgQWBBSqCuNjLuEFj02uRKSYSQvzk9sy
KzAfBgNVHSMEGDAWgBTtMOfuB6dX5MSwE4Vsixeb+LAutTANBgkqhkiG9w0BAQsF
AAOCAQEAfJT2uxSjAOfClYrj9atjxDz8EVaELLNTZEL1QNBqFs1nD82MHRQs2Fyr
mMwaeIlDyIO44kyL9A/jFFi9yQP5li0qCTNQZu7bz1PWhsvnfJCJEQkkRLS9X8ZL
b03jOlsWrse0dFSY3wHJk/SmBUnO/VAdx6wZu/jf6mxar48nSm+3lxvRKNRutiPZ
96wjwUKr1ExOz3Ju2hf+/akUn1byb4hgVsF17TCy6zP7rSfdOknhuKwNc0KQLhsL
PLS5Ur7caLG/BVSX/kYVreYeHVw7yU5BWL9iAktSmckC+CMk5iw9XAE152kK2JUH
la0yTEo/WpPQtAxgXuxLy5XiiLSfNQ==
-----END CERTIFICATE-----`,
						PrivateKey: `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDYaaFK8p5RWQ11
VZa+12GFGEVUZzwzbBxiqKRfNelet1qLACaSM/vG4hRQPjn8F9m/1V9lbBODWpIO
Uxp8bRHvixYuW6Ya2xpixbMaE/CDBbuKhCq8myPutY3ZL4Q1neQ/NnaIZBOKaDrx
nj6qxIRAkRFkNe3eTqnPXI9/Xa3qwX2No/ZPsND8WLXUKO/hawX0ATsTJg1amNbp
k2U7MH05/LF7LoKp+FKx6wbWHTtfzHhbMeHOt5+PA+Q3qarrwZsJ02vXHJx263SC
LVrOwcICSmw7ezSda7UxA/9RvcZ5P6ZXdoj8kqhi5T0cqhE7TDXHcsdyn5Dpx+0O
R9e1otZDAgMBAAECggEAAgSmKOqZ+FySErnhcIErWyWOoUq0gIRDFYEeG0yHkxxh
n0c5P4bK6SAQRsP1ys3heCJXbpIzR7fPVzaGL4qILsmHbkI+NSToROs0EDZcY/8T
MH0ACwc6rri0YcX0KoMrmZKlSKtU6qcDLwql6fYa3PadXhJfWAHiyoNsoShF/W5G
cVieewukl1SVg/k5zIggIL/TZZ5ac8gSffrbeW+D4/0eKqWK2ZifeIF+XAxWGFMx
VSIQKsWSJy75rp5YmwIW12zvaX9PGClDQE55tha3U8N0/a49IVbr67D/SxeHhsPK
dVxIyVqK69jfq8nxnpk/NnhkHi7QFWw8g8JapY9lvQKBgQD0N1Dw8Wbj6qh3Ilo/
aF1jGDeWsF/1PcnuoFom85Bu4YDAT92uzxPT/is8ceJcL90lEVSn3EdDM30oJGr0
tONDKN1kB4iIayt/dBsjWCfOPW0jvJHjSXa4PTeqdxWUIrA5f5YklTbLodfA0EQq
RMDhXgEMujpDwCW9wbWXxiRP1wKBgQDi2t7/Fg0zzLX21YUQN3Y3nDVezoc73ph6
qNvsjAuRuthyjzLhM2zhFTiw9mdaDu7XKo/1ZCHse3JEKz+YZVn5I05XNxYrTkVD
xhHuEG1grxqKjMcMq4yUiUy4yY68TX3PEy9JrYV+n8S75hNaaf/I+fsm5c1rOP9o
5U4DX/FPdQKBgQCjPta8OKGueI1kFXJ+MCU8uFNwRzXdmRACku2wW9+QPuzxoHFv
CL0YWC5OmVHWjaglvw/3pSd9pE1lJ/LW4JOJsSdMVjzN89V/vPznA2aYVjc+TC64
38KcJU+wgynJe+aQiNi0W4nlVKoEGTN3jb3g6BWLjHCmGSshTPs2GRzswQKBgD3h
gFTK2h0YKUbEpcBvsJKozLIo2iDNroA/EYasCPfepO5S+4kMsxWO6WD0Rer+Cc6t
sIk6oDpWziukNHvIocthAxytTSHQ/vnmzLtIxd1Kxo2mqyFcpkNaVJBPgt0AsmHL
FOofKDwLLuomb38JTRmwfv70Tp2B9cHSUv5+rF+FAoGBALW3gmfRI3+Ax1scn5QQ
CIKJ5Fd6mrntdmXkW1NWz/DR0a04wSB/eiCz8J8KGfx78/S44JUyDZQq6S+1JMzL
+Cv2dgc5wG2swzUeTgA/0khXuJ6r17zLDIolXnTfXQ77y6dW/li6qUJyfTCFXe+A
9Ncpwbw8KmFw9wHm5eVAk/nz
-----END PRIVATE KEY-----`,
						CertificateEndDate:    time.Now().AddDate(1, 0, 0),
						CertificateCommonName: "example.com",
					},
				},
				monitor: &iaas.MonitorConnectionValue{
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
				{
					desc: c.CertificateInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "proxylb",
						"cert_index":  "0",
						"common_name": "example.com",
						"issuer_name": "example.com",
					}),
				},
				{
					desc: c.CertificateExpireDate,
					metric: createGaugeMetric(float64(time.Now().AddDate(1, 0, 0).Unix())*1000, map[string]string{
						"id":         "101",
						"name":       "proxylb",
						"cert_index": "0",
					}),
				},
			},
		},
		{
			name: "activity monitor APIs return error",
			in: &dummyProxyLBClient{
				find: []*iaas.ProxyLB{
					{
						ID:           101,
						Name:         "proxylb",
						Description:  "desc",
						Tags:         types.Tags{"tag1", "tag2"},
						Availability: types.Availabilities.Available,
						Plan:         types.ProxyLBPlans.CPS100,
						HealthCheck: &iaas.ProxyLBHealthCheck{
							Protocol:  types.ProxyLBProtocols.HTTP,
							Path:      "/",
							DelayLoop: 10,
						},
						SorryServer: &iaas.ProxyLBSorryServer{
							IPAddress: "192.168.0.21",
							Port:      80,
						},
						BindPorts: []*iaas.ProxyLBBindPort{
							{
								ProxyMode: types.ProxyLBProxyModes.HTTP,
								Port:      80,
							},
							{
								ProxyMode: types.ProxyLBProxyModes.HTTPS,
								Port:      443,
							},
						},
						Servers: []*iaas.ProxyLBServer{
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
				`level=WARN msg="can't get certificate: proxyLB=101" err=dummy2`,
				`level=WARN msg="can't get proxyLB's metrics: ProxyLBID=101" err=dummy3`,
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
