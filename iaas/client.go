package iaas

import (
	"log"
	"net/http"
	"sync"
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"
	"github.com/sacloud/sakuracloud_exporter/config"
	"go.uber.org/ratelimit"

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
	client := sakuraAPI.NewClient(c.Token, c.Secret, "is1a")
	client.UserAgent = fmt.Sprintf("sakuracloud_exporter/%s", version)
	client.TraceMode = c.Trace

	transport := &rateLimitRoundTripper{rateLimitPerSec: c.RateLimit}
	if c.Debug {
		transport.transport = &loggingRoundTripper{}
	}
	client.HTTPClient = &http.Client{Transport: transport}

	// TODO make configurable
	client.RetryMax = 9
	client.RetryInterval = 1 * time.Second

	return &Client{
		authStatus:    getAuthStatusClient(client),
		AutoBackup:    getAutoBackupClient(client),
		Coupon:        getCouponClient(client),
		Database:      getDatabaseClient(client, c.Zones),
		Internet:      getInternetClient(client, c.Zones),
		LoadBalancer:  getLoadBalancerClient(client, c.Zones),
		MobileGateway: getMobileGatewayClient(client, c.Zones),
		NFS:           getNFSClient(client, c.Zones),
		ProxyLB:       getProxyLBClient(client),
		Server:        getServerClient(client, c.Zones),
		SIM:           getSIMClient(client),
		VPCRouter:     getVPCRouterClient(client, c.Zones),
		Zone:          getZoneClient(client),
	}
}

func (c *Client) HasValidAPIKeys() bool {
	res, err := c.authStatus.Read()
	return res != nil && err == nil
}

type rateLimitRoundTripper struct {
	transport       http.RoundTripper
	rateLimitPerSec int

	once      sync.Once
	rateLimit ratelimit.Limiter
}

func (r *rateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.once.Do(func() {
		r.rateLimit = ratelimit.New(r.rateLimitPerSec)
	})
	if r.transport == nil {
		r.transport = http.DefaultTransport
	}

	r.rateLimit.Take()
	return r.transport.RoundTrip(req)
}

type loggingRoundTripper struct {
	transport http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if l.transport == nil {
		l.transport = http.DefaultTransport
	}
	log.Printf("request: %s", req.URL.Path)
	return l.transport.RoundTrip(req)
}
