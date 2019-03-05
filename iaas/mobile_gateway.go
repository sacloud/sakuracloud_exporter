package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type MobileGateway struct {
	*sacloud.MobileGateway
	ZoneName string
}

type MobileGatewayClient interface {
	Find() ([]*MobileGateway, error)
	TrafficStatus(zone string, mobileGatewayID int64) (*sacloud.TrafficStatus, error)
	TrafficControl(zone string, mobileGatewayID int64) (*sacloud.TrafficMonitoringConfig, error)
	MonitorNIC(zone string, mobileGatewayID int64, index int, end time.Time) ([]*NICMetrics, error)
}

func getMobileGatewayClient(client *sakuraAPI.Client, zones []string) MobileGatewayClient {
	return &mobileGatewayClient{
		rawClient: client,
		zones:     zones,
	}
}

type mobileGatewayClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *mobileGatewayClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.MobileGateway.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.MobileGateways {
		results = append(results, &MobileGateway{
			MobileGateway: &res.MobileGateways[i],
			ZoneName:      c.Zone,
		})
	}
	return results, err
}

func (s *mobileGatewayClient) Find() ([]*MobileGateway, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*MobileGateway
	for _, s := range res {
		results = append(results, s.(*MobileGateway))
	}
	return results, nil
}

func (s *mobileGatewayClient) MonitorNIC(zone string, mobileGatewayID int64, index int, end time.Time) ([]*NICMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.MobileGateway.MonitorBy(mobileGatewayID, index, param)
	}

	return queryNICMonitorValue(s.rawClient, zone, end, query)
}

func (s *mobileGatewayClient) TrafficStatus(zone string, mobileGatewayID int64) (*sacloud.TrafficStatus, error) {
	c := s.rawClient.Clone()
	c.Zone = zone

	return c.MobileGateway.GetTrafficStatus(mobileGatewayID)
}

func (s *mobileGatewayClient) TrafficControl(zone string, mobileGatewayID int64) (*sacloud.TrafficMonitoringConfig, error) {
	c := s.rawClient.Clone()
	c.Zone = zone

	return c.MobileGateway.GetTrafficMonitoringConfig(mobileGatewayID)
}
