// Package cmd Copyright 2023 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/spf13/cobra"
)

type JsonImportableObject interface {
	keyfactor.KeyfactorApiPAMProviderTypeCreateRequest |
		keyfactor.CSSCMSDataModelModelsProvider
}

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
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)

		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamTypes, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderGetPamProviderTypes(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Get PAM Types", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(pamTypes)
		if marshallError != nil {
			log.Printf("%sError: %s", ColorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
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
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamConfigFile, _ := cmd.Flags().GetString("from-file")
		providerName, _ := cmd.Flags().GetString("name")
		repoName, _ := cmd.Flags().GetString("repo")
		branchName, _ := cmd.Flags().GetString("branch")

		if pamConfigFile == "" && repoName == "" {
			log.Printf("%sError - must supply either a config file or GitHub repository to get file from.", ColorRed)
			return
		}

		var pamProviderType *keyfactor.KeyfactorApiPAMProviderTypeCreateRequest
		var errors error
		if repoName != "" {
			// get JSON config from integration-manifest on GitHub
			pamProviderType, errors = GetTypeFromInternet(providerName, repoName, branchName, pamProviderType)
			if errors != nil {
				log.Printf("%sError reading from GitHub %s/%s: %s", ColorRed, repoName, branchName, errors)
				fmt.Println("Please check the repository name and branch name and try again.")
				fmt.Println(errors)
				return
			}
		} else {
			pamProviderType, errors = GetTypeFromConfigFile(pamConfigFile, pamProviderType)
			if errors != nil {
				log.Printf("%sError reading from file %s: %s", ColorRed, pamConfigFile, errors)
				return
			}
		}

		// pamType, errors :=
		createdPamProviderType, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderCreatePamProviderType(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Type_(*pamProviderType).
			Execute()
		if errors != nil {
			WriteApiError("Create PAM Provider type", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(createdPamProviderType)
		if marshallError != nil {
			log.Printf("%sError: %s", ColorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersListCmd = &cobra.Command{
	Use:   "list",
	Short: "Returns a list of all the configured PAM providers.",
	Long:  "Returns a list of all the configured PAM providers.",
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)

		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamProviders, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderGetPamProviders(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Get PAM Providers", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(pamProviders)
		if marshallError != nil {
			log.Printf("%sError: %s", ColorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific defined PAM Provider by ID.",
	Long:  "Get a specific defined PAM Provider by ID.",
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamProviderId, _ := cmd.Flags().GetInt32("id")
		// pamProviderName := cmd.Flags().GetString("name")

		pamProvider, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderGetPamProvider(context.Background(), pamProviderId).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Get PAM Provider", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(pamProvider)
		if marshallError != nil {
			log.Printf("%sError: %s", ColorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new PAM Provider, currently only supported from file.",
	Long:  "Create a new PAM Provider, currently only supported from file.",
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamConfigFile, _ := cmd.Flags().GetString("from-file")

		var pamProvider *keyfactor.CSSCMSDataModelModelsProvider
		pamProvider, errors := GetTypeFromConfigFile(pamConfigFile, pamProvider)
		if errors != nil {
			log.Printf("%sError reading from file %s: %s", ColorRed, pamConfigFile, errors)
			return
		}

		// pamType, errors :=
		createdPamProvider, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderCreatePamProvider(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Provider(*pamProvider).
			Execute()
		if errors != nil {
			WriteApiError("Create PAM Provider", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(createdPamProvider)
		if marshallError != nil {
			log.Printf("%sError: %s", ColorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates an existing PAM Provider, currently only supported from file.",
	Long:  "Updates an existing PAM Provider, currently only supported from file.",
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamConfigFile, _ := cmd.Flags().GetString("from-file")

		var pamProvider *keyfactor.CSSCMSDataModelModelsProvider
		pamProvider, errors := GetTypeFromConfigFile(pamConfigFile, pamProvider)
		if errors != nil {
			log.Printf("%sError reading from file %s: %s", ColorRed, pamConfigFile, errors)
			return
		}

		// pamType, errors :=
		createdPamProvider, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderUpdatePamProvider(context.Background()).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Provider(*pamProvider).
			Execute()
		if errors != nil {
			WriteApiError("Create PAM Provider", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(createdPamProvider)
		if marshallError != nil {
			log.Printf("%sError: %s", ColorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a defined PAM Provider by ID.",
	Long:  "Delete a defined PAM Provider by ID.",
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		sdkClient, _ := initGenClient(configFile, profile, noPrompt, authConfig, false)
		pamProviderId, _ := cmd.Flags().GetInt32("id")
		// pamProviderName := cmd.Flags().GetString("name")

		httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderDeletePamProvider(context.Background(), pamProviderId).
			XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Delete PAM Provider", httpResponse, errors)
			return
		}

		fmt.Printf("Deleted PAM Provider %d", pamProviderId)
	},
}

func GetPamTypeInternet(providerName string, repo string, branch string) (interface{}, error) {
	if branch == "" {
		branch = "main"
	}
	response, errors := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/Keyfactor/%s/%s/integration-manifest.json", repo, branch))
	if errors != nil {
		return nil, errors
	}

	//check response status code is 200
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error: %s", response.Status)
	}

	defer response.Body.Close()

	manifest, errors := io.ReadAll(response.Body)
	if errors != nil {
		return nil, errors
	}

	var manifestJson map[string]interface{}
	errors = json.Unmarshal(manifest, &manifestJson)
	if errors != nil {
		log.Printf("%sError during Unmarshal() of PAM integration-manifest", ColorRed)
		return nil, errors
	}
	pamTypeJson := manifestJson["about"].(map[string]interface{})["pam"].(map[string]interface{})["pam_types"].(map[string]interface{})[providerName]

	return pamTypeJson, nil
}

func WriteApiError(process string, httpResponse *http.Response, errors error) {
	fmt.Printf("%s Error processing request for %s - %s - %s", ColorRed, process, errors, parseError(httpResponse.Body))
}

func GetTypeFromInternet[T JsonImportableObject](providerName string, repo string, branch string, returnType *T) (*T, error) {
	manifestJson, errors := GetPamTypeInternet(providerName, repo, branch)
	if errors != nil {
		return new(T), errors
	}

	manifestJsonBytes, errors := json.Marshal(manifestJson)
	if errors != nil {
		log.Printf("Error during Marshal() of PAM Type from manifest: %s", errors)
		return new(T), errors
	}

	var objectFromJson T
	errors = json.Unmarshal(manifestJsonBytes, &objectFromJson)
	if errors != nil {
		log.Printf("Error during Unmarshal(): %s", errors)
		return new(T), errors
	}

	return &objectFromJson, nil
}

func GetTypeFromConfigFile[T JsonImportableObject](filename string, returnType *T) (*T, error) {
	file, errors := os.Open(filename)
	if errors != nil {
		return new(T), errors
	}

	var objectFromFile T
	decoder := json.NewDecoder(file)
	errors = decoder.Decode(&objectFromFile)
	if errors != nil {
		return new(T), errors
	}

	return &objectFromFile, nil
}

func init() {
	var filePath string
	var name string
	var repo string
	var branch string
	var id int32
	RootCmd.AddCommand(pamCmd)

	// PAM Provider Types
	pamCmd.AddCommand(pamTypesListCmd)
	pamCmd.AddCommand(pamTypesCreateCmd)
	pamTypesCreateCmd.Flags().StringVarP(&filePath, "from-file", "f", "", "Path to a JSON file containing the PAM Type Object Data.")
	pamTypesCreateCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesCreateCmd.Flags().StringVarP(&repo, "repo", "r", "", "Keyfactor repository name of the PAM Provider Type.")
	pamTypesCreateCmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch name for the repository. Can be left blank for 'main' by default.")

	// PAM Providers
	pamCmd.AddCommand(pamProvidersListCmd)
	pamCmd.AddCommand(pamProvidersGetCmd)
	pamProvidersGetCmd.Flags().Int32VarP(&id, "id", "i", 0, "Integer ID of the PAM Provider.")
	pamProvidersGetCmd.MarkFlagRequired("id")

	pamCmd.AddCommand(pamProvidersCreateCmd)
	pamProvidersCreateCmd.Flags().StringVarP(&filePath, "from-file", "f", "", "Path to a JSON file containing the PAM Provider Object Data.")
	pamProvidersCreateCmd.MarkFlagRequired("from-file")

	pamCmd.AddCommand(pamProvidersUpdateCmd)
	pamProvidersUpdateCmd.Flags().StringVarP(&filePath, "from-file", "f", "", "Path to a JSON file containing the PAM Provider Object Data.")
	pamProvidersUpdateCmd.MarkFlagRequired("from-file")

	pamCmd.AddCommand(pamProvidersDeleteCmd)
	pamProvidersDeleteCmd.Flags().Int32VarP(&id, "id", "i", 0, "Integer ID of the PAM Provider.")
	pamProvidersDeleteCmd.MarkFlagRequired("id")

}
