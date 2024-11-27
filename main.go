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
		_, err = network.NewSubnet(ctx, configName("snet-bastion"), &network.SubnetArgs{
			AddressPrefix:                     pulumi.String(bastionSubnet),
			ResourceGroupName:                 resourceGroup.Name,
			SubnetName:                        pulumi.String(configName("snet-bastion")),
			VirtualNetworkName:                vnet.Name,
			PrivateEndpointNetworkPolicies:    pulumi.String("Enabled"),
			PrivateLinkServiceNetworkPolicies: pulumi.String("Enabled"),
		})
		if err != nil {
			return err
		}

		// Create a Subnet (VM)
		_, err = network.NewSubnet(ctx, configName("snet-vm"), &network.SubnetArgs{
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

		return nil
	})
}
