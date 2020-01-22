// Copyright 2019-2020 The sakuracloud_exporter Authors
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
	"github.com/sacloud/libsacloud/v2/sacloud/fake"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/trace"
	"github.com/sacloud/sakuracloud_exporter/config"

	"fmt"
)

type Client struct {
	authStatus    authStatusClient
	AutoBackup    AutoBackupClient
	Coupon        CouponClient
	Database      DatabaseClient
	Internet      InternetClient
	LoadBalancer  LoadBalancerClient
	LocalRouter   LocalRouterClient
	MobileGateway MobileGatewayClient
	NFS           NFSClient
	ProxyLB       ProxyLBClient
	Server        ServerClient
	SIM           SIMClient
	VPCRouter     VPCRouterClient
	Zone          ZoneClient
}

func NewSakuraCloucClient(c config.Config, version string) *Client {
	caller := &sacloud.Client{
		AccessToken:       c.Token,
		AccessTokenSecret: c.Secret,
		UserAgent:         fmt.Sprintf("sakuracloud_exporter/%s", version),
		RetryMax:          9,
		RetryWaitMin:      1 * time.Second,
		RetryWaitMax:      5 * time.Second,
		HTTPClient: &http.Client{
			Transport: &sacloud.RateLimitRoundTripper{RateLimitPerSec: c.RateLimit},
		},
	}
	if c.FakeMode != "" {
		fakeStorePath := c.FakeMode
		if stat, err := os.Stat(fakeStorePath); err == nil {
			if stat.IsDir() {
				fakeStorePath = filepath.Join(fakeStorePath, "fake-store.json")
			}
		}
		fake.DataStore = fake.NewJSONFileStore(fakeStorePath)
		fake.SwitchFactoryFuncToFake()
		fake.InitDataStore()
	}

	if c.Debug {
		trace.AddClientFactoryHooks()
	}
	if c.Trace {
		caller.HTTPClient.Transport = &sacloud.TracingRoundTripper{
			Transport: caller.HTTPClient.Transport,
		}
	}

	return &Client{
		authStatus:    getAuthStatusClient(caller),
		AutoBackup:    getAutoBackupClient(caller, c.Zones),
		Coupon:        getCouponClient(caller),
		Database:      getDatabaseClient(caller, c.Zones),
		Internet:      getInternetClient(caller, c.Zones),
		LoadBalancer:  getLoadBalancerClient(caller, c.Zones),
		LocalRouter:   getLocalRouterClient(caller),
		MobileGateway: getMobileGatewayClient(caller, c.Zones),
		NFS:           getNFSClient(caller, c.Zones),
		ProxyLB:       getProxyLBClient(caller),
		Server:        getServerClient(caller, c.Zones),
		SIM:           getSIMClient(caller),
		VPCRouter:     getVPCRouterClient(caller, c.Zones),
		Zone:          getZoneClient(caller),
	}
}

func (c *Client) HasValidAPIKeys(ctx context.Context) bool {
	res, err := c.authStatus.Read(ctx)
	return res != nil && err == nil
}
