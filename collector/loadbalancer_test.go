package collector

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyLoadBalancerClient struct {
	find       []*iaas.LoadBalancer
	findErr    error
	status     []*sacloud.LoadBalancerStatus
	statusErr  error
	monitor    *sacloud.MonitorInterfaceValue
	monitorErr error
}

func (d *dummyLoadBalancerClient) Find(ctx context.Context) ([]*iaas.LoadBalancer, error) {
	return d.find, d.findErr
}
func (d *dummyLoadBalancerClient) Status(ctx context.Context, zone string, id types.ID) ([]*sacloud.LoadBalancerStatus, error) {
	return d.status, d.statusErr
}
func (d *dummyLoadBalancerClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	return d.monitor, d.monitorErr
}

func TestLoadBalancerCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewLoadBalancerCollector(context.Background(), testLogger, testErrors, &dummyLoadBalancerClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.LoadBalancerInfo,
		c.Receive,
		c.Send,
		c.VIPInfo,
		c.VIPCPS,
		c.ServerInfo,
		c.ServerUp,
		c.ServerConnection,
		c.ServerCPS,
	}))
}

func TestLoadBalancerCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewLoadBalancerCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             iaas.LoadBalancerClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyLoadBalancerClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list loadbalancers" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyLoadBalancerClient{},
			wantMetrics: nil,
		},
		{
			name: "a load balancer",
			in: &dummyLoadBalancerClient{
				find: []*iaas.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &sacloud.LoadBalancer{
							ID:             101,
							Name:           "loadbalancer",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							PlanID:         types.LoadBalancerPlans.Standard,
							VRID:           1,
							IPAddresses:    []string{"192.168.0.11"},
							DefaultRoute:   "192.168.0.1",
							NetworkMaskLen: 24,
							InstanceStatus: types.ServerInstanceStatuses.Up,
						},
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "standard",
						"ha":          "0",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
			},
		},
		{
			name: "a highspec load balancer with activity monitors",
			in: &dummyLoadBalancerClient{
				find: []*iaas.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &sacloud.LoadBalancer{
							ID:             101,
							Name:           "loadbalancer",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							PlanID:         types.LoadBalancerPlans.Premium,
							VRID:           1,
							IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
							DefaultRoute:   "192.168.0.1",
							NetworkMaskLen: 24,
							InstanceStatus: types.ServerInstanceStatuses.Up,
							VirtualIPAddresses: []*sacloud.LoadBalancerVirtualIPAddress{
								{
									VirtualIPAddress: "192.168.0.101",
									Port:             80,
									DelayLoop:        100,
									SorryServer:      "192.168.0.21",
									Description:      "vip-desc",
									Servers: []*sacloud.LoadBalancerServer{
										{
											IPAddress: "192.168.0.201",
											Port:      80,
											Enabled:   true,
											HealthCheck: &sacloud.LoadBalancerServerHealthCheck{
												Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
												ResponseCode: http.StatusOK,
												Path:         "/index.html",
											},
										},
									},
								},
							},
						},
					},
				},
				status: []*sacloud.LoadBalancerStatus{
					{
						VirtualIPAddress: "192.168.0.101",
						Port:             80,
						CPS:              100,
						Servers: []*sacloud.LoadBalancerServerStatus{
							{
								IPAddress:  "192.168.0.201",
								Port:       80,
								Status:     types.ServerInstanceStatuses.Up,
								CPS:        200,
								ActiveConn: 300,
							},
						},
					},
				},
				monitor: &sacloud.MonitorInterfaceValue{
					Time:    monitorTime,
					Receive: 100,
					Send:    200,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "highspec",
						"ha":          "1",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "192.168.0.12",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.VIPInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"port":         "80",
						"interval":     "100",
						"sorry_server": "192.168.0.21",
						"description":  "vip-desc",
					}),
				},
				{
					desc: c.Receive,
					metric: createGaugeWithTimestamp(float64(100)*8/1000, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.Send,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.VIPCPS,
					metric: createGaugeMetric(100, map[string]string{
						"id":        "101",
						"name":      "loadbalancer",
						"zone":      "is1a",
						"vip_index": "0",
						"vip":       "192.168.0.101",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "loadbalancer",
						"zone":          "is1a",
						"vip_index":     "0",
						"vip":           "192.168.0.101",
						"server_index":  "0",
						"ipaddress":     "192.168.0.201",
						"monitor":       "http",
						"path":          "/index.html",
						"response_code": "200",
					}),
				},
				{
					desc: c.ServerUp,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"server_index": "0",
						"ipaddress":    "192.168.0.201",
					}),
				},
				{
					desc: c.ServerCPS,
					metric: createGaugeMetric(200, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"server_index": "0",
						"ipaddress":    "192.168.0.201",
					}),
				},
				{
					desc: c.ServerConnection,
					metric: createGaugeMetric(300, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"server_index": "0",
						"ipaddress":    "192.168.0.201",
					}),
				},
			},
		},
		{
			name: "status and monitor API return error",
			in: &dummyLoadBalancerClient{
				find: []*iaas.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &sacloud.LoadBalancer{
							ID:             101,
							Name:           "loadbalancer",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							PlanID:         types.LoadBalancerPlans.Premium,
							VRID:           1,
							IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
							DefaultRoute:   "192.168.0.1",
							NetworkMaskLen: 24,
							InstanceStatus: types.ServerInstanceStatuses.Up,
							VirtualIPAddresses: []*sacloud.LoadBalancerVirtualIPAddress{
								{
									VirtualIPAddress: "192.168.0.101",
									Port:             80,
									DelayLoop:        100,
									SorryServer:      "192.168.0.21",
									Description:      "vip-desc",
									Servers: []*sacloud.LoadBalancerServer{
										{
											IPAddress: "192.168.0.201",
											Port:      80,
											Enabled:   true,
											HealthCheck: &sacloud.LoadBalancerServerHealthCheck{
												Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
												ResponseCode: http.StatusOK,
												Path:         "/index.html",
											},
										},
									},
								},
							},
						},
					},
				},
				statusErr:  errors.New("dummy1"),
				monitorErr: errors.New("dummy2"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "highspec",
						"ha":          "1",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "192.168.0.12",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.VIPInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"port":         "80",
						"interval":     "100",
						"sorry_server": "192.168.0.21",
						"description":  "vip-desc",
					}),
				},
			},
			wantLogs: []string{
				`level=warn msg="can't fetch loadbalancer's status: ID: 101" err=dummy1`,
				`level=warn msg="can't get loadbalancer's NIC metrics: ID=101" err=dummy2`,
			},
			wantErrCounter: 2,
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "loadbalancer")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
