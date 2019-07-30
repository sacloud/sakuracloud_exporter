package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/libsacloud/v2/utils/newsfeed"
)

type ServerClient interface {
	Find(ctx context.Context) ([]*Server, error)
	MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorCPUTimeValue, error)
	MonitorDisk(ctx context.Context, zone string, diskID types.ID, end time.Time) (*sacloud.MonitorDiskValue, error)
	MonitorNIC(ctx context.Context, zone string, nicID types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error)
	MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error)
}

type Server struct {
	*sacloud.Server
	ZoneName string
}

func getServerClient(caller sacloud.APICaller, zones []string) ServerClient {
	return &serverClient{
		serverOp:    sacloud.NewServerOp(caller),
		diskOp:      sacloud.NewDiskOp(caller),
		interfaceOp: sacloud.NewInterfaceOp(caller),
		zones:       zones,
	}
}

type serverClient struct {
	serverOp    sacloud.ServerAPI
	diskOp      sacloud.DiskAPI
	interfaceOp sacloud.InterfaceAPI
	zones       []string
}

func (c *serverClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.serverOp.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, s := range res.Servers {
		results = append(results, &Server{
			Server:   s,
			ZoneName: zone,
		})
	}
	return results, err
}

func (c *serverClient) Find(ctx context.Context) ([]*Server, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*Server
	for _, s := range res {
		results = append(results, s.(*Server))
	}
	return results, nil
}

func (c *serverClient) MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorCPUTimeValue, error) {
	mvs, err := c.serverOp.Monitor(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorCPUTimeValue(mvs.Values), nil
}

func (c *serverClient) MonitorDisk(ctx context.Context, zone string, diskID types.ID, end time.Time) (*sacloud.MonitorDiskValue, error) {
	mvs, err := c.diskOp.Monitor(ctx, zone, diskID, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorDiskValue(mvs.Values), nil
}

func (c *serverClient) MonitorNIC(ctx context.Context, zone string, nicID types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	mvs, err := c.interfaceOp.Monitor(ctx, zone, nicID, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}

func (c *serverClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return newsfeed.GetByURL(infoURL)
}
