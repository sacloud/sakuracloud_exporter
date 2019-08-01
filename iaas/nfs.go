package iaas

import (
	"context"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/libsacloud/v2/utils/nfs"
)

type NFS struct {
	*sacloud.NFS
	Plan     *nfs.PlanInfo
	PlanName string
	ZoneName string
}

type NFSClient interface {
	Find(ctx context.Context) ([]*NFS, error)
	MonitorFreeDiskSize(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorFreeDiskSizeValue, error)
	MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error)
}

func getNFSClient(caller sacloud.APICaller, zones []string) NFSClient {
	return &nfsClient{
		noteOp: sacloud.NewNoteOp(caller),
		nfsOp:  sacloud.NewNFSOp(caller),
		zones:  zones,
	}
}

type nfsClient struct {
	noteOp sacloud.NoteAPI
	nfsOp  sacloud.NFSAPI
	zones  []string
}

func (c *nfsClient) find(ctx context.Context, zone string) ([]interface{}, error) {
	var results []interface{}
	res, err := c.nfsOp.Find(ctx, zone, &sacloud.FindCondition{
		Count: 10000,
	})
	if err != nil {
		return results, err
	}
	for _, v := range res.NFS {
		planInfo, err := nfs.GetPlanInfo(ctx, c.noteOp, v.PlanID)
		if err != nil {
			return nil, err
		}
		planName := ""
		switch planInfo.DiskPlanID {
		case types.NFSPlans.HDD:
			planName = "HDD"
		case types.NFSPlans.SSD:
			planName = "SSD"
		}
		results = append(results, &NFS{
			NFS:      v,
			PlanName: planName,
			Plan:     planInfo,
			ZoneName: zone,
		})
	}
	return results, err
}

func (c *nfsClient) Find(ctx context.Context) ([]*NFS, error) {
	res, err := queryToZones(ctx, c.zones, c.find)
	if err != nil {
		return nil, err
	}
	var results []*NFS
	for _, s := range res {
		results = append(results, s.(*NFS))
	}
	return results, nil
}

func (c *nfsClient) MonitorFreeDiskSize(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorFreeDiskSizeValue, error) {
	mvs, err := c.nfsOp.MonitorFreeDiskSize(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorFreeDiskSizeValue(mvs.Values), nil
}

func (c *nfsClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	mvs, err := c.nfsOp.MonitorInterface(ctx, zone, id, monitorCondition(end))
	if err != nil {
		return nil, err
	}
	return monitorInterfaceValue(mvs.Values), nil
}
