// Copyright 2019-2025 The sakuracloud_exporter Authors
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

type ProxyLBClient interface {
	Find(ctx context.Context) ([]*iaas.ProxyLB, error)
	GetCertificate(ctx context.Context, id types.ID) (*iaas.ProxyLBCertificates, error)
	Monitor(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorConnectionValue, error)
}

func getProxyLBClient(caller iaas.APICaller) ProxyLBClient {
	return &proxyLBClient{
		client: iaas.NewProxyLBOp(caller),
	}
}

type proxyLBClient struct {
	client iaas.ProxyLBAPI
}

func (c *proxyLBClient) Find(ctx context.Context) ([]*iaas.ProxyLB, error) {
	var results []*iaas.ProxyLB
	res, err := c.client.Find(ctx, &iaas.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	return res.ProxyLBs, nil
}

func (c *proxyLBClient) GetCertificate(ctx context.Context, id types.ID) (*iaas.ProxyLBCertificates, error) {
	return c.client.GetCertificates(ctx, id)
}

func (c *proxyLBClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorConnectionValue, error) {
	mvs, err := c.client.MonitorConnection(ctx, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorConnectionValue(mvs.Values), nil
}
