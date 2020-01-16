// Copyright 2019-2020 The sakuracloud_exporter Authors
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
