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

	"github.com/sacloud/libsacloud/v2/helper/query"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type NFS struct {
	*sacloud.NFS
	Plan     *query.NFSPlanInfo
	PlanName string
	ZoneName string
}

type NFSClient interface {
	Find(ctx context.Context) ([]*NFS, error)
	MonitorFreeDiskSize(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorFreeDiskSizeValue, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error)
}

func getNFSClient(caller sacloud.APICaller, zones []string) NFSClient {
	return &nfsClient{
		noteOp: sacloud.NewNoteOp(caller),
		nfsOp:  sacloud.NewNFSOp(caller),
		zones:  zones,
	}
}

type nfsClient struct {
	noteOp sacloud.NoteAPI
	nfsOp  sacloud.NFSAPI
	zones  []string
}

func (c *nfsClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.nfsOp.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, v := range res.NFS {
		planInfo, err := query.GetNFSPlanInfo(ctx, c.noteOp, v.PlanID)
		if err != nil {
			return nil, err
		}
		planName := ""
		switch planInfo.DiskPlanID {
		case types.NFSPlans.HDD:
			planName = "HDD"
		case types.NFSPlans.SSD:
			planName = "SSD"
		}
		results = append(results, &NFS{
			NFS:      v,
			PlanName: planName,
			Plan:     planInfo,
			ZoneName: zone,
		})
	}
	return results, err
}

func (c *nfsClient) Find(ctx context.Context) ([]*NFS, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*NFS
	for _, s := range res {
		results = append(results, s.(*NFS))
	}
	return results, nil
}

func (c *nfsClient) MonitorFreeDiskSize(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorFreeDiskSizeValue, error) {
	mvs, err := c.nfsOp.MonitorFreeDiskSize(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorFreeDiskSizeValue(mvs.Values), nil
}

func (c *nfsClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	mvs, err := c.nfsOp.MonitorInterface(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}
