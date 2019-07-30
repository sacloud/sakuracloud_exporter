package iaas

import (
	"context"
	"fmt"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/search"
	"github.com/sacloud/libsacloud/v2/sacloud/search/keys"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type AutoBackupClient interface {
	Find(ctx context.Context) ([]*sacloud.AutoBackup, error)
	ListBackups(ctx context.Context, zone string, autoBackupID types.ID) ([]*sacloud.Archive, error)
}

func getAutoBackupClient(caller sacloud.APICaller, zones []string) AutoBackupClient {
	return &autoBackupClient{
		caller: caller,
		zones:  zones,
	}
}

type autoBackupClient struct {
	caller sacloud.APICaller
	zones  []string
}

func (c *autoBackupClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	client := sacloud.NewAutoBackupOp(c.caller)
	searched, err := client.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return nil, err
	}
	var res []interface{}
	for _, v := range searched.AutoBackups {
		res = append(res, v)
	}
	return res, nil
}

func (c *autoBackupClient) Find(ctx context.Context) ([]*sacloud.AutoBackup, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*sacloud.AutoBackup
	for _, v := range res {
		results = append(results, v.(*sacloud.AutoBackup))
	}
	return results, nil
}

func (c *autoBackupClient) ListBackups(ctx context.Context, zone string, autoBackupID types.ID) ([]*sacloud.Archive, error) {

	client := sacloud.NewArchiveOp(c.caller)
	tagName := fmt.Sprintf("autobackup-%d", autoBackupID)

	searched, err := client.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
		Filter: search.Filter{
			search.Key(keys.Tags): search.TagsAndEqual(tagName),
		},
	})
	if err != nil {
		return nil, err
	}

	var res []*sacloud.Archive
	for _, v := range searched.Archives {
		if v.Availability.IsAvailable() {
			res = append(res, v)
		}
	}
	return res, err
}
