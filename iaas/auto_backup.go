package iaas

import (
	"fmt"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type AutoBackupClient interface {
	Find() ([]*sacloud.AutoBackup, error)
	ListBackups(zone string, autoBackupID int64) ([]*sacloud.Archive, error)
}

func getAutoBackupClient(client *sakuraAPI.Client) AutoBackupClient {
	return &autoBackupClient{
		rawClient: client,
	}
}

type autoBackupClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *autoBackupClient) Find() ([]*sacloud.AutoBackup, error) {
	client := s.rawClient.Clone()
	client.Zone = "is1a"

	res, err := client.AutoBackup.Reset().Limit(10000).Find()
	if err != nil {
		return nil, err
	}
	var results []*sacloud.AutoBackup
	for i := range res.CommonServiceAutoBackupItems {
		results = append(results, &res.CommonServiceAutoBackupItems[i])
	}
	return results, nil
}

func (s *autoBackupClient) ListBackups(zone string, autoBackupID int64) ([]*sacloud.Archive, error) {

	client := s.rawClient.Clone()
	client.Zone = zone

	tagName := fmt.Sprintf("autobackup-%d", autoBackupID)

	res, err := client.Archive.Reset().Limit(100).WithTag(tagName).Find()
	if err != nil {
		return nil, err
	}

	var results []*sacloud.Archive
	for i := range res.Archives {
		if res.Archives[i].IsAvailable() {
			results = append(results, &res.Archives[i])
		}
	}
	return results, err
}
