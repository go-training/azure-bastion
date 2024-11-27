package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/storage/v2"
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
		storageAccount := conf.Require("storageAccount")
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

		// Create an Azure resource (Storage Account)
		account, err := storage.NewStorageAccount(ctx, configName("sa"), &storage.StorageAccountArgs{
			ResourceGroupName: resourceGroup.Name,
			AccountName:       pulumi.String(storageAccount),
			Sku: &storage.SkuArgs{
				Name: pulumi.String("Standard_LRS"),
			},
			Kind: pulumi.String("StorageV2"),
		})
		if err != nil {
			return err
		}

		// Export the primary key of the Storage Account
		ctx.Export("primaryStorageKey", pulumi.All(resourceGroup.Name, account.Name).ApplyT(
			func(args []interface{}) (string, error) {
				resourceGroupName := args[0].(string)
				accountName := args[1].(string)
				accountKeys, err := storage.ListStorageAccountKeys(ctx, &storage.ListStorageAccountKeysArgs{
					ResourceGroupName: resourceGroupName,
					AccountName:       accountName,
				})
				if err != nil {
					return "", err
				}

				return accountKeys.Keys[0].Value, nil
			},
		))

		return nil
	})
}
