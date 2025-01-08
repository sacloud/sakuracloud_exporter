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

type SIMClient interface {
	Find(ctx context.Context) ([]*iaas.SIM, error)
	GetNetworkOperatorConfig(ctx context.Context, id types.ID) ([]*iaas.SIMNetworkOperatorConfig, error)
	MonitorTraffic(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorLinkValue, error)
}

func getSIMClient(caller iaas.APICaller) SIMClient {
	return &simClient{
		client: iaas.NewSIMOp(caller),
	}
}

type simClient struct {
	client iaas.SIMAPI
}

func (c *simClient) Find(ctx context.Context) ([]*iaas.SIM, error) {
	var results []*iaas.SIM
	res, err := c.client.Find(ctx, &iaas.FindCondition{
		Include: []string{"*", "Status.sim"},
		Count:   10000,
	})
	if err != nil {
		return results, err
	}
	return res.SIMs, nil
}

func (c *simClient) GetNetworkOperatorConfig(ctx context.Context, id types.ID) ([]*iaas.SIMNetworkOperatorConfig, error) {
	return c.client.GetNetworkOperator(ctx, id)
}

func (c *simClient) MonitorTraffic(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorLinkValue, error) {
	mvs, err := c.client.MonitorSIM(ctx, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorLinkValue(mvs.Values), nil
}
