// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
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
		_, err := kfClient.GetCertificateStoreByID(storeID)
		if err != nil {
			log.Error().Err(err).Send()
			return err
		}

		dErr := kfClient.DeleteCertificateStore(storeID)
		if dErr != nil {
			log.Error().Err(dErr).Send()
			return dErr
		}
		outputResult(fmt.Sprintf("successfully deleted store %s", storeID), outputFormat)
		return nil
	},
}

var storesImportCmd = &cobra.Command{
	Use:   "import --file <file name to import> --store-type-id <store type id> --store-type-name <store type name> --results-path <filepath for results> --dry-run <check fields only>",
	Short: "Create certificate stores from CSV file.",
	Long: `Certificate stores: Will parse a CSV and attempt to create a certificate store for each row with the provided parameters.
'store-type-name' OR 'store-type-id' are required.
'file' is the path to the file to be imported.
'resultspath' is where the import results will be written to.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg("Importing certificate stores")

		// Specific flags
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeID, _ := cmd.Flags().GetInt("store-type-id")
		filePath, _ := cmd.Flags().GetString("file")
		outPath, _ := cmd.Flags().GetString("results-path")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		log.Debug().Str("storeTypeName", storeTypeName).
			Int("storeTypeId", storeTypeID).
			Str("filePath", filePath).
			Str("outPath", outPath).
			Bool("dryRun", dryRun).Msg("Specific flags")

		// expEnabled checks
		isExperimental := true
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		// CLI Logic

		// Check inputs
		st, stErr := validateStoreTypeInputs(storeTypeID, storeTypeName, outputFormat)
		if stErr != nil {
			log.Error().Err(stErr).Msg("Error validating store type inputs")
			return stErr
		}

		if outPath == "" {
			outPath = strings.Split(filePath, ".")[0] + "_results.csv" // todo: make this configurable
		}

		log.Debug().Str("filePath", filePath).
			Str("outPath", outPath).
			Bool("dryRun", dryRun).
			Int("storeTypeId", storeTypeID).Send()

		// get file headers
		log.Info().Str("filePath", filePath).
			Msg("Opening file")
		csvFile, err := os.Open(filePath)
		if err != nil {
			log.Error().Err(err).Msgf("unable to open file: '%s'", filePath)
			//outputError(err, true, outputFormat)
			cmd.SilenceUsage = true
			return err
		}

		// read file
		log.Info().Msgf("Reading file '%s' as CSV", filePath)
		inFile, cErr := csv.NewReader(csvFile).ReadAll()
		if cErr != nil {
			log.Error().Err(cErr).
				Str("filePath", filePath).
				Msg("unable to read file")
			//outputError(cErr, true, outputFormat)
			cmd.SilenceUsage = true
			return cErr
		}

		// check for minimum necessary required fields for creating certificate stores
		log.Info().Msgf("Checking for minimum required fields for creating certificate stores")
		intId, reqPropertiesForStoreType := getRequiredProperties(st, *kfClient)

		// if not present in header, throw error.
		headerRow := inFile[0]
		log.Debug().Msgf("Header row: %v", headerRow)
		missingFields := make([]string, 0)

		//check fields
		for _, reqField := range reqPropertiesForStoreType {
			exists := false
			for _, headerField := range headerRow {
				log.Debug().Msgf("Checking for required field %s in header '%s'", reqField, headerField)
				if strings.EqualFold(headerField, "Properties."+reqField) {
					log.Debug().Msgf("Found required field %s in header '%s'", reqField, headerField)
					exists = true
					continue
				}
			}
			if !exists {
				log.Debug().Msgf("Missing required field '%s'", reqField)
				missingFields = append(missingFields, reqField)
			}
		}

		if len(missingFields) > 0 {
			missingFieldsError := fmt.Errorf("missing required fields in headers: '%v'", missingFields)
			//fmt.Printf("Missing Required Fields in headers: %v", missingFields)
			log.Error().Err(missingFieldsError).Send()
			//outputError(missingFieldsError, true, outputFormat)
			cmd.SilenceUsage = true
			return missingFieldsError
		}

		//foreach row attempt to create the store
		//track errors
		resultsMap := make(map[int]string)
		originalMap := make(map[int][]string)
		errorCount := 0

		log.Info().Msgf("Processing CSV rows from file '%s'", filePath)
		for idx, row := range inFile {
			log.Debug().Msgf("Processing row '%d'", idx)
			originalMap[idx] = row

			if idx == 0 {
				// skip header row
				log.Debug().Msgf("Skipping header row")
				continue
			}
			reqJson := getJsonForRequest(headerRow, row)
			reqJson.Set(intId, "CertStoreType")

			// cannot send in 0 as ContainerId, need to omit
			containerId, _ := strconv.Atoi(reqJson.S("ContainerId").String())
			if containerId == 0 {
				log.Debug().Msgf("ContainerId is 0, omitting from request")
				reqJson.Set(nil, "ContainerId")
			}
			log.Debug().Msgf("Request JSON: %s", reqJson.String())

			// parse properties
			var createStoreReqParameters api.CreateStoreFctArgs
			props := unmarshalPropertiesString(reqJson.S("Properties").String())
			reqJson.Delete("Properties") // todo: why is this deleting the properties from the request json?
			mJson := reqJson.String()
			conversionError := json.Unmarshal([]byte(mJson), &createStoreReqParameters)

			if conversionError != nil {
				//outputError(conversionError, true, outputFormat)
				cmd.SilenceUsage = true
				log.Error().Err(conversionError).Msgf("Unable to convert the json into the request parameters object.  %s", conversionError.Error())
				return conversionError
			}

			createStoreReqParameters.Properties = props
			log.Debug().Msgf("Request parameters: %v", createStoreReqParameters)

			//make request.
			log.Info().Msgf("Calling Command to create store from row '%d'", idx)
			res, err := kfClient.CreateStore(&createStoreReqParameters)

			if err != nil {
				log.Error().Err(err).Msgf("Error creating store from row '%d'", idx)
				resultsMap[idx] = err.Error()
				errorCount++
			} else {
				log.Info().Msgf("Successfully created store from row '%d' as '%s'", idx, res.Id)
				resultsMap[idx] = fmt.Sprintf("Success.  CertStoreId = %s", res.Id)
			}
		}

		log.Debug().Msg("Appending results to original CSV")
		for oIdx, oRow := range originalMap {
			extendedRow := append(oRow, resultsMap[oIdx])
			originalMap[oIdx] = extendedRow
		}
		totalRows := len(resultsMap)
		totalSuccess := totalRows - errorCount
		log.Debug().Int("totalRows", totalRows).
			Int("totalSuccess", totalSuccess).Send()

		log.Info().Msgf("Writing results to file '%s'", outPath)
		writeCsvFile(outPath, originalMap)
		log.Info().Int("totalRows", totalRows).
			Int("totalSuccesses", totalSuccess).
			Int("errorCount", errorCount).
			Msgf("Wrote results to file '%s'", outPath)
		outputResult(fmt.Sprintf("%d records processed.", totalRows), outputFormat)
		if totalSuccess > 0 {
			//fmt.Printf("\n%d certificate stores successfully created.", totalSuccess)
			outputResult(fmt.Sprintf("%d certificate stores successfully created.", totalSuccess), outputFormat)
		}
		if errorCount > 0 {
			//fmt.Printf("\n%d rows had errors.", errorCount)
			outputResult(fmt.Sprintf("%d rows had errors.", errorCount), outputFormat)
		}
		//fmt.Printf("\nImport results written to %s\n\n", outPath)
		outputResult(fmt.Sprintf("Import results written to %s", outPath), outputFormat)
		return nil
	}}

func init() {
	var (
		storeID string
	)
	RootCmd.AddCommand(storesCmd)
	storesCmd.AddCommand(storesListCmd)
	storesCmd.AddCommand(storesGetCmd)

	// get cmd
	storesGetCmd.Flags().StringVarP(&storeID, "id", "i", "", "ID of the certificate store to get.")

}
