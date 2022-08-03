// Copyright 2019-2022 The sacloud/sakuracloud_exporter Authors
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

//go:build e2e
// +build e2e

package minimum

import (
	"strings"
	"testing"
	"time"

	"github.com/sacloud/packages-go/e2e"
	"github.com/stretchr/testify/require"
)

const commandName = "sakuracloud_exporter"

func TestE2E_minimum(t *testing.T) {
	err := e2e.RunCommand(t, commandName, "-h")

	require.NoError(t, err)
}

func TestE2E_output(t *testing.T) {
	reader, err := e2e.StartCommandWithStdErr(t, commandName, "--webaddr", "localhost:9542")
	if err != nil {
		t.Fatal(err)
	}

	output := e2e.NewOutput(reader, "")
	if err := output.WaitOutput("msg=listening addr=localhost:9542", 5*time.Second); err != nil {
		t.Fatal(err)
	}

	response, err := e2e.HttpGetWithResponse("http://localhost:9542/metrics")
	if err != nil {
		t.Fatal(err)
	}

	// go
	require.True(t, strings.Contains(string(response), `# TYPE go_gc_duration_seconds summary`))
	// sakuracloud
	require.True(t, strings.Contains(string(response), `sakuracloud_exporter_errors_total{collector="zone"} 0`))
}
