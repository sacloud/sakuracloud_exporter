package iaas

import (
	"context"
	"net/http"
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
		RetryInterval:     1 * time.Second,
		HTTPClient: &http.Client{
			Transport: &sacloud.RateLimitRoundTripper{RateLimitPerSec: c.RateLimit},
		},
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
