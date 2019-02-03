package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type ServerClient interface {
	Find() ([]*sacloud.Server, error)
	MonitorCPU(zone string, serverID int64, end time.Time) (*sacloud.FlatMonitorValue, error)
	MonitorDisk(zone string, diskID int64, end time.Time) (*DiskMetrics, error)
	MonitorNIC(zone string, nicID int64, end time.Time) (*NICMetrics, error)
}

func getServerClient(client *sakuraAPI.Client, zones []string) ServerClient {
	return &serverClient{
		rawClient: client,
		zones:     zones,
	}
}

type serverClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *serverClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.Server.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.Servers {
		results = append(results, &res.Servers[i])
	}
	return results, err
}

func (s *serverClient) Find() ([]*sacloud.Server, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*sacloud.Server
	for _, s := range res {
		results = append(results, s.(*sacloud.Server))
	}
	return results, nil
}

func (s *serverClient) MonitorCPU(zone string, serverID int64, end time.Time) (*sacloud.FlatMonitorValue, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Server.Monitor(serverID, param)
	}

	return queryCPUTimeMonitorValue(s.rawClient, zone, end, query)
}

func (s *serverClient) MonitorDisk(zone string, diskID int64, end time.Time) (*DiskMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Disk.Monitor(diskID, param)
	}

	return queryDiskMonitorValue(s.rawClient, zone, end, query)
}

func (s *serverClient) MonitorNIC(zone string, nicID int64, end time.Time) (*NICMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Interface.Monitor(nicID, param)
	}

	return queryNICMonitorValue(s.rawClient, zone, end, query)
}
