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

	"github.com/sacloud/webaccel-api-go"
)

// WebAccelClient calls SakuraCloud webAccel API
type WebAccelClient interface {
	Find(ctx context.Context) ([]*webaccel.Site, error)
	Usage(ctx context.Context) (*webaccel.MonthlyUsageResults, error)
}

func getWebAccelClient(caller webaccel.APICaller) WebAccelClient {
	return &webAccelClient{
		client: webaccel.NewOp(caller),
	}
}

type webAccelClient struct {
	client webaccel.API
}

func (c *webAccelClient) Find(ctx context.Context) ([]*webaccel.Site, error) {
	res, err := c.client.List(ctx)
	if err != nil {
		return nil, err
	}
	return res.Sites, nil
}

func (c *webAccelClient) Usage(ctx context.Context) (*webaccel.MonthlyUsageResults, error) {
	return c.client.MonthlyUsage(ctx, "")
}
