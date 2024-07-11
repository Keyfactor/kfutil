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
	"io"
	"net/http"
	"os"

	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type JSONImportableObject interface {
	keyfactor.KeyfactorApiPAMProviderTypeCreateRequest |
		keyfactor.CSSCMSDataModelModelsProvider
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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		sdkClient, clientErr := initGenClient(configFile, profile, noPrompt, authConfig, false)
		if clientErr != nil {
			return clientErr
		}

		// CLI Logic
		log.Debug().Msg("call: PAMProviderGetPamProviderTypes()")
		pamTypes, httpResponse, err := sdkClient.PAMProviderApi.
			PAMProviderGetPamProviderTypes(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).
			XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		log.Debug().Msg("returned: PAMProviderGetPamProviderTypes()")
		log.Trace().Interface("httpResponse", httpResponse).
			Msg("PAMProviderGetPamProviderTypes")
		if err != nil {
			var status string
			if httpResponse != nil {
				status = httpResponse.Status
			} else {
				status = "No HTTP response received from Keyfactor Command."
			}
			log.Error().Err(err).
				Str("httpResponseCode", status).
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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		//Client, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)

		// Check required flags
		if pamConfigFile == "" && repoName == "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a config `--from-file` or a `--repo` GitHub repository to get file from")
		} else if pamConfigFile != "" && repoName != "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a config `--from-file` or a `--repo` GitHub repository to get file from, not both")
		}

		// CLI Logic

		var pamProviderType *keyfactor.KeyfactorApiPAMProviderTypeCreateRequest
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
		createdPamProviderType, httpResponse, rErr := sdkClient.PAMProviderApi.PAMProviderCreatePamProviderType(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Type_(*pamProviderType).
			Execute()
		log.Debug().Msg("returned: PAMProviderCreatePamProviderType()")
		log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderCreatePamProviderType")
		if rErr != nil {
			log.Error().Err(rErr).Send()
			return returnHttpErr(httpResponse, rErr)
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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		//Client, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)

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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		//Client, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)

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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		// Client, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)

		// CLI Logic
		var pamProvider *keyfactor.CSSCMSDataModelModelsProvider
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
		createdPamProvider, httpResponse, cErr := sdkClient.PAMProviderApi.PAMProviderCreatePamProvider(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Provider(*pamProvider).
			Execute()
		log.Debug().Msg("returned: PAMProviderCreatePamProvider()")
		log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderCreatePamProvider")
		if cErr != nil {
			// output response body
			log.Debug().Msg("Converting PAM Provider response body to string")
			return returnHttpErr(httpResponse, cErr)
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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		//Client, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)

		// CLI Logic
		var pamProvider *keyfactor.CSSCMSDataModelModelsProvider
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
		createdPamProvider, httpResponse, err := sdkClient.PAMProviderApi.PAMProviderUpdatePamProvider(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Provider(*pamProvider).
			Execute()
		log.Debug().Msg("returned: PAMProviderUpdatePamProvider()")
		log.Trace().Interface("httpResponse", httpResponse).Msg("PAMProviderUpdatePamProvider")
		if err != nil {
			returnHttpErr(httpResponse, err)
		}

		log.Debug().Msg(convertResponseMsg)
		jsonString, mErr := json.Marshal(createdPamProvider)
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
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		//Client, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)

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

	var objectFromFile T
	log.Debug().Msg("Decoding PAM Type config file")
	decoder := json.NewDecoder(file)
	dErr := decoder.Decode(&objectFromFile)
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
	RootCmd.AddCommand(pamCmd)

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
