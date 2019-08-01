package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type LoadBalancer struct {
	*sacloud.LoadBalancer
	ZoneName string
}

type LoadBalancerClient interface {
	Find(ctx context.Context) ([]*LoadBalancer, error)
	Status(ctx context.Context, zone string, id types.ID) ([]*sacloud.LoadBalancerStatus, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error)
}

func getLoadBalancerClient(caller sacloud.APICaller, zones []string) LoadBalancerClient {
	return &loadBalancerClient{
		client: sacloud.NewLoadBalancerOp(caller),
		zones:  zones,
	}
}

type loadBalancerClient struct {
	client sacloud.LoadBalancerAPI
	zones  []string
}

func (c *loadBalancerClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.client.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, lb := range res.LoadBalancers {
		results = append(results, &LoadBalancer{
			LoadBalancer: lb,
			ZoneName:     zone,
		})
	}
	return results, err
}

func (c *loadBalancerClient) Find(ctx context.Context) ([]*LoadBalancer, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*LoadBalancer
	for _, s := range res {
		results = append(results, s.(*LoadBalancer))
	}
	return results, nil
}

func (c *loadBalancerClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	mvs, err := c.client.MonitorInterface(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *loadBalancerClient) Status(ctx context.Context, zone string, id types.ID) ([]*sacloud.LoadBalancerStatus, error) {
	res, err := c.client.Status(ctx, zone, id)
	if err != nil {
		return nil, err
	}
	return res.Status, nil
}
