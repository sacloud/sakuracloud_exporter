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
	"fmt"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/iaas-api-go/search/keys"
	"github.com/sacloud/iaas-api-go/types"
)

type AutoBackupClient interface {
	Find(ctx context.Context) ([]*iaas.AutoBackup, error)
	ListBackups(ctx context.Context, zone string, autoBackupID types.ID) ([]*iaas.Archive, error)
}

func getAutoBackupClient(caller iaas.APICaller, zones []string) AutoBackupClient {
	return &autoBackupClient{
		caller: caller,
	}
}

type autoBackupClient struct {
	caller iaas.APICaller
}

func (c *autoBackupClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	client := iaas.NewAutoBackupOp(c.caller)
	searched, err := client.Find(ctx, zone, &iaas.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return nil, err
	}
	var res []interface{}
	for _, v := range searched.AutoBackups {
		res = append(res, v)
	}
	return res, nil
}

func (c *autoBackupClient) Find(ctx context.Context) ([]*iaas.AutoBackup, error) {
	res, err := c.find(ctx, "is1a")
	if err != nil {
		return nil, err
	}
	var results []*iaas.AutoBackup
	for _, v := range res {
		results = append(results, v.(*iaas.AutoBackup))
	}
	return results, nil
}

func (c *autoBackupClient) ListBackups(ctx context.Context, zone string, autoBackupID types.ID) ([]*iaas.Archive, error) {
	client := iaas.NewArchiveOp(c.caller)
	tagName := fmt.Sprintf("autobackup-%d", autoBackupID)

	searched, err := client.Find(ctx, zone, &iaas.FindCondition{
		Count: 10000,
		Filter: search.Filter{
			search.Key(keys.Tags): search.TagsAndEqual(tagName),
		},
	})
	if err != nil {
		return nil, err
	}

	var res []*iaas.Archive
	for _, v := range searched.Archives {
		if v.Availability.IsAvailable() {
			res = append(res, v)
		}
	}
	return res, err
}
