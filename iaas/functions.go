// Copyright 2019-2021 The sakuracloud_exporter Authors
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
	"sync"
)

type perZoneQueryFunc func(ctx context.Context, zone string) ([]interface{}, error)

func queryToZones(ctx context.Context, zones []string, query perZoneQueryFunc) ([]interface{}, error) {
	var wg sync.WaitGroup
	wg.Add(len(zones))

	type result struct {
		results []interface{}
		err     error
	}

	resCh := make(chan *result)
	defer close(resCh)

	for i := range zones {
		go func(zone string) {
			res, err := query(ctx, zone)
			resCh <- &result{
				results: res,
				err:     err,
			}
		}(zones[i])
	}

	var results []interface{}
	var err error
	go func() {
		for res := range resCh {
			if err == nil {
				if res.err != nil {
					err = res.err
				} else {
					results = append(results, res.results...)
				}
			}
			wg.Done()
		}
	}()

	wg.Wait()
	return results, err
}
