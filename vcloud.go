package main

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/govcd"
	"github.com/vmware/go-vcloud-director/types/v56"
)

type vCloudConfig struct {
	VcdClient *govcd.VCDClient
	Client    *govcd.Client
	Org       *govcd.Org
	Vdc       *govcd.Vdc
}

var vConfig vCloudConfig

func initVCD() error {
	apiEndPoint, err := url.Parse(vHref)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("[vCloud] Creating VCD client...")
	vcdClient := govcd.NewVCDClient(*apiEndPoint, vInsecure)
	fmt.Println("[vCloud] Authenticating...")
	err = vcdClient.Authenticate(vCloudUsername, vCloudPassword, vOrg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("[vCloud] Getting org...")
	org, err := vcdClient.GetOrgByName(vOrg)
	if err != nil {
		fmt.Println("entity", vOrg, "not found:", err)
		return err
	}
	fmt.Println("[vCloud] Getting vdc...")
	vdc, err := org.GetVDCByName(vVDC, true)
	if err != nil {
		fmt.Println(err)
		return err
	}

	vConfig = vCloudConfig{
		VcdClient: vcdClient,
		Client:    &vcdClient.Client,
		Org:       org,
		Vdc:       vdc,
	}
	return nil
}

func vappDeploy(vapp string, user string) (string, error) {
	if !validateName(vapp) {
		return "", errors.New("Invalid name")
	}
	return "id? vapp obj?", nil
	// DEPLOY VAPP TO USER
}

func vappPowerAndIPs(id string) (string, error) {
	// POWER ON VAPP AND GET IPS.
	return "ip1", nil
}

func takeBoolPointer(value bool) *bool {
	return &value
}

func takeIntPointer(value int) *int {
	return &value
}

func testvCloud() {

	vappName := "CoolBroMan"
	vmName := "yep"
	err := vConfig.Vdc.ComposeRawVApp(vappName)
	if err != nil {
		fmt.Println("error creating vApp: %s", err)
	}

	vapp, err := vConfig.Vdc.GetVAppByName(vappName, true)
	if err != nil {
		fmt.Println("unable to find vApp by name %s: %s", vappName, err)
	}
	// must wait until the vApp exits
	err = vapp.BlockWhileStatus("UNRESOLVED", vConfig.VcdClient.Client.MaxRetryTimeout)
	if err != nil {
		fmt.Println("error waiting for created test vApp to have working state: %s", err)
	}

	desiredNetConfig := &types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 2
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeNone,
			Network:                 types.NoneNetwork,
			NetworkConnectionIndex:  1,
		},
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModeNone,
			Network:                 types.NoneNetwork,
			NetworkConnectionIndex:  2,
		})

	newDisk := types.DiskSettings{
		AdapterType:       "5",
		SizeMb:            int64(16384),
		BusNumber:         0,
		UnitNumber:        0,
		ThinProvisioned:   takeBoolPointer(true),
		OverrideVmDefault: true,
	}
	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      vmName,
			NetworkConnectionSection:  desiredNetConfig,
			Description:               "created by test",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntPointer(2),
				NumCoresPerSocket: takeIntPointer(1),
				CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
				MemoryResourceMb:  &types.MemoryResourceMb{Configured: 512},
				MediaSection:      nil,
				DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
				HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"}, // need support older version vCD
				VirtualCpuType:    "VM32",
			},
		},
		AllEULAsAccepted: true,
	}

	vm, err := vapp.AddEmptyVm(requestDetails)
	if err != nil {
		fmt.Println("error creating empty VM: %s", err)
	}
	fmt.Println(vm)
}

func vcloudAuth(username, password string) error {
	_, err := vConfig.VcdClient.GetAuthResponse(username, password, vOrg)
	return err
}
