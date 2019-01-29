package iaas

import (
	sakuraAPI "github.com/sacloud/libsacloud/api"
	"github.com/sacloud/sakuracloud_exporter/config"

	"fmt"
	"time"
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
	Server        ServerClient
	SIM           SIMClient
	VPCRouter     VPCRouterClient
	Zone          ZoneClient
}

func NewSakuraCloucClient(c config.Config, version string) *Client {
	client := sakuraAPI.NewClient(c.Token, c.Secret, "is1a")
	client.UserAgent = fmt.Sprintf("sakuracloud_exporter/%s", version)

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
