// Copyright 2019-2023 The sacloud/sakuracloud_exporter Authors
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

package collector

import (
	"context"
	"testing"

	"github.com/sacloud/webaccel-api-go"
	"github.com/stretchr/testify/require"
)

type dummyWebAccelClient struct {
	sites []*webaccel.Site
	usage *webaccel.MonthlyUsageResults
	err   error
}

func (d *dummyWebAccelClient) Find(ctx context.Context) ([]*webaccel.Site, error) {
	return d.sites, d.err
}

func (d *dummyWebAccelClient) Usage(ctx context.Context) (*webaccel.MonthlyUsageResults, error) {
	return d.usage, d.err
}

func TestWebAccelCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewWebAccelCollector(context.Background(), testLogger, testErrors, &dummyWebAccelClient{})

	descs := collectDescs(c)
	require.Len(t, descs, 8)
}
