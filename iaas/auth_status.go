package iaas

import (
	sakuraAPI "github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

type authStatusClient interface {
	Read() (*sacloud.AuthStatus, error)
}

func getAuthStatusClient(client *sakuraAPI.Client) authStatusClient {
	return client.AuthStatus
}
