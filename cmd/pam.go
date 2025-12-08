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
	"encoding/json"
	"fmt"

	keyfactor "github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type JSONImportableObject interface {
	keyfactor.Provider |
		keyfactor.ProviderType |
		keyfactor.ProviderTypeCreateRequest
}

const (
	convertResponseMsg = "Converting PAM Provider response to JSON"
)

var pamCmd = &cobra.Command{
	Use:   "pam",
	Short: "Keyfactor PAM Provider APIs.",
	Long: `Privileged Access Management (PAM) functionality in Keyfactor Web APIs allows for configuration of third 
party PAM providers to secure certificate stores. The PAM component of the Keyfactor API includes methods necessary to 
programmatically create, delete, edit, and list PAM Providers.`,
}

var pamProvidersListCmd = &cobra.Command{
	Use:   "list",
	Short: "Returns a list of all the configured PAM providers.",
	Long:  "Returns a list of all the configured PAM providers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		// Specific flags

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Msg("list PAM Providers")

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		log.Debug().Msg("call: PAMProviderGetPamProviders()")
		pamProviders, err := kfClient.ListPAMProviders(nil)
		//pamProviders, httpResponse, err := sdkClient.PAMProviderApi.PAMProviderGetPamProviders(context.Background()).
		//	XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
		//	Execute()
		//log.Debug().Msg("returned: PAMProviderGetPamProviders()")
		//log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderGetPamProviders")
		if err != nil {
			log.Error().Err(err).Send()
			return err
		}

		log.Debug().Msg("Converting PAM Providers response to JSON")
		jsonString, mErr := json.Marshal(pamProviders)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().Str("output", string(jsonString)).
			Msg("successfully listed PAM providers")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamProvidersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific defined PAM Provider by ID.",
	Long:  "Get a specific defined PAM Provider by ID.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		// Specific flags
		pamProviderId, _ := cmd.Flags().GetInt32("id")
		pamProviderName, _ := cmd.Flags().GetString("name")

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		log.Info().Str("name", pamProviderName).
			Int32("id", pamProviderId).
			Msg("get PAM Provider")

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		var (
			pamProvider *keyfactor.ProviderResponseLegacy
			err         error
		)

		if pamProviderId == 0 && pamProviderName != "" {
			log.Debug().Str("name", pamProviderName).Msg("resolving PAM Provider ID from name")
			pamProvider, err = kfClient.GetPamProviderByName(pamProviderName)
			if err != nil {
				log.Error().Err(err).Str(
					"name",
					pamProviderName,
				).Msg("error listing PAM providers to resolve ID from name")
				return err
			}
		} else {
			pamProvider, err = kfClient.GetPAMProvider(int(pamProviderId))
			if err != nil {
				log.Error().Err(err).Int32("id", pamProviderId).Msg("error getting PAM provider")
				return err
			}
		}

		// CLI Logic
		//log.Debug().Msg("call: PAMProviderGetPamProvider()")
		//pamProvider, httpResponse, err := sdkClient.PAMProviderApi.PAMProviderGetPamProvider(
		//	context.Background(),
		//	pamProviderId,
		//).
		//	XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
		//	Execute()
		//log.Debug().Msg("returned: PAMProviderGetPamProvider()")
		//log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderGetPamProvider")
		//
		//if err != nil {
		//	log.Error().Err(err).Str("httpResponseCode", httpResponse.Status).Msg("error getting PAM provider")
		//	return err
		//}

		log.Debug().Msg(convertResponseMsg)
		jsonString, mErr := json.Marshal(pamProvider)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().Str("output", string(jsonString)).
			Msg("successfully retrieved PAM provider")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamProvidersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new PAM Provider, currently only supported from file.",
	Long:  "Create a new PAM Provider, currently only supported from file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		// Specific flags
		pamConfigFile, _ := cmd.Flags().GetString(FlagFromFile)

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Str("file", pamConfigFile).
			Msg("create PAM Provider from file")

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)

		//_, cmdResp, sErr := sdkClient.StatusApi.StatusGetEndpoints(context.Background()).Execute()
		//if sErr != nil {
		//	log.Error().Err(sErr).Msg("failed to get Keyfactor Command version")
		//} else {
		//	bug63171 := checkBug63171(cmdResp, "CREATE")
		//	if bug63171 != nil {
		//		return bug63171
		//	}
		//}

		if cErr != nil {
			return cErr
		}

		// CLI Logic
		var pamProvider *keyfactor.Provider
		log.Debug().Msg("call: GetTypeFromConfigFile()")
		pamProvider, err := GetTypeFromConfigFile(pamConfigFile, pamProvider)
		log.Debug().Msg("returned: GetTypeFromConfigFile()")
		if err != nil {
			log.Error().Err(err).
				Str("file", pamConfigFile).
				Msg("failed parsing PAM Provider config from file")
			return err
		}

		log.Debug().Msg("call: PAMProviderCreatePamProvider()")
		createRequest := keyfactor.ProviderCreateRequest{
			Name:                    pamProvider.Name,
			Remote:                  pamProvider.Remote,
			Area:                    pamProvider.Area,
			ProviderType:            pamProvider.ProviderType,
			ProviderTypeParamValues: pamProvider.ProviderTypeParamValues,
			SecuredAreaId:           pamProvider.SecuredAreaId,
		}

		createdPamProvider, cErr := kfClient.CreatePAMProvider(&createRequest)
		log.Debug().Msg("returned: PAMProviderCreatePamProvider()")
		if cErr != nil {
			// output response body
			log.Debug().Msg("Converting PAM Provider response body to string")
			return cErr
		}

		log.Debug().Msg(convertResponseMsg)
		jsonString, mErr := json.Marshal(createdPamProvider)
		if mErr != nil {
			log.Error().Err(mErr).Msg("invalid API response from Keyfactor Command")
			return mErr
		}
		log.Info().Str("output", string(jsonString)).Msg("successfully created PAM provider")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamProvidersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates an existing PAM Provider, currently only supported from file.",
	Long:  "Updates an existing PAM Provider, currently only supported from file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		// Specific flags
		pamConfigFile, _ := cmd.Flags().GetString(FlagFromFile)

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Str("file", pamConfigFile).
			Msg("update PAM Provider from file")

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		var pamProvider *keyfactor.Provider
		log.Debug().Str("file", pamConfigFile).
			Msg("call: GetTypeFromConfigFile()")
		pamProvider, err := GetTypeFromConfigFile(pamConfigFile, pamProvider)
		log.Debug().Msg("returned: GetTypeFromConfigFile()")
		if err != nil {
			//log.Printf("%sError reading from file %s: %s", ColorRed, pamConfigFile, err)
			log.Error().Err(err).Str("file", pamConfigFile).Msg("failed parsing PAM Provider config from file")
			return err
		}

		log.Debug().Msg("call: PAMProviderUpdatePamProvider()")
		updateRequest := keyfactor.ProviderUpdateRequestLegacy{
			Name:                    pamProvider.Name,
			Remote:                  pamProvider.Remote,
			Area:                    pamProvider.Area,
			ProviderType:            pamProvider.ProviderType,
			ProviderTypeParamValues: pamProvider.ProviderTypeParamValues,
			SecuredAreaId:           pamProvider.SecuredAreaId,
		}

		updatedPamProvider, cErr := kfClient.UpdatePAMProvider(&updateRequest)

		log.Debug().Msg("returned: PAMProviderUpdatePamProvider()")
		if err != nil {
			return err
		}

		log.Debug().Msg(convertResponseMsg)
		jsonString, mErr := json.Marshal(updatedPamProvider)
		if mErr != nil {
			log.Error().Err(mErr).Msg("invalid API response from Keyfactor Command")
			return mErr
		}

		log.Info().
			Str("pamConfigFile", pamConfigFile).
			Str("output", string(jsonString)).
			Msg("successfully updated PAM provider")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamProvidersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a defined PAM Provider by ID.",
	Long:  "Delete a defined PAM Provider by ID.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		// Specific flags
		pamProviderId, _ := cmd.Flags().GetInt32("id")
		pamProviderName, _ := cmd.Flags().GetString("name")

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Int32("id", pamProviderId).
			Msg("delete PAM Provider")

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		//log.Debug().
		//	Int32("id", pamProviderId).
		//	Msg("call: PAMProviderDeletePamProvider()")
		//httpResponse, err := sdkClient.PAMProviderApi.PAMProviderDeletePamProvider(context.Background(), pamProviderId).
		//	XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
		//	Execute()
		//log.Debug().Msg("returned: PAMProviderDeletePamProvider()")
		//log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderDeletePamProvider")
		//if err != nil {
		//	log.Error().Err(err).Int32("id", pamProviderId).Msg("failed to delete PAM provider")
		//	return err
		//}

		if pamProviderId == 0 && pamProviderName != "" {
			log.Debug().Str("name", pamProviderName).Msg("resolving PAM Provider ID from name")
			pamProvider, err := kfClient.GetPamProviderByName(pamProviderName)
			if err != nil {
				log.Error().Err(err).Str(
					"name",
					pamProviderName,
				).Msg("error listing PAM providers to resolve ID from name")
				return err
			} else if pamProvider == nil {
				log.Error().Str(
					"name",
					pamProviderName,
				).Msg("PAM provider not found to resolve ID from name")
				return fmt.Errorf("PAM provider not found with name '%s'", pamProviderName)
			}
			pamProviderId = int32(pamProvider.Id)
		}

		delErr := kfClient.DeletePAMProvider(int(pamProviderId))
		if delErr != nil {
			log.Error().Err(delErr).Int32("id", pamProviderId).Msg("failed to delete PAM provider")
			return delErr
		}

		log.Info().Int32("id", pamProviderId).Msg("successfully deleted PAM provider")
		outputResult(fmt.Sprintf("Deleted PAM provider with ID %d", pamProviderId), outputFormat)
		return nil
	},
}

