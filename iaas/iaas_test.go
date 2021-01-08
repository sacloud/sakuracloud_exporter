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
	"os"
	"testing"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/testutil"
)

var testZone string
var testCaller *sacloud.Client

func TestMain(m *testing.M) {
	// this is for to use fake driver on libsacloud
	os.Setenv("TESTACC", "")

	testZone = testutil.TestZone()
	testCaller = testutil.SingletonAPICaller().(*sacloud.Client)
	testCaller.UserAgent = "test-sakuracloud_exporter/dev"

	ret := m.Run()
	os.Exit(ret)
}
