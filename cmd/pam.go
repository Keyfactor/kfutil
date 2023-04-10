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

	"github.com/spf13/cobra"
)

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

func WriteApiError(process string, httpResponse *http.Response, errors error) {
	fmt.Printf("%s Error processing request for %s - %s - %s", colorRed, process, errors, parseError(httpResponse.Body))
}

func init() {
	RootCmd.AddCommand(pamCmd)

	// GET PAM Provider Types
	pamCmd.AddCommand(pamTypesListCmd)
}
