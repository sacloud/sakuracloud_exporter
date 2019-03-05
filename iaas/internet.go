package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type Internet struct {
	*sacloud.Internet
	ZoneName string
}

type InternetClient interface {
	Find() ([]*Internet, error)
	MonitorTraffic(zone string, internetID int64, end time.Time) ([]*RouterMetrics, error)
}

func getInternetClient(client *sakuraAPI.Client, zones []string) InternetClient {
	return &internetClient{
		rawClient: client,
		zones:     zones,
	}
}

type internetClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *internetClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.Internet.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.Internet {
		results = append(results, &Internet{
			Internet: &res.Internet[i],
			ZoneName: c.Zone,
		})
	}
	return results, err
}

func (s *internetClient) Find() ([]*Internet, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*Internet
	for _, s := range res {
		results = append(results, s.(*Internet))
	}
	return results, nil
}

func (s *internetClient) MonitorTraffic(zone string, internetID int64, end time.Time) ([]*RouterMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Internet.Monitor(internetID, param)
	}

	return queryRouterMonitorValue(s.rawClient, zone, end, query)
}
