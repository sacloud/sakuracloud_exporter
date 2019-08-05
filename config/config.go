package config

import (
	"errors"
	"fmt"

	"github.com/alexflint/go-arg"
)

const (
	maximumRateLimit = 10
	defaultRateLimit = 5
)

// Config gets its content from env and passes it on to different packages
type Config struct {
	Trace     bool     `arg:"env:TRACE"`
	Debug     bool     `arg:"env:DEBUG"`
	FakeMode  string   `arg:"env:FAKE_MODE"`
	Token     string   `arg:"required,env:SAKURACLOUD_ACCESS_TOKEN"`
	Secret    string   `arg:"required,env:SAKURACLOUD_ACCESS_TOKEN_SECRET"`
	Zones     []string // TODO zones parameter is not implements.
	WebAddr   string   `arg:"env:WEB_ADDR"`
	WebPath   string   `arg:"env:WEB_PATH"`
	RateLimit int      `arg:"env:SAKURACLOUD_RATE_LIMIT"`

	NoCollectorAutoBackup    bool `arg:"--no-collector.auto-backup" help:"Disable the AutoBackup collector"`
	NoCollectorCoupon        bool `arg:"--no-collector.coupon" help:"Disable the Coupon collector"`
	NoCollectorDatabase      bool `arg:"--no-collector.database" help:"Disable the Database collector"`
	NoCollectorInternet      bool `arg:"--no-collector.internet" help:"Disable the Internet(Switch+Router) collector"`
	NoCollectorLoadBalancer  bool `arg:"--no-collector.load-balancer" help:"Disable the LoadBalancer collector"`
	NoCollectorMobileGateway bool `arg:"--no-collector.mobile-gateway" help:"Disable the MobileGateway collector"`
	NoCollectorNFS           bool `arg:"--no-collector.nfs" help:"Disable the NFS collector"`
	NoCollectorProxyLB       bool `arg:"--no-collector.proxy-lb" help:"Disable the ProxyLB(Enhanced LoadBalancer) collector"`
	NoCollectorServer        bool `arg:"--no-collector.server" help:"Disable the Server collector"`
	NoCollectorSIM           bool `arg:"--no-collector.sim" help:"Disable the SIM collector"`
	NoCollectorVPCRouter     bool `arg:"--no-collector.vpc-router" help:"Disable the VPCRouter collector"`
	NoCollectorZone          bool `arg:"--no-collector.zone" help:"Disable the Zone collector"`
}

func InitConfig() (Config, error) {
	c := Config{
		WebPath:   "/metrics",
		WebAddr:   ":9542",
		Zones:     []string{"is1a", "is1b", "tk1a", "tk1v"},
		RateLimit: defaultRateLimit,
	}
	arg.MustParse(&c)

	if c.Token == "" {
		return c, errors.New("SakuraCloud API Token is required")
	}
	if c.Secret == "" {
		return c, errors.New("SakuraCloud API Secret is required")
	}
	if c.RateLimit <= 0 {
		c.RateLimit = defaultRateLimit
	}
	if c.RateLimit > maximumRateLimit {
		return c, fmt.Errorf("--ratelimit must be 1 to %d", maximumRateLimit)
	}

	return c, nil
}
