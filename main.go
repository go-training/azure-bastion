package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func ModuleName(rg, name string) string {
	return rg + "-" + name
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "azure")
		resourceGroupName := conf.Require("resourceGroupName")
		vnetIP := conf.Require("vnet")
		bastionSubnet := conf.Require("bastionSubnet")
		vmSubnet := conf.Require("vmSubnet")

		configName := func(module string) string {
			return ModuleName(resourceGroupName, module)
		}
		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(ctx, configName("rg"), &resources.ResourceGroupArgs{
			ResourceGroupName: pulumi.String(resourceGroupName),
		})
		if err != nil {
			return err
		}

		// Create a Virtual Network
		vnet, err := network.NewVirtualNetwork(ctx, configName("vnet"), &network.VirtualNetworkArgs{
			AddressSpace: &network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String(vnetIP),
				},
			},
			FlowTimeoutInMinutes: pulumi.Int(10),
			Location:             resourceGroup.Location,
			ResourceGroupName:    resourceGroup.Name,
			VirtualNetworkName:   pulumi.String(configName("vnet")),
		})
		if err != nil {
			return err
		}

		// Create a Subnet (bastion)
		bastionNetwork, err := network.NewSubnet(ctx, configName("snet-bastion"), &network.SubnetArgs{
			AddressPrefix:                     pulumi.String(bastionSubnet),
			ResourceGroupName:                 resourceGroup.Name,
			SubnetName:                        pulumi.String("AzureBastionSubnet"),
			VirtualNetworkName:                vnet.Name,
			PrivateEndpointNetworkPolicies:    pulumi.String("Enabled"),
			PrivateLinkServiceNetworkPolicies: pulumi.String("Enabled"),
		})
		if err != nil {
			return err
		}

		// Create a Subnet (VM)
		vmNetwork, err := network.NewSubnet(ctx, configName("snet-vm"), &network.SubnetArgs{
			AddressPrefix:                     pulumi.String(vmSubnet),
			ResourceGroupName:                 resourceGroup.Name,
			SubnetName:                        pulumi.String(configName("snet-vm")),
			VirtualNetworkName:                vnet.Name,
			PrivateEndpointNetworkPolicies:    pulumi.String("Enabled"),
			PrivateLinkServiceNetworkPolicies: pulumi.String("Enabled"),
		})
		if err != nil {
			return err
		}

		// create vm netowrk interface
		_, err = network.NewNetworkInterface(ctx, configName("nic-vm"), &network.NetworkInterfaceArgs{
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name: pulumi.String("internal"),
					Subnet: &network.SubnetTypeArgs{
						Id: vmNetwork.ID(),
					},
					PrivateIPAllocationMethod: pulumi.String("Dynamic"),
				},
			},
			Location:             resourceGroup.Location,
			NetworkInterfaceName: pulumi.String(configName("nic-vm")),
			ResourceGroupName:    resourceGroup.Name,
		})
		if err != nil {
			return err
		}

		// create public ip
		bastionIP, err := network.NewPublicIPAddress(ctx, configName("pip-bastion"), &network.PublicIPAddressArgs{
			IdleTimeoutInMinutes:     pulumi.Int(5),
			Location:                 resourceGroup.Location,
			PublicIpAddressName:      pulumi.String(configName("pip-bastion")),
			ResourceGroupName:        resourceGroup.Name,
			PublicIPAllocationMethod: pulumi.String(network.IPAllocationMethodStatic),
			PublicIPAddressVersion:   pulumi.String(network.IPVersionIPv4),
			Sku: &network.PublicIPAddressSkuArgs{
				Name: pulumi.String(network.PublicIPAddressSkuNameStandard),
				Tier: pulumi.String(network.PublicIPAddressSkuTierGlobal),
			},
		})
		if err != nil {
			return err
		}

		// create bastion host
		_, err = network.NewBastionHost(ctx, configName("bastion-host"), &network.BastionHostArgs{
			BastionHostName:     pulumi.String(configName("bastion-host")),
			EnableTunneling:     pulumi.Bool(true),
			EnableShareableLink: pulumi.Bool(false),
			EnableFileCopy:      pulumi.Bool(false),
			EnableIpConnect:     pulumi.Bool(false),
			EnableKerberos:      pulumi.Bool(false),
			IpConfigurations: network.BastionHostIPConfigurationArray{
				&network.BastionHostIPConfigurationArgs{
					Name: pulumi.String(configName("bastionz-host-config")),
					PublicIPAddress: &network.SubResourceArgs{
						Id: bastionIP.ID(),
					},
					Subnet: &network.SubResourceArgs{
						Id: bastionNetwork.ID(),
					},
				},
			},
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			Sku: &network.SkuArgs{
				Name: pulumi.String(network.PublicIPAddressSkuNameStandard),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
