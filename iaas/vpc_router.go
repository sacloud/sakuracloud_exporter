package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type VPCRouter struct {
	*sacloud.VPCRouter
	ZoneName string
}

type VPCRouterClient interface {
	Find(ctx context.Context) ([]*VPCRouter, error)
	Status(ctx context.Context, zone string, id types.ID) (*sacloud.VPCRouterStatus, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*sacloud.MonitorInterfaceValue, error)
}

func getVPCRouterClient(caller sacloud.APICaller, zones []string) VPCRouterClient {
	return &vpcRouterClient{
		client: sacloud.NewVPCRouterOp(caller),
		zones:  zones,
	}
}

type vpcRouterClient struct {
	client sacloud.VPCRouterAPI
	zones  []string
}

func (c *vpcRouterClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.client.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, v := range res.VPCRouters {
		results = append(results, &VPCRouter{
			VPCRouter: v,
			ZoneName:  zone,
		})
	}
	return results, err
}

func (c *vpcRouterClient) Find(ctx context.Context) ([]*VPCRouter, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*VPCRouter
	for _, s := range res {
		results = append(results, s.(*VPCRouter))
	}
	return results, nil
}

func (c *vpcRouterClient) MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	mvs, err := c.client.MonitorInterface(ctx, zone, id, index, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *vpcRouterClient) Status(ctx context.Context, zone string, id types.ID) (*sacloud.VPCRouterStatus, error) {
	return c.client.Status(ctx, zone, id)
}
