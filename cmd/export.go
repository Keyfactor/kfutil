package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	keyfactor_command_client_api "github.com/Keyfactor/keyfactor-go-client-sdk"
	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
)

var exportPath string
var fCollections bool
var fMetadata bool
var fExpirationAlerts bool
var fIssuedAlerts bool
var fDeniedAlerts bool
var fPendingAlerts bool
var fNetworks bool
var fWorkflowDefinitions bool
var fReports bool
var fSecurityRoles bool
var fAll bool

type exportModelsReport struct {
	Id                      *int32                                                `json:"-"`
	Scheduled               *int32                                                `json:"Scheduled,omitempty"`
	DisplayName             *string                                               `json:"DisplayName,omitempty"`
	Description             *string                                               `json:"Description,omitempty"`
	ReportPath              *string                                               `json:"ReportPath,omitempty"`
	VersionNumber           *string                                               `json:"VersionNumber,omitempty"`
	Categories              *string                                               `json:"Categories,omitempty"`
	ShortName               *string                                               `json:"ShortName,omitempty"`
	InNavigator             *bool                                                 `json:"InNavigator,omitempty"`
	Favorite                *bool                                                 `json:"Favorite,omitempty"`
	RemoveDuplicates        *bool                                                 `json:"RemoveDuplicates,omitempty"`
	UsesCollection          *bool                                                 `json:"UsesCollection,omitempty"`
	ReportParameter         []keyfactor_command_client_api.ModelsReportParameters `json:"ReportParameter,omitempty"`
	Schedules               []keyfactor_command_client_api.ModelsReportSchedule   `json:"Schedules,omitempty"`
	AcceptedScheduleFormats []string                                              `json:"AcceptedScheduleFormats,omitempty"`
}

type exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest struct {
	// Display name of the Definition
	DisplayName *string `json:"DisplayName,omitempty"`
	// Description of the Definition
	Description *string `json:"Description,omitempty"`
	// Key to be used to look up definition when starting a new workflow.  For enrollment workflowTypes, this should be a template
	Key *string `json:"Key,omitempty"`
	//Name of Template corresponding to key value
	KeyName *string `json:"KeyName,omitempty"`
	// The Type of Workflow
	WorkflowType *string `json:"WorkflowType,omitempty"`
}

type outJson struct {
	Collections         []keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest `json:"Collections"`
	MetadataFields      []keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest                  `json:"MetadataFields"`
	ExpirationAlerts    []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest           `json:"ExpirationAlerts"`
	IssuedCertAlerts    []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest                   `json:"IssuedCertAlerts"`
	DeniedCertAlerts    []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest                   `json:"DeniedCertAlerts"`
	PendingCertAlerts   []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest                 `json:"PendingCertAlerts"`
	Networks            []keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest                                  `json:"Networks"`
	WorkflowDefinitions []exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest                                                `json:"WorkflowDefinitions"`
	BuiltInReports      []exportModelsReport                                                                                      `json:"BuiltInReports"`
	CustomReports       []keyfactor_command_client_api.ModelsCustomReportCreationRequest                                          `json:"CustomReports"`
	SecurityRoles       []api.CreateSecurityRoleArg                                                                               `json:"SecurityRoles"`
}

