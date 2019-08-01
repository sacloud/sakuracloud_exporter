package iaas

import (
	"context"

	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ZoneClient calls SakuraCloud zone API
type ZoneClient interface {
	Find(ctx context.Context) ([]*sacloud.Zone, error)
}

func getZoneClient(caller sacloud.APICaller) ZoneClient {
	return &zoneClient{
		client: sacloud.NewZoneOp(caller),
	}
}

type zoneClient struct {
	client sacloud.ZoneAPI
}

func (c *zoneClient) Find(ctx context.Context) ([]*sacloud.Zone, error) {
	res, err := c.client.Find(ctx, nil)
	if err != nil {
		return nil, err
	}
	return res.Zones, nil
}
