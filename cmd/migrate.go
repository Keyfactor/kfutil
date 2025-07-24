// Copyright 2025 Keyfactor
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
	"errors"
	"fmt"
	"strings"

	// "github.com/Keyfactor/keyfactor-go-client-sdk/v24/api/keyfactor/v2"
	"github.com/Keyfactor/keyfactor-go-client-sdk/v2/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Keyfactor Migration Tools.",
	Long: `Migrating to new Types and Extension implementations in Keyfactor is possible but not currently automated
	in the platform. This tool aims to assist in performing the necessary steps to migrate, in limited scenarios,
	to new Extension implementations that have definitions that differ from prior releases.`,
}

var migrateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check usage of a feature to migrate. Currently only PAM is supported.",
	Long:  "Check usage of a feature to migrate. Currently only PAM is supported",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := true

		// load specified flags
		fromCheck, _ := cmd.Flags().GetString("from") // name of entity, e.g. PAM Provider
		pamCheck, _ := cmd.Flags().GetBool("pam-usage")

		if pamCheck == false {
			return errors.New("Flag --pam-usage was not specified, but this is the only currently supported use case.")
		}

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Str("from", fromCheck).
			Bool("pam-usage", pamCheck).
			Msg("migrate PAM Provider")

		sdkClient, err := initGenClient(false)
		if err != nil {
			return err
		}

		// get all secret GUIDs for PAM Provider
		found, pamProvider, err := getExistingPamProvider(sdkClient, fromCheck)

		activePamSecretGuids := map[string]bool{}
		for _, param := range pamProvider.ProviderTypeParamValues {
			if param.InstanceGuid != nil {
				// enter every instance guid as a key with value true
				// represents an active Secret being managed in this pam provider
				// the same key will be set multiple times for each parameter for a particular Secret, but this should be no issue
				activePamSecretGuids[*param.InstanceGuid] = true
			}
		}

		if err != nil {
			log.Error().Err(err).Send()
			return err
		}

		if found == false {
			return errors.New("Named entity in 'from' argument was not found, no check can be run.")
		}

		legacyClient, err := initClient(false)
		if err != nil {
			return err
		}

		// get all certificate stores
		certStoreList, err := legacyClient.ListCertificateStores(nil)

		if err != nil {
			log.Error().Err(err).Send()
			return err
		}

		certStoreGuids := map[string]bool{}
		// loop through every found certificate store
		for _, store := range *certStoreList {
			// get properties field, as this will contain the Secret GUID for one of our active Instances if the PAM provider is in use
			storeProperties := store.PropertiesString

			// loop through all found Instance GUIDs of the PAM Provider
			// if the GUID is present in the Properties field, add this Store ID to the list to return
			for instanceGuid, _ := range activePamSecretGuids {
				if strings.Contains(storeProperties, instanceGuid) {
					certStoreGuids[store.Id] = true
				}
			}
		}

		// print out list of Cert Store GUIDs
		fmt.Println("\nThe following Cert Store Ids are using the PAM Provider with name '" + fromCheck + "'\n")
		for storeId, _ := range certStoreGuids {
			fmt.Println(storeId)
		}

		return nil
	},
}

