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
	Trace     bool     `arg:"env:TRACE" help:"Enable output of trace log of Sakura cloud API call"`
	Debug     bool     `arg:"env:DEBUG" help:"Enable output of debug level log"`
	FakeMode  string   `arg:"--fake-mode,env:FAKE_MODE" help:"File path to fetch/store fake data. If this flag is specified, enable fake-mode"`
	Token     string   `arg:"required,env:SAKURACLOUD_ACCESS_TOKEN" help:"Token for using the SakuraCloud API"`
	Secret    string   `arg:"required,env:SAKURACLOUD_ACCESS_TOKEN_SECRET" help:"Secret for using the SakuraCloud API"`
	Zones     []string `arg:"-"` // TODO zones parameter is not implements.
	WebAddr   string   `arg:"env:WEB_ADDR"`
	WebPath   string   `arg:"env:WEB_PATH"`
	RateLimit int      `arg:"env:SAKURACLOUD_RATE_LIMIT" help:"Rate limit per second for SakuraCloud API calls"`

	NoCollectorAutoBackup              bool `arg:"--no-collector.auto-backup" help:"Disable the AutoBackup collector"`
	NoCollectorCoupon                  bool `arg:"--no-collector.coupon" help:"Disable the Coupon collector"`
	NoCollectorDatabase                bool `arg:"--no-collector.database" help:"Disable the Database collector"`
	NoCollectorESME                    bool `arg:"--no-collector.esme" help:"Disable the ESME collector"`
	NoCollectorInternet                bool `arg:"--no-collector.internet" help:"Disable the Internet(Switch+Router) collector"`
	NoCollectorLoadBalancer            bool `arg:"--no-collector.load-balancer" help:"Disable the LoadBalancer collector"`
	NoCollectorLocalRouter             bool `arg:"--no-collector.local-router" help:"Disable the LocalRouter collector"`
	NoCollectorMobileGateway           bool `arg:"--no-collector.mobile-gateway" help:"Disable the MobileGateway collector"`
	NoCollectorNFS                     bool `arg:"--no-collector.nfs" help:"Disable the NFS collector"`
	NoCollectorProxyLB                 bool `arg:"--no-collector.proxy-lb" help:"Disable the ProxyLB(Enhanced LoadBalancer) collector"`
	NoCollectorServer                  bool `arg:"--no-collector.server" help:"Disable the Server collector"`
	NoCollectorServerExceptMaintenance bool `arg:"--no-collector.server.except-maintenance" help:"Disable the Server collector except for maintenance information"`
	NoCollectorSIM                     bool `arg:"--no-collector.sim" help:"Disable the SIM collector"`
	NoCollectorVPCRouter               bool `arg:"--no-collector.vpc-router" help:"Disable the VPCRouter collector"`
	NoCollectorZone                    bool `arg:"--no-collector.zone" help:"Disable the Zone collector"`
}

func InitConfig() (Config, error) {
	c := Config{
		WebPath:   "/metrics",
		WebAddr:   ":9542",
		Zones:     []string{"is1a", "is1b", "tk1a", "tk1b", "tk1v"},
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
	if c.NoCollectorServerExceptMaintenance && c.NoCollectorServer {
		return c, fmt.Errorf("--no-collector.server.except-maintenance enabled and --no-collector-server are both enabled")
	}

	return c, nil
}
