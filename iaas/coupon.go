package iaas

import (
	"context"
	"errors"
	"sync"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// CouponClient calls SakuraCloud coupon API
type CouponClient interface {
	Find(context.Context) ([]*sacloud.Coupon, error)
}

func getCouponClient(caller sacloud.APICaller) CouponClient {
	return &couponClient{caller: caller}
}

type couponClient struct {
	caller    sacloud.APICaller
	accountID types.ID
	once      sync.Once
}

func (c *couponClient) Find(ctx context.Context) ([]*sacloud.Coupon, error) {
	var err error
	c.once.Do(func() {
		var auth *sacloud.AuthStatus

		authStatusOp := sacloud.NewAuthStatusOp(c.caller)
		auth, err = authStatusOp.Read(ctx)
		if err != nil {
			return
		}
		c.accountID = auth.AccountID
	})
	if err != nil {
		return nil, err
	}
	if c.accountID.IsEmpty() {
		return nil, errors.New("getting AccountID is failed. please check your API Key settings")
	}

	couponOp := sacloud.NewCouponOp(c.caller)
	searched, err := couponOp.Find(ctx, c.accountID)
	if err != nil {
		return nil, err
	}

	var res []*sacloud.Coupon
	for _, v := range searched.Coupons {
		res = append(res, v)
	}
	return res, nil
}
