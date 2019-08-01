package iaas

import (
	"context"

	"github.com/sacloud/libsacloud/v2/sacloud"
)

type authStatusClient interface {
	Read(context.Context) (*sacloud.AuthStatus, error)
}

func getAuthStatusClient(caller sacloud.APICaller) authStatusClient {
	return sacloud.NewAuthStatusOp(caller)
}
