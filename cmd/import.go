package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	keyfactor_command_client_api "github.com/Keyfactor/keyfactor-go-client-sdk"
	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

// TODO print actual error messages
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Keyfactor instance import utilities.",
	Long:  `A collection of APIs and utilities for importing Keyfactor instance data.`,
	Run: func(cmd *cobra.Command, args []string) {
		exportPath := cmd.Flag("file").Value.String()
		jsonFile, oErr := os.Open(exportPath)
		if oErr != nil {
			fmt.Printf("Error opening exported file: %s\n", oErr)
			log.Fatalf("Error: %s", oErr)
		}
		defer jsonFile.Close()
		var out outJson
		bJson, _ := io.ReadAll(jsonFile)
		jErr := json.Unmarshal(bJson, &out)
		if jErr != nil {
			fmt.Printf("Error reading exported file: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		kfClient := initGenClient()
		if len(out.Collections) != 0 {
			importCollections(out.Collections, kfClient)
		}
		if len(out.MetadataFields) != 0 {
			importMetadataFields(out.MetadataFields, kfClient)
		}
		if len(out.ExpirationAlerts) != 0 {
			importExpirationAlerts(out.ExpirationAlerts, kfClient)
		}
		if len(out.IssuedCertAlerts) != 0 {
			importIssuedCertAlerts(out.IssuedCertAlerts, kfClient)
		}
		if len(out.DeniedCertAlerts) != 0 {
			importDeniedCertAlerts(out.DeniedCertAlerts, kfClient)
		}
		if len(out.PendingCertAlerts) != 0 {
			importPendingCertAlerts(out.PendingCertAlerts, kfClient)
		}
		if len(out.Networks) != 0 {
			importNetworks(out.Networks, kfClient)
		}
		if len(out.WorkflowDefinitions) != 0 {
			importWorkflowDefinitions(out.WorkflowDefinitions, kfClient)
		}
		if len(out.BuiltInReports) != 0 {
			importBuiltInReports(out.BuiltInReports, kfClient)
		}
		if len(out.CustomReports) != 0 {
			importCustomReports(out.CustomReports, kfClient)
		}
		if len(out.SecurityRoles) != 0 {
			kfClient, _ := initClient()
			importSecurityRoles(out.SecurityRoles, kfClient)
		}
	},
}

func importCollections(collections []keyfactor_command_client_api.ModelsCertificateQuery, kfClient *keyfactor_command_client_api.APIClient) {
	//TODO error message if collection name already exists
	for _, collection := range collections {
		cJson, _ := json.Marshal(collection)
		var collectionReq keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest
		jErr := json.Unmarshal(cJson, &collectionReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		collectionReq.Query = collection.Content
		_, httpResp, reqErr := kfClient.CertificateCollectionApi.CertificateCollectionCreateCollection(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).
			Request(collectionReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create collection: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(collectionReq.Name)
			fmt.Println("Added", string(name), "to collections")
		}
	}
}

func importMetadataFields(metadataFields []keyfactor_command_client_api.ModelsMetadataFieldTypeModel, kfClient *keyfactor_command_client_api.APIClient) {
	for _, metadata := range metadataFields {
		mJson, _ := json.Marshal(metadata)
		var metadataReq keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest
		jErr := json.Unmarshal(mJson, &metadataReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.MetadataFieldApi.MetadataFieldCreateMetadataField(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).MetadataFieldType(metadataReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create metadata field type: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(metadataReq.Name)
			fmt.Println("Added", string(name), "to metadata field types.")
		}
	}
}

func importExpirationAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertDefinitionResponse, kfClient *keyfactor_command_client_api.APIClient) {
	//TODO do I need to check that it corresponds to the correct collection? Ask JD
	for _, alert := range alerts {
		aJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest
		jErr := json.Unmarshal(aJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		alertReq.CertificateQueryId = alert.CertificateQuery.Id
		_, httpResp, reqErr := kfClient.ExpirationAlertApi.ExpirationAlertAddExpirationAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alertReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create expiration alert: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(alertReq.DisplayName)
			fmt.Println("Added", string(name), "to expiration alerts.")
		}
	}
}

func importIssuedCertAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertDefinitionResponse, kfClient *keyfactor_command_client_api.APIClient) {
	for _, alert := range alerts {
		aJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest
		jErr := json.Unmarshal(aJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.IssuedAlertApi.IssuedAlertAddIssuedAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alertReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create issued cert alert: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(alertReq.DisplayName)
			fmt.Println("Added", string(name), "to issued cert alerts.")
		}
	}
}

func importDeniedCertAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertDefinitionResponse, kfClient *keyfactor_command_client_api.APIClient) {
	for _, alert := range alerts {
		aJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest
		jErr := json.Unmarshal(aJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.DeniedAlertApi.DeniedAlertAddDeniedAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alertReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create denied cert alert: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(alertReq.DisplayName)
			fmt.Println("Added", string(name), "to denied cert alerts.")
		}
	}
}

func importPendingCertAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertDefinitionResponse, kfClient *keyfactor_command_client_api.APIClient) {
	for _, alert := range alerts {
		aJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest
		jErr := json.Unmarshal(aJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.PendingAlertApi.PendingAlertAddPendingAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alertReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create pending cert alert: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(alertReq.DisplayName)
			fmt.Println("Added", string(name), "to pending cert alerts.")
		}
	}
}

func importNetworks(networks []keyfactor_command_client_api.KeyfactorApiModelsSslNetworkQueryResponse, kfClient *keyfactor_command_client_api.APIClient) {
	for _, network := range networks {
		nJson, _ := json.Marshal(network)
		var networkReq keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest
		jErr := json.Unmarshal(nJson, &networkReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.SslApi.SslCreateNetwork(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Network(networkReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create SSL network: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(networkReq.Name)
			fmt.Println("Added", string(name), "to SSL networks.")
		}
	}
}

func importWorkflowDefinitions(workflowDefs []keyfactor_command_client_api.KeyfactorApiModelsWorkflowsDefinitionQueryResponse, kfClient *keyfactor_command_client_api.APIClient) {
	for _, workflowDef := range workflowDefs {
		wJson, _ := json.Marshal(workflowDef)
		var workflowDefReq keyfactor_command_client_api.KeyfactorApiModelsWorkflowsDefinitionCreateRequest
		jErr := json.Unmarshal(wJson, &workflowDefReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.WorkflowDefinitionApi.WorkflowDefinitionCreateNewDefinition(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(workflowDefReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create a new workflow definition: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(workflowDefReq.DisplayName)
			fmt.Println("Added", string(name), "to workflow definitions.")
		}
	}
}

// identified built-in reports where UsesCollections is false
func checkBuiltInReportDiffs(exportedReport keyfactor_command_client_api.ModelsReport) *int32 {
	bReports, _ := getReports()
	//check if built in report was modified from default in exported instance; if modified, update built-in report in new instance
	for _, bReport := range bReports {
		iJson, _ := json.Marshal(bReport.DisplayName)
		eJson, _ := json.Marshal(exportedReport.DisplayName)
		var usesCollectionBool = "false"
		if string(iJson) == string(eJson) {
			usesCollectionJson, _ := json.Marshal(exportedReport.UsesCollection)
			if string(usesCollectionJson) == usesCollectionBool {
				return bReport.Id
			} else {
				return nil
			}
		}
	}
	return nil
}

// only imports built in reports where UsesCollections is false
func importBuiltInReports(reports []keyfactor_command_client_api.ModelsReport, kfClient *keyfactor_command_client_api.APIClient) {
	//TODO PUT /Reports, only for reports that don't use a collection, ask JD
	for _, report := range reports {
		newReportId := checkBuiltInReportDiffs(report)
		if newReportId != nil {
			rJson, _ := json.Marshal(report)
			var reportReq keyfactor_command_client_api.ModelsReportRequestModel
			jErr := json.Unmarshal(rJson, &reportReq)
			if jErr != nil {
				fmt.Printf("Error: %s\n", jErr)
				log.Fatalf("Error: %s", jErr)
			}
			_, httpResp, reqErr := kfClient.ReportsApi.ReportsUpdateReport(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(reportReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
			if reqErr != nil {
				fmt.Printf("Error, unable to update built-in report: %s\n", httpResp.Body)
				log.Fatalf("Error: %s", httpResp.Body)
			} else {
				name, _ := json.Marshal(report.DisplayName)
				fmt.Println("Updated", string(name), "in built-in reports.")
			}
		}
	}
}

func importCustomReports(reports []keyfactor_command_client_api.ModelsCustomReport, kfClient *keyfactor_command_client_api.APIClient) {
	for _, report := range reports {
		rJson, _ := json.Marshal(report)
		var reportReq keyfactor_command_client_api.ModelsCustomReportCreationRequest
		jErr := json.Unmarshal(rJson, &reportReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, httpResp, reqErr := kfClient.ReportsApi.ReportsCreateCustomReport(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(reportReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		if reqErr != nil {
			fmt.Printf("Error, unable to create custom report: %s\n", httpResp.Body)
			log.Fatalf("Error: %s", httpResp.Body)
		} else {
			name, _ := json.Marshal(reportReq.DisplayName)
			fmt.Println("Added", string(name), "to custom reports.")
		}
	}
}

func importSecurityRoles(roles api.GetSecurityRolesResponse, kfClient *api.Client) {
	for _, role := range roles {
		rJson, _ := json.Marshal(role)
		var roleReq api.CreateSecurityRoleArg
		jErr := json.Unmarshal(rJson, &roleReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		_, reqErr := kfClient.CreateSecurityRole(&roleReq)
		if reqErr != nil {
			fmt.Printf("Error, unable to create security role: %s\n", reqErr)
			log.Fatalf("Error: %s", reqErr)
		} else {
			name, _ := json.Marshal(roleReq.Name)
			fmt.Println("Added", string(name), "to security roles.")
		}
	}
}

func init() {
	var exportPath string

	RootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(&exportPath, "file", "f", "", "export JSON to a specified filepath")
	importCmd.MarkFlagRequired("file")
}
