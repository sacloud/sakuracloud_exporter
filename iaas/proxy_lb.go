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

package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ProxyLBClient interface {
	Find(ctx context.Context) ([]*sacloud.ProxyLB, error)
	GetCertificate(ctx context.Context, id types.ID) (*sacloud.ProxyLBCertificates, error)
	Monitor(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorConnectionValue, error)
}

func getProxyLBClient(caller sacloud.APICaller) ProxyLBClient {
	return &proxyLBClient{
		client: sacloud.NewProxyLBOp(caller),
	}
}

type proxyLBClient struct {
	client sacloud.ProxyLBAPI
}

func (c *proxyLBClient) Find(ctx context.Context) ([]*sacloud.ProxyLB, error) {
	var results []*sacloud.ProxyLB
	res, err := c.client.Find(ctx, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	return res.ProxyLBs, nil
}

func (c *proxyLBClient) GetCertificate(ctx context.Context, id types.ID) (*sacloud.ProxyLBCertificates, error) {
	return c.client.GetCertificates(ctx, id)
}

func (c *proxyLBClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorConnectionValue, error) {
	mvs, err := c.client.MonitorConnection(ctx, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorConnectionValue(mvs.Values), nil
}
