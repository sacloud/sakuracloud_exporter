package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type VPCRouter struct {
	*sacloud.VPCRouter
	ZoneName string
}

type VPCRouterClient interface {
	Find() ([]*VPCRouter, error)
	Status(zone string, vpcRouterID int64) (*sacloud.VPCRouterStatus, error)
	MonitorNIC(zone string, vpcRouterID int64, index int, end time.Time) ([]*NICMetrics, error)
}

func getVPCRouterClient(client *sakuraAPI.Client, zones []string) VPCRouterClient {
	return &vpcRouterClient{
		rawClient: client,
		zones:     zones,
	}
}

type vpcRouterClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *vpcRouterClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.VPCRouter.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.VPCRouters {
		results = append(results, &VPCRouter{
			VPCRouter: &res.VPCRouters[i],
			ZoneName:  c.Zone,
		})
	}
	return results, err
}

func (s *vpcRouterClient) Find() ([]*VPCRouter, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*VPCRouter
	for _, s := range res {
		results = append(results, s.(*VPCRouter))
	}
	return results, nil
}

func (s *vpcRouterClient) MonitorNIC(zone string, vpcRouterID int64, index int, end time.Time) ([]*NICMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.VPCRouter.MonitorBy(vpcRouterID, index, param)
	}

	return queryNICMonitorValue(s.rawClient, zone, end, query)
}

func (s *vpcRouterClient) Status(zone string, vpcRouterID int64) (*sacloud.VPCRouterStatus, error) {
	c := s.rawClient.Clone()
	c.Zone = zone

	res, err := c.VPCRouter.Status(vpcRouterID)
	if err != nil {
		return nil, err
	}
	return res, nil
}
