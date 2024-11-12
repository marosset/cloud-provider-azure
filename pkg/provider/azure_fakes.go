/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	"time"

	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/mock_azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/privatezoneclient/mock_privatezoneclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/securitygroupclient/mock_securitygroupclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/subnetclient/mock_subnetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualnetworklinkclient/mock_virtualnetworklinkclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/diskclient/mockdiskclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/interfaceclient/mockinterfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/loadbalancerclient/mockloadbalancerclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/publicipclient/mockpublicipclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/subnetclient/mocksubnetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient/mockvmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"
	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/config"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/privatelinkservice"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/routetable"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/securitygroup"
	"sigs.k8s.io/cloud-provider-azure/pkg/provider/subnet"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

// NewTestScaleSet creates a fake ScaleSet for unit test
func NewTestScaleSet(ctrl *gomock.Controller) (*ScaleSet, error) {
	return newTestScaleSetWithState(ctrl)
}

func newTestScaleSetWithState(ctrl *gomock.Controller) (*ScaleSet, error) {
	cloud := GetTestCloud(ctrl)
	ss, err := newScaleSet(cloud)
	if err != nil {
		return nil, err
	}

	return ss.(*ScaleSet), nil
}

func NewTestFlexScaleSet(ctrl *gomock.Controller) (*FlexScaleSet, error) {
	cloud := GetTestCloud(ctrl)
	fs, err := newFlexScaleSet(cloud)
	if err != nil {
		return nil, err
	}

	return fs.(*FlexScaleSet), nil
}

// GetTestCloud returns a fake azure cloud for unit tests in Azure related CSI drivers
func GetTestCloud(ctrl *gomock.Controller) (az *Cloud) {
	az = &Cloud{
		Config: Config{
			AzureClientConfig: config.AzureClientConfig{
				ARMClientConfig: azclient.ARMClientConfig{
					TenantID: "TenantID",
				},
				AzureAuthConfig: azclient.AzureAuthConfig{},
				SubscriptionID:  "subscription",
			},
			ResourceGroup:                            "rg",
			VnetResourceGroup:                        "rg",
			RouteTableResourceGroup:                  "rg",
			SecurityGroupResourceGroup:               "rg",
			PrivateLinkServiceResourceGroup:          "rg",
			Location:                                 "westus",
			VnetName:                                 "vnet",
			SubnetName:                               "subnet",
			SecurityGroupName:                        "nsg",
			RouteTableName:                           "rt",
			PrimaryAvailabilitySetName:               "as",
			PrimaryScaleSetName:                      "vmss",
			MaximumLoadBalancerRuleCount:             250,
			VMType:                                   consts.VMTypeStandard,
			LoadBalancerBackendPoolConfigurationType: consts.LoadBalancerBackendPoolConfigurationTypeNodeIPConfiguration,
		},
		nodeZones:                map[string]*utilsets.IgnoreCaseSet{},
		nodeInformerSynced:       func() bool { return true },
		nodeResourceGroups:       map[string]string{},
		unmanagedNodes:           utilsets.NewString(),
		excludeLoadBalancerNodes: utilsets.NewString(),
		nodePrivateIPs:           map[string]*utilsets.IgnoreCaseSet{},
		routeCIDRs:               map[string]string{},
		eventRecorder:            &record.FakeRecorder{},
		lockMap:                  newLockMap(),
	}
	az.DisksClient = mockdiskclient.NewMockInterface(ctrl)
	az.InterfacesClient = mockinterfaceclient.NewMockInterface(ctrl)
	az.LoadBalancerClient = mockloadbalancerclient.NewMockInterface(ctrl)
	az.PublicIPAddressesClient = mockpublicipclient.NewMockInterface(ctrl)
	az.SubnetsClient = mocksubnetclient.NewMockInterface(ctrl)
	az.VirtualMachineScaleSetsClient = mockvmssclient.NewMockInterface(ctrl)
	az.VirtualMachineScaleSetVMsClient = mockvmssvmclient.NewMockInterface(ctrl)
	az.VirtualMachinesClient = mockvmclient.NewMockInterface(ctrl)
	clientFactory := mock_azclient.NewMockClientFactory(ctrl)
	az.ComputeClientFactory = clientFactory
	az.NetworkClientFactory = clientFactory
	securtyGrouptrack2Client := mock_securitygroupclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetSecurityGroupClient().Return(securtyGrouptrack2Client).AnyTimes()
	mockPrivateDNSClient := mock_privatezoneclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetPrivateZoneClient().Return(mockPrivateDNSClient).AnyTimes()
	virtualNetworkLinkClient := mock_virtualnetworklinkclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetVirtualNetworkLinkClient().Return(virtualNetworkLinkClient).AnyTimes()
	subnetTrack2Client := mock_subnetclient.NewMockInterface(ctrl)
	clientFactory.EXPECT().GetSubnetClient().Return(subnetTrack2Client).AnyTimes()
	az.AuthProvider = &azclient.AuthProvider{
		ComputeCredential: mock_azclient.NewMockTokenCredential(ctrl),
	}
	az.VMSet, _ = newAvailabilitySet(az)
	az.vmCache, _ = az.newVMCache()
	az.lbCache, _ = az.newLBCache()
	az.nsgRepo, _ = securitygroup.NewSecurityGroupRepo(az.SecurityGroupResourceGroup, az.SecurityGroupName, az.NsgCacheTTLInSeconds, az.Config.DisableAPICallCache, securtyGrouptrack2Client)
	az.subnetRepo = subnet.NewMockRepository(ctrl)
	az.pipCache, _ = az.newPIPCache()
	az.LoadBalancerBackendPool = NewMockBackendPool(ctrl)

	az.plsRepo = privatelinkservice.NewMockRepository(ctrl)
	az.routeTableRepo = routetable.NewMockRepository(ctrl)

	getter := func(_ context.Context, _ string) (interface{}, error) { return nil, nil }
	az.storageAccountCache, _ = azcache.NewTimedCache(time.Minute, getter, az.Config.DisableAPICallCache)
	az.fileServicePropertiesCache, _ = azcache.NewTimedCache(5*time.Minute, getter, az.Config.DisableAPICallCache)

	az.regionZonesMap = map[string][]string{az.Location: {"1", "2", "3"}}

	{
		kubeClient := fake.NewSimpleClientset() // FIXME: inject kubeClient
		informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
		az.serviceLister = informerFactory.Core().V1().Services().Lister()
		informerFactory.Start(wait.NeverStop)
		informerFactory.WaitForCacheSync(wait.NeverStop)
	}

	return az
}

// GetTestCloudWithExtendedLocation returns a fake azure cloud for unit tests in Azure related CSI drivers with extended location.
func GetTestCloudWithExtendedLocation(ctrl *gomock.Controller) (az *Cloud) {
	az = GetTestCloud(ctrl)
	az.Config.ExtendedLocationName = "microsoftlosangeles1"
	az.Config.ExtendedLocationType = "EdgeZone"
	return az
}
