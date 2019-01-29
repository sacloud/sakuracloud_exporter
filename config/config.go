package config

import (
	"errors"

	"github.com/alexflint/go-arg"
)

// Config gets its content from env and passes it on to different packages
type Config struct {
	Debug   bool     `arg:"env:DEBUG"`
	Token   string   `arg:"env:SAKURACLOUD_ACCESS_TOKEN"`
	Secret  string   `arg:"env:SAKURACLOUD_ACCESS_TOKEN_SECRET"`
	Zones   []string // TODO zones parameter is not implements.
	WebAddr string   `arg:"env:WEB_ADDR"`
	WebPath string   `arg:"env:WEB_PATH"`
}

func InitConfig() (Config, error) {
	c := Config{
		WebPath: "/metrics",
		WebAddr: ":9542",
		Zones:   []string{"is1a", "is1b", "tk1a", "tk1v"},
	}
	arg.MustParse(&c)

	if c.Token == "" {
		return c, errors.New("SakuraCloud API Token is required")
	}
	if c.Secret == "" {
		return c, errors.New("SakuraCloud API Secret is required")
	}
	return c, nil
}