func exportToJSON(out outJson, exportPath string) {
	mOut, jErr := json.MarshalIndent(out, "", "    ")
	if jErr != nil {
		fmt.Printf("Error processing JSON object. %s\n", jErr)
		log.Fatalf("[ERROR]: %s", jErr)
	}
	wErr := os.WriteFile(exportPath, mOut, 0666)
	if wErr != nil {
		fmt.Printf("Error writing files to %s: %s\n", exportPath, wErr)
		log.Fatalf("[ERROR]: %s", wErr)
	} else {
		fmt.Printf("Content successfully written to %s", exportPath)
	}
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Keyfactor instance export utilities.",
	Long:  `A collection of APIs and utilities for exporting Keyfactor instance data.`,
	Run: func(cmd *cobra.Command, args []string) {
		// initialize each entry as an empty list in the event it is not requested by the flags
		out := outJson{
			Collections:         []keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest{},
			MetadataFields:      []keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest{},
			ExpirationAlerts:    []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest{},
			IssuedCertAlerts:    []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest{},
			DeniedCertAlerts:    []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest{},
			PendingCertAlerts:   []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest{},
			Networks:            []keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest{},
			WorkflowDefinitions: []exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest{},
			BuiltInReports:      []exportModelsReport{},
			CustomReports:       []keyfactor_command_client_api.ModelsCustomReportCreationRequest{},
			SecurityRoles:       []api.CreateSecurityRoleArg{},
		}
		exportPath := cmd.Flag("file").Value.String()
		if cmd.Flag("all").Value.String() == "true" {
			out.Collections = getCollections()
			out.MetadataFields = getMetadata()
			out.ExpirationAlerts = getExpirationAlerts()
			out.IssuedCertAlerts = getIssuedAlerts()
			out.DeniedCertAlerts = getDeniedAlerts()
			out.PendingCertAlerts = getPendingAlerts()
			out.Networks = getSslNetworks()
			out.WorkflowDefinitions = getWorkflowDefinitions()
			out.BuiltInReports, out.CustomReports = getReports()
			out.SecurityRoles = getRoles()
		} else {
			if cmd.Flag("collections").Value.String() == "true" {
				out.Collections = getCollections()
			}
			if cmd.Flag("metadata").Value.String() == "true" {
				out.MetadataFields = getMetadata()
			}
			if cmd.Flag("expiration-alerts").Value.String() == "true" {
				out.ExpirationAlerts = getExpirationAlerts()
			}
			if cmd.Flag("issued-alerts").Value.String() == "true" {
				out.IssuedCertAlerts = getIssuedAlerts()
			}
			if cmd.Flag("denied-alerts").Value.String() == "true" {
				out.DeniedCertAlerts = getDeniedAlerts()
			}
			if cmd.Flag("pending-alerts").Value.String() == "true" {
				out.PendingCertAlerts = getPendingAlerts()
			}
			if cmd.Flag("networks").Value.String() == "true" {
				out.Networks = getSslNetworks()
			}
			if cmd.Flag("workflow-definitions").Value.String() == "true" {
				out.WorkflowDefinitions = getWorkflowDefinitions()
			}
			if cmd.Flag("reports").Value.String() == "true" {
				out.BuiltInReports, out.CustomReports = getReports()
			}
			if cmd.Flag("security-roles").Value.String() == "true" {
				out.SecurityRoles = getRoles()
			}
		}
		exportToJSON(out, exportPath)
	},
}

