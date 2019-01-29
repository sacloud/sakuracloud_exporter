package iaas

import (
	sakuraAPI "github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

// CouponClient calls SakuraCloud coupon API
type CouponClient interface {
	Find() ([]*sacloud.Coupon, error)
}

func getCouponClient(client *sakuraAPI.Client) CouponClient {
	return client.Coupon
}
