package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type ProxyLBClient interface {
	Find() ([]*sacloud.ProxyLB, error)
	GetCertificate(proxyLBID int64) (*sacloud.ProxyLBCertificates, error)
	Monitor(proxyLBID int64, end time.Time) (*ProxyLBMetrics, error)
}

func getProxyLBClient(client *sakuraAPI.Client) ProxyLBClient {
	return &proxyLBClient{
		rawClient: client,
	}
}

type proxyLBClient struct {
	rawClient *sakuraAPI.Client
}

func (s *proxyLBClient) Find() ([]*sacloud.ProxyLB, error) {
	client := s.rawClient.Clone()

	res, err := client.ProxyLB.Reset().Limit(10000).Include("*").Find()
	if err != nil {
		return nil, err
	}
	var results []*sacloud.ProxyLB
	for i := range res.CommonServiceProxyLBItems {
		results = append(results, &res.CommonServiceProxyLBItems[i])
	}
	return results, nil
}

func (s *proxyLBClient) GetCertificate(proxyLBID int64) (*sacloud.ProxyLBCertificates, error) {
	client := s.rawClient.Clone()
	return client.ProxyLB.GetCertificates(proxyLBID)
}

func (s *proxyLBClient) Monitor(proxyLBID int64, end time.Time) (*ProxyLBMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.ProxyLB.Monitor(proxyLBID, param)
	}

	return queryProxyLBMonitorValue(s.rawClient, "is1a", end, query)
}