func init() {
	var (
		filePath string
		name     string
		id       int32
	)

	RootCmd.AddCommand(pamCmd)

	// PAM Providers

	// PAM Providers List
	pamCmd.AddCommand(pamProvidersListCmd)

	// PAM Providers Get
	pamCmd.AddCommand(pamProvidersGetCmd)
	pamProvidersGetCmd.Flags().Int32VarP(&id, "id", "i", 0, "Integer ID of the PAM Provider.")
	pamProvidersGetCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider.")
	pamProvidersGetCmd.MarkFlagsMutuallyExclusive("id", "name")

	// PAM Providers Create
	pamCmd.AddCommand(pamProvidersCreateCmd)
	pamProvidersCreateCmd.Flags().StringVarP(
		&filePath,
		FlagFromFile,
		"f",
		"",
		"Path to a JSON file containing the PAM Provider Object Data.",
	)
	pamProvidersCreateCmd.MarkFlagRequired(FlagFromFile)
	pamCmd.AddCommand(pamProvidersUpdateCmd)
	pamProvidersUpdateCmd.Flags().StringVarP(
		&filePath,
		FlagFromFile,
		"f",
		"",
		"Path to a JSON file containing the PAM Provider Object Data.",
	)

	// PAM Providers Update
	pamProvidersUpdateCmd.MarkFlagRequired(FlagFromFile)

	// PAM Providers Delete
	pamCmd.AddCommand(pamProvidersDeleteCmd)
	pamProvidersDeleteCmd.Flags().Int32VarP(&id, "id", "i", 0, "Integer ID of the PAM Provider.")
	pamProvidersDeleteCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider.")
	pamProvidersDeleteCmd.MarkFlagsMutuallyExclusive("id", "name")

}
