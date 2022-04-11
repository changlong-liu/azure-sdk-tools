//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	err                        error
	ctx                        context.Context
	cred                       azcore.TokenCredential
	appName                    = "app01"
	ascDomainName              = ".azuremicroservices.io"
	dnsCname                   = "asc"
	insightsInstrumentationKey string
	serviceName                = "test-scenario-instance"
	blobUrl                    = getEnv("BLOB_URL", "")
	customDomainName           = getEnv("CUSTOM_DOMAIN_NAME", "")
	dnsResourceGroup           = getEnv("DNS_RESOURCE_GROUP", "")
	dnsSubscriptionId          = getEnv("DNS_SUBSCRIPTION_ID", "")
	location                   = getEnv("LOCATION", "westus")
	mysqlKey                   = getEnv("MYSQL_KEY", "")
	resourceGroupName          = getEnv("RESOURCE_GROUP_NAME", "scenarioTestTempGroup")
	subscriptionId             = getEnv("AZURE_SUBSCRIPTION_ID", "")
	userAssignedIdentity       = getEnv("USER_ASSIGNED_IDENTITY", "")
)

func main() {
	ctx = context.Background()
	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}
	createResourceGroup()
	prepare()
	springSample()
	cleanup()
	deleteResourceGroup()
}

func prepare() {
	// From step Generate_Unique_ServiceName
	template := map[string]interface{}{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"outputs": map[string]interface{}{
			"serviceName": map[string]interface{}{
				"type":  "string",
				"value": "[substring(variables('serviceNameLong'), 0, 12)]",
			},
		},
		"parameters": map[string]interface{}{
			"serviceNamePrefix": map[string]interface{}{
				"type":         "string",
				"defaultValue": "asc-",
			},
		},
		"resources": []interface{}{},
		"variables": map[string]interface{}{
			"serviceNameLong": "[concat(parameters('serviceNamePrefix'), uniqueString(resourceGroup().id))]",
		},
	}
	params := map[string]interface{}{}
	deployment := armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template:   template,
			Parameters: params,
			Mode:       to.Ptr(armresources.DeploymentModeIncremental),
		},
	}
	deploymentExtend := createDeployment("Generate_Unique_ServiceName", &deployment)
	serviceName = deploymentExtend.Properties.Outputs.(map[string]interface{})["serviceName"].(map[string]interface{})["value"].(string)

	// From step Create_Application_Insight_Instance
	template = map[string]interface{}{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"outputs": map[string]interface{}{
			"insightsInstrumentationKey": map[string]interface{}{
				"type":  "string",
				"value": "[reference(resourceId('Microsoft.Insights/components', parameters('name')), '2014-04-01').InstrumentationKey]",
			},
		},
		"parameters": map[string]interface{}{
			"name": map[string]interface{}{
				"type":         "string",
				"defaultValue": "asc-api-ai-instance",
				"metadata": map[string]interface{}{
					"description": "Name of Application Insights resource.",
				},
			},
		},
		"resources": []interface{}{
			map[string]interface{}{
				"name":       "[parameters('name')]",
				"type":       "microsoft.insights/components",
				"apiVersion": "2014-04-01",
				"location":   "eastus",
				"properties": map[string]interface{}{
					"ApplicationId":    "[parameters('name')]",
					"Application_Type": "web",
					"Flow_Type":        "Redfield",
					"Request_Source":   "CustomDeployment",
				},
				"tags": map[string]interface{}{},
			},
		},
	}
	params = map[string]interface{}{}
	deployment = armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template:   template,
			Parameters: params,
			Mode:       to.Ptr(armresources.DeploymentModeIncremental),
		},
	}
	deploymentExtend = createDeployment("Create_Application_Insight_Instance", &deployment)
	insightsInstrumentationKey = deploymentExtend.Properties.Outputs.(map[string]interface{})["insightsInstrumentationKey"].(map[string]interface{})["value"].(string)

	// From step Add_Dns_Cname_Record
	template = map[string]interface{}{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]interface{}{
			"userAssignedIdentity": map[string]interface{}{
				"type":         "string",
				"defaultValue": "$(userAssignedIdentity)",
			},
			"utcValue": map[string]interface{}{
				"type":         "string",
				"defaultValue": "[utcNow()]",
			},
		},
		"resources": []interface{}{
			map[string]interface{}{
				"name":       "Add_Dns_Cname_Record",
				"type":       "Microsoft.Resources/deploymentScripts",
				"apiVersion": "2020-10-01",
				"identity": map[string]interface{}{
					"type": "UserAssigned",
					"userAssignedIdentities": map[string]interface{}{
						"[parameters('userAssignedIdentity')]": map[string]interface{}{},
					},
				},
				"kind":     "AzurePowerShell",
				"location": "[resourceGroup().location]",
				"properties": map[string]interface{}{
					"azPowerShellVersion": "6.2",
					"cleanupPreference":   "OnSuccess",
					"environmentVariables": []interface{}{
						map[string]interface{}{
							"name":  "resourceGroupName",
							"value": dnsResourceGroup,
						},
						map[string]interface{}{
							"name":  "dnsZoneName",
							"value": customDomainName,
						},
						map[string]interface{}{
							"name":  "dnsCname",
							"value": "asc",
						},
						map[string]interface{}{
							"name":  "dnsCnameAlias",
							"value": serviceName + ".azuremicroservices.io",
						},
					},
					"forceUpdateTag":    "[parameters('utcValue')]",
					"retentionInterval": "P1D",
					"scriptContent":     "# Copyright (c) 2021 Microsoft Corporation\n# \n# This software is released under the MIT License.\n# https://opensource.org/licenses/MIT\n$resourceGroupName = ${Env:resourceGroupName}\n$dnsCname = ${Env:dnsCname}\n$dnsZoneName = ${Env:dnsZoneName}\n$dnsCnameAlias = ${Env:dnsCnameAlias}\nConnect-AzAccount -Identity\nNew-AzDnsRecordSet -Name $dnsCname -RecordType CNAME -ZoneName $dnsZoneName -ResourceGroupName $resourceGroupName -Ttl 3600 -DnsRecords (New-AzDnsRecordConfig -Cname $dnsCnameAlias) -Overwrite\n$RecordSet = Get-AzDnsRecordSet -Name $dnsCname -RecordType CNAME -ResourceGroupName $resourceGroupName -ZoneName $dnsZoneName\n$RecordSet",
					"timeout":           "PT1H",
				},
			},
		},
	}
	params = map[string]interface{}{
		"userAssignedIdentity": map[string]interface{}{"value": userAssignedIdentity},
	}
	deployment = armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template:   template,
			Parameters: params,
			Mode:       to.Ptr(armresources.DeploymentModeIncremental),
		},
	}
	_ = createDeployment("Add_Dns_Cname_Record", &deployment)
}

