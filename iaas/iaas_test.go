package iaas

import (
	sakuraAPI "github.com/sacloud/libsacloud/api"

	"log"
	"os"
	"testing"
)

var testClient *sakuraAPI.Client

func TestMain(m *testing.M) {

	accessToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

	if accessToken == "" || accessTokenSecret == "" {
		log.Println("Please Set ENV 'SAKURACLOUD_ACCESS_TOKEN' and 'SAKURACLOUD_ACCESS_TOKEN_SECRET'")
		os.Exit(0) // exit normal
	}

	zone := os.Getenv("SAKURACLOUD_ZONE")
	if zone == "" {
		zone = "is1b"
	}

	traceMode := false
	if os.Getenv("SAKURACLOUD_TRACE_MODE") != "" {
		traceMode = true
	}

	c := sakuraAPI.NewClient(accessToken, accessTokenSecret, zone)
	c.TraceMode = traceMode
	c.UserAgent = "sakuracloud_exporter/go-test"

	testClient = c
	ret := m.Run()
	os.Exit(ret)
}
