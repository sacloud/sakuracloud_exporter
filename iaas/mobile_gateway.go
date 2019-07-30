package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type MobileGateway struct {
	*sacloud.MobileGateway
	ZoneName string
}

type MobileGatewayClient interface {
	Find(ctx context.Context) ([]*MobileGateway, error)
	TrafficStatus(ctx context.Context, zone string, id types.ID) (*sacloud.MobileGatewayTrafficStatus, error)
	TrafficControl(ctx context.Context, zone string, id types.ID) (*sacloud.MobileGatewayTrafficControl, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*sacloud.MonitorInterfaceValue, error)
}

func getMobileGatewayClient(caller sacloud.APICaller, zones []string) MobileGatewayClient {
	return &mobileGatewayClient{
		client: sacloud.NewMobileGatewayOp(caller),
		zones:  zones,
	}
}

type mobileGatewayClient struct {
	client sacloud.MobileGatewayAPI
	zones  []string
}

func (c *mobileGatewayClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.client.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, mgw := range res.MobileGateways {
		results = append(results, &MobileGateway{
			MobileGateway: mgw,
			ZoneName:      zone,
		})
	}
	return results, err
}

func (c *mobileGatewayClient) Find(ctx context.Context) ([]*MobileGateway, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*MobileGateway
	for _, s := range res {
		results = append(results, s.(*MobileGateway))
	}
	return results, nil
}

func (c *mobileGatewayClient) MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	mvs, err := c.client.MonitorInterface(ctx, zone, id, index, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *mobileGatewayClient) TrafficStatus(ctx context.Context, zone string, id types.ID) (*sacloud.MobileGatewayTrafficStatus, error) {
	return c.client.TrafficStatus(ctx, zone, id)
}

func (c *mobileGatewayClient) TrafficControl(ctx context.Context, zone string, id types.ID) (*sacloud.MobileGatewayTrafficControl, error) {
	return c.client.GetTrafficConfig(ctx, zone, id)
}