var migratePamCmd = &cobra.Command{
	Use:   "pam",
	Short: "Migrate existing PAM Provider usage to a new PAM Provider",
	Long:  "Migrate existing PAM Provider usage to a new PAM Provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := true

		// load specified flags
		migrateFrom, _ := cmd.Flags().GetString("from")       // defined pam provider
		migrateTo, _ := cmd.Flags().GetString("to")           // target pam provider type
		appendName, _ := cmd.Flags().GetString("append-name") // text to append to <<FROM>> name
		storeUsingPam, _ := cmd.Flags().GetString("store")

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// <<TO>> Pam Type must be CyberArk-SdkCredentialProvider

		// Log flags
		log.Info().Str("from", migrateFrom).
			Str("to", migrateTo).
			Str("append-name", appendName).
			Msg("migrate PAM Provider")

		// Authenticate
		sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// TODO: assign error and check
		legacyClient, _ := initClient(false)

		found, fromPamProvider, processedError := getExistingPamProvider(sdkClient, migrateFrom)

		if processedError != nil {
			return processedError
		}

		if found == false {
			return errors.New("Original PAM Provider to migrate from was not found by Name")
		}

		// get PAM Type definition for PAM Provider migrating <<FROM>>
		log.Debug().Msg("call: PAMProviderGetPamProviders()")
		pamTypes, httpResponse, rErr := sdkClient.PAMProviderApi.PAMProviderGetPamProviderTypes(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		log.Debug().Msg("returned: PAMProviderGetPamProviders()")

		if rErr != nil {
			log.Error().Err(rErr).Send()
			return returnHttpErr(httpResponse, rErr)
		}

		// jobject, _ = json.MarshalIndent(pamTypes, "", "    ")
		// fmt.Println(string(jobject))

		// assess <<FROM>> source PAM Type to create map for storing existing data
		// this map has the first string key record the parameter field name
		// the inner map tracks PAM instance GUIDs to that instances value for the field
		// map[fieldname] -> map[InstanceGuid] = set value
		inUsePamParamValues := map[string]map[string]string{}
		fromProviderLevelParamValues := map[string]string{}
		var fromPamType keyfactor.CSSCMSDataModelModelsProviderType
		var toPamType keyfactor.CSSCMSDataModelModelsProviderType
		for _, pamType := range pamTypes {
			if *pamType.Id == *fromPamProvider.ProviderType.Id {
				fromPamType = pamType

				if debugFlag {
					fmt.Println("vvv FROM TYPE vvv")
					jobject, _ := json.MarshalIndent(pamType, "", "    ")
					fmt.Println(string(jobject))
					jobject, _ = json.MarshalIndent(pamType.AdditionalProperties["Parameters"], "", "    ")
					fmt.Println(string(jobject))
					fmt.Println("^^^ FROM TYPE ^^^")
				}
			}

			if *pamType.Name == migrateTo {
				toPamType = pamType

				if debugFlag {
					fmt.Println("vvv TO TYPE vvv")
					jobject, _ := json.MarshalIndent(pamType, "", "    ")
					fmt.Println(string(jobject))
					jobject, _ = json.MarshalIndent(pamType.AdditionalProperties["Parameters"], "", "    ")
					fmt.Println(string(jobject))
					fmt.Println("^^^ TO TYPE ^^^")
				}
			}
		}

		// TODO: check typing, have to access "Parameters" instead of ProviderTypeParams
		for _, pamParamType := range fromPamType.AdditionalProperties["Parameters"].([]interface{}) {
			if pamParamType.(map[string]interface{})["InstanceLevel"].(bool) {
				// found definition of an instance level param for the type in question
				// create key in map for the field name
				inUsePamParamValues[pamParamType.(map[string]interface{})["Name"].(string)] = map[string]string{}
			}
		}

		if debugFlag {
			fmt.Println("vvv IN USE PAM PROVIDER PARAMETER INSTANCES vvv")
			jobject, _ := json.MarshalIndent(inUsePamParamValues, "", "    ")
			fmt.Println(string(jobject))
			fmt.Println("^^^ IN USE PAM PROVIDER PARAMETER INSTANCES ^^^")
		}

		// step through list of every defined param value
		// record unique GUIDs of every param value on InstanceLevel : true
		// InstanceLevel : true is for per-secret fields
		// InstanceLevel : false is provider level secrets - these are also recorded for migration
		for _, pamParam := range fromPamProvider.ProviderTypeParamValues {
			// jobject, _ = json.MarshalIndent(pamParam, "", "    ")
			// fmt.Println(string(jobject))
			fieldName := *pamParam.ProviderTypeParam.Name
			usageGuid := *pamParam.InstanceGuid
			if *pamParam.ProviderTypeParam.InstanceLevel {
				inUsePamParamValues[fieldName][usageGuid] = *pamParam.Value
			} else {
				fromProviderLevelParamValues[fieldName] = *pamParam.Value
			}
		}

		if debugFlag {
			fmt.Println("vvv IN USE PAM PROVIDER PARAMETER VALUES vvv")
			jobject, _ := json.MarshalIndent(inUsePamParamValues, "", "    ")
			fmt.Println(string(jobject))
			fmt.Println("^^^ IN USE PAM PROVIDER PARAMETER VALUES ^^^")
		}

		// GET all PAM Types
		// select array entry with matching Name field of <<TO>>
		// mark GUID ID for pam type
		// mark integer IDs for each Parameter type

		var migrationTargetPamProvider keyfactor.CSSCMSDataModelModelsProvider

		// check if target PAM Provider already exists
		found, migrationTargetPamProvider, processedError = getExistingPamProvider(sdkClient, fromPamProvider.Name+appendName)

		if processedError != nil {
			return processedError
		}

		// create PAM Provider if it does not exist already
		if found == false {
			migrationTargetPamProvider, processedError = createMigrationTargetPamProvider(sdkClient, fromPamProvider, fromPamType, toPamType, appendName)

			if processedError != nil {
				return processedError
			}
		}

		// foreach store GUID pass in as a parameter-----
		// GET Store by GUID (instance GUID matches Store Id GUID)
		// output some store info to confirm

		// TODO: use updated client when API endpoint available
		certStore, rErr := legacyClient.GetCertificateStoreByID(storeUsingPam)
		if rErr != nil {
			log.Error().Err(rErr).Send()
			return rErr
		}

		if debugFlag {
			fmt.Println("vvv FOUND CERT STORE vvv")
			jobject, _ := json.MarshalIndent(certStore, "", "    ")
			fmt.Println(string(jobject))
			fmt.Println("^^^ FOUND CERT STORE ^^^")

			fmt.Println("vvv CERT STORE PROPERTIES vvv")
			jobject, _ = json.MarshalIndent(certStore.Properties, "", "    ")
			fmt.Println(string(jobject))
			fmt.Println("^^^ CERT STORE PROPERTIES ^^^")
		}

		// foreach property key (properties is an object not an array)
		// if value is an object, and object has an InstanceGuid
		// property object is a match for a secret
		// instead, can check if there is a ProviderId set, and if that
		// matches integer id of original Provider <<FROM>>

		for propName, prop := range certStore.Properties {
			propSecret, isSecret := prop.(map[string]interface{})
			if isSecret {
				formattedSecret := map[string]map[string]interface{}{
					"Value": map[string]interface{}{},
				}
				isManaged := propSecret["IsManaged"].(bool)
				if isManaged { // managed secret, i.e. PAM Provider in use

					// check if Pam Secret is using our migrating provider
					if *fromPamProvider.Id == int32(propSecret["ProviderId"].(float64)) {
						// Pam Secret that Needs to be migrated
						formattedSecret["Value"] = buildMigratedPamSecret(propSecret, fromProviderLevelParamValues, *migrationTargetPamProvider.Id)
					} else {
						// reformat to required POST format for properties
						formattedSecret["Value"] = reformatPamSecretForPost(propSecret)
					}
				} else {
					// non-managed secret i.e. a KF-encrypted secret, or no value
					// still needs to be reformatted to required POST format
					formattedSecret["Value"] = map[string]interface{}{
						"SecretValue": propSecret["Value"],
					}
				}

				// update Properties object with newly formatted secret, compliant with POST requirements
				certStore.Properties[propName] = formattedSecret
			}
		}

		if debugFlag {
			fmt.Println("vvv SECRETS REFORMATTED vvv")
			jobject, _ := json.MarshalIndent(certStore.Properties, "", "    ")
			fmt.Println(string(jobject))
			fmt.Println("^^^ SECRETS REFORMATTED ^^^")
		}

		// update property object
		// set required fields, and new Properties
		updateStoreArgs := api.UpdateStoreFctArgs{
			Id:                      certStore.Id,
			ClientMachine:           certStore.ClientMachine,
			StorePath:               certStore.StorePath,
			AgentId:                 certStore.AgentId,
			Properties:              certStore.Properties,
			Password:                &certStore.Password, // TODO: secret field, needs to be processed the same as other secret fields
			InventorySchedule:       &certStore.InventorySchedule,
			CertStoreInventoryJobId: &certStore.CertStoreInventoryJobId,
		}

		// TODO: use updated client when API endpoint available
		updatedStore, rErr := legacyClient.UpdateStore(&updateStoreArgs)

		if rErr != nil {
			log.Error().Err(rErr).Send()
			return rErr
		}

		fmt.Println("vvv UPDATED STORE vvv")
		jobject, _ := json.MarshalIndent(updatedStore, "", "    ")
		fmt.Println(string(jobject))
		fmt.Println("^^^ UPDATED STORE ^^^")

		return nil
	},
}

