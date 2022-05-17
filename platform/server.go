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
	"github.com/sacloud/packages-go/newsfeed"
)

type ServerClient interface {
	Find(ctx context.Context) ([]*Server, error)
	ReadDisk(ctx context.Context, zone string, diskID types.ID) (*iaas.Disk, error)
	MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error)
	MonitorDisk(ctx context.Context, zone string, diskID types.ID, end time.Time) (*iaas.MonitorDiskValue, error)
	MonitorNIC(ctx context.Context, zone string, nicID types.ID, end time.Time) (*iaas.MonitorInterfaceValue, error)
	MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error)
}

type Server struct {
	*iaas.Server
	ZoneName string
}

func getServerClient(caller iaas.APICaller, zones []string) ServerClient {
	return &serverClient{
		serverOp:    iaas.NewServerOp(caller),
		diskOp:      iaas.NewDiskOp(caller),
		interfaceOp: iaas.NewInterfaceOp(caller),
		zones:       zones,
	}
}

type serverClient struct {
	serverOp    iaas.ServerAPI
	diskOp      iaas.DiskAPI
	interfaceOp iaas.InterfaceAPI
	zones       []string
}

func (c *serverClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.serverOp.Find(ctx, zone, &iaas.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, s := range res.Servers {
		results = append(results, &Server{
			Server:   s,
			ZoneName: zone,
		})
	}
	return results, err
}

func (c *serverClient) Find(ctx context.Context) ([]*Server, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*Server
	for _, s := range res {
		results = append(results, s.(*Server))
	}
	return results, nil
}

func (c *serverClient) ReadDisk(ctx context.Context, zone string, diskID types.ID) (*iaas.Disk, error) {
	return c.diskOp.Read(ctx, zone, diskID)
}

func (c *serverClient) MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error) {
	mvs, err := c.serverOp.Monitor(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorCPUTimeValue(mvs.Values), nil
}

func (c *serverClient) MonitorDisk(ctx context.Context, zone string, diskID types.ID, end time.Time) (*iaas.MonitorDiskValue, error) {
	mvs, err := c.diskOp.Monitor(ctx, zone, diskID, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorDiskValue(mvs.Values), nil
}

func (c *serverClient) MonitorNIC(ctx context.Context, zone string, nicID types.ID, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	mvs, err := c.interfaceOp.Monitor(ctx, zone, nicID, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *serverClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return newsfeed.GetByURL(infoURL)
}
