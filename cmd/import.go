// Copyright 2024 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type Body struct {
	ErrorCode string
	Message   string
}

func parseError(error io.ReadCloser) string {
	log.Debug().Msgf("%s: parseError", DebugFuncEnter)

	log.Debug().Msg("Reading error body")
	bytes, ioErr := io.ReadAll(error)
	if ioErr != nil {
		fmt.Printf("Error: %s\n", ioErr)
		log.Error().Err(ioErr).Send()
		return ioErr.Error()
	}
	var newError Body
	jErr := json.Unmarshal(bytes, &newError)
	if jErr != nil {
		fmt.Printf("Error: %s\n", jErr)
		log.Error().Err(jErr).Send()
		return jErr.Error()
	}
	log.Debug().Msgf("%s: parseError", DebugFuncExit)
	return newError.Message
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Keyfactor instance import utilities.",
	Long:  `A collection of APIs and utilities for importing Keyfactor instance data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug().Msgf("%s: importCmd", DebugFuncEnter)
		isExperimental := true

		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		log.Info().Msg("Running import...")

		log.Debug().Msgf("%s: createAuthConfigFromParams", DebugFuncCall)
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		if authConfig == nil {
			return fmt.Errorf("Error: %s", FailedAuthMsg)
		}

		exportPath := cmd.Flag("file").Value.String()
		log.Debug().Str("exportPath", exportPath).Msg("exportPath")

		log.Debug().Str("exportPath", exportPath).
			Msg("Reading exported file")

		jsonFile, oErr := os.Open(exportPath)
		if oErr != nil {
			fmt.Printf("Error opening exported file: %s\n", oErr)
			//log.Fatalf("Error: %s", oErr)
			log.Error().
				Str("exportPath", exportPath).
				Err(oErr).
				Send()
		}
		defer jsonFile.Close()
		var out outJson
		bJson, ioErr := io.ReadAll(jsonFile)
		if ioErr != nil {
			fmt.Printf("Error reading exported file: %s\n", ioErr)
			//log.Fatalf("Error: %s", ioErr)
			log.Error().Err(ioErr).Send()
			return ioErr
		}
		jErr := json.Unmarshal(bJson, &out)
		if jErr != nil {
			fmt.Printf("Error reading exported file: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
			return jErr
		}
		log.Debug().Msgf("%s: initGenClient", DebugFuncCall)
		kfClient, clientErr := initGenClient(configFile, profile, noPrompt, authConfig, false)
		log.Debug().Msgf("%s: initClient", DebugFuncExit)
		oldkfClient, oldClientErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		if clientErr != nil {
			log.Error().Err(clientErr).Send()
			return clientErr
		} else if oldClientErr != nil {
			log.Error().Err(oldClientErr).Send()
			return oldClientErr
		}

		if cmd.Flag("all").Value.String() == "true" {
			log.Debug().Msgf("%s: importCollections", DebugFuncCall)
			importCollections(out.Collections, kfClient)
			log.Debug().Msgf("%s: importMetadataFields", DebugFuncCall)
			importMetadataFields(out.MetadataFields, kfClient)

			log.Debug().Msgf("%s: importIssuedCertAlerts", DebugFuncCall)
			importIssuedCertAlerts(out.IssuedCertAlerts, kfClient)

			log.Debug().Msgf("%s: importDeniedCertAlerts", DebugFuncCall)
			importDeniedCertAlerts(out.DeniedCertAlerts, kfClient)

			log.Debug().Msgf("%s: importPendingCertAlerts", DebugFuncCall)
			importPendingCertAlerts(out.PendingCertAlerts, kfClient)

			log.Debug().Msgf("%s: importNetworks", DebugFuncCall)
			importNetworks(out.Networks, kfClient)

			log.Debug().Msgf("%s: importWorkflowDefinitions", DebugFuncCall)
			importWorkflowDefinitions(out.WorkflowDefinitions, kfClient)

			log.Debug().Msgf("%s: importBuiltInReports", DebugFuncCall)
			importBuiltInReports(out.BuiltInReports, kfClient)

			log.Debug().Msgf("%s: importCustomReports", DebugFuncCall)
			importCustomReports(out.CustomReports, kfClient)

			log.Debug().Msgf("%s: importSecurityRoles", DebugFuncCall)
			importSecurityRoles(out.SecurityRoles, oldkfClient)
		} else {
			if len(out.Collections) != 0 && cmd.Flag("collections").Value.String() == "true" {
				log.Debug().Msgf("%s: importCollections", DebugFuncCall)
				importCollections(out.Collections, kfClient)
			}
			if len(out.MetadataFields) != 0 && cmd.Flag("metadata").Value.String() == "true" {
				log.Debug().Msgf("%s: importMetadataFields", DebugFuncCall)
				importMetadataFields(out.MetadataFields, kfClient)
			}
			if len(out.IssuedCertAlerts) != 0 && cmd.Flag("issued-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: importIssuedCertAlerts", DebugFuncCall)
				importIssuedCertAlerts(out.IssuedCertAlerts, kfClient)
			}
			if len(out.DeniedCertAlerts) != 0 && cmd.Flag("denied-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: importDeniedCertAlerts", DebugFuncCall)
				importDeniedCertAlerts(out.DeniedCertAlerts, kfClient)
			}
			if len(out.PendingCertAlerts) != 0 && cmd.Flag("pending-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: importPendingCertAlerts", DebugFuncCall)
				importPendingCertAlerts(out.PendingCertAlerts, kfClient)
			}
			if len(out.Networks) != 0 && cmd.Flag("networks").Value.String() == "true" {
				log.Debug().Msgf("%s: importNetworks", DebugFuncCall)
				importNetworks(out.Networks, kfClient)
			}
			if len(out.WorkflowDefinitions) != 0 && cmd.Flag("workflow-definitions").Value.String() == "true" {
				log.Debug().Msgf("%s: importWorkflowDefinitions", DebugFuncCall)
				importWorkflowDefinitions(out.WorkflowDefinitions, kfClient)
			}
			if len(out.BuiltInReports) != 0 && cmd.Flag("reports").Value.String() == "true" {
				log.Debug().Msgf("%s: importBuiltInReports", DebugFuncCall)
				importBuiltInReports(out.BuiltInReports, kfClient)
			}
			if len(out.CustomReports) != 0 && cmd.Flag("reports").Value.String() == "true" {
				log.Debug().Msgf("%s: importCustomReports", DebugFuncCall)
				importCustomReports(out.CustomReports, kfClient)
			}
			if len(out.SecurityRoles) != 0 && cmd.Flag("security-roles").Value.String() == "true" {
				log.Debug().Msgf("%s: importSecurityRoles", DebugFuncCall)
				importSecurityRoles(out.SecurityRoles, oldkfClient)
			}
		}
		log.Debug().Msgf("%s: importCmd", DebugFuncExit)
		return nil
	},
}

func importCollections(collections []keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest, kfClient *keyfactor.APIClient) {
	for _, collection := range collections {
		_, httpResp, reqErr := kfClient.CertificateCollectionApi.
			CertificateCollectionCreateCollection(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).
			Request(collection).
			XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		name, jmErr := json.Marshal(collection.Name)
		if jmErr != nil {
			fmt.Printf("Error: %s\n", jmErr)
			//log.Fatalf("Error: %s", jmErr)
			log.Error().Err(jmErr).Send()
		}
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create collection %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
		} else {
			n, jnErr := json.Marshal(collection.Name)
			if jnErr != nil {
				fmt.Printf("Error: %s\n", jnErr)
				//log.Fatalf("Error: %s", jnErr)
				log.Error().Err(jnErr).Send()
			}
			fmt.Println("Added", string(n), "to collections")
		}
	}
}

func importMetadataFields(metadataFields []keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest, kfClient *keyfactor.APIClient) {
	for _, metadata := range metadataFields {
		_, httpResp, reqErr := kfClient.MetadataFieldApi.MetadataFieldCreateMetadataField(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).
			MetadataFieldType(metadata).
			XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		n, jmErr := json.Marshal(metadata.Name)

		if reqErr != nil {
			if jmErr != nil {
				fmt.Printf("Error: %s\n", jmErr)
				//log.Fatalf("Error: %s", jmErr)
				log.Error().Err(jmErr).Send()
			}
			log.Error().Err(reqErr).Send()
			fmt.Printf("%s Error! Unable to create metadata field type %s - %s%s\n", ColorRed, string(n), parseError(httpResp.Body), ColorWhite)
		} else {
			log.Info().Msgf("Added %s to metadata field types.", string(n))
			fmt.Println("Added", string(n), "to metadata field types.")
		}
	}
}

func importIssuedCertAlerts(alerts []keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest, kfClient *keyfactor.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.IssuedAlertApi.IssuedAlertAddIssuedAlert(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create issued cert alert %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
		} else {
			fmt.Println("Added", string(name), "to issued cert alerts.")
		}
	}
}

func importDeniedCertAlerts(alerts []keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest, kfClient *keyfactor.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.DeniedAlertApi.DeniedAlertAddDeniedAlert(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create denied cert alert %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
		} else {
			fmt.Println("Added", string(name), "to denied cert alerts.")
		}
	}
}

func importPendingCertAlerts(alerts []keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest, kfClient *keyfactor.APIClient) {
	for _, alert := range alerts {
		_, httpResp, reqErr := kfClient.PendingAlertApi.PendingAlertAddPendingAlert(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).Req(alert).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(alert.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create pending cert alert %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
		} else {
			fmt.Println("Added", string(name), "to pending cert alerts.")
		}
	}
}

func importNetworks(networks []keyfactor.KeyfactorApiModelsSslCreateNetworkRequest, kfClient *keyfactor.APIClient) {
	for _, network := range networks {
		_, httpResp, reqErr := kfClient.SslApi.SslCreateNetwork(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).Network(network).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(network.Name)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create SSL network %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
		} else {
			fmt.Println("Added", string(name), "to SSL networks.")
		}
	}
}

// identify matching templates between instances by name, then return the template Id of the matching template in the import instance
func findMatchingTemplates(exportedWorkflowDef exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest, kfClient *keyfactor.APIClient) *string {
	importInstanceTemplates, _, _ := kfClient.TemplateApi.TemplateGetTemplates(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
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
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
		}
		newTemplateId := findMatchingTemplates(workflowDef, kfClient)
		if newTemplateId != nil {
			workflowDefReq.Key = newTemplateId
		}
		_, httpResp, reqErr := kfClient.WorkflowDefinitionApi. // todo: Why is the object not being used?
									WorkflowDefinitionCreateNewDefinition(context.Background()).
									XKeyfactorRequestedWith(XKeyfactorRequestedWith).
									Request(workflowDefReq).
									XKeyfactorApiVersion(XKeyfactorApiVersion).
									Execute()
		name, jmErr := json.Marshal(workflowDef.DisplayName)
		if jmErr != nil {
			fmt.Printf("Error: %s\n", jmErr)
			//log.Fatalf("Error: %s", jmErr)
			log.Error().Err(jmErr).Send()
			return
		}

		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create workflow definition %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
			log.Error().Err(reqErr).Send()
		} else {
			fmt.Println("Added", string(name), "to workflow definitions.")
			log.Info().Msgf("Added %s to workflow definitions.", string(name))
		}
	}
}

// check for built-in report discrepancies between instances, return the report id of reports that need to be updated in import instance
func checkBuiltInReportDiffs(exportedReport exportModelsReport, kfClient *keyfactor.APIClient) *int32 {
	importInstanceReports, _, _ := kfClient.ReportsApi.ReportsQueryReports(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
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
				//log.Fatalf("Error: %s", jErr)
				log.Error().Err(jErr).Send()
			}
			reportReq.Id = newReportId
			_, httpResp, reqErr := kfClient.ReportsApi. //todo: Why is the object not being used?
									ReportsUpdateReport(context.Background()).
									XKeyfactorRequestedWith(XKeyfactorRequestedWith).
									Request(reportReq).
									XKeyfactorApiVersion(XKeyfactorApiVersion).
									Execute()
			name, jmErr := json.Marshal(report.DisplayName)
			if jmErr != nil {
				fmt.Printf("Error: %s\n", jmErr)
				//log.Fatalf("Error: %s", jmErr)
				log.Error().Err(jmErr).Send()
				return
			}
			if reqErr != nil {
				fmt.Printf("%s Error! Unable to update built-in report %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
				log.Error().Err(reqErr).Send()
			} else {
				fmt.Println("Updated", string(name), "in built-in reports.")
				log.Info().Msgf("Updated %s in built-in reports.", string(name))
			}
		}
	}
}

func importCustomReports(reports []keyfactor.ModelsCustomReportCreationRequest, kfClient *keyfactor.APIClient) {
	for _, report := range reports {
		_, httpResp, reqErr := kfClient.ReportsApi.ReportsCreateCustomReport(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).Request(report).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
		name, _ := json.Marshal(report.DisplayName)
		if reqErr != nil {
			fmt.Printf("%s Error! Unable to create custom report %s - %s%s\n", ColorRed, string(name), parseError(httpResp.Body), ColorWhite)
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
			fmt.Printf("%s Error! Unable to create security role %s - %s%s\n", ColorRed, string(name), reqErr, ColorWhite)
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
