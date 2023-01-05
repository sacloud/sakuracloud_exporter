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

package platform

import (
	"context"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
)

type VPCRouter struct {
	*iaas.VPCRouter
	ZoneName string
}

type VPCRouterClient interface {
	Find(ctx context.Context) ([]*VPCRouter, error)
	Status(ctx context.Context, zone string, id types.ID) (*iaas.VPCRouterStatus, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*iaas.MonitorInterfaceValue, error)
	MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error)
	MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error)
}

func getVPCRouterClient(caller iaas.APICaller, zones []string) VPCRouterClient {
	return &vpcRouterClient{
		client: iaas.NewVPCRouterOp(caller),
		zones:  zones,
	}
}

type vpcRouterClient struct {
	client iaas.VPCRouterAPI
	zones  []string
}

func (c *vpcRouterClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.client.Find(ctx, zone, &iaas.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, v := range res.VPCRouters {
		results = append(results, &VPCRouter{
			VPCRouter: v,
			ZoneName:  zone,
		})
	}
	return results, err
}

func (c *vpcRouterClient) Find(ctx context.Context) ([]*VPCRouter, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*VPCRouter
	for _, s := range res {
		results = append(results, s.(*VPCRouter))
	}
	return results, nil
}

func (c *vpcRouterClient) MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	mvs, err := c.client.MonitorInterface(ctx, zone, id, index, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *vpcRouterClient) MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error) {
	mvs, err := c.client.MonitorCPU(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorCPUTimeValue(mvs.Values), nil
}

func (c *vpcRouterClient) Status(ctx context.Context, zone string, id types.ID) (*iaas.VPCRouterStatus, error) {
	return c.client.Status(ctx, zone, id)
}

func (c *vpcRouterClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return newsfeed.GetByURL(infoURL)
}