func getExistingPamProvider(sdkClient *keyfactor.APIClient, name string) (bool, keyfactor.CSSCMSDataModelModelsProvider, error) {
	var pamProvider keyfactor.CSSCMSDataModelModelsProvider

	logMsg := fmt.Sprintf("Looking up usage of PAM Provider with name %s", name)
	log.Debug().Msg(logMsg)
	fmt.Println(logMsg)

	foundProvider, httpResponse, err := sdkClient.PAMProviderApi.PAMProviderGetPamProviders(context.Background()).
		XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
		PqQueryString(fmt.Sprintf("name -eq \"%s\"", name)).
		Execute()

	logMsg = fmt.Sprintf("Reading response for PAM Provider with name %s", name)
	log.Debug().Msg(logMsg)
	fmt.Println(logMsg)

	if err != nil {
		log.Error().Err(err).Send()
		return false, pamProvider, returnHttpErr(httpResponse, err)
	}

	if len(foundProvider) > 1 {
		logMsg = "More than one PAM Provider returned for the same name. This is not supported behavior."
		log.Error().Msg(logMsg)
		return false, pamProvider, errors.New(logMsg)
	}

	if len(foundProvider) == 0 {
		logMsg = "No PAM Provider was found with the given name."
		log.Warn().Msg(logMsg)
		return false, pamProvider, nil
	}

	return true, foundProvider[0], nil
}

