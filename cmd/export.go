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
	"os"
	"strconv"

	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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
	ID                      *int32                             `json:"-"`
	Scheduled               *int32                             `json:"Scheduled,omitempty"`
	DisplayName             *string                            `json:"DisplayName,omitempty"`
	Description             *string                            `json:"Description,omitempty"`
	ReportPath              *string                            `json:"ReportPath,omitempty"`
	VersionNumber           *string                            `json:"VersionNumber,omitempty"`
	Categories              *string                            `json:"Categories,omitempty"`
	ShortName               *string                            `json:"ShortName,omitempty"`
	InNavigator             *bool                              `json:"InNavigator,omitempty"`
	Favorite                *bool                              `json:"Favorite,omitempty"`
	RemoveDuplicates        *bool                              `json:"RemoveDuplicates,omitempty"`
	UsesCollection          *bool                              `json:"UsesCollection,omitempty"`
	ReportParameter         []keyfactor.ModelsReportParameters `json:"ReportParameter,omitempty"`
	Schedules               []keyfactor.ModelsReportSchedule   `json:"Schedules,omitempty"`
	AcceptedScheduleFormats []string                           `json:"AcceptedScheduleFormats,omitempty"`
}

type exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest struct {
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
	Collections         []keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest `json:"Collections"`
	MetadataFields      []keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest                  `json:"MetadataFields"`
	ExpirationAlerts    []keyfactor.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest           `json:"ExpirationAlerts"`
	IssuedCertAlerts    []keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest                   `json:"IssuedCertAlerts"`
	DeniedCertAlerts    []keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest                   `json:"DeniedCertAlerts"`
	PendingCertAlerts   []keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest                 `json:"PendingCertAlerts"`
	Networks            []keyfactor.KeyfactorApiModelsSslCreateNetworkRequest                                  `json:"Networks"`
	WorkflowDefinitions []exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest                             `json:"WorkflowDefinitions"`
	BuiltInReports      []exportModelsReport                                                                   `json:"BuiltInReports"`
	CustomReports       []keyfactor.ModelsCustomReportCreationRequest                                          `json:"CustomReports"`
	SecurityRoles       []api.CreateSecurityRoleArg                                                            `json:"SecurityRoles"`
}

