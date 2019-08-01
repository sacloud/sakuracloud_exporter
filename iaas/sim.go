package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type SIMClient interface {
	Find(ctx context.Context) ([]*sacloud.SIM, error)
	GetNetworkOperatorConfig(ctx context.Context, id types.ID) ([]*sacloud.SIMNetworkOperatorConfig, error)
	MonitorTraffic(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorLinkValue, error)
}

func getSIMClient(caller sacloud.APICaller) SIMClient {
	return &simClient{
		client: sacloud.NewSIMOp(caller),
	}
}

type simClient struct {
	client sacloud.SIMAPI
}

func (c *simClient) Find(ctx context.Context) ([]*sacloud.SIM, error) {
	var results []*sacloud.SIM
	res, err := c.client.Find(ctx, &sacloud.FindCondition{
		Include: []string{"*", "Status.sim"},
		Count:   10000,
	})
	if err != nil {
		return results, err
	}
	for _, lb := range res.SIMs {
		results = append(results, lb)
	}
	return results, err
}

func (c *simClient) GetNetworkOperatorConfig(ctx context.Context, id types.ID) ([]*sacloud.SIMNetworkOperatorConfig, error) {
	return c.client.GetNetworkOperator(ctx, id)
}

func (c *simClient) MonitorTraffic(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorLinkValue, error) {
	mvs, err := c.client.MonitorSIM(ctx, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorLinkValue(mvs.Values), nil
}
