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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
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
		inputFile, _ := cmd.Flags().GetString("file")
		//outPath, _ := cmd.Flags().GetString("outpath")

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
		} else if storeID != "" {
			stores = append(stores, storeID)
		} else if inputFile != "" {
			// check that the input file is a valid csv file and contains the "Id" column in the header
			// if it is, read the file and add the Ids to the stores list
			// if it is not, return an error

			csvFile, ioErr := os.Open(inputFile)
			if ioErr != nil {
				log.Error().Err(ioErr).Msgf("unable to open file: '%s'", inputFile)
				//outputError(err, true, outputFormat)
				cmd.SilenceUsage = true
				return ioErr
			}
			log.Info().Msgf("Reading file '%s' as CSV", inputFile)
			inFile, cErr := csv.NewReader(csvFile).ReadAll()
			inputMap, _ := csvToMap(inputFile)
			if cErr != nil {
				log.Error().Err(cErr).
					Str("filePath", inputFile).
					Msg("unable to read file")
				//outputError(cErr, true, outputFormat)
				cmd.SilenceUsage = true
				return cErr
			}
			if len(inFile) < 1 {
				log.Error().Msg("No data in file")
				//outputError(errors.New("no data in file"), true, outputFormat)
				cmd.SilenceUsage = true
				return fmt.Errorf("no data in file %s", inputFile)
			}
			headerRow := inFile[0]
			//check that the header row contains the "Id" column
			containsID := false
			for _, h := range headerRow {
				if h == "Id" {
					containsID = true
					break
				}
			}
			if !containsID {
				log.Error().Msg("File does not contain 'Id' column")
				//outputError(errors.New("file does not contain 'Id' column"), true, outputFormat)
				cmd.SilenceUsage = true
				return fmt.Errorf("file does not contain 'Id' column, unable to delete stores")
			}

			for row, data := range inputMap {
				log.Trace().
					Int("row", row).Str("file", inputFile).
					Msg("Reading row from file")
				log.Debug().Str("storeID", data["Id"]).Msg("Adding store to delete list")
				stores = append(stores, data["Id"])
			}

		} else {
			//prompt as a multi-select for list of stores retrieved from command to delete
			log.Info().Msg("No store ID provided, prompting for store to delete")
			storesResp, err := kfClient.ListCertificateStores(nil)
			if err != nil {
				log.Error().Err(err).Send()
				return err
			}
			storesMap := make(map[string]string)
			sTypesMap := make(map[int]string)
			for _, s := range *storesResp {
				//lookup storetype if not already in map
				if _, ok := sTypesMap[s.CertStoreType]; !ok {
					sTypeResp, err := kfClient.GetCertificateStoreType(s.CertStoreType)
					if err != nil {
						log.Error().Err(err).Send()
						continue
					}
					sTypesMap[sTypeResp.StoreType] = sTypeResp.Name
				}
				name := fmt.Sprintf("%s/%s/%s (%s)", sTypesMap[s.CertStoreType], s.ClientMachine, s.StorePath, s.Id)
				storesMap[name] = s.Id
			}

			// get all keys from storesMap
			storeOptions := make([]string, 0, len(storesMap))
			for k := range storesMap {
				storeOptions = append(storeOptions, k)
			}

			prompt := &survey.MultiSelect{
				Message: "Choose 1 or more stores to delete:",
				Options: storeOptions,
			}

			var selectedStores []string
			askErr := survey.AskOne(prompt, &selectedStores)
			if askErr != nil {
				log.Error().Err(askErr).Send()
				return askErr
			}
			for _, s := range selectedStores {
				log.Debug().Str("storeID", storesMap[s]).Msg("Adding store to delete list")
				stores = append(stores, storesMap[s])
			}
		}

		var errs []error
		for _, st := range stores {
			_, err := kfClient.GetCertificateStoreByID(st)
			if err != nil {
				log.Error().Err(err).Send()
				errs = append(errs, fmt.Errorf("Store ID '%s': '%s'", st, err.Error()))
				continue
			}

			dErr := kfClient.DeleteCertificateStore(st)
			if dErr != nil {
				log.Error().Err(dErr).Send()
				errs = append(errs, dErr)
				continue
			}
			outputResult(fmt.Sprintf("successfully deleted store %s", st), outputFormat)
		}
		if len(errs) > 0 {
			errsStr := ""
			for _, e := range errs {
				errsStr += e.Error() + "\n\t"
			}
			return fmt.Errorf("occurred while deleting stores:\n\t%s", errsStr)
		}
		return nil
	},
}

func init() {
	var (
		storeID   string
		deleteAll bool
		inputFile string
	)
	RootCmd.AddCommand(storesCmd)
	storesCmd.AddCommand(storesListCmd)
	storesCmd.AddCommand(storesGetCmd)
	storesCmd.AddCommand(storesDeleteCmd)

	// get cmd
	storesGetCmd.Flags().StringVarP(&storeID, "id", "i", "", "ID of the certificate store to get.")

	// delete cmd
	storesDeleteCmd.Flags().StringVarP(&storeID, "id", "i", "", "ID of the certificate store to delete.")
	storesDeleteCmd.Flags().StringVarP(&inputFile, "file", "f", "", "The path to a CSV file containing the Ids of the stores to delete.")
	storesDeleteCmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Attempt to delete ALL stores.")
	storesDeleteCmd.MarkFlagsMutuallyExclusive("id", "all")

}