func exportToJSON(out outJson, exportPath string) error {
	mOut, jErr := json.MarshalIndent(out, "", "    ")
	if jErr != nil {
		fmt.Printf("Error processing JSON object. %s\n", jErr)
		//log.Fatalf("[ERROR]: %s", jErr)
		log.Error().Err(jErr)
		return jErr
	}
	wErr := os.WriteFile(exportPath, mOut, 0666)
	if wErr != nil {
		fmt.Printf("Error writing files to %s: %s\n", exportPath, wErr)
		//log.Fatalf("[ERROR]: %s", wErr)
		log.Error().Err(wErr)
		return wErr
	} else {
		fmt.Printf("Content successfully written to %s", exportPath)
		return nil
	}
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Keyfactor instance export utilities.",
	Long:  `A collection of APIs and utilities for exporting Keyfactor instance data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug().Msgf("%s: exportCmd", DebugFuncEnter)
		isExperimental := true

		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		log.Info().Msg("Exporting data from Keyfactor instance")

		// initialize each entry as an empty list in the event it is not requested by the flags
		out := outJson{
			Collections:         []keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest{},
			MetadataFields:      []keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest{},
			ExpirationAlerts:    []keyfactor.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest{},
			IssuedCertAlerts:    []keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest{},
			DeniedCertAlerts:    []keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest{},
			PendingCertAlerts:   []keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest{},
			Networks:            []keyfactor.KeyfactorApiModelsSslCreateNetworkRequest{},
			WorkflowDefinitions: []exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest{},
			BuiltInReports:      []exportModelsReport{},
			CustomReports:       []keyfactor.ModelsCustomReportCreationRequest{},
			SecurityRoles:       []api.CreateSecurityRoleArg{},
		}

		log.Debug().Msgf("%s: createAuthConfigFromParams", DebugFuncCall)
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)

		if authConfig == nil {
			log.Error().Msg("auth config is nil, invalid client configuration")
			return fmt.Errorf(FailedAuthMsg)
		}

		exportPath := cmd.Flag("file").Value.String()
		log.Debug().Str("exportPath", exportPath).Msg("exportPath")

		log.Debug().Msgf("%s: initGenClient", DebugFuncCall)
		kfClient, clientErr := initGenClient(configFile, profile, noPrompt, authConfig, false)
		log.Debug().Msgf("%s: initClient", DebugFuncCall)
		oldkfClient, oldClientErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		if clientErr != nil {
			log.Error().Err(clientErr).Send()
			return clientErr
		} else if oldClientErr != nil {
			log.Error().Err(oldClientErr).Send()
			return oldClientErr
		}

		if cmd.Flag("all").Value.String() == "true" {
			log.Debug().Msgf("%s: getCollections", DebugFuncCall)
			out.Collections = getCollections(kfClient)

			log.Debug().Msgf("%s: getMetadata", DebugFuncCall)
			out.MetadataFields = getMetadata(kfClient)

			log.Debug().Msgf("%s: getExpirationAlerts", DebugFuncCall)
			out.ExpirationAlerts = getExpirationAlerts(kfClient)

			log.Debug().Msgf("%s: getIssuedAlerts", DebugFuncCall)
			out.IssuedCertAlerts = getIssuedAlerts(kfClient)

			log.Debug().Msgf("%s: getDeniedAlerts", DebugFuncCall)
			out.DeniedCertAlerts = getDeniedAlerts(kfClient)

			log.Debug().Msgf("%s: getPendingAlerts", DebugFuncCall)
			out.PendingCertAlerts = getPendingAlerts(kfClient)

			log.Debug().Msgf("%s: getSslNetworks", DebugFuncCall)
			out.Networks = getSslNetworks(kfClient)

			log.Debug().Msgf("%s: getWorkflowDefinitions", DebugFuncCall)
			out.WorkflowDefinitions = getWorkflowDefinitions(kfClient)

			log.Debug().Msgf("%s: getReports", DebugFuncCall)
			out.BuiltInReports, out.CustomReports = getReports(kfClient)

			log.Debug().Msgf("%s: getRoles", DebugFuncCall)
			out.SecurityRoles = getRoles(oldkfClient)
		} else {
			if cmd.Flag("collections").Value.String() == "true" {
				log.Debug().Msgf("%s: getCollections", DebugFuncCall)
				out.Collections = getCollections(kfClient)
			}
			if cmd.Flag("metadata").Value.String() == "true" {
				log.Debug().Msgf("%s: getMetadata", DebugFuncCall)
				out.MetadataFields = getMetadata(kfClient)
			}
			if cmd.Flag("expiration-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: getExpirationAlerts", DebugFuncCall)
				out.ExpirationAlerts = getExpirationAlerts(kfClient)
			}
			if cmd.Flag("issued-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: getIssuedAlerts", DebugFuncCall)
				out.IssuedCertAlerts = getIssuedAlerts(kfClient)
			}
			if cmd.Flag("denied-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: getDeniedAlerts", DebugFuncCall)
				out.DeniedCertAlerts = getDeniedAlerts(kfClient)
			}
			if cmd.Flag("pending-alerts").Value.String() == "true" {
				log.Debug().Msgf("%s: getPendingAlerts", DebugFuncCall)
				out.PendingCertAlerts = getPendingAlerts(kfClient)
			}
			if cmd.Flag("networks").Value.String() == "true" {
				log.Debug().Msgf("%s: getSslNetworks", DebugFuncCall)
				out.Networks = getSslNetworks(kfClient)
			}
			if cmd.Flag("workflow-definitions").Value.String() == "true" {
				log.Debug().Msgf("%s: getWorkflowDefinitions", DebugFuncCall)
				out.WorkflowDefinitions = getWorkflowDefinitions(kfClient)
			}
			if cmd.Flag("reports").Value.String() == "true" {
				log.Debug().Msgf("%s: getReports", DebugFuncCall)
				out.BuiltInReports, out.CustomReports = getReports(kfClient)
			}
			if cmd.Flag("security-roles").Value.String() == "true" {
				log.Debug().Msgf("%s: getRoles", DebugFuncCall)
				out.SecurityRoles = getRoles(oldkfClient)
			}
		}
		log.Debug().Msgf("%s: exportToJSON", DebugFuncCall)
		exportToJSON(out, exportPath)

		log.Debug().Msgf("%s: exportCmd", DebugFuncExit)
		log.Info().Msg("Export complete")
		return nil
	},
}

func getCollections(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest {
	log.Debug().Msgf("%s: getCollections", DebugFuncEnter)

	log.Debug().Msgf("%s: CertificateCollectionGetCollections", DebugFuncCall)
	collections, _, reqErr := kfClient.CertificateCollectionApi.CertificateCollectionGetCollections(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()

	if reqErr != nil {
		log.Error().Err(reqErr).Send()
		fmt.Printf("%s Error! Unable to get collections %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lCollectionReq []keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest
	for _, collection := range collections {
		log.Debug().Msgf("Marshalling collection %s", *collection.Name)
		cJson, jmErr := json.Marshal(collection)
		if jmErr != nil {
			if collection.Name != nil && collection.Id != nil {
				log.Error().Err(jmErr).Msgf("Error marshalling collection %s(%d)", *collection.Name, *collection.Id)
			}
			fmt.Printf("Error: %s\n", jmErr)
			continue
		}

		log.Debug().Msgf("Unmarshalling collection %s", *collection.Name)
		var collectionReq keyfactor.KeyfactorApiModelsCertificateCollectionsCertificateCollectionCreateRequest
		jErr := json.Unmarshal(cJson, &collectionReq)
		if jErr != nil {
			log.Error().Err(jErr).Send()
			fmt.Printf("Error: %s\n", jErr)
		}
		collectionReq.Query = collection.Content
		collectionReq.Id = nil

		log.Debug().Msgf("Appending collection %s", *collection.Name)
		lCollectionReq = append(lCollectionReq, collectionReq)
	}
	log.Debug().Msgf("%s: getCollections", DebugFuncExit)
	return lCollectionReq
}

func getMetadata(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest {
	log.Debug().Msgf("%s: getMetadata", DebugFuncEnter)

	log.Debug().Msgf("%s: MetadataFieldGetAllMetadataFields", DebugFuncCall)
	metadata, _, reqErr := kfClient.MetadataFieldApi.MetadataFieldGetAllMetadataFields(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if reqErr != nil {
		log.Error().Err(reqErr).Send()
		fmt.Printf("%s Error! Unable to get metadata %s%s\n", ColorRed, reqErr, ColorWhite)
		return nil
	}

	var lMetadataReq []keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest
	for _, metadataItem := range metadata {
		mName := ""
		if metadataItem.Name != nil {
			mName = *metadataItem.Name
		} else if metadataItem.Id != nil {
			mName = fmt.Sprintf("%d", *metadataItem.Id)
		}
		log.Debug().Str("mName", mName).Msg("Marshalling metadata")
		mJson, jmErr := json.Marshal(metadataItem)
		if jmErr != nil {
			log.Error().Err(jmErr).Send()
			fmt.Printf("Error: %s\n", jmErr)
			continue
		}

		log.Debug().Msgf("Unmarshalling metadata '%s'", mName)
		var metadataReq keyfactor.KeyfactorApiModelsMetadataFieldMetadataFieldCreateRequest
		jErr := json.Unmarshal(mJson, &metadataReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
			continue
		}
		metadataItem.Id = nil

		log.Debug().Msgf("Appending metadata '%s'", mName)
		lMetadataReq = append(lMetadataReq, metadataReq)
	}
	return lMetadataReq
}

func getExpirationAlerts(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest {

	alerts, _, reqErr := kfClient.ExpirationAlertApi.ExpirationAlertGetExpirationAlerts(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get expiration alerts %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lAlertReq []keyfactor.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor.KeyfactorApiModelsAlertsExpirationExpirationAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			log.Error().Err(jErr).Send()
			return nil // todo: maybe return the error instead?
		}
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getIssuedAlerts(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest {

	alerts, _, reqErr := kfClient.IssuedAlertApi.IssuedAlertGetIssuedAlerts(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get issued cert alerts %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lAlertReq []keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor.KeyfactorApiModelsAlertsIssuedIssuedAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
			return nil // todo: maybe return the error instead?
		}
		alertReq.TemplateId = nil
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getDeniedAlerts(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest {

	alerts, _, reqErr := kfClient.DeniedAlertApi.DeniedAlertGetDeniedAlerts(
		context.Background(),
	).XKeyfactorRequestedWith(
		XKeyfactorRequestedWith,
	).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get denied cert alerts %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lAlertReq []keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor.KeyfactorApiModelsAlertsDeniedDeniedAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
			return nil // todo: maybe return the error instead?
		}
		alertReq.TemplateId = nil
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getPendingAlerts(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest {

	alerts, _, reqErr := kfClient.PendingAlertApi.PendingAlertGetPendingAlerts(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get pending cert alerts %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lAlertReq []keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest
	for _, alert := range alerts {
		mJson, _ := json.Marshal(alert)
		var alertReq keyfactor.KeyfactorApiModelsAlertsPendingPendingAlertCreationRequest
		jErr := json.Unmarshal(mJson, &alertReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
		}
		alertReq.TemplateId = nil
		lAlertReq = append(lAlertReq, alertReq)
	}
	return lAlertReq
}

func getSslNetworks(kfClient *keyfactor.APIClient) []keyfactor.KeyfactorApiModelsSslCreateNetworkRequest {

	networks, _, reqErr := kfClient.SslApi.
		SslGetNetworks(context.Background()).
		XKeyfactorRequestedWith(XKeyfactorRequestedWith).
		XKeyfactorApiVersion(XKeyfactorApiVersion).
		Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get SSL networks %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lNetworkReq []keyfactor.KeyfactorApiModelsSslCreateNetworkRequest
	for _, network := range networks {
		mJson, _ := json.Marshal(network)
		var networkReq keyfactor.KeyfactorApiModelsSslCreateNetworkRequest
		jErr := json.Unmarshal(mJson, &networkReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send()
			continue
		}
		lNetworkReq = append(lNetworkReq, networkReq)
	}
	return lNetworkReq
}

func getWorkflowDefinitions(kfClient *keyfactor.APIClient) []exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest {

	workflowDefs, _, reqErr := kfClient.WorkflowDefinitionApi.
		WorkflowDefinitionQuery(context.Background()).
		XKeyfactorRequestedWith(XKeyfactorRequestedWith).
		XKeyfactorApiVersion(XKeyfactorApiVersion).
		Execute()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get workflow definitions %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lWorkflowReq []exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest
	for _, workflowDef := range workflowDefs {
		mJson, mErr := json.Marshal(workflowDef)
		if mErr != nil {
			fmt.Printf("Error: %s\n", mErr)
			//log.Fatalf("Error: %s", mErr)
			log.Error().Err(mErr).Send() //todo: better error message?
			continue
		}
		var workflowReq exportKeyfactorAPIModelsWorkflowsDefinitionCreateRequest
		jErr := json.Unmarshal(mJson, &workflowReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send() //todo: better error message?
			continue
		}
		if workflowDef.Key != nil {
			key, convErr := strconv.ParseInt(*workflowDef.Key, 10, 64)
			if convErr != nil {
				fmt.Printf("Error: %s\n", convErr)
				//log.Fatalf("Error: %s", convErr)
				log.Error().Err(convErr).Send() //todo: better error message?
				continue
			}
			key32 := int32(key)
			template, _, tErr := kfClient.TemplateApi.
				TemplateGetTemplate(context.Background(), key32).
				XKeyfactorRequestedWith(XKeyfactorRequestedWith).
				XKeyfactorApiVersion(XKeyfactorApiVersion).
				Execute()
			if tErr != nil {
				log.Error().Err(tErr).Send() //todo: better error message?
				continue
			}
			workflowReq.KeyName = template.TemplateName
		}
		workflowReq.Key = nil
		lWorkflowReq = append(lWorkflowReq, workflowReq)
	}
	return lWorkflowReq
}

func getReports(kfClient *keyfactor.APIClient) ([]exportModelsReport, []keyfactor.ModelsCustomReportCreationRequest) {

	//Gets all built-in reports
	bReports, _, bErr := kfClient.ReportsApi.ReportsQueryReports(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if bErr != nil {
		fmt.Printf("%s Error! Unable to get built-in reports %s%s\n", ColorRed, bErr, ColorWhite)
	}
	var lbReportsReq []exportModelsReport
	for _, bReport := range bReports {
		mJson, mErr := json.Marshal(bReport)
		if mErr != nil {
			fmt.Printf("Error: %s\n", mErr)
			//log.Fatalf("Error: %s", mErr)
			log.Error().Err(mErr).Send() //todo: better error message?
			continue
		}
		var newbReport exportModelsReport
		jErr := json.Unmarshal(mJson, &newbReport)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send() //todo: better error message?
			continue
		}
		newbReport.ID = nil
		lbReportsReq = append(lbReportsReq, newbReport)
	}
	//Gets all custom reports
	cReports, _, cErr := kfClient.ReportsApi.ReportsQueryCustomReports(context.Background()).XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).Execute()
	if cErr != nil {
		fmt.Printf("%s Error! Unable to get custom reports %s%s\n", ColorRed, cErr, ColorWhite)
	}
	var lcReportReq []keyfactor.ModelsCustomReportCreationRequest
	for _, cReport := range cReports {
		mJson, _ := json.Marshal(cReport)
		var cReportReq keyfactor.ModelsCustomReportCreationRequest
		jErr := json.Unmarshal(mJson, &cReportReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send() //todo: better error message?
			continue
		}
		lcReportReq = append(lcReportReq, cReportReq)
	}
	return lbReportsReq, lcReportReq
}

func getRoles(kfClient *api.Client) []api.CreateSecurityRoleArg {
	roles, reqErr := kfClient.GetSecurityRoles()
	if reqErr != nil {
		fmt.Printf("%s Error! Unable to get security roles %s%s\n", ColorRed, reqErr, ColorWhite)
	}
	var lRoleReq []api.CreateSecurityRoleArg
	for _, role := range roles {
		mJson, _ := json.Marshal(role)
		var cRoleReq api.CreateSecurityRoleArg
		jErr := json.Unmarshal(mJson, &cRoleReq)
		if jErr != nil {
			fmt.Printf("Error: %s\n", jErr)
			//log.Fatalf("Error: %s", jErr)
			log.Error().Err(jErr).Send() //todo: better error message?
			continue
		}
		lRoleReq = append(lRoleReq, cRoleReq)
	}
	return lRoleReq
}

func init() {
	RootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportPath, "file", "f", "", "path to JSON output file with exported data")
	exportCmd.MarkFlagRequired("file")

	exportCmd.Flags().BoolVarP(&fAll, "all", "a", false, "export all exportable data to JSON file")
	exportCmd.Flags().Lookup("all").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fCollections, "collections", "c", false, "export collections to JSON file")
	exportCmd.Flags().Lookup("collections").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fMetadata, "metadata", "m", false, "export metadata to JSON file")
	exportCmd.Flags().Lookup("metadata").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(
		&fExpirationAlerts,
		"expiration-alerts",
		"e",
		false,
		"export expiration cert alerts to JSON file",
	)
	exportCmd.Flags().Lookup("expiration-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fIssuedAlerts, "issued-alerts", "i", false, "export issued cert alerts to JSON file")
	exportCmd.Flags().Lookup("issued-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fDeniedAlerts, "denied-alerts", "d", false, "export denied cert alerts to JSON file")
	exportCmd.Flags().Lookup("denied-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fPendingAlerts, "pending-alerts", "p", false, "export pending cert alerts to JSON file")
	exportCmd.Flags().Lookup("pending-alerts").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fNetworks, "networks", "n", false, "export SSL networks to JSON file")
	exportCmd.Flags().Lookup("networks").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(
		&fWorkflowDefinitions,
		"workflow-definitions",
		"w",
		false,
		"export workflow definitions to JSON file",
	)
	exportCmd.Flags().Lookup("workflow-definitions").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fReports, "reports", "r", false, "export reports to JSON file")
	exportCmd.Flags().Lookup("reports").NoOptDefVal = "true"
	exportCmd.Flags().BoolVarP(&fSecurityRoles, "security-roles", "s", false, "export security roles to JSON file")
	exportCmd.Flags().Lookup("security-roles").NoOptDefVal = "true"
}
