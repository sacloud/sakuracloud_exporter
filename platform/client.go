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
	"fmt"
	"os"
	"path/filepath"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/iaas-api-go/fake"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/sakuracloud_exporter/config"
	"github.com/sacloud/webaccel-api-go"
)

type Client struct {
	authStatus    authStatusClient
	AutoBackup    AutoBackupClient
	Bill          BillClient
	Coupon        CouponClient
	Database      DatabaseClient
	ESME          ESMEClient
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

	WebAccel WebAccelClient
}

func NewSakuraCloudClient(c config.Config, version string) *Client {
	fakeStorePath := c.FakeMode
	if stat, err := os.Stat(fakeStorePath); err == nil {
		if stat.IsDir() {
			fakeStorePath = filepath.Join(fakeStorePath, "fake-store.json")
		}
	}
	caller := api.NewCallerWithOptions(&api.CallerOptions{
		Options: &client.Options{
			AccessToken:          c.Token,
			AccessTokenSecret:    c.Secret,
			HttpRequestRateLimit: c.RateLimit,
			UserAgent:            fmt.Sprintf("sakuracloud_exporter/%s", version),
			Trace:                c.Trace,
		},
		TraceAPI:      c.Debug,
		FakeMode:      c.FakeMode != "",
		FakeStorePath: fakeStorePath,
	})
	if c.FakeMode != "" {
		fake.InitDataStore()
	}

	webaccelCaller := &webaccel.Client{
		Options: &client.Options{
			AccessToken:          c.Token,
			AccessTokenSecret:    c.Secret,
			HttpRequestRateLimit: c.RateLimit,
			UserAgent:            fmt.Sprintf("sakuracloud_exporter/%s", version),
			Trace:                c.Trace,
		},
	}

	return &Client{
		authStatus:    getAuthStatusClient(caller),
		AutoBackup:    getAutoBackupClient(caller, c.Zones),
		Bill:          getBillClient(caller),
		Coupon:        getCouponClient(caller),
		Database:      getDatabaseClient(caller, c.Zones),
		ESME:          getESMEClient(caller),
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

		WebAccel: getWebAccelClient(webaccelCaller),
	}
}

func (c *Client) HasValidAPIKeys(ctx context.Context) bool {
	res, err := c.authStatus.Read(ctx)
	return res != nil && err == nil
}

func (c *Client) HasWebAccelPermission(ctx context.Context) bool {
	res, err := c.authStatus.Read(ctx)
	if res == nil || err != nil {
		return false
	}

	return res.ExternalPermission.PermittedWebAccel()
}
