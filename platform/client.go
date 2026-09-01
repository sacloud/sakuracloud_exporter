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

	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/fake"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/helper/api"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/trace"
	"github.com/sacloud/sacloud-sdk-go/api/webaccel"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
	"github.com/sacloud/sakuracloud_exporter/config"
)

type Client struct {
	authContext   authContextClient
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

func apiRequestRateLimit(rateLimit int) (uint16, error) {
	if rateLimit < 0 || rateLimit > 1<<16-1 {
		return 0, fmt.Errorf("API request rate limit must be between 0 and %d", 1<<16-1)
	}
	return uint16(rateLimit), nil
}

func NewSakuraCloudClient(c config.Config, version string) (*Client, error) {
	fakeStorePath := c.FakeMode
	if stat, err := os.Stat(fakeStorePath); err == nil {
		if stat.IsDir() {
			fakeStorePath = filepath.Join(fakeStorePath, "fake-store.json")
		}
	}

	rateLimit, err := apiRequestRateLimit(c.RateLimit)
	if err != nil {
		return nil, err
	}

	saClient := &saclient.Client{}
	if err := saClient.SetEnviron(os.Environ()); err != nil {
		return nil, fmt.Errorf("failed to set environment variables to saclient: %w", err)
	}
	if err := saClient.SetWith(
		saclient.WithUserAgent(fmt.Sprintf("sakuracloud_exporter/%s", version)),
		saclient.WithAPIRequestRateLimit(rateLimit),
	); err != nil {
		return nil, fmt.Errorf("failed to set saclient options: %w", err)
	}
	if c.Token != "" && c.Secret != "" {
		if err := saClient.SetWith(saclient.WithBasicAuth(c.Token, c.Secret)); err != nil {
			return nil, fmt.Errorf("failed to set saclient authentication: %w", err)
		}
	}
	if c.Trace {
		if err := saClient.SetWith(saclient.WithTraceMode("all")); err != nil {
			return nil, fmt.Errorf("failed to set saclient trace mode: %w", err)
		}
	}

	caller := iaas.NewClientFromSaclient(saClient)
	if c.Debug {
		trace.AddClientFactoryHooks()
	}
	if c.FakeMode != "" {
		if fakeStorePath != "" {
			fake.DataStore = fake.NewJSONFileStore(fakeStorePath)
		}
		fake.InitDataStore()
		fake.SwitchFactoryFuncToFake()
		api.SetupFakeDefaults()
	}

	webaccelClient := &webaccel.Client{
		Saclient: saClient,
	}

	authClient, err := getAuthContextClient(saClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get AuthContext client: %w", err)
	}

	return &Client{
		authContext:   authClient,
		AutoBackup:    getAutoBackupClient(caller, c.Zones),
		Bill:          getBillClient(caller, authClient),
		Coupon:        getCouponClient(caller, authClient),
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

		WebAccel: getWebAccelClient(webaccelClient),
	}, nil
}

func (c *Client) HasValidCredentials(ctx context.Context) bool {
	res, err := c.authContext.ReadAuthContext(ctx)
	return res != nil && err == nil
}
