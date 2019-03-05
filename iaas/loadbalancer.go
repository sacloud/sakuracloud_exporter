package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type LoadBalancer struct {
	*sacloud.LoadBalancer
	ZoneName string
}

type LoadBalancerClient interface {
	Find() ([]*LoadBalancer, error)
	Status(zone string, loadBalancerID int64) (*sacloud.LoadBalancerStatusResult, error)
	MonitorNIC(zone string, loadBalancerID int64, end time.Time) ([]*NICMetrics, error)
}

func getLoadBalancerClient(client *sakuraAPI.Client, zones []string) LoadBalancerClient {
	return &loadBalancerClient{
		rawClient: client,
		zones:     zones,
	}
}

type loadBalancerClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *loadBalancerClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.LoadBalancer.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.LoadBalancers {
		results = append(results, &LoadBalancer{
			LoadBalancer: &res.LoadBalancers[i],
			ZoneName:     c.Zone,
		})
	}
	return results, err
}

func (s *loadBalancerClient) Find() ([]*LoadBalancer, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*LoadBalancer
	for _, s := range res {
		results = append(results, s.(*LoadBalancer))
	}
	return results, nil
}

func (s *loadBalancerClient) MonitorNIC(zone string, loadBalancerID int64, end time.Time) ([]*NICMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.LoadBalancer.Monitor(loadBalancerID, param)
	}

	return queryNICMonitorValue(s.rawClient, zone, end, query)
}

func (s *loadBalancerClient) Status(zone string, loadBalancerID int64) (*sacloud.LoadBalancerStatusResult, error) {
	c := s.rawClient.Clone()
	c.Zone = zone

	res, err := c.LoadBalancer.Status(loadBalancerID)
	if err != nil {
		return nil, err
	}
	return res, nil
}
