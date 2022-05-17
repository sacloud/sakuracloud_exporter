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

type LocalRouterClient interface {
	Find(ctx context.Context) ([]*iaas.LocalRouter, error)
	Health(ctx context.Context, id types.ID) (*iaas.LocalRouterHealth, error)
	Monitor(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorLocalRouterValue, error)
}

func getLocalRouterClient(caller iaas.APICaller) LocalRouterClient {
	return &localRouterClient{
		client: iaas.NewLocalRouterOp(caller),
	}
}

type localRouterClient struct {
	client iaas.LocalRouterAPI
}

func (c *localRouterClient) Find(ctx context.Context) ([]*iaas.LocalRouter, error) {
	var results []*iaas.LocalRouter
	res, err := c.client.Find(ctx, &iaas.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	return res.LocalRouters, nil
}

func (c *localRouterClient) Health(ctx context.Context, id types.ID) (*iaas.LocalRouterHealth, error) {
	return c.client.HealthStatus(ctx, id)
}

func (c *localRouterClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*iaas.MonitorLocalRouterValue, error) {
	mvs, err := c.client.MonitorLocalRouter(ctx, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorLocalRouterValue(mvs.Values), nil
}
