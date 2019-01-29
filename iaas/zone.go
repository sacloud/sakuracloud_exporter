package iaas

import (
	sakuraAPI "github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

// ZoneClient calls SakuraCloud coupon API
type ZoneClient interface {
	Find() ([]*sacloud.Zone, error)
}

func getZoneClient(client *sakuraAPI.Client) ZoneClient {
	return &zoneClient{rawClient: client}
}

type zoneClient struct {
	rawClient *sakuraAPI.Client
}

func (s *zoneClient) Find() ([]*sacloud.Zone, error) {
	client := s.rawClient.Clone()
	client.Zone = "is1a"

	res, err := client.GetZoneAPI().Find()
	if err != nil {
		return nil, err
	}

	var results []*sacloud.Zone
	for i := range res.Zones {
		results = append(results, &res.Zones[i])
	}
	return results, nil
}