func springSample() {
	var relativePath string
	var uploadUrl string
	// From step Services_CheckNameAvailability
	servicesClient, err := test.NewServicesClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	_, err = servicesClient.CheckNameAvailability(ctx,
		location,
		test.NameAvailabilityParameters{
			Name: to.Ptr(serviceName),
			Type: to.Ptr("Microsoft.AppPlatform/Spring"),
		},
		nil)
	if err != nil {
		panic(err)
	}

	// From step Services_CreateOrUpdate
	servicesClientCreateOrUpdateResponse, err := servicesClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		test.ServiceResource{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"key1": to.Ptr("value1"),
			},
			Properties: &test.ClusterResourceProperties{},
			SKU: &test.SKU{
				Name: to.Ptr("S0"),
				Tier: to.Ptr("Standard"),
			},
		},
		&test.ServicesClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = servicesClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Services_Get
	_, err = servicesClient.Get(ctx,
		resourceGroupName,
		serviceName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Services_Update
	servicesClientUpdateResponse, err := servicesClient.BeginUpdate(ctx,
		resourceGroupName,
		serviceName,
		test.ServiceResource{
			Tags: map[string]*string{
				"created-by": to.Ptr("api-test"),
				"hello":      to.Ptr("world"),
			},
			SKU: &test.SKU{
				Name: to.Ptr("S0"),
				Tier: to.Ptr("Standard"),
			},
		},
		&test.ServicesClientBeginUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = servicesClientUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Services_DisableTestEndpoint
	_, err = servicesClient.DisableTestEndpoint(ctx,
		resourceGroupName,
		serviceName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Services_EnableTestEndpoint
	_, err = servicesClient.EnableTestEndpoint(ctx,
		resourceGroupName,
		serviceName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Services_RegenerateTestKey
	_, err = servicesClient.RegenerateTestKey(ctx,
		resourceGroupName,
		serviceName,
		test.RegenerateTestKeyRequestPayload{
			KeyType: to.Ptr(test.TestKeyTypePrimary),
		},
		nil)
	if err != nil {
		panic(err)
	}

	// From step Services_ListTestKeys
	_, err = servicesClient.ListTestKeys(ctx,
		resourceGroupName,
		serviceName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Certificates_CreateOrUpdate
	certificatesClient, err := test.NewCertificatesClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	certificateName := "asc-certificate"
	certificatesClientCreateOrUpdateResponse, err := certificatesClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		certificateName,
		test.CertificateResource{
			Properties: &test.CertificateProperties{
				KeyVaultCertName: to.Ptr("pfx-cert"),
				VaultURI:         to.Ptr("https://integration-test-prod.vault.azure.net/"),
			},
		},
		&test.CertificatesClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = certificatesClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Certificates_Get
	certificateName = "asc-certificate"
	_, err = certificatesClient.Get(ctx,
		resourceGroupName,
		serviceName,
		certificateName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Certificates_List
	certificatesClientListPager := certificatesClient.List(resourceGroupName,
		serviceName,
		nil)
	for certificatesClientListPager.More() {
	}

	// From step ConfigServers_Validate
	configServersClient, err := test.NewConfigServersClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	configServersClientValidateResponse, err := configServersClient.BeginValidate(ctx,
		resourceGroupName,
		serviceName,
		test.ConfigServerSettings{
			GitProperty: &test.ConfigServerGitProperty{
				Label: to.Ptr("master"),
				SearchPaths: []*string{
					to.Ptr("/")},
				URI: to.Ptr("https://github.com/VSChina/asc-config-server-test-public.git"),
			},
		},
		&test.ConfigServersClientBeginValidateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = configServersClientValidateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step ConfigServers_UpdatePut
	configServersClientUpdatePutResponse, err := configServersClient.BeginUpdatePut(ctx,
		resourceGroupName,
		serviceName,
		test.ConfigServerResource{
			Properties: &test.ConfigServerProperties{
				ConfigServer: &test.ConfigServerSettings{
					GitProperty: &test.ConfigServerGitProperty{
						Label: to.Ptr("master"),
						SearchPaths: []*string{
							to.Ptr("/")},
						URI: to.Ptr("https://github.com/VSChina/asc-config-server-test-public.git"),
					},
				},
			},
		},
		&test.ConfigServersClientBeginUpdatePutOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = configServersClientUpdatePutResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step ConfigServers_UpdatePatch
	configServersClientUpdatePatchResponse, err := configServersClient.BeginUpdatePatch(ctx,
		resourceGroupName,
		serviceName,
		test.ConfigServerResource{
			Properties: &test.ConfigServerProperties{
				ConfigServer: &test.ConfigServerSettings{
					GitProperty: &test.ConfigServerGitProperty{
						URI: to.Ptr("https://github.com/azure-samples/spring-petclinic-microservices-config"),
					},
				},
			},
		},
		&test.ConfigServersClientBeginUpdatePatchOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = configServersClientUpdatePatchResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step ConfigServers_Get
	_, err = configServersClient.Get(ctx,
		resourceGroupName,
		serviceName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step MonitoringSettings_UpdatePut
	monitoringSettingsClient, err := test.NewMonitoringSettingsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	monitoringSettingsClientUpdatePutResponse, err := monitoringSettingsClient.BeginUpdatePut(ctx,
		resourceGroupName,
		serviceName,
		test.MonitoringSettingResource{
			Properties: &test.MonitoringSettingProperties{
				AppInsightsInstrumentationKey: to.Ptr(insightsInstrumentationKey),
				AppInsightsSamplingRate:       to.Ptr[float64](50),
				TraceEnabled:                  to.Ptr(true),
			},
		},
		&test.MonitoringSettingsClientBeginUpdatePutOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = monitoringSettingsClientUpdatePutResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step MonitoringSettings_Get
	_, err = monitoringSettingsClient.Get(ctx,
		resourceGroupName,
		serviceName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step MonitoringSettings_UpdatePatch
	monitoringSettingsClientUpdatePatchResponse, err := monitoringSettingsClient.BeginUpdatePatch(ctx,
		resourceGroupName,
		serviceName,
		test.MonitoringSettingResource{
			Properties: &test.MonitoringSettingProperties{
				AppInsightsSamplingRate: to.Ptr[float64](100),
			},
		},
		&test.MonitoringSettingsClientBeginUpdatePatchOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = monitoringSettingsClientUpdatePatchResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_Create
	appsClient, err := test.NewAppsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	appsClientCreateOrUpdateResponse, err := appsClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		test.AppResource{
			Identity: &test.ManagedIdentityProperties{
				Type:        to.Ptr(test.ManagedIdentityTypeSystemAssigned),
				PrincipalID: to.Ptr("principalid"),
				TenantID:    to.Ptr("tenantid"),
			},
			Location: to.Ptr(location),
			Properties: &test.AppResourceProperties{
				ActiveDeploymentName: to.Ptr("mydeployment1"),
				EnableEndToEndTLS:    to.Ptr(false),
				Fqdn:                 to.Ptr(appName + ".mydomain.com"),
				HTTPSOnly:            to.Ptr(false),
				Public:               to.Ptr(false),
			},
		},
		&test.AppsClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = appsClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_Get
	_, err = appsClient.Get(ctx,
		resourceGroupName,
		serviceName,
		appName,
		&test.AppsClientGetOptions{SyncStatus: nil})
	if err != nil {
		panic(err)
	}

	// From step Deployments_CreateOrUpdate_Default
	deploymentsClient, err := test.NewDeploymentsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	deploymentName := "default"
	deploymentsClientCreateOrUpdateResponse, err := deploymentsClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		test.DeploymentResource{
			Properties: &test.DeploymentResourceProperties{
				DeploymentSettings: &test.DeploymentSettings{
					CPU: to.Ptr[int32](1),
					EnvironmentVariables: map[string]*string{
						"env": to.Ptr("test"),
					},
					JvmOptions:     to.Ptr("-Xms1G -Xmx3G"),
					MemoryInGB:     to.Ptr[int32](3),
					RuntimeVersion: to.Ptr(test.RuntimeVersionJava8),
				},
				Source: &test.UserSourceInfo{
					Type:             to.Ptr(test.UserSourceTypeJar),
					ArtifactSelector: to.Ptr("sub-module-1"),
					RelativePath:     to.Ptr("<default>"),
					Version:          to.Ptr("1.0"),
				},
			},
			SKU: &test.SKU{
				Name:     to.Ptr("S0"),
				Capacity: to.Ptr[int32](1),
				Tier:     to.Ptr("Standard"),
			},
		},
		&test.DeploymentsClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = deploymentsClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Deployments_Get
	deploymentName = "default"
	_, err = deploymentsClient.Get(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Apps_Update_ActiveDeployment
	appsClientUpdateResponse, err := appsClient.BeginUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		test.AppResource{
			Identity: &test.ManagedIdentityProperties{
				Type:        to.Ptr(test.ManagedIdentityTypeSystemAssigned),
				PrincipalID: to.Ptr("principalid"),
				TenantID:    to.Ptr("tenantid"),
			},
			Properties: &test.AppResourceProperties{
				ActiveDeploymentName: to.Ptr("default"),
			},
		},
		&test.AppsClientBeginUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = appsClientUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_Update_Disk
	appsClientUpdateResponse, err = appsClient.BeginUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		test.AppResource{
			Identity: &test.ManagedIdentityProperties{
				Type:        to.Ptr(test.ManagedIdentityTypeSystemAssigned),
				PrincipalID: to.Ptr("principalid"),
				TenantID:    to.Ptr("tenantid"),
			},
			Properties: &test.AppResourceProperties{
				PersistentDisk: &test.PersistentDisk{
					MountPath: to.Ptr("/data"),
					SizeInGB:  to.Ptr[int32](10),
				},
				TemporaryDisk: &test.TemporaryDisk{
					MountPath: to.Ptr("/tmpdisk"),
					SizeInGB:  to.Ptr[int32](3),
				},
			},
		},
		&test.AppsClientBeginUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = appsClientUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_List
	appsClientListPager := appsClient.List(resourceGroupName,
		serviceName,
		nil)
	for appsClientListPager.More() {
	}

	// From step Bindings_Create
	bindingsClient, err := test.NewBindingsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	bindingName := "mysql-binding"
	bindingsClientCreateOrUpdateResponse, err := bindingsClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		bindingName,
		test.BindingResource{
			Properties: &test.BindingResourceProperties{
				BindingParameters: map[string]interface{}{
					"databaseName": "mysqldb",
					"username":     "test",
				},
				Key:        to.Ptr(mysqlKey),
				ResourceID: to.Ptr("/subscriptions/b46590cb-a111-4b84-935f-c305aaf1f424/resourceGroups/mary-west/providers/Microsoft.DBforMySQL/servers/fake-sql"),
			},
		},
		&test.BindingsClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = bindingsClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Bindings_Update
	bindingName = "mysql-binding"
	bindingsClientUpdateResponse, err := bindingsClient.BeginUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		bindingName,
		test.BindingResource{
			Properties: &test.BindingResourceProperties{
				BindingParameters: map[string]interface{}{
					"databaseName": "mysqldb2",
					"username":     "test2",
				},
				Key:        to.Ptr(mysqlKey),
				ResourceID: to.Ptr("/subscriptions/" + subscriptionId + "/resourceGroups/" + resourceGroupName + "/providers/Microsoft.DocumentDB/databaseAccounts/my-cosmosdb-1"),
			},
		},
		&test.BindingsClientBeginUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = bindingsClientUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Bindings_Get
	bindingName = "mysql-binding"
	_, err = bindingsClient.Get(ctx,
		resourceGroupName,
		serviceName,
		appName,
		bindingName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Bindings_List
	bindingsClientListPager := bindingsClient.List(resourceGroupName,
		serviceName,
		appName,
		nil)
	for bindingsClientListPager.More() {
	}

	// From step Bindings_Delete
	bindingName = "mysql-binding"
	bindingsClientDeleteResponse, err := bindingsClient.BeginDelete(ctx,
		resourceGroupName,
		serviceName,
		appName,
		bindingName,
		&test.BindingsClientBeginDeleteOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = bindingsClientDeleteResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_ValidateDomain
	_, err = appsClient.ValidateDomain(ctx,
		resourceGroupName,
		serviceName,
		appName,
		test.CustomDomainValidatePayload{
			Name: to.Ptr(customDomainName),
		},
		nil)
	if err != nil {
		panic(err)
	}

	// From step CustomDomains_CreateOrUpdate
	customDomainsClient, err := test.NewCustomDomainsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	domainName := dnsCname + "." + customDomainName
	customDomainsClientCreateOrUpdateResponse, err := customDomainsClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		domainName,
		test.CustomDomainResource{
			Properties: &test.CustomDomainProperties{
				CertName: to.Ptr("asc-certificate"),
			},
		},
		&test.CustomDomainsClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = customDomainsClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step CustomDomains_Update
	domainName = dnsCname + "." + customDomainName
	customDomainsClientUpdateResponse, err := customDomainsClient.BeginUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		domainName,
		test.CustomDomainResource{
			Properties: &test.CustomDomainProperties{
				CertName: to.Ptr("asc-certificate"),
			},
		},
		&test.CustomDomainsClientBeginUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = customDomainsClientUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step CustomDomains_Get
	domainName = dnsCname + "." + customDomainName
	_, err = customDomainsClient.Get(ctx,
		resourceGroupName,
		serviceName,
		appName,
		domainName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step CustomDomains_List
	customDomainsClientListPager := customDomainsClient.List(resourceGroupName,
		serviceName,
		appName,
		nil)
	for customDomainsClientListPager.More() {
	}

	// From step Apps_GetResourceUploadUrl
	appsClientGetResourceUploadURLResponse, err := appsClient.GetResourceUploadURL(ctx,
		resourceGroupName,
		serviceName,
		appName,
		nil)
	if err != nil {
		panic(err)
	}
	relativePath = *appsClientGetResourceUploadURLResponse.RelativePath
	uploadUrl = *appsClientGetResourceUploadURLResponse.UploadURL

	// From step Upload_File
	template := map[string]interface{}{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]interface{}{
			"userAssignedIdentity": map[string]interface{}{
				"type":         "string",
				"defaultValue": "$(userAssignedIdentity)",
			},
			"utcValue": map[string]interface{}{
				"type":         "string",
				"defaultValue": "[utcNow()]",
			},
		},
		"resources": []interface{}{
			map[string]interface{}{
				"name":       "Upload_File",
				"type":       "Microsoft.Resources/deploymentScripts",
				"apiVersion": "2020-10-01",
				"identity": map[string]interface{}{
					"type": "UserAssigned",
					"userAssignedIdentities": map[string]interface{}{
						"[parameters('userAssignedIdentity')]": map[string]interface{}{},
					},
				},
				"kind":     "AzurePowerShell",
				"location": "[resourceGroup().location]",
				"properties": map[string]interface{}{
					"azPowerShellVersion": "6.2",
					"cleanupPreference":   "OnSuccess",
					"environmentVariables": []interface{}{
						map[string]interface{}{
							"name":        "uploadUrl",
							"secureValue": uploadUrl,
						},
						map[string]interface{}{
							"name":        "blobUrl",
							"secureValue": blobUrl,
						},
					},
					"forceUpdateTag":    "[parameters('utcValue')]",
					"retentionInterval": "P1D",
					"scriptContent": "$uploadUri = ${Env:uploadUrl}\n$blobUri = ${Env:blobUrl}\n$localFilePath = '/tmp/temp.file'\nfunction DownloadJarFromBlob([string]$blobUri, [string]$localOutputFilePath) {\n	$BlobFile = [Microsoft.WindowsAzure.Storage.Blob.CloudBlob]::new($blobUri)\n	$DownLoadTask = $BlobFile.DownloadToFileAsync($localOutputFilePath, 4)\n	$DownLoadTask\n}\n\nfunction UploadToFileShare([string]$uploadUri, [string]$localFilePath) {\n	$CloudFile = [Microsoft.WindowsAzure.Storage.File.CloudFile]::New($uploadUri)\n	$UploadTask = $CloudFile.UploadFromFileAsync($localFilePath)\n	$UploadTask\n}\n\nConnect-AzAccount -Identity\nDownloadJarFromBlob $blobUri $localFilePath\nUploadToFileShare $uploadUri $localFilePath",
					"timeout": "PT1H",
				},
			},
		},
	}
	params := map[string]interface{}{
		"userAssignedIdentity": map[string]interface{}{"value": userAssignedIdentity},
	}
	deployment := armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template:   template,
			Parameters: params,
			Mode:       to.Ptr(armresources.DeploymentModeIncremental),
		},
	}
	_ = createDeployment("Upload_File", &deployment)

	// From step Deployments_CreateOrUpdate
	deploymentName = "blue"
	deploymentsClientCreateOrUpdateResponse, err = deploymentsClient.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		test.DeploymentResource{
			Properties: &test.DeploymentResourceProperties{
				DeploymentSettings: &test.DeploymentSettings{
					CPU: to.Ptr[int32](1),
					EnvironmentVariables: map[string]*string{
						"env": to.Ptr("test"),
					},
					JvmOptions:     to.Ptr("-Xms1G -Xmx3G"),
					MemoryInGB:     to.Ptr[int32](3),
					RuntimeVersion: to.Ptr(test.RuntimeVersionJava8),
				},
				Source: &test.UserSourceInfo{
					Type:             to.Ptr(test.UserSourceTypeJar),
					ArtifactSelector: to.Ptr("sub-module-1"),
					RelativePath:     to.Ptr(relativePath),
					Version:          to.Ptr("1.0"),
				},
			},
			SKU: &test.SKU{
				Name:     to.Ptr("S0"),
				Capacity: to.Ptr[int32](2),
				Tier:     to.Ptr("Standard"),
			},
		},
		&test.DeploymentsClientBeginCreateOrUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = deploymentsClientCreateOrUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_Update
	appsClientUpdateResponse, err = appsClient.BeginUpdate(ctx,
		resourceGroupName,
		serviceName,
		appName,
		test.AppResource{
			Identity: &test.ManagedIdentityProperties{
				Type:        to.Ptr(test.ManagedIdentityTypeSystemAssigned),
				PrincipalID: to.Ptr("principalid"),
				TenantID:    to.Ptr("tenantid"),
			},
			Properties: &test.AppResourceProperties{
				ActiveDeploymentName: to.Ptr("blue"),
			},
		},
		&test.AppsClientBeginUpdateOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = appsClientUpdateResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Deployments_Restart
	deploymentName = "blue"
	deploymentsClientRestartResponse, err := deploymentsClient.BeginRestart(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		&test.DeploymentsClientBeginRestartOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = deploymentsClientRestartResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Deployments_Stop
	deploymentName = "blue"
	deploymentsClientStopResponse, err := deploymentsClient.BeginStop(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		&test.DeploymentsClientBeginStopOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = deploymentsClientStopResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Deployments_Start
	deploymentName = "blue"
	deploymentsClientStartResponse, err := deploymentsClient.BeginStart(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		&test.DeploymentsClientBeginStartOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = deploymentsClientStartResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Deployments_GetLogFileUrl
	deploymentName = "blue"
	_, err = deploymentsClient.GetLogFileURL(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		nil)
	if err != nil {
		panic(err)
	}

	// From step Deployments_List
	deploymentsClientListPager := deploymentsClient.List(resourceGroupName,
		serviceName,
		appName,
		&test.DeploymentsClientListOptions{Version: []string{}})
	for deploymentsClientListPager.More() {
	}

	// From step Deployments_ListForCluster
	deploymentsClientListForClusterPager := deploymentsClient.ListForCluster(resourceGroupName,
		serviceName,
		&test.DeploymentsClientListForClusterOptions{Version: []string{}})
	for deploymentsClientListForClusterPager.More() {
	}

	// From step Services_List
	servicesClientListPager := servicesClient.List(resourceGroupName,
		nil)
	for servicesClientListPager.More() {
	}

	// From step Services_ListBySubscription
	servicesClientListBySubscriptionPager := servicesClient.ListBySubscription(nil)
	for servicesClientListBySubscriptionPager.More() {
	}

	// From step Deployments_Delete
	deploymentName = "blue"
	deploymentsClientDeleteResponse, err := deploymentsClient.BeginDelete(ctx,
		resourceGroupName,
		serviceName,
		appName,
		deploymentName,
		&test.DeploymentsClientBeginDeleteOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = deploymentsClientDeleteResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step CustomDomains_Delete
	domainName = dnsCname + "." + customDomainName
	customDomainsClientDeleteResponse, err := customDomainsClient.BeginDelete(ctx,
		resourceGroupName,
		serviceName,
		appName,
		domainName,
		&test.CustomDomainsClientBeginDeleteOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = customDomainsClientDeleteResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Apps_Delete
	appName := "app01"
	appsClientDeleteResponse, err := appsClient.BeginDelete(ctx,
		resourceGroupName,
		serviceName,
		appName,
		&test.AppsClientBeginDeleteOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = appsClientDeleteResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Certificates_Delete
	certificateName = "asc-certificate"
	certificatesClientDeleteResponse, err := certificatesClient.BeginDelete(ctx,
		resourceGroupName,
		serviceName,
		certificateName,
		&test.CertificatesClientBeginDeleteOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = certificatesClientDeleteResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Services_Delete
	servicesClientDeleteResponse, err := servicesClient.BeginDelete(ctx,
		resourceGroupName,
		serviceName,
		&test.ServicesClientBeginDeleteOptions{ResumeToken: ""})
	if err != nil {
		panic(err)
	}
	_, err = servicesClientDeleteResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// From step Skus_List
	sKUsClient, err := test.NewSKUsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	sKUsClientListPager := sKUsClient.List(nil)
	for sKUsClientListPager.More() {
	}

	// From step Operations_List
	operationsClient, err := test.NewOperationsClient(cred, nil)
	if err != nil {
		panic(err)
	}
	operationsClientListPager := operationsClient.List(nil)
	for operationsClientListPager.More() {
	}
}

func cleanup() {
	// From step delete_cname_record
	template := map[string]interface{}{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": map[string]interface{}{
			"userAssignedIdentity": map[string]interface{}{
				"type":         "string",
				"defaultValue": "$(userAssignedIdentity)",
			},
			"utcValue": map[string]interface{}{
				"type":         "string",
				"defaultValue": "[utcNow()]",
			},
		},
		"resources": []interface{}{
			map[string]interface{}{
				"name":       "delete_cname_record",
				"type":       "Microsoft.Resources/deploymentScripts",
				"apiVersion": "2020-10-01",
				"identity": map[string]interface{}{
					"type": "UserAssigned",
					"userAssignedIdentities": map[string]interface{}{
						"[parameters('userAssignedIdentity')]": map[string]interface{}{},
					},
				},
				"kind":     "AzurePowerShell",
				"location": "[resourceGroup().location]",
				"properties": map[string]interface{}{
					"azPowerShellVersion": "6.2",
					"cleanupPreference":   "OnSuccess",
					"environmentVariables": []interface{}{
						map[string]interface{}{
							"name":  "resourceGroupName",
							"value": dnsResourceGroup,
						},
						map[string]interface{}{
							"name":  "dnsCname",
							"value": "asc",
						},
						map[string]interface{}{
							"name":  "dnsZoneName",
							"value": customDomainName,
						},
					},
					"forceUpdateTag":    "[parameters('utcValue')]",
					"retentionInterval": "P1D",
					"scriptContent":     "$resourceGroupName = ${Env:resourceGroupName}\n$dnsCNAME = ${Env:dnsCname}\n$dnsZoneName = ${Env:dnsZoneName}\n\nConnect-AzAccount -Identity\n\n$RecordSet = Get-AzDnsRecordSet -Name $dnsCname -RecordType CNAME -ResourceGroupName $resourceGroupName -ZoneName $dnsZoneName\n$Result = Remove-AzDnsRecordSet -RecordSet $RecordSet\n$Result",
					"timeout":           "PT1H",
				},
			},
		},
	}
	params := map[string]interface{}{
		"userAssignedIdentity": map[string]interface{}{"value": userAssignedIdentity},
	}
	deployment := armresources.Deployment{
		Properties: &armresources.DeploymentProperties{
			Template:   template,
			Parameters: params,
			Mode:       to.Ptr(armresources.DeploymentModeIncremental),
		},
	}
	_ = createDeployment("delete_cname_record", &deployment)
}

func createResourceGroup() error {
	rand.Seed(time.Now().UnixNano())
	resourceGroupName = fmt.Sprintf("go-sdk-sample-%d", rand.Intn(1000))
	rgClient, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	param := armresources.ResourceGroup{
		Location: to.Ptr(location),
	}
	_, err = rgClient.CreateOrUpdate(ctx, resourceGroupName, param, nil)
	if err != nil {
		panic(err)
	}
	return nil
}

func deleteResourceGroup() error {
	rgClient, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	pollerResponse, err := rgClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		panic(err)
	}
	_, err = pollerResponse.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func createDeployment(deploymentName string, deployment *armresources.Deployment) *armresources.DeploymentExtended {
	deployClient, err := armresources.NewDeploymentsClient(subscriptionId, cred, nil)
	if err != nil {
		panic(err)
	}
	poller, err := deployClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		deploymentName,
		*deployment,
		&armresources.DeploymentsClientBeginCreateOrUpdateOptions{},
	)
	if err != nil {
		panic(err)
	}
	res, err := poller.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		panic(err)
	}
	return &res.DeploymentExtended
}
