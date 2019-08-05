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