func createMigrationTargetPamProvider(sdkClient *keyfactor.APIClient, fromPamProvider keyfactor.CSSCMSDataModelModelsProvider, fromPamType keyfactor.CSSCMSDataModelModelsProviderType, toPamType keyfactor.CSSCMSDataModelModelsProviderType, appendName string) (keyfactor.CSSCMSDataModelModelsProvider, error) {
	fmt.Println("creating new Provider of migration target PAM Type")
	var migrationPamProvider keyfactor.CSSCMSDataModelModelsProvider
	migrationPamProvider.Name = fromPamProvider.Name + appendName
	migrationPamProvider.ProviderType = keyfactor.CSSCMSDataModelModelsProviderType{
		Id: toPamType.Id,
	}
	var onevalue int32 = 1
	migrationPamProvider.Area = &onevalue
	migrationPamProvider.SecuredAreaId = nil

	// need to init AdditionalProperties map when setting value
	migrationPamProvider.AdditionalProperties = map[string]interface{}{
		"Remote": false, // this property is not on the model for some reason
	}

	fmt.Println("getting migration target PAM Type parameter definitions, InstanceLevel : false")
	// TODO: check typing, have to access "Parameters" instead of ProviderTypeParams
	for _, pamParamType := range fromPamType.AdditionalProperties["Parameters"].([]interface{}) {
		if !pamParamType.(map[string]interface{})["InstanceLevel"].(bool) {
			// found a provider level parameter
			// need to find the value to map over
			// then create an object with that value and TypeParam settings
			paramName := pamParamType.(map[string]interface{})["Name"].(string)
			paramValue := selectProviderParamValue(paramName, fromPamProvider.ProviderTypeParamValues)
			paramTypeId := selectProviderTypeParamId(paramName, toPamType.AdditionalProperties["Parameters"].([]interface{}))
			falsevalue := false
			providerLevelParameter := keyfactor.CSSCMSDataModelModelsPamProviderTypeParamValue{
				Value: &paramValue,
				ProviderTypeParam: &keyfactor.CSSCMSDataModelModelsProviderTypeParam{
					Id:            &paramTypeId,
					Name:          &paramName,
					InstanceLevel: &falsevalue,
				},
			}

			// append filled out provider type parameter object, which contains the Provider-level parameter values
			// migrationPamProvider.ProviderTypeParamValues = append(migrationPamProvider.ProviderTypeParamValues, providerLevelParameter)

			// TODO: need to explicit filter for CyberArk expected params, i.e. not map over Safe
			// this needs to be done programatically for other provider types
			if paramName == "AppId" {
				migrationPamProvider.ProviderTypeParamValues = append(migrationPamProvider.ProviderTypeParamValues, providerLevelParameter)
			}
		}
	}

	msg := "Created new PAM Provider definition to be created."
	fmt.Println(msg)
	log.Info().Msg(msg)

	if debugFlag {
		fmt.Println("vvv PAM PROVIDER TO BE CREATED vvv")
		jobject, _ := json.MarshalIndent(migrationPamProvider, "", "    ")
		fmt.Println(string(jobject))
		fmt.Println("^^^ PAM PROVIDER TO BE CREATED ^^^")
	}

	// POST new PAM Provider
	// create new PAM Instance of designated <<TO>> type
	// set area = 1 or previous value
	// name = old name plus append parameter
	// providertype.id = Type GUID
	// for all provider level values:
	// set value to migrating value
	// null for instanceid, instanceguid
	// providertypeparam should be set to all matching values from GET TYPES
	// ignoring datatype

	createdPamProvider, httpResponse, rErr := sdkClient.PAMProviderApi.PAMProviderCreatePamProvider(context.Background()).
		Provider(migrationPamProvider).
		XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
		Execute()

	if rErr != nil {
		log.Error().Err(rErr).Send()
		return *createdPamProvider, returnHttpErr(httpResponse, rErr)
	}

	if debugFlag {
		fmt.Println("vvv CREATED MIGRATION PAM PROVIDER vvv")
		jobject, _ := json.MarshalIndent(createdPamProvider, "", "    ")
		fmt.Println(string(jobject))
		fmt.Println("^^^ CREATED MIGRATION PAM PROVIDER ^^^")
	}

	return *createdPamProvider, nil
}

