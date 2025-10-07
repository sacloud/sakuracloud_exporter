// Copyright 2019-2025 The sakuracloud_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/fake"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/iaas-api-go/types"
	diskBuilders "github.com/sacloud/iaas-service-go/disk/builder"
	serverBuilders "github.com/sacloud/iaas-service-go/server/builder"
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

	caller := iaas.NewClient("dummy-token", "dummy-secret")
	createFuncs := []func(caller iaas.APICaller){
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
		go func(f func(caller iaas.APICaller)) {
			f(caller)
			wg.Done()
		}(f)
	}

	wg.Wait()
	log.Println("Done.")
}

func createAutoBackup(caller iaas.APICaller) {
	diskOp := iaas.NewDiskOp(caller)
	disk, err := diskOp.Create(context.Background(), "is1a", &iaas.DiskCreateRequest{
		Name:        "example-disk-for-auto-backup",
		DiskPlanID:  types.DiskPlans.SSD,
		SizeMB:      40 * 1024,
		Description: "desc",
		Tags:        types.Tags{"example", "auto-backup"},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	backupOp := iaas.NewAutoBackupOp(caller)
	_, err = backupOp.Create(context.Background(), "is1a", &iaas.AutoBackupCreateRequest{
		Name:   "example",
		DiskID: disk.ID,
		BackupSpanWeekdays: []types.EDayOfTheWeek{
			types.DaysOfTheWeek.Monday,
			types.DaysOfTheWeek.Wednesday,
			types.DaysOfTheWeek.Friday,
		},
		MaximumNumberOfArchives: 5,
		Description:             "desc",
		Tags:                    types.Tags{"example", "auto-backup"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createDatabase(caller iaas.APICaller) {
	swOp := iaas.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &iaas.SwitchCreateRequest{
		Name:        "example-switch-for-database",
		Description: "desc",
		Tags:        types.Tags{"example", "database"},
	})
	if err != nil {
		log.Fatal(err)
	}

	dbOp := iaas.NewDatabaseOp(caller)
	db, err := dbOp.Create(context.Background(), "is1a", &iaas.DatabaseCreateRequest{
		PlanID:         types.DatabasePlans.DB30GB,
		SwitchID:       sw.ID,
		IPAddresses:    []string{"192.168.0.11"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName: types.RDBMSTypesPostgreSQL.String(),
			DefaultUser:  "user01",
			UserPassword: "dummy-password-01",
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:   5432,
			SourceNetwork: []string{"192.168.0.0/24", "192.168.1.0/24"},
			DefaultUser:   "user01",
			UserPassword:  "dummy-password-01",
		},
		BackupSetting: &iaas.DatabaseSettingBackup{
			Rotate: 3,
			Time:   "00:00",
			DayOfWeek: []types.EDayOfTheWeek{
				types.DaysOfTheWeek.Sunday,
			},
		},
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "database"},
	})
	if err != nil {
		log.Fatal(err)
	}

	waiter := iaas.WaiterForApplianceUp(func() (interface{}, error) {
		return dbOp.Read(context.Background(), "is1a", db.ID)
	}, 10)
	if _, err := waiter.WaitForState(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func createInternet(caller iaas.APICaller) {
	op := iaas.NewInternetOp(caller)
	_, err := op.Create(context.Background(), "is1a", &iaas.InternetCreateRequest{
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

func createLoadBalancer(caller iaas.APICaller) {
	swOp := iaas.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &iaas.SwitchCreateRequest{
		Name:        "example-switch-for-load-balancer-standard",
		Description: "dest",
		Tags:        types.Tags{"example", "load-balancer", "plan=standard"},
	})
	if err != nil {
		log.Fatal(err)
	}

	lbOp := iaas.NewLoadBalancerOp(caller)
	lb, err := lbOp.Create(context.Background(), "is1a", &iaas.LoadBalancerCreateRequest{
		SwitchID:       sw.ID,
		PlanID:         types.LoadBalancerPlans.Standard,
		VRID:           10,
		IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Name:           "example",
		Description:    "desc",
		Tags:           types.Tags{"example", "load-balancer", "plan=standard"},
		VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.101",
				Port:             80,
				DelayLoop:        10,
				SorryServer:      "192.168.0.21",
				Description:      "desc",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.201",
						Port:      80,
						Enabled:   true,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:         "/status",
							ResponseCode: 200,
						},
					},
					{
						IPAddress: "192.168.0.202",
						Port:      80,
						Enabled:   true,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
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

	waiter := iaas.WaiterForApplianceUp(func() (interface{}, error) {
		return lbOp.Read(context.Background(), "is1a", lb.ID)
	}, 10)
	if _, err := waiter.WaitForState(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func createMobileGateway(caller iaas.APICaller) {
	simOp := iaas.NewSIMOp(caller)
	sim, err := simOp.Create(context.Background(), &iaas.SIMCreateRequest{
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "mobile-gateway"},
		ICCID:       "123456789012345",
		PassCode:    "dummy-pass-code",
	})
	if err != nil {
		log.Fatal(err)
	}

	mgwOp := iaas.NewMobileGatewayOp(caller)
	mgw, err := mgwOp.Create(context.Background(), "is1a", &iaas.MobileGatewayCreateRequest{
		Name:                            "example",
		Description:                     "desc",
		Tags:                            types.Tags{"example", "mobile-gateway"},
		InternetConnectionEnabled:       true,
		InterDeviceCommunicationEnabled: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = iaas.WaiterForReady(func() (interface{}, error) {
		return mgwOp.Read(context.Background(), "is1a", mgw.ID)
	}).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.SetTrafficConfig(context.Background(), "is1a", mgw.ID, &iaas.MobileGatewayTrafficControl{
		TrafficQuotaInMB:       1024,
		BandWidthLimitInKbps:   64,
		EmailNotifyEnabled:     true,
		SlackNotifyEnabled:     true,
		SlackNotifyWebhooksURL: "https://xxxxxx.slack.com/example-webhook-url",
		AutoTrafficShaping:     true,
	}); err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.SetDNS(context.Background(), "is1a", mgw.ID, &iaas.MobileGatewayDNSSetting{
		DNS1: "133.242.0.1",
		DNS2: "133.242.0.2",
	}); err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.AddSIM(context.Background(), "is1a", mgw.ID, &iaas.MobileGatewayAddSIMRequest{
		SIMID: sim.ID.String(),
	}); err != nil {
		log.Fatal(err)
	}
	if err := simOp.AssignIP(context.Background(), sim.ID, &iaas.SIMAssignIPRequest{IP: "10.0.0.123"}); err != nil {
		log.Fatal(err)
	}

	if err := mgwOp.Boot(context.Background(), "is1a", mgw.ID); err != nil {
		log.Fatal(err)
	}
	_, err = iaas.WaiterForApplianceUp(func() (interface{}, error) {
		return mgwOp.Read(context.Background(), "is1a", mgw.ID)
	}, 10).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func createNFS(caller iaas.APICaller) {
	swOp := iaas.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &iaas.SwitchCreateRequest{
		Name:        "example-for-nfs",
		Description: "desc",
		Tags:        types.Tags{"example", "nfs"},
	})
	if err != nil {
		log.Fatal(err)
	}

	nfsOp := iaas.NewNFSOp(caller)
	planID, err := query.FindNFSPlanID(context.Background(), iaas.NewNoteOp(caller), types.NFSPlans.HDD, types.NFSHDDSizes.Size100GB)
	if err != nil {
		log.Fatal(err)
	}

	n, err := nfsOp.Create(context.Background(), "is1a", &iaas.NFSCreateRequest{
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

	if _, err := iaas.WaiterForApplianceUp(func() (interface{}, error) {
		return nfsOp.Read(context.Background(), "is1a", n.ID)
	}, 10).WaitForState(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func createProxyLB(caller iaas.APICaller) {
	lbOp := iaas.NewProxyLBOp(caller)

	_, err := lbOp.Create(context.Background(), &iaas.ProxyLBCreateRequest{
		Plan: types.ProxyLBPlans.CPS500,
		HealthCheck: &iaas.ProxyLBHealthCheck{
			Protocol:  types.ProxyLBProtocols.HTTP,
			Path:      "/status",
			DelayLoop: 10,
		},
		SorryServer: &iaas.ProxyLBSorryServer{
			IPAddress: "192.168.0.1",
			Port:      80,
		},
		BindPorts: []*iaas.ProxyLBBindPort{
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
		Servers: []*iaas.ProxyLBServer{
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
		LetsEncrypt: &iaas.ProxyLBACMESetting{
			CommonName: "example.com",
			Enabled:    true,
		},
		StickySession: &iaas.ProxyLBStickySession{
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

func createServer(caller iaas.APICaller) {
	swOp := iaas.NewSwitchOp(caller)
	sw, err := swOp.Create(context.Background(), "is1a", &iaas.SwitchCreateRequest{
		Name:        "example-for-server",
		Description: "desc",
		Tags:        types.Tags{"example", "server"},
	})
	if err != nil {
		log.Fatal(err)
	}

	builder := &serverBuilders.Builder{
		Name:            "example",
		CPU:             2,
		MemoryGB:        4,
		InterfaceDriver: types.InterfaceDrivers.VirtIO,
		Description:     "desc",
		Tags:            types.Tags{"example", "server"},
		BootAfterCreate: true,
		NIC:             &serverBuilders.SharedNICSetting{},
		AdditionalNICs: []serverBuilders.AdditionalNICSettingHolder{
			&serverBuilders.ConnectedNICSetting{SwitchID: sw.ID},
		},
		DiskBuilders: []diskBuilders.Builder{
			&diskBuilders.FromUnixBuilder{
				OSType:      ostype.Ubuntu,
				Name:        "example-disk-for-server",
				SizeGB:      20,
				DistantFrom: nil,
				PlanID:      types.DiskPlans.SSD,
				Connection:  types.DiskConnections.VirtIO,
				Description: "desc",
				Tags:        types.Tags{"example", "server"},
				Client:      diskBuilders.NewBuildersAPIClient(caller),
			},
		},
		Client: serverBuilders.NewBuildersAPIClient(caller),
	}
	if _, err := builder.Build(context.Background(), "is1a"); err != nil {
		log.Fatal(err)
	}
}

func createVPCRouter(caller iaas.APICaller) {
	routerOp := iaas.NewInternetOp(caller)
	router, err := routerOp.Create(context.Background(), "is1a", &iaas.InternetCreateRequest{
		Name:           "example-router-for-vpc",
		Description:    "desc",
		Tags:           types.Tags{"example", "vpc-router"},
		NetworkMaskLen: 28,
		BandWidthMbps:  500,
	})
	if err != nil {
		log.Fatal(err)
	}

	swOp := iaas.NewSwitchOp(caller)
	sw, err := swOp.Read(context.Background(), "is1a", router.Switch.ID)
	if err != nil {
		log.Fatal(err)
	}
	ipaddresses := sw.Subnets[0].GetAssignedIPAddresses()

	vpcOp := iaas.NewVPCRouterOp(caller)
	vpcRouter, err := vpcOp.Create(context.Background(), "is1a", &iaas.VPCRouterCreateRequest{
		Name:        "example",
		Description: "desc",
		Tags:        types.Tags{"example", "vpc-router"},
		PlanID:      types.VPCRouterPlans.HighSpec,
		Switch:      &iaas.ApplianceConnectedSwitch{ID: router.Switch.ID},
		Settings: &iaas.VPCRouterSetting{
			VRID:                      5,
			InternetConnectionEnabled: true,
			Interfaces: []*iaas.VPCRouterInterfaceSetting{
				{
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

	_, err = iaas.WaiterForReady(func() (interface{}, error) {
		return vpcOp.Read(context.Background(), "is1a", vpcRouter.ID)
	}).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := vpcOp.Boot(context.Background(), "is1a", vpcRouter.ID); err != nil {
		log.Fatal(err)
	}

	_, err = iaas.WaiterForApplianceUp(func() (interface{}, error) {
		return vpcOp.Read(context.Background(), "is1a", vpcRouter.ID)
	}, 10).WaitForState(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
