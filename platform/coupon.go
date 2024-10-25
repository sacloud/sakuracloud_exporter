// Copyright 2019-2023 The sakuracloud_exporter Authors
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

package platform

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

// CouponClient calls SakuraCloud coupon API
type CouponClient interface {
	Find(context.Context) ([]*iaas.Coupon, error)
}

func getCouponClient(caller iaas.APICaller) CouponClient {
	return &couponClient{
		caller: caller,
		cache:  *newCache(30 * time.Minute),
	}
}

type couponClient struct {
	caller    iaas.APICaller
	accountID types.ID
	once      sync.Once
	cache     cache
}

func (c *couponClient) Find(ctx context.Context) ([]*iaas.Coupon, error) {
	ca := c.cache.get()
	if ca != nil {
		return ca.([]*iaas.Coupon), nil
	}

	var err error
	c.once.Do(func() {
		var auth *iaas.AuthStatus

		authStatusOp := iaas.NewAuthStatusOp(c.caller)
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

	couponOp := iaas.NewCouponOp(c.caller)
	searched, err := couponOp.Find(ctx, c.accountID)
	if err != nil {
		return nil, err
	}

	n, err := c.nextCacheExpiresAt()
	if err != nil {
		return nil, err
	}
	err = c.cache.set(searched.Coupons, n)
	if err != nil {
		return nil, err
	}

	return searched.Coupons, nil
}

// キャッシュの有効期限を算出する
//
// Billing APIは1日1回 AM4:30 (JST) にデータが更新される。
// このため、現在時刻がAM4:30 (JST) よりも早ければ当日のAM4:30 (JST)、
// 現在時刻がAM4:30 (JST) よりも遅ければ翌日のAM4:30 (JST) を有効期限として扱う。
func (c *couponClient) nextCacheExpiresAt() (time.Time, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, err
	}

	// 実行環境のタイムゾーンは不定のためJSTを基準にする
	now := time.Now().In(jst)
	expiresAt := time.Date(now.Year(), now.Month(), now.Day(), BillAPIUpdateHourJST, BillAPIUpdateMinuteJST, 0, 0, jst)
	if now.Equal(expiresAt) || now.After(expiresAt) {
		expiresAt = expiresAt.Add(24 * time.Hour)
	}

	return expiresAt, nil
}
