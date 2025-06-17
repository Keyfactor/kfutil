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
	"fmt"

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

var migratePamCmd = &cobra.Command{
	Use:   "pam",
	Short: "Migrate existing PAM Provider usage to a new PAM Provider",
	Long:  "Migrate existing PAM Provider usage to a new PAM Provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := true

		// load specified flags
		migrateFrom, _ := cmd.Flags().GetString("from")       // defined pam provider
		migrateTo, _ := cmd.Flags().GetString("to")           // target pam provider typefffffff
		appendName, _ := cmd.Flags().GetString("append-name") // text to append to <<FROM>> name
		// TODO: define stores flag to pass in GUIDs of stores to migrate

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

		log.Info().Msg("looking up usage of <<from>> PAM Provider")

		log.Debug().Msg("call: PAMProviderGetPamProviders()")
		listPamProvidersInUse, httpResponse, rErr := sdkClient.PAMProviderApi.PAMProviderGetPamProviders(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			PqQueryString(fmt.Sprintf("name -eq \"%s\"", migrateFrom)).
			Execute()
		log.Debug().Msg("returned: PAMProviderGetPamProviders()")

		if rErr != nil {
			log.Error().Err(rErr).Send()
			return returnHttpErr(httpResponse, rErr)
		}

		jobject, _ := json.MarshalIndent(listPamProvidersInUse, "", "    ")
		fmt.Println(string(jobject))

		// TODO: ensure only 1 returned PAM Provider definition

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
		for _, pamType := range pamTypes {
			if *pamType.Id == *listPamProvidersInUse[0].ProviderType.Id {
				// TODO: remove debugging
				jobject, _ := json.MarshalIndent(pamType, "", "    ")
				fmt.Println(string(jobject))
				jobject, _ = json.MarshalIndent(pamType.AdditionalProperties["Parameters"], "", "    ")
				fmt.Println(string(jobject))
				// TODO: check typing, have to access "Parameters" instead of ProviderTypeParams
				for _, pamParamType := range pamType.AdditionalProperties["Parameters"].([]interface{}) {
					jobject, _ := json.MarshalIndent(pamParamType, "", "    ")
					fmt.Println(string(jobject))
					if pamParamType.(map[string]interface{})["InstanceLevel"].(bool) {
						// found definition of an instance level param for the type in question
						// create key in map for the field name
						inUsePamParamValues[pamParamType.(map[string]interface{})["Name"].(string)] = map[string]string{}
						fmt.Println("made it!")
					}
				}
			}
		}
		jobject, _ = json.MarshalIndent(inUsePamParamValues, "", "    ")
		fmt.Println(string(jobject))

		// step through list of every defined param value
		// record unique GUIDs of every param value on InstanceLevel : true
		// don't count InstanceLevel : false because those are Secret (DataType:2) instances for the Provider itself, not actual usages
		for _, pamParam := range listPamProvidersInUse[0].ProviderTypeParamValues {
			jobject, _ = json.MarshalIndent(pamParam, "", "    ")
			fmt.Println(string(jobject))
			if *pamParam.ProviderTypeParam.InstanceLevel {
				fieldName := *pamParam.ProviderTypeParam.Name
				usageGuid := *pamParam.InstanceGuid
				inUsePamParamValues[fieldName][usageGuid] = *pamParam.Value
			}
		}
		jobject, _ = json.MarshalIndent(inUsePamParamValues, "", "    ")
		fmt.Println(string(jobject))

		return nil

		// TODO: make sure every field has the same number of GUIDs tracked
		// tally GUID count for logging

		// log.Info().Msgf("Found %d PAM Provider usages of Provider %s",)

		// GET all PAM Types
		// select array entry with matching Name field of <<TO>>
		// mark GUID ID for pam type
		// mark integer IDs for each Parameter type

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

		// foreach store GUID pass in as a parameter-----
		// GET Store by GUID (instance GUID matches Store Id GUID)
		// output some store info to confirm

		// parse Properties field into interactable object

		// foreach property key (properties is an object not an array)
		// if value is an object, and object has an InstanceGuid
		// property object is a match for a secret
		// instead, can check if there is a ProviderId set, and if that
		// matches integer id of original Provider <<FROM>>

		// update property object
		// foreach ProviderTypeParameterValues
		// where ProviderTypeParam.Name = first map key (map is map[fieldname]map[InstanceGuid]value)
		// create new PAM value for this secret
		// json object:
		// value: {
		// provider: integer id of new provider
		// Parameters: {
		// fieldname: new value
		// }}
		//
		// leave all other fields untouched
		// IMPORTANT: other secret fields need to match value:{secretvalue:"*****" or secretvalue:null}

		// marshal json back to string for Properties field
		// make sure quotes are escaped
		// submit PUT for updating Store definition
	},
}

func init() {
	var from string
	var to string
	var appendName string

	RootCmd.AddCommand(migrateCmd)

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

	migratePamCmd.MarkFlagRequired("from")
	migratePamCmd.MarkFlagRequired("to")
	migratePamCmd.MarkFlagRequired("append-name")
}
