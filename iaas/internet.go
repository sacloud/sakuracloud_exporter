package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type Internet struct {
	*sacloud.Internet
	ZoneName string
}

type InternetClient interface {
	Find(ctx context.Context) ([]*Internet, error)
	MonitorTraffic(ctx context.Context, zone string, internetID types.ID, end time.Time) (*sacloud.MonitorRouterValue, error)
}

func getInternetClient(caller sacloud.APICaller, zones []string) InternetClient {
	return &internetClient{
		client: sacloud.NewInternetOp(caller),
		zones:  zones,
	}
}

type internetClient struct {
	client sacloud.InternetAPI
	zones  []string
}

func (c *internetClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.client.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, router := range res.Internets {
		results = append(results, &Internet{
			Internet: router,
			ZoneName: zone,
		})
	}
	return results, err
}

func (c *internetClient) Find(ctx context.Context) ([]*Internet, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*Internet
	for _, s := range res {
		results = append(results, s.(*Internet))
	}
	return results, nil
}

func (c *internetClient) MonitorTraffic(ctx context.Context, zone string, internetID types.ID, end time.Time) (*sacloud.MonitorRouterValue, error) {
	mvs, err := c.client.Monitor(ctx, zone, internetID, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorRouterValue(mvs.Values), nil
}
