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
	"os"
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
)

var testZone string
var testCaller *iaas.Client

func TestMain(m *testing.M) {
	// this is for to use fake driver on iaas-api-go
	os.Setenv("TESTACC", "")

	testZone = testutil.TestZone()
	testCaller = testutil.SingletonAPICaller().(*iaas.Client)

	ret := m.Run()
	os.Exit(ret)
}
