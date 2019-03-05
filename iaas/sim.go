package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type SIMClient interface {
	Find() ([]*sacloud.SIM, error)
	GetNetworkOperatorConfig(simID int64) (*sacloud.SIMNetworkOperatorConfigs, error)
	MonitorTraffic(simID int64, end time.Time) ([]*SIMMetrics, error)
}

func getSIMClient(client *sakuraAPI.Client) SIMClient {
	return &simClient{
		rawClient: client,
	}
}

type simClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *simClient) Find() ([]*sacloud.SIM, error) {
	client := s.rawClient.Clone()

	res, err := client.SIM.Reset().Limit(10000).Include("*").Include("Status.sim").Find()
	if err != nil {
		return nil, err
	}
	var results []*sacloud.SIM
	for i := range res.CommonServiceSIMItems {
		results = append(results, &res.CommonServiceSIMItems[i])
	}
	return results, nil
}

func (s *simClient) GetNetworkOperatorConfig(simID int64) (*sacloud.SIMNetworkOperatorConfigs, error) {
	client := s.rawClient.Clone()

	return client.SIM.GetNetworkOperator(simID)
}

func (s *simClient) MonitorTraffic(simID int64, end time.Time) ([]*SIMMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.SIM.Monitor(simID, param)
	}

	return querySIMMonitorValue(s.rawClient, "is1a", end, query)
}
