package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type vCloudConfig struct {
	VcdClient *govcd.VCDClient
	Client    *govcd.Client
	Org       *govcd.Org
	Vdc       *govcd.Vdc
}

var vConfig vCloudConfig

func vcloudAuth(username, password string) error {
	_, err := vConfig.VcdClient.GetAuthResponse(username, password, vOrg)
	return err
}

func initVCD() error {
	apiEndPoint, err := url.Parse(vHref)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("[vCloud] Creating VCD client...")
	vcdClient := govcd.NewVCDClient(*apiEndPoint, vInsecure)
	fmt.Println("[vCloud] Authenticating...")
	err = vcdClient.Authenticate(codyConf.VCloudUsername, codyConf.VCloudPassword, vOrg)
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

	vapp = strings.TrimSpace(vapp)
	user = strings.TrimSpace(user)
	name := user + "-" + vapp
	storageProfileName := "Defsec"
	networkName := "DefSec_01"

	// get catalocg
	catalog, err := vConfig.Org.FindCatalog("Other") // TODO change to DefSec_Lessons
	if err != nil {
		return "", err
	} else if catalog.Catalog == nil {
		return "", errors.New("Catalog not found")
	}

	item, err := catalog.GetCatalogItemByName(strings.TrimSpace(vapp), true)
	if err != nil {
		return "", err
	} else if item == nil {
		return "", errors.New("Catalog item not found")
	}
	fmt.Println(item)

	// find vapp template
	vappTemplate, err := item.GetVAppTemplate()
	if err != nil {
		panic(err)
	}

	var networks []*types.OrgVDCNetwork
	net, err := vConfig.Vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return "", fmt.Errorf("error finding network : %s, err: %s",
			networkName, err)
	}
	networks = append(networks, net.OrgVDCNetwork)

	// Get StorageProfileReference
	storageProfileRef, err := vConfig.Vdc.FindStorageProfileReference(storageProfileName)
	if err != nil {
		return "", fmt.Errorf("error finding storage profile: %s", err)
	}

	task, err := vConfig.Vdc.ComposeVApp(networks, vappTemplate, storageProfileRef, name, "description", true)
	if err != nil {
		//return "", fmt.Errorf("error composing vapp: %s", err)
		return "", errors.New("VApp already exists!")
	}

	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	err = task.WaitTaskCompletion()
	if err != nil {
		return "", fmt.Errorf("error composing vapp: %s", err)
	}
	// Get VApp
	vappLive, err := vConfig.Vdc.GetVAppByName(name, true)
	if err != nil {
		return "", fmt.Errorf("error getting vapp: %s", err)
	}

	err = vappLive.BlockWhileStatus("UNRESOLVED", vConfig.Client.MaxRetryTimeout)
	if err != nil {
		return "", fmt.Errorf("error waiting for created test vApp to have working state: %s", err)
	}
	if err != nil {
		fmt.Println("error creating vApp: %s", err)
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

/*
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
*/
