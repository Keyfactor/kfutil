package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

type Body struct {
	ErrorCode string
	Message   string
}

func parseError(error io.ReadCloser) string {
	bytes, _ := io.ReadAll(error)
	var newError Body
	json.Unmarshal(bytes, &newError)
	return newError.Message
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Keyfactor instance import utilities.",
	Long:  `A collection of APIs and utilities for importing Keyfactor instance data.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("hostname")
		kfcUsername, _ := cmd.Flags().GetString("username")
		kfcPassword, _ := cmd.Flags().GetString("password")
		kfcDomain, _ := cmd.Flags().GetString("domain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)

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
		kfClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		oldkfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
		if cmd.Flag("all").Value.String() == "true" {
			importCollections(out.Collections, kfClient)
			importMetadataFields(out.MetadataFields, kfClient)
			importIssuedCertAlerts(out.IssuedCertAlerts, kfClient)
			importDeniedCertAlerts(out.DeniedCertAlerts, kfClient)
			importPendingCertAlerts(out.PendingCertAlerts, kfClient)
			importNetworks(out.Networks, kfClient)
			importWorkflowDefinitions(out.WorkflowDefinitions, kfClient)
			importBuiltInReports(out.BuiltInReports, kfClient)
			importCustomReports(out.CustomReports, kfClient)
			importSecurityRoles(out.SecurityRoles, oldkfClient)
		} else {
			if len(out.Collections) != 0 && cmd.Flag("collections").Value.String() == "true" {
				importCollections(out.Collections, kfClient)
			}
			if len(out.MetadataFields) != 0 && cmd.Flag("metadata").Value.String() == "true" {
				importMetadataFields(out.MetadataFields, kfClient)
			}
			if len(out.IssuedCertAlerts) != 0 && cmd.Flag("issued-alerts").Value.String() == "true" {
				importIssuedCertAlerts(out.IssuedCertAlerts, kfClient)
			}
			if len(out.DeniedCertAlerts) != 0 && cmd.Flag("denied-alerts").Value.String() == "true" {
				importDeniedCertAlerts(out.DeniedCertAlerts, kfClient)
			}
			if len(out.PendingCertAlerts) != 0 && cmd.Flag("pending-alerts").Value.String() == "true" {
				importPendingCertAlerts(out.PendingCertAlerts, kfClient)
			}
			if len(out.Networks) != 0 && cmd.Flag("networks").Value.String() == "true" {
				importNetworks(out.Networks, kfClient)
			}
			if len(out.WorkflowDefinitions) != 0 && cmd.Flag("workflow-definitions").Value.String() == "true" {
				importWorkflowDefinitions(out.WorkflowDefinitions, kfClient)
			}
			if len(out.BuiltInReports) != 0 && cmd.Flag("reports").Value.String() == "true" {
				importBuiltInReports(out.BuiltInReports, kfClient)
			}
			if len(out.CustomReports) != 0 && cmd.Flag("reports").Value.String() == "true" {
				importCustomReports(out.CustomReports, kfClient)
			}
			if len(out.SecurityRoles) != 0 && cmd.Flag("security-roles").Value.String() == "true" {
				importSecurityRoles(out.SecurityRoles, oldkfClient)
			}
		}
	},
}

func importCollections(collections []keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest, kfClient *keyfactor.APIClient) {
	for _, collection := range collections {
		_, httpResp, reqErr := kfClient.CertificateCollectionApi.CertificateCollectionCreateCollection(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).
			Request(collection).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(collection.Name)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create collection %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			name, _ := json.Marshal(collection.Name)
			fmt.Println("Added", string(name), "to collections")
		}
	}
}

func importMetadataFields(metadataFields []keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest, kfClient *keyfactor.APIClient) {
	for _, metadata := range metadataFields {
		_, httpResp, reqErr := kfClient.MetadataFieldApi.MetadataFieldCreateMetadataField(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).MetadataFieldType(metadata).
			XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(metadata.Name)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create metadata field type %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			fmt.Println("Added", string(name), "to metadata field types.")
		}
	}
}

func importIssuedCertAlerts(alerts []keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest, kfClient *keyfactor.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.IssuedAlertApi.IssuedAlertAddIssuedAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create issued cert alert %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			fmt.Println("Added", string(name), "to issued cert alerts.")
		}
	}
}

func importDeniedCertAlerts(alerts []keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest, kfClient *keyfactor.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.DeniedAlertApi.DeniedAlertAddDeniedAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create denied cert alert %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			fmt.Println("Added", string(name), "to denied cert alerts.")
		}
	}
}

func importPendingCertAlerts(alerts []keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest, kfClient *keyfactor.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.PendingAlertApi.PendingAlertAddPendingAlert(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create pending cert alert %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			fmt.Println("Added", string(name), "to pending cert alerts.")
		}
	}
}

func importNetworks(networks []keyfactor.KeyfactorApiModelsSslCreateNetworkRequest, kfClient *keyfactor.APIClient) {
	for _, network := range networks {
		_, httpResp, reqErr := kfClient.SslApi.SslCreateNetwork(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Network(network).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(network.Name)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create SSL network %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			fmt.Println("Added", string(name), "to SSL networks.")
		}
	}
}

// identify matching templates between instances by name, then return the template Id of the matching template in the import instance
func findMatchingTemplates(exportedWorkflowDef exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest, kfClient *keyfactor.APIClient) *string {
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

func importWorkflowDefinitions(workflowDefs []exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest, kfClient *keyfactor.APIClient) {
	for _, workflowDef := range workflowDefs {
		wJson, _ := json.Marshal(workflowDef)
		var workflowDefReq keyfactor.KeyfactorApiModelsWorkflowsDefinitionCreateRequest
		jErr := json.Unmarshal(wJson, &workflowDefReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		newTemplateId := findMatchingTemplates(workflowDef, kfClient)
		if newTemplateId != nil {
			workflowDefReq.Key = newTemplateId
		}
		_, httpResp, reqErr := kfClient.WorkflowDefinitionApi.WorkflowDefinitionCreateNewDefinition(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(workflowDefReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(workflowDef.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create workflow definition %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
		} else {
			fmt.Println("Added", string(name), "to workflow definitions.")
		}
	}
}

// check for built-in report discrepancies between instances, return the report id of reports that need to be updated in import instance
func checkBuiltInReportDiffs(exportedReport exportModelsReport, kfClient *keyfactor.APIClient) *int32 {
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
func importBuiltInReports(reports []exportModelsReport, kfClient *keyfactor.APIClient) {
	for _, report := range reports {
		newReportId := checkBuiltInReportDiffs(report, kfClient)
		if newReportId != nil {
			rJson, _ := json.Marshal(report)
			var reportReq keyfactor.ModelsReportRequestModel
			jErr := json.Unmarshal(rJson, &reportReq)
			if jErr != nil {
				fmt.Printf("Error: %s\n", jErr)
				log.Fatalf("Error: %s", jErr)
			}
			reportReq.Id = newReportId
			_, httpResp, reqErr := kfClient.ReportsApi.ReportsUpdateReport(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(reportReq).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
			name, _ := json.Marshal(report.DisplayName)
			if reqErr != nil {
				fmt.Printf("%s Error! Unable to update built-in report %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
			} else {
				fmt.Println("Updated", string(name), "in built-in reports.")
			}
		}
	}
}

func importCustomReports(reports []keyfactor.ModelsCustomReportCreationRequest, kfClient *keyfactor.APIClient) {
	for _, report := range reports {
		_, httpResp, reqErr := kfClient.ReportsApi.ReportsCreateCustomReport(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).Request(report).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(report.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create custom report %s - %s%s\n", colorRed, string(name), parseError(httpResp.Body), colorWhite)
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
			fmt.Printf("%s Error! Unable to create security role %s - %s%s\n", colorRed, string(name), reqErr, colorWhite)
		} else {
			fmt.Println("Added", string(name), "to security roles.")
		}
	}
}

func init() {
	RootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(&exportPath, "file", "f", "", "path to JSON file containing exported data")
	importCmd.MarkFlagRequired("file")

	importCmd.Flags().BoolVarP(&fAll, "all", "a", false, "import all importable data to JSON file")
	importCmd.Flags().Lookup("all").NoOptDefVal = "true"

	importCmd.Flags().BoolVarP(&fCollections, "collections", "c", false, "import collections to JSON file")
	importCmd.Flags().Lookup("collections").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fMetadata, "metadata", "m", false, "import metadata to JSON file")
	importCmd.Flags().Lookup("metadata").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fIssuedAlerts, "issued-alerts", "i", false, "import issued cert alerts to JSON file")
	importCmd.Flags().Lookup("issued-alerts").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fDeniedAlerts, "denied-alerts", "d", false, "import denied cert alerts to JSON file")
	importCmd.Flags().Lookup("denied-alerts").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fPendingAlerts, "pending-alerts", "p", false, "import pending cert alerts to JSON file")
	importCmd.Flags().Lookup("pending-alerts").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fNetworks, "networks", "n", false, "import SSL networks to JSON file")
	importCmd.Flags().Lookup("networks").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fWorkflowDefinitions, "workflow-definitions", "w", false, "import workflow definitions to JSON file")
	importCmd.Flags().Lookup("workflow-definitions").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fReports, "reports", "r", false, "import reports to JSON file")
	importCmd.Flags().Lookup("reports").NoOptDefVal = "true"
	importCmd.Flags().BoolVarP(&fSecurityRoles, "security-roles", "s", false, "import security roles to JSON file")
	importCmd.Flags().Lookup("security-roles").NoOptDefVal = "true"
}
