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
	Long:  `A collections of APIs for interacting with Keyfactor PAM Providers and creating new PAM Provider types.`,
}

var pamTypesListCmd = &cobra.Command{
	Use:   "types-list",
	Short: "List defined PAM Provider types.",
	Long:  "List defined PAM Provider types.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		pamTypes, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderGetPamProviderTypes(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Get PAM Types", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(pamTypes)
		if marshallError != nil {
			log.Printf("%sError: %s", colorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamTypesCreateCmd = &cobra.Command{
	Use:   "types-create",
	Short: "Create a new PAM Provider type, currently only supported from file.",
	Long:  "Create a new PAM Provider type, currently only supported from file.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		configFile, _ := cmd.Flags().GetString("from-file")

		var pamProviderType *keyfactor.KeyfactorApiPAMProviderTypeCreateRequest
		pamProviderType, errors := GetTypeFromConfigFile(configFile, pamProviderType)
		if errors != nil {
			log.Printf("%sError reading from file %s: %s", colorRed, configFile, errors)
			return
		}

		// pamType, errors :=
		createdPamProviderType, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderCreatePamProviderType(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Type_(*pamProviderType).
			Execute()
		if errors != nil {
			WriteApiError("Create PAM Provider type", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(createdPamProviderType)
		if marshallError != nil {
			log.Printf("%sError: %s", colorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List defined PAM Providers.",
	Long:  "List defined PAM Providers.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		pamProviders, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderGetPamProviders(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Get PAM Providers", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(pamProviders)
		if marshallError != nil {
			log.Printf("%sError: %s", colorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}
var pamProvidersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific defined PAM Provider by ID.",
	Long:  "Get a specific defined PAM Provider by ID.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		pamProviderId, _ := cmd.Flags().GetInt32("id")
		// pamProviderName := cmd.Flags().GetString("name")

		pamProvider, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderGetPamProvider(context.Background(), pamProviderId).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Get PAM Provider", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(pamProvider)
		if marshallError != nil {
			log.Printf("%sError: %s", colorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new PAM Provider, currently only supported from file.",
	Long:  "Create a new PAM Provider, currently only supported from file.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		configFile, _ := cmd.Flags().GetString("from-file")

		var pamProvider *keyfactor.CSSCMSDataModelModelsProvider
		pamProvider, errors := GetTypeFromConfigFile(configFile, pamProvider)
		if errors != nil {
			log.Printf("%sError reading from file %s: %s", colorRed, configFile, errors)
			return
		}

		// pamType, errors :=
		createdPamProvider, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderCreatePamProvider(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Provider(*pamProvider).
			Execute()
		if errors != nil {
			WriteApiError("Create PAM Provider", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(createdPamProvider)
		if marshallError != nil {
			log.Printf("%sError: %s", colorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing PAM Provider, currently only supported from file.",
	Long:  "Update an existing PAM Provider, currently only supported from file.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		configFile, _ := cmd.Flags().GetString("from-file")

		var pamProvider *keyfactor.CSSCMSDataModelModelsProvider
		pamProvider, errors := GetTypeFromConfigFile(configFile, pamProvider)
		if errors != nil {
			log.Printf("%sError reading from file %s: %s", colorRed, configFile, errors)
			return
		}

		// pamType, errors :=
		createdPamProvider, httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderUpdatePamProvider(context.Background()).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Provider(*pamProvider).
			Execute()
		if errors != nil {
			WriteApiError("Create PAM Provider", httpResponse, errors)
			return
		}

		jsonString, marshallError := json.Marshal(createdPamProvider)
		if marshallError != nil {
			log.Printf("%sError: %s", colorRed, marshallError)
		}
		fmt.Printf("%s", jsonString)
	},
}

var pamProvidersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a defined PAM Provider by ID.",
	Long:  "Delete a defined PAM Provider by ID.",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		sdkClient := initGenClient()
		pamProviderId, _ := cmd.Flags().GetInt32("id")
		// pamProviderName := cmd.Flags().GetString("name")

		httpResponse, errors := sdkClient.PAMProviderApi.PAMProviderDeletePamProvider(context.Background(), pamProviderId).
			XKeyfactorRequestedWith(xKeyfactorRequestedWith).XKeyfactorApiVersion(xKeyfactorApiVersion).
			Execute()
		if errors != nil {
			WriteApiError("Delete PAM Provider", httpResponse, errors)
			return
		}

		fmt.Printf("Deleted PAM Provider %d", pamProviderId)
	},
}

func WriteApiError(process string, httpResponse *http.Response, errors error) {
	fmt.Printf("%s Error processing request for %s - %s - %s", colorRed, process, errors, parseError(httpResponse.Body))
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
	var id int32
	RootCmd.AddCommand(pamCmd)

	// PAM Provider Types
	pamCmd.AddCommand(pamTypesListCmd)
	pamCmd.AddCommand(pamTypesCreateCmd)
	pamTypesCreateCmd.Flags().StringVarP(&filePath, "from-file", "f", "", "Path to a JSON file containing the PAM Type Object Data.")
	pamTypesCreateCmd.MarkFlagRequired("from-file")

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
