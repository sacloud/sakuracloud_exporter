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

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

type ESMEClient interface {
	Find(ctx context.Context) ([]*iaas.ESME, error)
	Logs(ctx context.Context, esmeID types.ID) ([]*iaas.ESMELogs, error)
}

func getESMEClient(caller iaas.APICaller) ESMEClient {
	return &esmeClient{
		caller: caller,
	}
}

type esmeClient struct {
	caller iaas.APICaller
}

func (c *esmeClient) Find(ctx context.Context) ([]*iaas.ESME, error) {
	client := iaas.NewESMEOp(c.caller)
	searched, err := client.Find(ctx, &iaas.FindCondition{})
	if err != nil {
		return nil, err
	}
	return searched.ESME, nil
}

func (c *esmeClient) Logs(ctx context.Context, esmeID types.ID) ([]*iaas.ESMELogs, error) {
	client := iaas.NewESMEOp(c.caller)
	return client.Logs(ctx, esmeID)
}