func getCollections() []keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest {
	kfClient := initGenClient()
	collections, _, reqErr := kfClient.CertificateCollectionApi.CertificateCollectionGetCollections(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get collections %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lCollectionReq []keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest
	for _, collection := range collections {
		cJson, _ := json.Marshal(collection)
		var collectionReq keyfactor_command_client_api.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest
		jErr := json.Unmarshal(cJson, &collectionReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		collectionReq.Query = collection.Content
		collectionReq.Id = nil
		lCollectionReq = append(lCollectionReq, collectionReq)
	}
	return lCollectionReq
}

func getMetadata() []keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest {
	kfClient := initGenClient()
	metadata, _, reqErr := kfClient.MetadataFieldApi.MetadataFieldGetAllMetadataFields(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get metadata %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lMetadataReq []keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest
	for _, metadataItem := range metadata {
		mJson, _ := json.Marshal(metadataItem)
		var metadataReq keyfactor_command_client_api.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest
		jErr := json.Unmarshal(mJson, &metadataReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		metadataItem.Id = nil
		lMetadataReq = append(lMetadataReq, metadataReq)
	}
	return lMetadataReq
}

func getExpirationAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest {
	kfClient := initGenClient()
	alerts, _, reqErr := kfClient.ExpirationAlertApi.ExpirationAlertGetExpirationAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get expiration alerts %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lAlertReq []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getIssuedAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest {
	kfClient := initGenClient()
	alerts, _, reqErr := kfClient.IssuedAlertApi.IssuedAlertGetIssuedAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get issued cert alerts %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lAlertReq []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		alertReq.TemplateId = nil
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getDeniedAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest {
	kfClient := initGenClient()
	alerts, _, reqErr := kfClient.DeniedAlertApi.DeniedAlertGetDeniedAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get denied cert alerts %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lAlertReq []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		alertReq.TemplateId = nil
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getPendingAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest {
	kfClient := initGenClient()
	alerts, _, reqErr := kfClient.PendingAlertApi.PendingAlertGetPendingAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get pending cert alerts %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lAlertReq []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		alertReq.TemplateId = nil
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getSslNetworks() []keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest {
	kfClient := initGenClient()
	networks, _, reqErr := kfClient.SslApi.SslGetNetworks(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get SSL networks %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lNetworkReq []keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest
	for _, network := range networks {
		mJson, _ := json.Marshal(network)
		var networkReq keyfactor_command_client_api.KeyfactorApiModelsSslCreateNetworkRequest
		jErr := json.Unmarshal(mJson, &networkReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		lNetworkReq = append(lNetworkReq, networkReq)
	}
	return lNetworkReq
}

func getWorkflowDefinitions() []exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest {
	kfClient := initGenClient()
	workflowDefs, _, reqErr := kfClient.WorkflowDefinitionApi.WorkflowDefinitionQuery(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get workflow definitions %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lWorkflowReq []exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest
	for _, workflowDef := range workflowDefs {
		mJson, _ := json.Marshal(workflowDef)
		var workflowReq exportKeyfactorApiModelsWorkflowsDefinitionCreateRequest
		jErr := json.Unmarshal(mJson, &workflowReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		if workflowDef.Key != nil {
			key, _ := strconv.ParseInt(*workflowDef.Key, 10, 64)
			key32 := int32(key)
			template, _, _ := kfClient.TemplateApi.TemplateGetTemplate(context.Background(), key32).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
			workflowReq.KeyName = template.TemplateName
		}
		workflowReq.Key = nil
		lWorkflowReq = append(lWorkflowReq, workflowReq)
	}
	return lWorkflowReq
}

func getReports() ([]exportModelsReport, []keyfactor_command_client_api.ModelsCustomReportCreationRequest) {
	kfClient := initGenClient()
	//Gets all built-in reports
	bReports, _, bErr := kfClient.ReportsApi.ReportsQueryReports(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if bErr != nil {
		fmt.Printf("%s Error! Unable to get built-in reports %s%s\n", colorRed, bErr, colorWhite)
	}
	var lbReportsReq []exportModelsReport
	for _, bReport := range bReports {
		mJson, _ := json.Marshal(bReport)
		var newbReport exportModelsReport
		jErr := json.Unmarshal(mJson, &newbReport)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		newbReport.Id = nil
		lbReportsReq = append(lbReportsReq, newbReport)
	}
	//Gets all custom reports
	cReports, _, cErr := kfClient.ReportsApi.ReportsQueryCustomReports(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if cErr != nil {
		fmt.Printf("%s Error! Unable to get custom reports %s%s\n", colorRed, cErr, colorWhite)
	}
	var lcReportReq []keyfactor_command_client_api.ModelsCustomReportCreationRequest
	for _, cReport := range cReports {
		mJson, _ := json.Marshal(cReport)
		var cReportReq keyfactor_command_client_api.ModelsCustomReportCreationRequest
		jErr := json.Unmarshal(mJson, &cReportReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		lcReportReq = append(lcReportReq, cReportReq)
	}
	return lbReportsReq, lcReportReq
}

func getRoles() []api.CreateSecurityRoleArg {
	kfClient, _ := initClient()
	roles, reqErr := kfClient.GetSecurityRoles()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get security roles %s%s\n", colorRed, reqErr, colorWhite)
	}
	var lRoleReq []api.CreateSecurityRoleArg
	for _, role := range roles {
		mJson, _ := json.Marshal(role)
		var cRoleReq api.CreateSecurityRoleArg
		jErr := json.Unmarshal(mJson, &cRoleReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Fatalf("Error: %s", jErr)
		}
		lRoleReq = append(lRoleReq, cRoleReq)
	}
	return lRoleReq
}

func init() {
	RootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportPath, "file", "f", "", "export JSON to a specified filepath")
	exportCmd.MarkFlagRequired("file")

	exportCmd.Flags().BoolVarP(&fAll, "all", "a", false, "export all exportable data to JSON file")
	exportCmd.Flags().Lookup("all").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fCollections, "collections", "c", false, "export collections to JSON file")
	exportCmd.Flags().Lookup("collections").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fMetadata, "metadata", "m", false, "export metadata to JSON file")
	exportCmd.Flags().Lookup("metadata").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fExpirationAlerts, "expiration-alerts", "e", false, "export expiration cert alerts to JSON file")
	exportCmd.Flags().Lookup("expiration-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fIssuedAlerts, "issued-alerts", "i", false, "export issued cert alerts to JSON file")
	exportCmd.Flags().Lookup("issued-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fDeniedAlerts, "denied-alerts", "d", false, "export denied cert alerts to JSON file")
	exportCmd.Flags().Lookup("denied-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fPendingAlerts, "pending-alerts", "p", false, "export pending cert alerts to JSON file")
	exportCmd.Flags().Lookup("pending-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fNetworks, "networks", "n", false, "export SSL networks to JSON file")
	exportCmd.Flags().Lookup("networks").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fWorkflowDefinitions, "workflow-definitions", "w", false, "export workflow definitions to JSON file")
	exportCmd.Flags().Lookup("workflow-definitions").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fReports, "reports", "r", false, "export reports to JSON file")
	exportCmd.Flags().Lookup("reports").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fSecurityRoles, "security-roles", "s", false, "export security roles to JSON file")
	exportCmd.Flags().Lookup("security-roles").NoOptDefVal = "true"
}
