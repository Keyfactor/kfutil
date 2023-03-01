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

// TODO import flags --all (export too)
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

func importCollections(collections []keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, collection := range collections {
		_, httpResp, reqErr := kfClient.CertificateCollectionApi.CertificateCollectionCreateCollection(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).
			Request(collection).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(collection.Name)
		if reqErr != nil {
			fmt.Printf("Error! Unable to create collection %s: %s\n", string(name), httpResp.Body)
		} else {
			name, _ := json.Marshal(collection.Name)
			fmt.Println("Added", string(name), "to collections")
		}
	}
}

func importMetadataFields(metadataFields []keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, metadata := range metadataFields {
		_, httpResp, reqErr := kfClient.MetadataFieldApi.MetadataFieldCreateMetadataField(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).MetadataFieldType(metadata).
			XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(metadata.Name)
		if reqErr != nil {
			fmt.Printf("Error! Unable to create metadata field type %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to metadata field types.")
		}
	}
}

func importIssuedCertAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.IssuedAlertApi.IssuedAlertAddIssuedAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("Error! Unable to create issued cert alert %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to issued cert alerts.")
		}
	}
}

func importDeniedCertAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.DeniedAlertApi.DeniedAlertAddDeniedAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("Error, unable to create denied cert alert %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to denied cert alerts.")
		}
	}
}

func importPendingCertAlerts(alerts []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.PendingAlertApi.PendingAlertAddPendingAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("Error, unable to create pending cert alert %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to pending cert alerts.")
		}
	}
}

func importNetworks(networks []keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, network := range networks {
		_, httpResp, reqErr := kfClient.SslApi.SslCreateNetwork(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Network(network).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(network.Name)
		if reqErr != nil {
			fmt.Printf("Error, unable to create SSL network %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to SSL networks.")
		}
	}
}

func findMatchingTempIds(exportedWorkflowDef exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest, kfClient *keyfactor_command_client_api.APIClient) *string {
	importInstanceTemplates, _, _ := kfClient.TemplateApi.TemplateGetTemplates(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	for _, template := range importInstanceTemplates {
		importInstTempNameJson, _ := json.Marshal(template.TemplateName)
		exportInstTempNameJson, _ := json.Marshal(exportedWorkflowDef.KeyName)
		if string(importInstTempNameJson) == string(exportInstTempNameJson) {
			importInstMatchingTemplateId, _ := json.Marshal(template.Id)
			importInstMatchingTemplateIdStr := string(importInstMatchingTemplateId)
			return &importInstMatchingTemplateIdStr
		}
	}
	return nil
}

func importWorkflowDefinitions(workflowDefs []exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, workflowDef := range workflowDefs {
		wJson, _ := json.Marshal(workflowDef)
		var workflowDefReq keyfactor_command_client_api.KeyfactorApiModelsWorkflowsDefinitionCreateRequest
		jErr := json.Unmarshal(wJson, &workflowDefReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		newTemplateId := findMatchingTempIds(workflowDef, kfClient)
		if newTemplateId != nil {
			workflowDefReq.Key = newTemplateId
		}
		_, httpResp, reqErr := kfClient.WorkflowDefinitionApi.WorkflowDefinitionCreateNewDefinition(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(workflowDefReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(workflowDef.DisplayName)
		if reqErr != nil {
			fmt.Printf("Error! Unable to create a new workflow definition %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to workflow definitions.")
		}
	}
}

func checkBuiltInReportDiffs(exportedReport exportModelsReport, kfClient *keyfactor_command_client_api.APIClient) *int32 {
	importInstanceReports, _, _ := kfClient.ReportsApi.ReportsQueryReports(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	//check if built in report was modified from default in exported instance; if modified, update built-in report in new instance
	for _, report := range importInstanceReports {
		importInstDispNameJson, _ := json.Marshal(report.DisplayName)
		exportInstDispNameJson, _ := json.Marshal(exportedReport.DisplayName)
		importInstInNavJson, _ := json.Marshal(report.InNavigator)
		exportInstInNavJson, _ := json.Marshal(exportedReport.InNavigator)
		importInstFavJson, _ := json.Marshal(report.Favorite)
		exportInstFavJson, _ := json.Marshal(exportedReport.Favorite)
		importInstRemDupJson, _ := json.Marshal(report.RemoveDuplicates)
		exportInstRemDupJson, _ := json.Marshal(exportedReport.RemoveDuplicates)
		var usesCollectionBool = "false"
		if string(importInstDispNameJson) == string(exportInstDispNameJson) {
			if (string(importInstFavJson) != string(exportInstFavJson)) || (string(importInstInNavJson) != string(exportInstInNavJson)) || (string(importInstRemDupJson) != string(exportInstRemDupJson)) {
				usesCollectionJson, _ := json.Marshal(exportedReport.UsesCollection)
				if string(usesCollectionJson) == usesCollectionBool {
					return report.Id
				}
			}
		}
	}
	return nil
}

// only imports built in reports where UsesCollections is false
func importBuiltInReports(reports []exportModelsReport, kfClient *keyfactor_command_client_api.APIClient) {
	for _, report := range reports {
		newReportId := checkBuiltInReportDiffs(report, kfClient)
		if newReportId != nil {
			rJson, _ := json.Marshal(report)
			var reportReq keyfactor_command_client_api.ModelsReportRequestModel
			jErr := json.Unmarshal(rJson, &reportReq)
			if jErr != nil {
				fmt.Printf("Error: %s\n", jErr)
				log.Fatalf("Error: %s", jErr)
			}
			reportReq.Id = newReportId
			_, httpResp, reqErr := kfClient.ReportsApi.ReportsUpdateReport(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(reportReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
			name, _ := json.Marshal(report.DisplayName)
			if reqErr != nil {
				fmt.Printf("Error! Unable to update built-in report %s: %s\n", string(name), httpResp.Body)
			} else {
				fmt.Println("Updated", string(name), "in built-in reports.")
			}
		}
	}
}

func importCustomReports(reports []keyfactor_command_client_api.ModelsCustomReportCreationRequest, kfClient *keyfactor_command_client_api.APIClient) {
	for _, report := range reports {
		_, httpResp, reqErr := kfClient.ReportsApi.ReportsCreateCustomReport(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(report).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(report.DisplayName)
		if reqErr != nil {
			fmt.Printf("Error! Unable to create custom report %s: %s\n", string(name), httpResp.Body)
		} else {
			fmt.Println("Added", string(name), "to custom reports.")
		}
	}
}

func importSecurityRoles(roles []api.CreateSecurityRoleArg, kfClient *api.Client) {
	for _, role := range roles {
		_, reqErr := kfClient.CreateSecurityRole(&role)
		name, _ := json.Marshal(role.Name)
		if reqErr != nil {
			fmt.Printf("Error! Unable to create security role %s: %s\n", string(name), reqErr)
		} else {
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
