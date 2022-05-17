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

	"github.com/sacloud/iaas-api-go"
)

// ZoneClient calls SakuraCloud zone API
type ZoneClient interface {
	Find(ctx context.Context) ([]*iaas.Zone, error)
}

func getZoneClient(caller iaas.APICaller) ZoneClient {
	return &zoneClient{
		client: iaas.NewZoneOp(caller),
	}
}

type zoneClient struct {
	client iaas.ZoneAPI
}

func (c *zoneClient) Find(ctx context.Context) ([]*iaas.Zone, error) {
	res, err := c.client.Find(ctx, nil)
	if err != nil {
		return nil, err
	}
	return res.Zones, nil
}
