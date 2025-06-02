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
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// type JSONImportableObject interface {
// 	keyfactor.KeyfactorApiPAMProviderTypeCreateRequest |
// 		keyfactor.CSSCMSDataModelModelsProvider
// }

// const (
// 	convertResponseMsg = "Converting PAM Provider response to JSON"
// )

type PAMParameterValue struct {
	GUID  string
	Value string
}

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
		migrateFrom, _ := cmd.Flags().GetString("from")
		migrateTo, _ := cmd.Flags().GetString("to")
		appendName, _ := cmd.Flags().GetString("append-name")

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

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

		// assess <<FROM>> source PAM Type to create map for storing existing data
		// this map has the first string key record the parameter field name
		// the inner map tracks PAM instance GUIDs to that instances value for the field
		inUsePamParamValues := map[string]map[string]string{}
		for _, pamType := range pamTypes {
			if pamType.Id == listPamProvidersInUse[0].ProviderType.Id {
				for _, pamParamType := range pamType.ProviderTypeParams {
					if *pamParamType.InstanceLevel {
						// found definition of an instance level param for the type in question
						// create key in map for the field name
						inUsePamParamValues[*pamParamType.Name] = map[string]string{}
					}
				}
			}
		}

		// step through list of every defined param value
		// record unique GUIDs of every param value on InstanceLevel : true
		// don't count InstanceLevel : false because those are Secret (DataType:2) instances for the Provider itself, not actual usages
		for _, pamParam := range listPamProvidersInUse[0].ProviderTypeParamValues {
			if *pamParam.ProviderTypeParam.InstanceLevel {
				fieldName := *pamParam.ProviderTypeParam.Name
				usageGuid := *pamParam.InstanceGuid
				inUsePamParamValues[fieldName][usageGuid] = *pamParam.Value
			}
		}

		// TODO: make sure every field has the same number of GUIDs tracked
		// tally GUID count for logging

		// log.Info().Msgf("Found %d PAM Provider usages of Provider %s",)

		// implement migration logic
		// create new PAM Instance of designated <<TO>> type
	},
}