func selectProviderParamValue(name string, providerParameters []keyfactor.CSSCMSDataModelModelsPamProviderTypeParamValue) string {
	for _, parameter := range providerParameters {
		if name == *parameter.ProviderTypeParam.Name {
			return *parameter.Value
		}
	}
	return "NOTFOUND" // TODO: throw error when not found
}

// TODO(check, remove): might need to select DisplayName as well, if required for input in API
func selectProviderTypeParamId(name string, pamTypeParameterDefinitions []interface{}) int32 {
	for _, parameterDefinition := range pamTypeParameterDefinitions {
		if name == parameterDefinition.(map[string]interface{})["Name"].(string) {
			return int32(parameterDefinition.(map[string]interface{})["Id"].(float64)) // interface returns value as float64, needs to be cast to int32
		}
	}

	return -1 // TODO: throw error when not found
}

func reformatPamSecretForPost(secretProp map[string]interface{}) map[string]interface{} {
	reformatted := map[string]interface{}{
		"Provider": secretProp["ProviderId"],
	}

	providerParams := secretProp["ProviderTypeParameterValues"].([]interface{})
	reformattedParams := map[string]string{}

	for _, param := range providerParams {
		providerTypeParam := param.(map[string]interface{})["ProviderTypeParam"].(map[string]interface{})
		name := providerTypeParam["Name"].(string)
		value := param.(map[string]interface{})["Value"].(string)
		reformattedParams[name] = value
	}

	reformatted["Parameters"] = reformattedParams
	return reformatted
}

// Inputs:
// secretProp: existing Pam config for property
// migratingValues: map of existing values for matched GUID of this field
// fromProvider: previous provider, to get type level values
// pamProvider: newly created Pam Provider for the migration, with Provider Id
func buildMigratedPamSecret(secretProp map[string]interface{}, fromProviderLevelValues map[string]string, providerId int32) map[string]interface{} {
	migrated := map[string]interface{}{
		"Provider": providerId,
	}

	providerParams := secretProp["ProviderTypeParameterValues"].([]interface{})
	reformattedParams := map[string]string{}

	// NOTE: this is making an assumption that the property names have not changed
	// and should be mapped back to the same value
	for _, param := range providerParams {
		providerTypeParam := param.(map[string]interface{})["ProviderTypeParam"].(map[string]interface{})
		name := providerTypeParam["Name"].(string)
		value := param.(map[string]interface{})["Value"].(string)
		reformattedParams[name] = value
	}

	// TODO: this logic needs to not be hard-coded, and evaluated for actual migrations other than legacy CyberArk
	reformattedParams["Safe"] = fromProviderLevelValues["Safe"]

	migrated["Parameters"] = reformattedParams

	return migrated
}

func init() {
	RootCmd.AddCommand(migrateCmd)

	// migrate check
	var pamCheck bool
	var fromCheck string

	migrateCmd.AddCommand(migrateCheckCmd)

	migrateCheckCmd.Flags().BoolVar(
		&pamCheck,
		"pam-usage",
		true,
		"Specify this flag to check usage of a PAM Provider named with the 'from' argument. Returns a list of Certificate Store GUIDs using that provider.",
	)

	migrateCheckCmd.Flags().StringVarP(
		&fromCheck,
		"from",
		"f",
		"",
		"The name of the KF entity to search for usage of. Behavior will be different depending on type of check specified.",
	)

	migrateCheckCmd.MarkFlagRequired("from")

	// migrate pam
	var from string
	var to string
	var appendName string
	var store string

	migrateCmd.AddCommand(migratePamCmd)

	migratePamCmd.Flags().StringVarP(
		&from,
		"from",
		"f",
		"",
		"Name of the defined PAM Provider to migrate to a new type",
	)

	migratePamCmd.Flags().StringVarP(
		&to,
		"to",
		"t",
		"",
		"Name of the PAM Provider Type to migrate to",
	)

	migratePamCmd.Flags().StringVarP(
		&appendName,
		"append-name",
		"a",
		"",
		"Text to append to current PAM Provider Name in newly created Provider",
	)

	migratePamCmd.Flags().StringVarP(
		&store,
		"store",
		"s",
		"",
		"GUID of a Certificate Store, using a PAM Provider that should be migrated",
	)

	migratePamCmd.MarkFlagRequired("from")
	migratePamCmd.MarkFlagRequired("to")
	migratePamCmd.MarkFlagRequired("append-name")
	migratePamCmd.MarkFlagRequired("store")
}
