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
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	keyfactor "github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed pam_types.json
var EmbeddedPAMTypesJSON []byte

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

var pamTypesGetCmd = &cobra.Command{
	Use:   "types-get",
	Short: "Get a specific defined PAM Provider type by ID or Name.",
	Long:  "Get a specific defined PAM Provider type by ID or Name.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false
		// Specific flags
		pamProviderTypeId, _ := cmd.Flags().GetString("id")
		pamProviderTypeName, _ := cmd.Flags().GetString("name")
		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		// Log flags
		log.Info().Str("name", pamProviderTypeName).
			Str("id", pamProviderTypeId).
			Msg("get PAM Provider Type")
		// Authenticate
		kfClient, cErr := initClient(false)
		if cErr != nil {
			return cErr
		}
		if pamProviderTypeId == "" && pamProviderTypeName == "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a PAM Provider Type `--id` or `--name` to get")
		}

		// CLI Logic
		if pamProviderTypeId == "" && pamProviderTypeName != "" {
			// Get ID from Name
			log.Debug().Str("pamProviderTypeName", pamProviderTypeName).
				Msg("call: GetPAMProviderTypeByName()")
			pamProviderType, getErr := kfClient.GetPAMProviderTypeByName(pamProviderTypeName)
			log.Debug().Msg("returned: GetPAMProviderTypeByName()")
			if getErr != nil {
				log.Error().Err(getErr).Send()
				return getErr
			}
			if pamProviderType != nil {
				output, mErr := json.Marshal(pamProviderType)
				if mErr != nil {
					log.Error().Err(mErr).Send()
					return mErr
				}
				log.Info().Str("output", string(output)).
					Msg("successfully retrieved PAM provider type")
				outputResult(output, outputFormat)
				return nil
			}
		}
		pamProviderType, getErr := kfClient.GetPAMProviderType(pamProviderTypeId)
		if getErr != nil {
			log.Error().Err(getErr).Send()
			return getErr
		}
		output, mErr := json.Marshal(pamProviderType)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().Str("output", string(output)).
			Msg("successfully retrieved PAM provider type")
		outputResult(output, outputFormat)
		return nil
	},
}

