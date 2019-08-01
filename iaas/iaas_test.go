package iaas

import (
	"os"
	"testing"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/testutil"
)

var testZone string
var testCaller *sacloud.Client

func TestMain(m *testing.M) {
	// this is for to use fake driver on libsacloud
	os.Setenv("TESTACC", "")

	testZone = testutil.TestZone()
	testCaller = testutil.SingletonAPICaller()
	testCaller.UserAgent = "test-sakuracloud_exporter/dev"

	ret := m.Run()
	os.Exit(ret)
}
