package main

import (
	"context"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/fake"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/libsacloud/v2/utils/nfs"
	"github.com/sacloud/libsacloud/v2/utils/server"
	"github.com/sacloud/libsacloud/v2/utils/server/ostype"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const fakeStoreFileName = "example-fake-store.json"

func main() {

	log.Println("generate example fake-store.json: start")
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// setup fake store
	fake.DataStore = fake.NewJSONFileStore(filepath.Join(curDir, fakeStoreFileName))
	// switch sacloud client to fake driver
	fake.SwitchFactoryFuncToFake()

	caller := sacloud.NewClient("dummy-token", "dummy-secret")
	createFuncs := []func(caller sacloud.APICaller){
		createAutoBackup,
		createDatabase,
		createInternet,
		createLoadBalancer,
		createMobileGateway,
		createNFS,
		createProxyLB,
		createServer,
		createVPCRouter,
	}

	var wg sync.WaitGroup
	wg.Add(len(createFuncs))
	for _, f := range createFuncs {
		go func(f func(caller sacloud.APICaller)) {
			f(caller)
			wg.Done()
		}(f)
	}

	wg.Wait()
	log.Println("Done.")
}

func createAutoBackup(caller sacloud.APICaller) {
	diskOp := sacloud.NewDiskOp(caller)
	disk, err := diskOp.Create(context.Background(), "is1a", &sacloud.DiskCreateRequest{
		Name:        "example-disk-for-auto-backup",
		DiskPlanID:  types.DiskPlans.SSD,
		SizeMB:      40 * 1024,
		Description: "desc",
		Tags:        types.Tags{"example", "auto-backup"},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	backupOp := sacloud.NewAutoBackupOp(caller)
	_, err = backupOp.Create(context.Background(), "is1a", &sacloud.AutoBackupCreateRequest{
		Name:   "example",
		DiskID: disk.ID,
		BackupSpanWeekdays: []types.EBackupSpanWeekday{
			types.BackupSpanWeekdays.Monday,
			types.BackupSpanWeekdays.Wednesday,
			types.BackupSpanWeekdays.Friday,
		},
		MaximumNumberOfArchives: 5,
		Description:             "desc",
		Tags:                    types.Tags{"example", "auto-backup"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createDatabase(caller sacloud.APICaller) {
	swOp := sacloud.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &sacloud.SwitchCreateRequest{
		Name:        "example-switch-for-database",
		Description: "desc",
		Tags:        types.Tags{"example", "database"},
	})
	if err != nil {
		log.Fatal(err)
	}

	dbOp := sacloud.NewDatabaseOp(caller)
	db, err := dbOp.Create(context.Background(), "is1a", &sacloud.DatabaseCreateRequest{
		PlanID:         types.DatabasePlans.DB30GB,
		SwitchID:       sw.ID,
		IPAddresses:    []string{"192.168.0.11"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Conf: &sacloud.DatabaseRemarkDBConfCommon{
			DatabaseName:     types.RDBMSVersions[types.RDBMSTypesPostgreSQL].Name,
			DatabaseVersion:  types.RDBMSVersions[types.RDBMSTypesPostgreSQL].Version,
			DatabaseRevision: types.RDBMSVersions[types.RDBMSTypesPostgreSQL].Revision,
			DefaultUser:      "user01",
			UserPassword:     "dummy-password-01",
		},
		CommonSetting: &sacloud.DatabaseSettingCommon{
			ServicePort:   5432,
			SourceNetwork: []string{"192.168.0.0/24", "192.168.1.0/24"},
			DefaultUser:   "user01",
			UserPassword:  "dummy-password-01",
		},
		BackupSetting: &sacloud.DatabaseSettingBackup{
			Rotate: 3,
			Time:   "00:00",
			DayOfWeek: []types.EBackupSpanWeekday{
				types.BackupSpanWeekdays.Sunday,
			},
		},
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "database"},
	})
	if err != nil {
		log.Fatal(err)
	}

	waiter := sacloud.WaiterForApplianceUp(func() (interface{}, error) {
		return dbOp.Read(context.Background(), "is1a", db.ID)
	}, 10)
	if _, err := waiter.WaitForState(context.Background()); err != nil {
		log.Fatal(err)
	}

}

func createInternet(caller sacloud.APICaller) {
	op := sacloud.NewInternetOp(caller)
	_, err := op.Create(context.Background(), "is1a", &sacloud.InternetCreateRequest{
		Name:           "example",
		Description:    "desc",
		Tags:           types.Tags{"example", "switch+router"},
		NetworkMaskLen: 28,
		BandWidthMbps:  500,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createLoadBalancer(caller sacloud.APICaller) {
	swOp := sacloud.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &sacloud.SwitchCreateRequest{
		Name:        "example-switch-for-load-balancer-standard",
		Description: "dest",
		Tags:        types.Tags{"example", "load-balancer", "plan=standard"},
	})
	if err != nil {
		log.Fatal(err)
	}

	lbOp := sacloud.NewLoadBalancerOp(caller)
	lb, err := lbOp.Create(context.Background(), "is1a", &sacloud.LoadBalancerCreateRequest{
		SwitchID:       sw.ID,
		PlanID:         types.LoadBalancerPlans.Standard,
		VRID:           10,
		IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Name:           "example",
		Description:    "desc",
		Tags:           types.Tags{"example", "load-balancer", "plan=standard"},
		VirtualIPAddresses: []*sacloud.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.101",
				Port:             80,
				DelayLoop:        10,
				SorryServer:      "192.168.0.21",
				Description:      "desc",
				Servers: []*sacloud.LoadBalancerServer{
					{
						IPAddress: "192.168.0.201",
						Port:      80,
						Enabled:   true,
						HealthCheck: &sacloud.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:         "/status",
							ResponseCode: 200,
						},
					},
					{
						IPAddress: "192.168.0.202",
						Port:      80,
						Enabled:   true,
						HealthCheck: &sacloud.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:         "/status",
							ResponseCode: 200,
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	waiter := sacloud.WaiterForApplianceUp(func() (interface{}, error) {
		return lbOp.Read(context.Background(), "is1a", lb.ID)
	}, 10)
	if _, err := waiter.WaitForState(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func createMobileGateway(caller sacloud.APICaller) {
	simOp := sacloud.NewSIMOp(caller)
	sim, err := simOp.Create(context.Background(), &sacloud.SIMCreateRequest{
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "mobile-gateway"},
		ICCID:       "123456789012345",
		PassCode:    "dummy-pass-code",
	})
	if err != nil {
		log.Fatal(err)
	}

	mgwOp := sacloud.NewMobileGatewayOp(caller)
	mgw, err := mgwOp.Create(context.Background(), "is1a", &sacloud.MobileGatewayCreateRequest{
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "mobile-gateway"},
		Settings: &sacloud.MobileGatewaySettingCreate{
			InternetConnectionEnabled:       true,
			InterDeviceCommunicationEnabled: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = sacloud.WaiterForReady(func() (interface{}, error) {
		return mgwOp.Read(context.Background(), "is1a", mgw.ID)
	}).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.SetTrafficConfig(context.Background(), "is1a", mgw.ID, &sacloud.MobileGatewayTrafficControl{
		TrafficQuotaInMB:       1024,
		BandWidthLimitInKbps:   64,
		EmailNotifyEnabled:     true,
		SlackNotifyEnabled:     true,
		SlackNotifyWebhooksURL: "https://xxxxxx.slack.com/example-webhook-url",
		AutoTrafficShaping:     true,
	}); err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.SetDNS(context.Background(), "is1a", mgw.ID, &sacloud.MobileGatewayDNSSetting{
		DNS1: "133.242.0.1",
		DNS2: "133.242.0.2",
	}); err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.AddSIM(context.Background(), "is1a", mgw.ID, &sacloud.MobileGatewayAddSIMRequest{
		SIMID: sim.ID.String(),
	}); err != nil {
		log.Fatal(err)
	}
	if err := simOp.AssignIP(context.Background(), sim.ID, &sacloud.SIMAssignIPRequest{IP: "10.0.0.123"}); err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.Boot(context.Background(), "is1a", mgw.ID); err != nil {
		log.Fatal(err)
	}
	_, err = sacloud.WaiterForApplianceUp(func() (interface{}, error) {
		return mgwOp.Read(context.Background(), "is1a", mgw.ID)
	}, 10).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func createNFS(caller sacloud.APICaller) {
	swOp := sacloud.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &sacloud.SwitchCreateRequest{
		Name:        "example-for-nfs",
		Description: "desc",
		Tags:        types.Tags{"example", "nfs"},
	})
	if err != nil {
		log.Fatal(err)
	}

	nfsOp := sacloud.NewNFSOp(caller)
	planID, err := nfs.FindNFSPlanID(context.Background(), sacloud.NewNoteOp(caller), types.NFSPlans.HDD, types.NFSHDDSizes.Size100GB)
	if err != nil {
		log.Fatal(err)
	}

	n, err := nfsOp.Create(context.Background(), "is1a", &sacloud.NFSCreateRequest{
		SwitchID:       sw.ID,
		PlanID:         planID,
		IPAddresses:    []string{"192.168.0.11"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Name:           "example",
		Description:    "desc",
		Tags:           types.Tags{"example", "nfs"},
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := sacloud.WaiterForApplianceUp(func() (interface{}, error) {
		return nfsOp.Read(context.Background(), "is1a", n.ID)
	}, 10).WaitForState(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func createProxyLB(caller sacloud.APICaller) {
	lbOp := sacloud.NewProxyLBOp(caller)

	_, err := lbOp.Create(context.Background(), &sacloud.ProxyLBCreateRequest{
		Plan: types.ProxyLBPlans.CPS500,
		HealthCheck: &sacloud.ProxyLBHealthCheck{
			Protocol:  types.ProxyLBProtocols.HTTP,
			Path:      "/status",
			DelayLoop: 10,
		},
		SorryServer: &sacloud.ProxyLBSorryServer{
			IPAddress: "192.168.0.1",
			Port:      80,
		},
		BindPorts: []*sacloud.ProxyLBBindPort{
			{
				ProxyMode:       types.ProxyLBProxyModes.HTTP,
				Port:            80,
				RedirectToHTTPS: true,
			},
			{
				ProxyMode:    types.ProxyLBProxyModes.HTTPS,
				Port:         443,
				SupportHTTP2: true,
			},
		},
		Servers: []*sacloud.ProxyLBServer{
			{
				IPAddress: "192.0.2.1",
				Port:      80,
				Enabled:   true,
			},
			{
				IPAddress: "192.0.2.2",
				Port:      80,
				Enabled:   true,
			},
		},
		LetsEncrypt: &sacloud.ProxyLBACMESetting{
			CommonName: "example.com",
			Enabled:    true,
		},
		StickySession: &sacloud.ProxyLBStickySession{
			Enabled: true,
		},
		UseVIPFailover: true,
		Region:         types.ProxyLBRegions.IS1,
		Name:           "example",
		Description:    "desc",
		Tags:           types.Tags{"example", "proxylb"},
	})

	if err != nil {
		log.Fatal(err)
	}
}

func createServer(caller sacloud.APICaller) {
	swOp := sacloud.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &sacloud.SwitchCreateRequest{
		Name:        "example-for-server",
		Description: "desc",
		Tags:        types.Tags{"example", "server"},
	})
	if err != nil {
		log.Fatal(err)
	}

	client := server.NewBuildersAPIClient(caller)

	builder := &server.Builder{
		Name:            "example",
		CPU:             2,
		MemoryGB:        4,
		InterfaceDriver: types.InterfaceDrivers.VirtIO,
		Description:     "desc",
		Tags:            types.Tags{"example", "server"},
		BootAfterCreate: true,
		NIC:             &server.SharedNICSetting{},
		AdditionalNICs: []server.AdditionalNICSettingHolder{
			&server.ConnectedNICSetting{SwitchID: sw.ID},
		},
		DiskBuilders: []server.DiskBuilder{
			&server.FromUnixDiskBuilder{
				OSType:      ostype.Ubuntu,
				Name:        "example-disk-for-server",
				SizeGB:      20,
				DistantFrom: nil,
				PlanID:      types.DiskPlans.SSD,
				Connection:  types.DiskConnections.VirtIO,
				Description: "desc",
				Tags:        types.Tags{"example", "server"},
			},
		},
	}
	if _, err := builder.Build(context.Background(), client, "is1a"); err != nil {
		log.Fatal(err)
	}
}

func createVPCRouter(caller sacloud.APICaller) {
	routerOp := sacloud.NewInternetOp(caller)
	router, err := routerOp.Create(context.Background(), "is1a", &sacloud.InternetCreateRequest{
		Name:           "example-router-for-vpc",
		Description:    "desc",
		Tags:           types.Tags{"example", "vpc-router"},
		NetworkMaskLen: 28,
		BandWidthMbps:  500,
	})
	if err != nil {
		log.Fatal(err)
	}

	swOp := sacloud.NewSwitchOp(caller)
	sw, err := swOp.Read(context.Background(), "is1a", router.Switch.ID)
	if err != nil {
		log.Fatal(err)
	}
	ipaddresses := sw.Subnets[0].GetAssignedIPAddresses()

	vpcOp := sacloud.NewVPCRouterOp(caller)
	vpcRouter, err := vpcOp.Create(context.Background(), "is1a", &sacloud.VPCRouterCreateRequest{
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "vpc-router"},
		PlanID:      types.VPCRouterPlans.HighSpec,
		Switch:      &sacloud.ApplianceConnectedSwitch{ID: router.Switch.ID},
		Settings: &sacloud.VPCRouterSetting{
			VRID:                      5,
			InternetConnectionEnabled: true,
			Interfaces: []*sacloud.VPCRouterInterfaceSetting{
				{
					Enabled:          true,
					IPAddress:        []string{ipaddresses[1], ipaddresses[2]},
					VirtualIPAddress: ipaddresses[0],
					NetworkMaskLen:   router.NetworkMaskLen,
					Index:            0,
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = sacloud.WaiterForReady(func() (interface{}, error) {
		return vpcOp.Read(context.Background(), "is1a", vpcRouter.ID)
	}).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := vpcOp.Boot(context.Background(), "is1a", vpcRouter.ID); err != nil {
		log.Fatal(err)
	}

	_, err = sacloud.WaiterForApplianceUp(func() (interface{}, error) {
		return vpcOp.Read(context.Background(), "is1a", vpcRouter.ID)
	}, 10).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
