// Package cmd Copyright 2023 Keyfactor
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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// storesCmd represents the stores command
var storesCmd = &cobra.Command{
	Use:   "stores",
	Short: "Keyfactor certificate stores APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate stores.`,
	//Run: func(cmd *cobra.Command, args []string) {
	//	fmt.Println("stores called")
	//},
}

var storesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate stores.",
	Long:  `List certificate stores.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		// Debug + expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)

		// CLI Logic
		params := make(map[string]interface{})
		stores, err := kfClient.ListCertificateStores(&params)

		if err != nil {
			log.Error().Err(err).Send()
			return err
		}
		output, jErr := json.Marshal(stores)
		if jErr != nil {
			log.Error().Err(jErr).Send()
			return jErr
		}
		outputResult(output, outputFormat)
		return nil
	},
}

var storesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a certificate store by ID.",
	Long:  `Get a certificate store by ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		storeID, _ := cmd.Flags().GetString("id")

		// Debug + expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)

		// CLI Logic
		stores, err := kfClient.GetCertificateStoreByID(storeID)
		if err != nil {
			log.Error().Err(err).Send()
			return err
		}
		output, jErr := json.Marshal(stores)
		if jErr != nil {
			log.Error().Err(jErr).Send()
			return jErr
		}
		outputResult(output, outputFormat)
		return nil
	},
}

var storesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a certificate store by ID.",
	Long:  `Delete a certificate store by ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		storeID, _ := cmd.Flags().GetString("id")
		deleteAll, _ := cmd.Flags().GetBool("all")

		// Debug + expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)

		// CLI Logic
		log.Info().Str("storeID", storeID).Msg("Deleting certificate store")
		log.Debug().Str("storeID", storeID).Msg("Checking that store exists")
		var (
			stores []string
		)
		if deleteAll {
			isExperimental := true
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			storesResp, err := kfClient.ListCertificateStores(nil)
			if err != nil {
				log.Error().Err(err).Send()
				return err
			}
			for _, s := range *storesResp {
				stores = append(stores, s.Id)
			}
		} else {
			stores = append(stores, storeID)
		}

		for _, st := range stores {
			_, err := kfClient.GetCertificateStoreByID(st)
			if err != nil {
				log.Error().Err(err).Send()
				return err
			}

			dErr := kfClient.DeleteCertificateStore(st)
			if dErr != nil {
				log.Error().Err(dErr).Send()
				return dErr
			}
			outputResult(fmt.Sprintf("successfully deleted store %s", st), outputFormat)
		}
		return nil
	},
}

func init() {
	var (
		storeID   string
		deleteAll bool
	)
	RootCmd.AddCommand(storesCmd)
	storesCmd.AddCommand(storesListCmd)
	storesCmd.AddCommand(storesGetCmd)
	storesCmd.AddCommand(storesDeleteCmd)

	// get cmd
	storesGetCmd.Flags().StringVarP(&storeID, "id", "i", "", "ID of the certificate store to get.")

	// delete cmd
	storesDeleteCmd.Flags().StringVarP(&storeID, "id", "i", "", "ID of the certificate store to delete.")
	storesDeleteCmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Attempt to delete ALL stores.")
	storesDeleteCmd.MarkFlagsMutuallyExclusive("id", "all")

}
