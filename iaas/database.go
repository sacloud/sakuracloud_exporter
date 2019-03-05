package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type Database struct {
	*sacloud.Database
	ZoneName string
}

type DatabaseClient interface {
	Find() ([]*Database, error)
	MonitorDatabase(zone string, diskID int64, end time.Time) ([]*DatabaseMetrics, error)
	MonitorCPU(zone string, databaseID int64, end time.Time) ([]*sacloud.FlatMonitorValue, error)
	MonitorNIC(zone string, nicID int64, end time.Time) ([]*NICMetrics, error)
	MonitorDisk(zone string, diskID int64, end time.Time) ([]*DiskMetrics, error)
}

func getDatabaseClient(client *sakuraAPI.Client, zones []string) DatabaseClient {
	return &databaseClient{
		rawClient: client,
		zones:     zones,
	}
}

type databaseClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *databaseClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.Database.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.Databases {
		results = append(results, &Database{
			Database: &res.Databases[i],
			ZoneName: c.Zone,
		})
	}
	return results, err
}

func (s *databaseClient) Find() ([]*Database, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*Database
	for _, s := range res {
		results = append(results, s.(*Database))
	}
	return results, nil
}

func (s *databaseClient) MonitorDatabase(zone string, databaseID int64, end time.Time) ([]*DatabaseMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Database.MonitorDatabase(databaseID, param)
	}

	return queryDatabaseMonitorValue(s.rawClient, zone, end, query)
}

func (s *databaseClient) MonitorCPU(zone string, databaseID int64, end time.Time) ([]*sacloud.FlatMonitorValue, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Database.MonitorCPU(databaseID, param)
	}

	return queryCPUTimeMonitorValue(s.rawClient, zone, end, query)
}

func (s *databaseClient) MonitorDisk(zone string, databaseID int64, end time.Time) ([]*DiskMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Database.MonitorSystemDisk(databaseID, param)
	}

	return queryDiskMonitorValue(s.rawClient, zone, end, query)
}

func (s *databaseClient) MonitorNIC(zone string, databaseID int64, end time.Time) ([]*NICMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.Database.MonitorInterface(databaseID, param)
	}

	return queryNICMonitorValue(s.rawClient, zone, end, query)
}
