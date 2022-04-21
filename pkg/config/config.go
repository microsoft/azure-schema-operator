package config

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

import "github.com/spf13/viper"

const (
	// AzureUseMSIKey configuration key if we should use MSI
	AzureUseMSIKey = "azure_use_msi"
	// AzureClientIDKey configuration key holding the client ID
	AzureClientIDKey = "azure_client_id"
	// AzureClientSecretKey key holding the client secret
	AzureClientSecretKey = "azure_client_secret"
	// AzureTenantIDKey key holding the Azure tenant ID
	AzureTenantIDKey = "azure_tenant_id"
	// DeltaCMDKey path to the delta-kusto binary
	DeltaCMDKey = "schemaop_delta_cmd"
	// SQLPackageCMDKey path to the sqlpackage binary
	SQLPackageCMDKey = "schemaop_sqlpackage_cmd"
	// SQLPackageUser user to access SQL Servers
	SQLPackageUser = "schemaop_sqlpackage_user"
	// SQLPackagePass password to authenticate against SQL servers
	SQLPackagePass = "schemaop_sqlpackage_pass"
	// ParallelWorkers Number of parallel worker groups
	ParallelWorkers = "schemaop_parallel_workers"
)

func init() {
	viper.AutomaticEnv()
}
