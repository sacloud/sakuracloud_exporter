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

package platform

import (
	"context"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

type MobileGateway struct {
	*iaas.MobileGateway
	ZoneName string
}

type MobileGatewayClient interface {
	Find(ctx context.Context) ([]*MobileGateway, error)
	TrafficStatus(ctx context.Context, zone string, id types.ID) (*iaas.MobileGatewayTrafficStatus, error)
	TrafficControl(ctx context.Context, zone string, id types.ID) (*iaas.MobileGatewayTrafficControl, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*iaas.MonitorInterfaceValue, error)
}

func getMobileGatewayClient(caller iaas.APICaller, zones []string) MobileGatewayClient {
	return &mobileGatewayClient{
		client: iaas.NewMobileGatewayOp(caller),
		zones:  zones,
	}
}

type mobileGatewayClient struct {
	client iaas.MobileGatewayAPI
	zones  []string
}

func (c *mobileGatewayClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.client.Find(ctx, zone, &iaas.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, mgw := range res.MobileGateways {
		results = append(results, &MobileGateway{
			MobileGateway: mgw,
			ZoneName:      zone,
		})
	}
	return results, err
}

func (c *mobileGatewayClient) Find(ctx context.Context) ([]*MobileGateway, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*MobileGateway
	for _, s := range res {
		results = append(results, s.(*MobileGateway))
	}
	return results, nil
}

func (c *mobileGatewayClient) MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	mvs, err := c.client.MonitorInterface(ctx, zone, id, index, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *mobileGatewayClient) TrafficStatus(ctx context.Context, zone string, id types.ID) (*iaas.MobileGatewayTrafficStatus, error) {
	return c.client.TrafficStatus(ctx, zone, id)
}

func (c *mobileGatewayClient) TrafficControl(ctx context.Context, zone string, id types.ID) (*iaas.MobileGatewayTrafficControl, error) {
	return c.client.GetTrafficConfig(ctx, zone, id)
}
