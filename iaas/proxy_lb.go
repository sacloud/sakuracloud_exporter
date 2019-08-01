package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ProxyLBClient interface {
	Find(ctx context.Context) ([]*sacloud.ProxyLB, error)
	GetCertificate(ctx context.Context, id types.ID) (*sacloud.ProxyLBCertificates, error)
	Monitor(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorConnectionValue, error)
}

func getProxyLBClient(caller sacloud.APICaller) ProxyLBClient {
	return &proxyLBClient{
		client: sacloud.NewProxyLBOp(caller),
	}
}

type proxyLBClient struct {
	client sacloud.ProxyLBAPI
}

func (c *proxyLBClient) Find(ctx context.Context) ([]*sacloud.ProxyLB, error) {
	var results []*sacloud.ProxyLB
	res, err := c.client.Find(ctx, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, lb := range res.ProxyLBs {
		results = append(results, lb)
	}
	return results, err
}

func (c *proxyLBClient) GetCertificate(ctx context.Context, id types.ID) (*sacloud.ProxyLBCertificates, error) {
	return c.client.GetCertificates(ctx, id)
}

func (c *proxyLBClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorConnectionValue, error) {
	mvs, err := c.client.MonitorConnection(ctx, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorConnectionValue(mvs.Values), nil
}