var pamTypesListCmd = &cobra.Command{
	Use:   "types-list",
	Short: "Returns a list of all available PAM provider types.",
	Long:  "Returns a list of all available PAM provider types.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Msg("list PAM Provider Types")

		// Authenticate
		kfClient, clientErr := initClient(false)
		if clientErr != nil {
			return clientErr
		}

		//// CLI Logic
		//log.Debug().Msg("call: PAMProviderGetPamProviderTypes()")
		//pamTypes, httpResponse, err := sdkClient.PAMProviderApi.
		//	PAMProviderGetPamProviderTypes(context.Background()).
		//	XKeyfactorRequestedWith(XKeyfactorRequestedWith).
		//	XKeyfactorApiVersion(XKeyfactorApiVersion).
		//	Execute()
		pamTypes, err := kfClient.ListPAMProviderTypes()
		log.Debug().Msg("returned: PAMProviderGetPamProviderTypes()")
		if err != nil {
			log.Error().Err(err).
				Msg("error listing PAM provider types")
			return err
		}

		log.Debug().Msg("Converting PAM Provider Types response to JSON")
		jsonString, mErr := json.Marshal(pamTypes)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().
			Msg("successfully listed PAM provider types")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamTypesCreateCmd = &cobra.Command{
	Use:   "types-create",
	Short: "Creates a new PAM provider type.",
	Long: `Creates a new PAM Provider type, currently only supported from JSON file and from GitHub. To install from 
Github. To install from GitHub, use the --repo flag to specify the GitHub repository and optionally the branch to use. 
NOTE: the file from Github must be named integration-manifest.json and must use the same schema as 
https://github.com/Keyfactor/hashicorp-vault-pam/blob/main/integration-manifest.json. To install from a local file, use
--from-file to specify the path to the JSON file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		// Specific flags
		pamConfigFile, _ := cmd.Flags().GetString(FlagFromFile)
		pamProviderName, _ := cmd.Flags().GetString("name")
		repoName, _ := cmd.Flags().GetString("repo")
		branchName, _ := cmd.Flags().GetString("branch")

		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Str("name", pamProviderName).
			Str("repo", repoName).
			Str("branch", branchName).
			Msg("create PAM Provider Type")

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// Check required flags
		if pamConfigFile == "" && repoName == "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a config `--from-file` or a `--repo` GitHub repository to get file from")
		} else if pamConfigFile != "" && repoName != "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a config `--from-file` or a `--repo` GitHub repository to get file from, not both")
		}

		// CLI Logic

		var pamProviderType *keyfactor.ProviderTypeCreateRequest
		var err error
		if repoName != "" {
			// get JSON config from integration-manifest on GitHub
			log.Debug().
				Str("pamProviderName", pamProviderName).
				Str("repoName", repoName).
				Str("branchName", branchName).
				Msg("call: GetTypeFromInternet()")
			pamProviderType, err = GetTypeFromInternet(pamProviderName, repoName, branchName, pamProviderType)
			log.Debug().Msg("returned: GetTypeFromInternet()")
			if err != nil {
				log.Error().Err(err).Send()
				return err
			}
		} else {
			log.Debug().Str("pamConfigFile", pamConfigFile).
				Msg(fmt.Sprintf("call: %s", "GetTypeFromConfigFile()"))
			pamProviderType, err = GetTypeFromConfigFile(pamConfigFile, pamProviderType)
			log.Debug().Msg(fmt.Sprintf("returned: %s", "GetTypeFromConfigFile()"))
			if err != nil {
				log.Error().Err(err).Send()
				return err
			}
		}

		if pamProviderName != "" {
			pamProviderType.Name = pamProviderName
		}

		log.Info().Str("pamProviderName", pamProviderType.Name).
			Msg("creating PAM provider type")

		log.Debug().Msg("call: PAMProviderCreatePamProviderType()")
		createdPamProviderType, rErr := kfClient.CreatePAMProviderType(pamProviderType)
		log.Debug().Msg("returned: PAMProviderCreatePamProviderType()")
		if rErr != nil {
			log.Error().Err(rErr).Send()
			return rErr
		}

		log.Debug().Msg("Converting PAM Provider Type response to JSON")
		jsonString, mErr := json.Marshal(createdPamProviderType)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().Str("output", string(jsonString)).
			Msg("successfully created PAM provider type")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamTypesDeleteCmd = &cobra.Command{
	Use:   "types-delete",
	Short: "Deletes a defined PAM Provider type by ID or Name.",
	Long:  "Deletes a defined PAM Provider type by ID or Name.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false
		// Specific flags
		pamProviderTypeId, _ := cmd.Flags().GetString("id")
		pamProviderTypeName, _ := cmd.Flags().GetString("name")
		deleteAll, _ := cmd.Flags().GetBool("all")
		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr

		}
		// Log flags
		log.Info().Str("name", pamProviderTypeName).
			Str("id", pamProviderTypeId).
			Msg("delete PAM Provider Type")
		// Authenticate
		kfClient, cErr := initClient(false)
		if cErr != nil {
			return cErr
		}
		// CLI Logic

		if deleteAll {
			log.Info().Msg("deleting ALL PAM Provider Types")
			pamProviderTypes, listErr := kfClient.ListPAMProviderTypes()
			if listErr != nil {
				log.Error().Err(listErr).Send()
				return listErr
			}
			if pamProviderTypes == nil || len(*pamProviderTypes) == 0 {
				log.Info().Msg("no PAM provider types to delete")
				outputResult("No PAM provider types to delete", outputFormat)
			}
			for _, pamProviderType := range *pamProviderTypes {
				log.Debug().Str("pamProviderTypeId", pamProviderType.Id).
					Msg("call: PAMProviderDeletePamProviderType()")
				delErr := kfClient.DeletePAMProviderType(pamProviderType.Id)
				if delErr != nil {
					log.Error().Err(delErr).Send()
					outputError(delErr, false, outputFormat)
					continue
				}
				log.Info().Str("id", pamProviderType.Id).Str("name", *pamProviderType.Name).
					Msg("successfully deleted PAM provider type")
				outputResult(fmt.Sprintf("Deleted PAM provider type with ID %s", pamProviderType.Id), outputFormat)
			}
			log.Info().Msg("successfully deleted ALL PAM provider types")
			outputResult(fmt.Sprintf("Deleted ALL %d PAM provider types", len(*pamProviderTypes)), outputFormat)
			return nil
		}
		if pamProviderTypeId == "" && pamProviderTypeName == "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a PAM Provider Type `--id` or `--name` to delete")
		}

		if pamProviderTypeId == "" && pamProviderTypeName != "" {
			// Get ID from Name
			log.Debug().Str("pamProviderTypeName", pamProviderTypeName).
				Msg("call: GetPAMProviderTypeByName()")
			pamProviderType, getErr := kfClient.GetPAMProviderTypeByName(pamProviderTypeName)
			log.Debug().Msg("returned: GetPAMProviderTypeByName()")
			if getErr != nil {
				log.Error().Err(getErr).Send()
				return getErr
			}
			pamProviderTypeId = pamProviderType.Id
		}

		log.Debug().Str("pamProviderTypeId", pamProviderTypeId).
			Msg("call: PAMProviderDeletePamProviderType()")
		delErr := kfClient.DeletePAMProviderType(pamProviderTypeId)
		if delErr != nil {
			log.Error().Err(delErr).Send()
			return delErr
		}

		log.Info().Str("name", pamProviderTypeName).
			Str("id", pamProviderTypeId).
			Msg("successfully deleted PAM provider type")
		outputResult(fmt.Sprintf("Deleted PAM provider type with ID %s", pamProviderTypeId), outputFormat)
		return nil
	},
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
		//kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		log.Debug().Msg("call: PAMProviderGetPamProviders()")
		pamProviders, httpResponse, err := sdkClient.PAMProviderApi.PAMProviderGetPamProviders(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		log.Debug().Msg("returned: PAMProviderGetPamProviders()")
		log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderGetPamProviders")
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
		//kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		log.Debug().Msg("call: PAMProviderGetPamProvider()")
		pamProvider, httpResponse, err := sdkClient.PAMProviderApi.PAMProviderGetPamProvider(
			context.Background(),
			pamProviderId,
		).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		log.Debug().Msg("returned: PAMProviderGetPamProvider()")
		log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderGetPamProvider")

		if err != nil {
			log.Error().Err(err).Str("httpResponseCode", httpResponse.Status).Msg("error getting PAM provider")
			return err
		}

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

func checkBug63171(cmdResp *http.Response, operation string) error {
	if cmdResp != nil && cmdResp.StatusCode == 200 {
		defer cmdResp.Body.Close()
		// .\Admin
		productVersion := cmdResp.Header.Get("X-Keyfactor-Product-Version")
		log.Debug().Str("productVersion", productVersion).Msg("Keyfactor Command Version")
		majorVersionStr := strings.Split(productVersion, ".")[0]
		// Try to convert to int
		majorVersion, err := strconv.Atoi(majorVersionStr)
		if err == nil && majorVersion >= 12 {
			// TODO: Pending resolution of this bug: https://dev.azure.com/Keyfactor/Engineering/_workitems/edit/63171
			errMsg := fmt.Sprintf(
				"PAM Provider %s is not supported in Keyfactor Command version 12 and later, "+
					"please use the Keyfactor Command UI to create PAM Providers", operation,
			)
			oErr := fmt.Errorf(errMsg)
			log.Error().Err(oErr).Send()
			outputError(oErr, true, outputFormat)
			return oErr
		}
	}
	return nil
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
		// pamProviderName := cmd.Flags().GetString("name")

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
		//kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		log.Debug().
			Int32("id", pamProviderId).
			Msg("call: PAMProviderDeletePamProvider()")
		httpResponse, err := sdkClient.PAMProviderApi.PAMProviderDeletePamProvider(context.Background(), pamProviderId).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		log.Debug().Msg("returned: PAMProviderDeletePamProvider()")
		log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderDeletePamProvider")
		if err != nil {
			log.Error().Err(err).Int32("id", pamProviderId).Msg("failed to delete PAM provider")
			return err
		}

		log.Info().Int32("id", pamProviderId).Msg("successfully deleted PAM provider")
		outputResult(fmt.Sprintf("Deleted PAM provider with ID %d", pamProviderId), outputFormat)
		return nil
	},
}

func GetPAMTypeInternet(providerName string, repo string, branch string) (interface{}, error) {
	log.Debug().Str("providerName", providerName).
		Str("repo", repo).
		Str("branch", branch).
		Msg("entered: GetPAMTypeInternet()")

	if branch == "" {
		log.Info().Msg("branch not specified, using 'main' by default")
		branch = "main"
	}

	providerUrl := fmt.Sprintf(
		"https://raw.githubusercontent.com/Keyfactor/%s/%s/integration-manifest.json",
		repo,
		branch,
	)
	log.Debug().Str("providerUrl", providerUrl).
		Msg("Getting PAM Type from Internet")
	response, err := http.Get(providerUrl)
	if err != nil {
		log.Error().Err(err).
			Str("providerUrl", providerUrl).
			Msg("error getting PAM Type from Internet")
		return nil, err
	}
	log.Trace().Interface("httpResponse", response).
		Msg("GetPAMTypeInternet")

	//check response status code is 200
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response status: %s", response.Status)
	}

	defer response.Body.Close()

	log.Debug().Msg("Parsing PAM response")
	manifest, iErr := io.ReadAll(response.Body)
	if iErr != nil {
		log.Error().Err(iErr).
			Str("providerUrl", providerUrl).
			Msg("unable to read PAM response")
		return nil, iErr
	}
	log.Trace().Interface("manifest", manifest).Send()

	var manifestJson map[string]interface{}
	log.Debug().Msg("Converting PAM response to JSON")
	jErr := json.Unmarshal(manifest, &manifestJson)
	if jErr != nil {
		log.Error().Err(jErr).
			Str("providerUrl", providerUrl).
			Msg("invalid integration-manifest.json provided")
		return nil, jErr
	}
	log.Debug().Msg("Parsing manifest response for PAM type config")
	pamTypeJson := manifestJson["about"].(map[string]interface{})["pam"].(map[string]interface{})["pam_types"].(map[string]interface{})[providerName]
	if pamTypeJson == nil {
		// Check if only one PAM Type is defined
		pamTypeJson = manifestJson["about"].(map[string]interface{})["pam"].(map[string]interface{})["pam_types"].(map[string]interface{})
		if len(pamTypeJson.(map[string]interface{})) == 1 {
			for _, v := range pamTypeJson.(map[string]interface{}) {
				pamTypeJson = v
			}
		} else {
			return nil, fmt.Errorf("unable to find PAM type %s in manifest on %s", providerName, providerUrl)
		}
	}

	log.Trace().Interface("pamTypeJson", pamTypeJson).Send()
	log.Debug().Msg("returning: GetPAMTypeInternet()")
	return pamTypeJson, nil
}

func GetTypeFromInternet[T JSONImportableObject](providerName string, repo string, branch string, returnType *T) (
	*T,
	error,
) {
	log.Debug().Str("providerName", providerName).
		Str("repo", repo).
		Str("branch", branch).
		Msg("entered: GetTypeFromInternet()")

	log.Debug().Msg("call: GetPAMTypeInternet()")
	manifestJSON, err := GetPAMTypeInternet(providerName, repo, branch)
	log.Debug().Msg("returned: GetPAMTypeInternet()")
	if err != nil {
		log.Error().Err(err).Send()
		return new(T), err
	}

	log.Debug().Msg("Converting PAM Type from manifest to bytes")
	manifestJSONBytes, jErr := json.Marshal(manifestJSON)
	if jErr != nil {
		log.Error().Err(jErr).Send()
		return new(T), jErr
	}

	var objectFromJSON T
	log.Debug().Msg("Converting PAM Type from bytes to JSON")
	mErr := json.Unmarshal(manifestJSONBytes, &objectFromJSON)
	if mErr != nil {
		log.Error().Err(mErr).Send()
		return new(T), mErr
	}

	log.Debug().Msg("returning: GetTypeFromInternet()")
	return &objectFromJSON, nil
}

func GetTypeFromConfigFile[T JSONImportableObject](filename string, returnType *T) (*T, error) {
	log.Debug().Str("filename", filename).
		Msg("entered: GetTypeFromConfigFile()")

	log.Debug().Str("filename", filename).
		Msg("Opening PAM Type config file")
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Send()
		return new(T), err
	}

	var rawData map[string]interface{}
	decoder := json.NewDecoder(file)
	dErr := decoder.Decode(&rawData)
	if dErr != nil {
		log.Error().Err(dErr).Send()
		return new(T), dErr
	}

	if _, ok := rawData["about"]; ok {
		// If the file contains the full manifest, extract the PAM type config
		log.Debug().Msg("Parsing PAM Type config from manifest file")
		about := rawData["about"].(map[string]interface{})
		pam := about["pam"].(map[string]interface{})
		pamTypes := pam["pam_types"].(map[string]interface{})
		var pamTypeConfig interface{}
		if len(pamTypes) == 1 {
			for _, v := range pamTypes {
				pamTypeConfig = v
			}
		} else {
			return new(T), fmt.Errorf("multiple PAM types found in manifest file, please provide a file with a single PAM type definition")
		}

		log.Debug().Msg("Converting PAM Type config from manifest to bytes")
		pamTypeConfigBytes, jErr := json.Marshal(pamTypeConfig)
		if jErr != nil {
			log.Error().Err(jErr).Send()
			return new(T), jErr
		}

		var objectFromManifest T
		log.Debug().Msg("Converting PAM Type config from bytes to JSON")
		mErr := json.Unmarshal(pamTypeConfigBytes, &objectFromManifest)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return new(T), mErr
		}

		log.Debug().Msg("returning: GetTypeFromConfigFile()")
		return &objectFromManifest, nil
	}

	// Rewind file pointer to beginning
	_, sErr := file.Seek(0, io.SeekStart)
	if sErr != nil {
		log.Error().Err(sErr).Send()
		return new(T), sErr
	}

	// If the file contains only the PAM type config
	var objectFromFile T
	log.Debug().Msg("Decoding PAM Type config file")
	decoder = json.NewDecoder(file)
	dErr = decoder.Decode(&objectFromFile)
	if dErr != nil {
		log.Error().Err(dErr).Send()
		return new(T), dErr
	}

	log.Debug().Msg("returning: GetTypeFromConfigFile()")
	return &objectFromFile, nil
}

func init() {
	var filePath string
	var name string
	var repo string
	var branch string
	var id int32
	var providerTypeId string
	var all bool
	RootCmd.AddCommand(pamCmd)

	pamCmd.AddCommand(pamTypesGetCmd)
	pamTypesGetCmd.Flags().StringVarP(&providerTypeId, "id", "i", "", "ID of the PAM Provider Type.")
	pamTypesGetCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesGetCmd.MarkFlagsMutuallyExclusive("id", "name")

	// PAM Provider Types List
	pamCmd.AddCommand(pamTypesListCmd)

	// PAM Provider Types Create
	pamCmd.AddCommand(pamTypesCreateCmd)
	pamTypesCreateCmd.Flags().StringVarP(
		&filePath,
		FlagFromFile,
		"f",
		"",
		"Path to a JSON file containing the PAM Type Object Data.",
	)
	pamTypesCreateCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesCreateCmd.Flags().StringVarP(&repo, "repo", "r", "", "Keyfactor repository name of the PAM Provider Type.")
	pamTypesCreateCmd.Flags().StringVarP(
		&branch,
		"branch",
		"b",
		"",
		"Branch name for the repository. Defaults to 'main'.",
	)

	pamTypesDeleteCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesDeleteCmd.Flags().StringVarP(&providerTypeId, "id", "i", "", "ID of the PAM Provider Type.")
	pamTypesDeleteCmd.Flags().BoolVarP(&all, "all", "a", false, "Delete all PAM Provider Types.")
	pamTypesDeleteCmd.MarkFlagsMutuallyExclusive("id", "name", "all")
	pamCmd.AddCommand(pamTypesDeleteCmd)

	// PAM Providers
	pamCmd.AddCommand(pamProvidersListCmd)
	pamCmd.AddCommand(pamProvidersGetCmd)
	pamProvidersGetCmd.Flags().Int32VarP(&id, "id", "i", 0, "Integer ID of the PAM Provider.")
	pamProvidersGetCmd.MarkFlagRequired("id")

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
	pamProvidersUpdateCmd.MarkFlagRequired(FlagFromFile)

	pamCmd.AddCommand(pamProvidersDeleteCmd)
	pamProvidersDeleteCmd.Flags().Int32VarP(&id, "id", "i", 0, "Integer ID of the PAM Provider.")
	pamProvidersDeleteCmd.MarkFlagRequired("id")

}
