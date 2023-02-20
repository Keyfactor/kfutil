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
)

type outJson struct {
	Collections         []keyfactor_command_client_api.ModelsCertificateQuery                                              `json:"Collections"`
	MetadataFields      []keyfactor_command_client_api.ModelsMetadataFieldTypeModel                                        `json:"MetadataFields"`
	ExpirationAlerts    []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertDefinitionResponse `json:"ExpirationAlerts"`
	IssuedCertAlerts    []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertDefinitionResponse         `json:"IssuedCertAlerts"`
	DeniedCertAlerts    []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertDefinitionResponse         `json:"DeniedCertAlerts"`
	PendingCertAlerts   []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertDefinitionResponse       `json:"PendingCertAlerts"`
	Networks            []keyfactor_command_client_api.KeyfactorApiModelsSslNetworkQueryResponse                           `json:"Networks"`
	WorkflowDefinitions []keyfactor_command_client_api.KeyfactorApiModelsWorkflowsDefinitionQueryResponse                  `json:"WorkflowDefinitions"`
	BuiltInReports      []keyfactor_command_client_api.ModelsReport                                                        `json:"BuiltInReports"`
	CustomReports       []keyfactor_command_client_api.ModelsCustomReport                                                  `json:"CustomReports"`
	SecurityRoles       api.GetSecurityRolesResponse                                                                       `json:"SecurityRoles"`
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

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Keyfactor instance migrate utilities.",
	Long:  `A collection of APIs and utilities for migrating Keyfactor instance data.`,
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Keyfactor instance export utilities.",
	Long:  `A collection of APIs and utilities for exporting Keyfactor instance data.`,
	Run: func(cmd *cobra.Command, args []string) {
		// initialize each entry as an empty list in the event it is not requested by the flags
		out := outJson{
			Collections:         []keyfactor_command_client_api.ModelsCertificateQuery{},
			MetadataFields:      []keyfactor_command_client_api.ModelsMetadataFieldTypeModel{},
			ExpirationAlerts:    []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertDefinitionResponse{},
			IssuedCertAlerts:    []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertDefinitionResponse{},
			DeniedCertAlerts:    []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertDefinitionResponse{},
			PendingCertAlerts:   []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertDefinitionResponse{},
			Networks:            []keyfactor_command_client_api.KeyfactorApiModelsSslNetworkQueryResponse{},
			WorkflowDefinitions: []keyfactor_command_client_api.KeyfactorApiModelsWorkflowsDefinitionQueryResponse{},
			BuiltInReports:      []keyfactor_command_client_api.ModelsReport{},
			CustomReports:       []keyfactor_command_client_api.ModelsCustomReport{},
			SecurityRoles:       api.GetSecurityRolesResponse{},
		}
		exportPath := cmd.Flag("file").Value.String()
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
		exportToJSON(out, exportPath)
	},
}

func getCollections() []keyfactor_command_client_api.ModelsCertificateQuery {
	kfClient := initGenClient()
	collections, _, reqErr := kfClient.CertificateCollectionApi.CertificateCollectionGetCollections(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get collections %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return collections
}

func getMetadata() []keyfactor_command_client_api.ModelsMetadataFieldTypeModel {
	kfClient := initGenClient()
	metadata, _, reqErr := kfClient.MetadataFieldApi.MetadataFieldGetAllMetadataFields(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get metadata %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return metadata
}

func getExpirationAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsExpirationExpirationAlertDefinitionResponse {
	kfClient := initGenClient()
	expirAlerts, _, reqErr := kfClient.ExpirationAlertApi.ExpirationAlertGetExpirationAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get expiration alerts %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return expirAlerts
}

func getIssuedAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsIssuedIssuedAlertDefinitionResponse {
	kfClient := initGenClient()
	issuedAlerts, _, reqErr := kfClient.IssuedAlertApi.IssuedAlertGetIssuedAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get issued cert alerts %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return issuedAlerts
}

func getDeniedAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsDeniedDeniedAlertDefinitionResponse {
	kfClient := initGenClient()
	deniedAlerts, _, reqErr := kfClient.DeniedAlertApi.DeniedAlertGetDeniedAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get denied cert alerts %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return deniedAlerts
}

func getPendingAlerts() []keyfactor_command_client_api.KeyfactorApiModelsAlertsPendingPendingAlertDefinitionResponse {
	kfClient := initGenClient()
	pendingAlerts, _, reqErr := kfClient.PendingAlertApi.PendingAlertGetPendingAlerts(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get pending cert alerts %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return pendingAlerts
}

func getSslNetworks() []keyfactor_command_client_api.KeyfactorApiModelsSslNetworkQueryResponse {
	kfClient := initGenClient()
	networks, _, reqErr := kfClient.SslApi.SslGetNetworks(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get SSL networks %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return networks
}

func getWorkflowDefinitions() []keyfactor_command_client_api.KeyfactorApiModelsWorkflowsDefinitionQueryResponse {
	kfClient := initGenClient()
	workflowDefs, _, reqErr := kfClient.WorkflowDefinitionApi.WorkflowDefinitionQuery(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("Error, unable to get workflow definitions %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return workflowDefs
}

func getReports() ([]keyfactor_command_client_api.ModelsReport, []keyfactor_command_client_api.ModelsCustomReport) {
	kfClient := initGenClient()
	//Gets all built-in reports
	bReports, _, bErr := kfClient.ReportsApi.ReportsQueryReports(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if bErr != nil {
		fmt.Printf("Error, unable to get built-in reports %s\n", bErr)
		log.Fatalf("Error: %s", bErr)
	}
	//Gets all custom reports
	cReports, _, cErr := kfClient.ReportsApi.ReportsQueryCustomReports(context.Background()).XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).Execute()
	if cErr != nil {
		fmt.Printf("Error, unable to get custom reports %s\n", cErr)
		log.Fatalf("Error: %s", cErr)
	}
	return bReports, cReports
}

func getRoles() api.GetSecurityRolesResponse {
	kfClient, _ := initClient()
	roles, reqErr := kfClient.GetSecurityRoles()
	if reqErr != nil {
		fmt.Printf("Error, unable to get roles %s\n", reqErr)
		log.Fatalf("Error: %s", reqErr)
	}
	return roles
}

func init() {
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

	RootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportPath, "file", "f", "", "export JSON to a specified filepath")
	exportCmd.MarkFlagRequired("file")

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
