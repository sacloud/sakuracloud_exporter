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

package config

import (
	"os"
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	initEnvVars()
	tests := []struct {
		name    string
		args    []string
		envs    map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "minimum",
			args: []string{"--token", "token", "--secret", "secret"},
			envs: nil,
			want: Config{
				Token:  "token",
				Secret: "secret",

				// 以下はデフォルト値
				WebPath:   "/metrics",
				WebAddr:   ":9542",
				Zones:     iaas.SakuraCloudZones,
				RateLimit: defaultRateLimit,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append([]string{os.Args[0]}, tt.args...)
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}

			got, err := InitConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.EqualValues(t, tt.want, got)
		})
	}
}

func initEnvVars() {
	keys := []string{
		"TRACE",
		"DEBUG",
		"FAKE_MODE",
		"SAKURACLOUD_ACCESS_TOKEN",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET",
		"WEB_ADDR",
		"WEB_PATH",
		"SAKURACLOUD_RATE_LIMIT",
	}
	for _, key := range keys {
		os.Unsetenv(key)
	}
}
