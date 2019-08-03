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

type dummyMobileGatewayClient struct {
	find              []*iaas.MobileGateway
	findErr           error
	trafficStatus     *sacloud.MobileGatewayTrafficStatus
	trafficStatusErr  error
	trafficControl    *sacloud.MobileGatewayTrafficControl
	trafficControlErr error
	monitor           *sacloud.MonitorInterfaceValue
	monitorErr        error
}

func (d *dummyMobileGatewayClient) Find(ctx context.Context) ([]*iaas.MobileGateway, error) {
	return d.find, d.findErr
}
func (d *dummyMobileGatewayClient) TrafficStatus(ctx context.Context, zone string, id types.ID) (*sacloud.MobileGatewayTrafficStatus, error) {
	return d.trafficStatus, d.trafficStatusErr
}
func (d *dummyMobileGatewayClient) TrafficControl(ctx context.Context, zone string, id types.ID) (*sacloud.MobileGatewayTrafficControl, error) {
	return d.trafficControl, d.trafficControlErr
}
func (d *dummyMobileGatewayClient) MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	return d.monitor, d.monitorErr
}

func TestMobileGatewayCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewMobileGatewayCollector(context.Background(), testLogger, testErrors, &dummyMobileGatewayClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.MobileGatewayInfo,
		c.Receive,
		c.Send,
		c.TrafficControlInfo,
		c.TrafficUplink,
		c.TrafficDownlink,
		c.TrafficShaping,
	}))
}

func TestMobileGatewayCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewMobileGatewayCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             iaas.MobileGatewayClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyMobileGatewayClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list mobile_gateways" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyMobileGatewayClient{},
			wantMetrics: nil,
		},
		{
			name: "a mobile gateway",
			in: &dummyMobileGatewayClient{
				find: []*iaas.MobileGateway{
					{
						ZoneName: "is1a",
						MobileGateway: &sacloud.MobileGateway{
							ID:             101,
							Name:           "mobile-gateway",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							InstanceStatus: types.ServerInstanceStatuses.Down,
							Settings: &sacloud.MobileGatewaySetting{
								InternetConnectionEnabled:       false,
								InterDeviceCommunicationEnabled: false,
							},
						},
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MobileGatewayInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                         "101",
						"name":                       "mobile-gateway",
						"zone":                       "is1a",
						"internet_connection":        "0",
						"inter_device_communication": "0",
						"tags":                       ",tag1,tag2,",
						"description":                "desc",
					}),
				},
			},
		},
		{
			name: "a mobile gateway with status and activity monitor",
			in: &dummyMobileGatewayClient{
				find: []*iaas.MobileGateway{
					{
						ZoneName: "is1a",
						MobileGateway: &sacloud.MobileGateway{
							ID:             101,
							Name:           "mobile-gateway",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							InstanceStatus: types.ServerInstanceStatuses.Up,
							Availability:   types.Availabilities.Available,
							Settings: &sacloud.MobileGatewaySetting{
								InternetConnectionEnabled:       true,
								InterDeviceCommunicationEnabled: true,
							},
						},
					},
				},
				trafficControl: &sacloud.MobileGatewayTrafficControl{
					TrafficQuotaInMB:       1024,
					BandWidthLimitInKbps:   64,
					EmailNotifyEnabled:     true,
					SlackNotifyEnabled:     true,
					SlackNotifyWebhooksURL: "https://example.com",
					AutoTrafficShaping:     true,
				},
				trafficStatus: &sacloud.MobileGatewayTrafficStatus{
					UplinkBytes:    100,
					DownlinkBytes:  200,
					TrafficShaping: true,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MobileGatewayInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                         "101",
						"name":                       "mobile-gateway",
						"zone":                       "is1a",
						"internet_connection":        "1",
						"inter_device_communication": "1",
						"tags":                       ",tag1,tag2,",
						"description":                "desc",
					}),
				},
				{
					desc: c.TrafficControlInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                      "101",
						"name":                    "mobile-gateway",
						"zone":                    "is1a",
						"traffic_quota_in_mb":     "1024",
						"bandwidth_limit_in_kbps": "64",
						"enable_email":            "1",
						"enable_slack":            "1",
						"slack_url":               "https://example.com",
						"auto_traffic_shaping":    "1",
					}),
				},
				{
					desc: c.TrafficUplink,
					metric: createGaugeMetric(100, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.TrafficDownlink,
					metric: createGaugeMetric(200, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.TrafficShaping,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "a mobile gateway with multiple interface",
			in: &dummyMobileGatewayClient{
				find: []*iaas.MobileGateway{
					{
						ZoneName: "is1a",
						MobileGateway: &sacloud.MobileGateway{
							ID:             101,
							Name:           "mobile-gateway",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							InstanceStatus: types.ServerInstanceStatuses.Up,
							Availability:   types.Availabilities.Available,
							Settings: &sacloud.MobileGatewaySetting{
								InternetConnectionEnabled:       true,
								InterDeviceCommunicationEnabled: true,
							},
							Interfaces: []*sacloud.MobileGatewayInterface{
								{
									IPAddress:            "192.168.0.1",
									SubnetNetworkMaskLen: 24,
								},
								{
									IPAddress:            "192.168.1.1",
									SubnetNetworkMaskLen: 28,
								},
							},
						},
					},
				},
				trafficControl: &sacloud.MobileGatewayTrafficControl{
					TrafficQuotaInMB:       1024,
					BandWidthLimitInKbps:   64,
					EmailNotifyEnabled:     true,
					SlackNotifyEnabled:     true,
					SlackNotifyWebhooksURL: "https://example.com",
					AutoTrafficShaping:     true,
				},
				trafficStatus: &sacloud.MobileGatewayTrafficStatus{
					UplinkBytes:    100,
					DownlinkBytes:  200,
					TrafficShaping: true,
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
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MobileGatewayInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                         "101",
						"name":                       "mobile-gateway",
						"zone":                       "is1a",
						"internet_connection":        "1",
						"inter_device_communication": "1",
						"tags":                       ",tag1,tag2,",
						"description":                "desc",
					}),
				},
				{
					desc: c.TrafficControlInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                      "101",
						"name":                    "mobile-gateway",
						"zone":                    "is1a",
						"traffic_quota_in_mb":     "1024",
						"bandwidth_limit_in_kbps": "64",
						"enable_email":            "1",
						"enable_slack":            "1",
						"slack_url":               "https://example.com",
						"auto_traffic_shaping":    "1",
					}),
				},
				{
					desc: c.TrafficUplink,
					metric: createGaugeMetric(100, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.TrafficDownlink,
					metric: createGaugeMetric(200, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.TrafficShaping,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.Receive,
					metric: createGaugeWithTimestamp(float64(100)*8/1000, map[string]string{
						"id":          "101",
						"name":        "mobile-gateway",
						"zone":        "is1a",
						"nic_index":   "0",
						"ipaddress":   "192.168.0.1",
						"nw_mask_len": "24",
					}, monitorTime),
				},
				{
					desc: c.Send,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":          "101",
						"name":        "mobile-gateway",
						"zone":        "is1a",
						"nic_index":   "0",
						"ipaddress":   "192.168.0.1",
						"nw_mask_len": "24",
					}, monitorTime),
				},
				{
					desc: c.Receive,
					metric: createGaugeWithTimestamp(float64(100)*8/1000, map[string]string{
						"id":          "101",
						"name":        "mobile-gateway",
						"zone":        "is1a",
						"nic_index":   "1",
						"ipaddress":   "192.168.1.1",
						"nw_mask_len": "28",
					}, monitorTime),
				},
				{
					desc: c.Send,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":          "101",
						"name":        "mobile-gateway",
						"zone":        "is1a",
						"nic_index":   "1",
						"ipaddress":   "192.168.1.1",
						"nw_mask_len": "28",
					}, monitorTime),
				},
			},
		},
		{
			name: "status and monitor API returns error",
			in: &dummyMobileGatewayClient{
				find: []*iaas.MobileGateway{
					{
						ZoneName: "is1a",
						MobileGateway: &sacloud.MobileGateway{
							ID:             101,
							Name:           "mobile-gateway",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							InstanceStatus: types.ServerInstanceStatuses.Up,
							Availability:   types.Availabilities.Available,
							Settings: &sacloud.MobileGatewaySetting{
								InternetConnectionEnabled:       true,
								InterDeviceCommunicationEnabled: true,
							},
							Interfaces: []*sacloud.MobileGatewayInterface{
								{
									IPAddress:            "192.168.0.1",
									SubnetNetworkMaskLen: 24,
								},
								{
									IPAddress:            "192.168.1.1",
									SubnetNetworkMaskLen: 28,
								},
							},
						},
					},
				},
				trafficControlErr: errors.New("dummy1"),
				trafficStatusErr:  errors.New("dummy2"),
				monitorErr:        errors.New("dummy3"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "mobile-gateway",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MobileGatewayInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                         "101",
						"name":                       "mobile-gateway",
						"zone":                       "is1a",
						"internet_connection":        "1",
						"inter_device_communication": "1",
						"tags":                       ",tag1,tag2,",
						"description":                "desc",
					}),
				},
			},
			wantLogs: []string{
				`level=warn msg="can't get mobile_gateway's receive bytes: ID=101, NICIndex=0" err=dummy3`,
				`level=warn msg="can't get mobile_gateway's receive bytes: ID=101, NICIndex=1" err=dummy3`,
				`level=warn msg="can't get mobile_gateway's traffic control config: ID=101" err=dummy1`,
				`level=warn msg="can't get mobile_gateway's traffic status: ID=101" err=dummy2`,
			},
			wantErrCounter: 4, // traffic control + traffic status + nic monitor*2
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "mobile_gateway")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
