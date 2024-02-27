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
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Jeffail/gabs"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

var (
	bulkStoreImportCSVHeader = []string{
		//"CertStoreType", leaving this out since each file can only import stores of a single type.  We'll get it from the parameters.
		"ContainerId",
		"ClientMachine",
		"StorePath",
		"CreateIfMissing",
		"Properties",
		"AgentId",
		"InventorySchedule.Immediate",
		"InventorySchedule.Interval.Minutes",
		"InventorySchedule.Daily.Time",
		"InventorySchedule.Weekly.Days",
		"InventorySchedule.Weekly.Time",
	}
)

var importStoresCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a file with certificate store parameters and create them in keyfactor.",
	Long:  `Tools for generating import templates and importing certificate stores`,
}

var storesCreateFromCSVCmd = &cobra.Command{
	Use:   "csv --file <file name to import> --store-type-id <store type id> --store-type-name <store type name> --results-path <filepath for results> --dry-run <check fields only>",
	Short: "Create certificate stores from CSV file.",
	Long: `Certificate stores: Will parse a CSV and attempt to create a certificate store for each row with the provided parameters.
'store-type-name' OR 'store-type-id' are required.
'file' is the path to the file to be imported.
'resultspath' is where the import results will be written to.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Specific flags
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeID, _ := cmd.Flags().GetInt("store-type-id")
		filePath, _ := cmd.Flags().GetString("file")
		outPath, _ := cmd.Flags().GetString("results-path")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Flag Checks
		inputErr := storeTypeIdentifierFlagCheck(cmd)
		if inputErr != nil {
			return inputErr
		}

		// expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		// CLI Logic
		log.Info().Msg("Importing certificate stores")
		log.Debug().Str("storeTypeName", storeTypeName).
			Int("storeTypeId", storeTypeID).
			Str("filePath", filePath).
			Str("outPath", outPath).
			Bool("dryRun", dryRun).Msg("Specific flags")

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
		inputMap, _ := csvToMap(filePath)
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
		intID, reqPropertiesForStoreType := getRequiredProperties(st, *kfClient)

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
		var (
			resultsMap  [][]string
			originalMap [][]string
		)

		errorCount := 0

		log.Info().Msgf("Processing CSV rows from file '%s'", filePath)
		for idx, row := range inFile {
			log.Debug().Msgf("Processing row '%d'", idx)
			originalMap = append(originalMap, row)

			if idx == 0 {
				// skip header row
				log.Debug().Msgf("Skipping header row")
				continue
			}
			reqJson := getJsonForRequest(headerRow, row)
			reqJson.Set(intID, "CertStoreType")

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
			mJSON := reqJson.String()
			conversionError := json.Unmarshal([]byte(mJSON), &createStoreReqParameters)

			if conversionError != nil {
				//outputError(conversionError, true, outputFormat)
				log.Error().Err(conversionError).Msgf("Unable to convert the json into the request parameters object.  %s", conversionError.Error())
				return conversionError
			}

			createStoreReqParameters.Properties = props
			log.Debug().Msgf("Request parameters: %v", createStoreReqParameters)

			// make request.
			log.Info().Msgf("Calling Command to create store from row '%d'", idx)
			res, err := kfClient.CreateStore(&createStoreReqParameters)

			if err != nil {
				log.Error().Err(err).Msgf("Error creating store from row '%d'", idx)
				resultsMap = append(resultsMap, []string{err.Error()})
				inputMap[idx-1]["Errors"] = err.Error()
				inputMap[idx-1]["Id"] = "error"
				errorCount++
			} else {
				log.Info().Msgf("Successfully created store from row '%d' as '%s'", idx, res.Id)
				resultsMap = append(resultsMap, []string{fmt.Sprintf("%s", res.Id)})
				inputMap[idx-1]["Id"] = res.Id
			}
		}

		log.Debug().Msg("Appending results to original CSV")
		for oIdx, oRow := range originalMap {
			if oIdx == 0 {
				// skip header row
				continue
			}
			// combine slices
			extendedRow := append(oRow, resultsMap[oIdx-1]...)
			originalMap[oIdx] = extendedRow
		}
		totalRows := len(resultsMap)
		totalSuccess := totalRows - errorCount
		log.Debug().Int("totalRows", totalRows).
			Int("totalSuccess", totalSuccess).Send()

		log.Info().Msgf("Writing results to file '%s'", outPath)
		//writeCsvFile(outPath, originalMap)
		mapToCSV(inputMap, outPath)
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

var storesCreateImportTemplateCmd = &cobra.Command{
	Use:   "generate-template --store-type-id <store type id> --store-type-name <store-type-name> --outpath <output file path>",
	Short: "For generating a CSV template with headers for bulk store creation.",
	Long: `kfutil stores generate-template creates a csv file containing headers for a specific cert store type.
store-type-name OR store-type-id is required.
outpath is the path the template should be written to.
Store type IDs can be found by running the "store-types" command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeID, _ := cmd.Flags().GetInt("store-type-id")
		outpath, _ := cmd.Flags().GetString("outpath")

		if noPrompt {
			inputErr := storeTypeIdentifierFlagCheck(cmd)
			if inputErr != nil {
				return inputErr
			}
		}

		// expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, clientErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
		if clientErr != nil {
			log.Error().Err(clientErr).Msg("Error initializing client")
			return clientErr
		}

		// CLI Logic
		log.Info().Msg("Generating template for certificate stores")
		log.Debug().Str("storeTypeName", storeTypeName).
			Int("storeTypeId", storeTypeID).
			Str("outpath", outpath).Msg("Specific flags")

		// Check inputs
		var (
			st    interface{}
			stErr error
		)
		var validStoreTypes []string
		var removeStoreTypes []interface{}
		if storeTypeID < 0 && storeTypeName == "" && !noPrompt {
			// prompt for store type
			validStoreTypesResp, vstErr := kfClient.ListCertificateStoreTypes()
			if vstErr != nil {
				log.Error().Err(vstErr).Msg("unable to list certificate store types")
				validStoreTypes = getValidStoreTypes("", "main")
			} else {
				for _, v := range *validStoreTypesResp {
					validStoreTypes = append(validStoreTypes, v.ShortName)
					removeStoreTypes = append(removeStoreTypes, v.ShortName)
				}
			}
			log.Info().Msg("No store type specified, prompting user to select one")
			prompt := &survey.Select{
				Message: "Choose a store type to export:",
				Options: validStoreTypes,
			}
			var selected string
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				log.Error().Err(err).Msg("user select prompt failed")
				fmt.Println(err)
			}
			log.Info().Str("storeType", selected).Msg("User selected store type")
			st = []interface{}{selected}
		} else {
			log.Debug().Msg("calling validateStoreTypeInputs()")
			st, stErr = validateStoreTypeInputs(storeTypeID, storeTypeName, outputFormat)
			log.Debug().Msg("returned from validateStoreTypeInputs()")
			if stErr != nil {
				log.Error().Err(stErr).Msg("Error validating store type inputs")
				return stErr
			}
		}
		log.Trace().Interface("st", st).Send()
		// get storetype for the list of properties
		log.Debug().Msg("calling getHeadersForStoreType()")
		intID, sTypeShortName, csvHeaders := getHeadersForStoreType(st, *kfClient)
		log.Debug().Str("shortName", sTypeShortName).Msg("returned from getHeadersForStoreType()")
		log.Debug().Int64("intID", intID).
			Interface("csvHeaders", csvHeaders).
			Send()

		if storeTypeName != "" && sTypeShortName != "" && storeTypeName != sTypeShortName {
			log.Debug().Str("storeTypeName", storeTypeName).
				Str("sTypeShortName", sTypeShortName).
				Msg("storeTypeName does not match sTypeShortName, overwriting storeTypeName with sTypeShortName")
			sTypeShortName = storeTypeName
		}

		// write csv file header row
		var filePath string
		if outpath != "" {
			filePath = outpath
		} else {
			if sTypeShortName != "" {
				filePath = fmt.Sprintf("%s_bulk_import_template.%s", sTypeShortName, "csv")
			} else {
				filePath = fmt.Sprintf("%s_bulk_import_template_%d.%s", "createstores", intID, "csv")
			}
		}
		log.Debug().Str("filePath", filePath).Msg("Writing template file")

		var csvContent [][]string
		var row []string

		log.Debug().Msg("Writing header row")
		for k, v := range csvHeaders {
			log.Trace().Int("index", k).
				Str("header", v).
				Send()
			row = append(row, v)
		}
		csvContent = append(csvContent, row)

		log.Info().Str("filePath", filePath).Msg("Writing template file")
		csvWriteErr := writeCsvFile(filePath, csvContent)
		if csvWriteErr != nil {
			log.Error().Err(csvWriteErr).Msg("Error writing csv file")
			return csvWriteErr
		}
		log.Info().Str("filePath", filePath).Msg("Template file written")
		outputResult(fmt.Sprintf("Template file for store type with id %d written to %s", intID, filePath), outputFormat)
		return nil
	},
}

var storesExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export existing defined certificate stores by type or store Id.",
	Long:  "Export the parameter values of defined certificate stores either by type or a specific store by Id. These parameters are stored in CSV for importing later.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeID, _ := cmd.Flags().GetInt("store-type-id")
		outpath, _ := cmd.Flags().GetString("outpath")

		inputErr := storeTypeIdentifierFlagCheck(cmd)
		if inputErr != nil {
			return inputErr
		}

		// expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		// CLI Logic
		log.Info().
			Str("storeTypeName", storeTypeName).
			Int("storeTypeId", storeTypeID).
			Str("outpath", outpath).
			Msg("Exporting certificate stores of specified type to CSV")

		// Check inputs
		st, stErr := validateStoreTypeInputs(storeTypeID, storeTypeName, outputFormat)
		if stErr != nil {
			log.Error().Err(stErr).Msg("validating store type inputs")
			return stErr
		}

		// get storetype for the list of properties
		log.Debug().Msg("calling getHeadersForStoreType()")
		storeType, err := kfClient.GetCertificateStoreType(st)
		log.Debug().Msg("returned from getHeadersForStoreType()")
		log.Trace().Interface("storeType", storeType).Send()
		if err != nil {
			log.Error().Err(err).Msg("retrieving store type")
			return err
		}

		log.Debug().Msg("calling getHeadersForStoreType()")
		typeID, _, csvHeaders := getHeadersForStoreType(st, *kfClient)
		log.Debug().Msg("returned from getHeadersForStoreType()")

		query := map[string]interface{}{"Category": typeID}
		log.Debug().Interface("query", query).Msg("calling ListCertificateStores()")
		storeList, lErr := kfClient.ListCertificateStores(&query)
		log.Debug().Msg("returned from ListCertificateStores()")
		log.Trace().Interface("storeList", storeList).Send()
		if lErr != nil {
			log.Error().Err(lErr).
				Int64("typeId", typeID).
				Msg("listing stores of type")
			return lErr
		}

		// add Id header to csvHeaders at -1
		log.Debug().Msg("adding Id header to csvHeaders")
		csvHeaders[len(csvHeaders)] = "Id"
		log.Trace().Interface("csvHeaders", csvHeaders).Send()
		csvData := make(map[string]map[string]interface{}, len(*storeList))

		log.Debug().Msg("iterating through stores")
		for _, listedStore := range *storeList {
			if listedStore.CertStoreType != int(typeID) {
				log.Debug().Int("listedStore.CertStoreType", listedStore.CertStoreType).
					Msg("skipping store")
				continue
			}
			log.Debug().Str("listedStore.Id", listedStore.Id).
				Msg("calling GetCertificateStoreByID()")
			store, err := kfClient.GetCertificateStoreByID(listedStore.Id)
			log.Debug().Msg("returned from GetCertificateStoreByID()")
			log.Trace().Interface("store", store).Send()
			if err != nil {
				log.Error().Err(err).Msg("retrieving store by id")
				return err
			}

			// populate store data into csv
			log.Debug().Str("store.Id", store.Id).
				Int("store.ContainerId", store.ContainerId).
				Str("store.ClientMachine", store.ClientMachine).
				Str("store.StorePath", store.StorePath).
				Bool("store.CreateIfMissing", store.CreateIfMissing).
				Str("store.AgentId", store.AgentId).
				Msg("populating store data into csv")

			csvData[store.Id] = map[string]interface{}{
				"Id":              store.Id,
				"ContainerId":     store.ContainerId,
				"ClientMachine":   store.ClientMachine,
				"StorePath":       store.StorePath,
				"CreateIfMissing": store.CreateIfMissing,
				"AgentId":         store.AgentId,
			}

			log.Debug().Msg("checking for InventorySchedule")
			if store.InventorySchedule.Immediate != nil {
				log.Debug().Msg("found InventorySchedule.Immediate")
				csvData[store.Id]["InventorySchedule.Immediate"] = store.InventorySchedule.Immediate
			}
			if store.InventorySchedule.Interval != nil {
				log.Debug().Msg("found InventorySchedule.Interval")
				csvData[store.Id]["InventorySchedule.Interval.Minutes"] = store.InventorySchedule.Interval.Minutes
			}
			if store.InventorySchedule.Daily != nil {
				log.Debug().Msg("found InventorySchedule.Daily")
				csvData[store.Id]["InventorySchedule.Daily.Time"] = store.InventorySchedule.Daily.Time
			}

			log.Debug().Msg("checking Properties")
			for name, prop := range store.Properties {
				log.Debug().Str("name", name).
					Interface("prop", prop).
					Msg("adding to properties CSV data")
				if name != "ServerUsername" && name != "ServerPassword" { // Don't add ServerUsername and ServerPassword to properties as they can't be exported via API
					csvData[store.Id]["Properties."+name] = prop
				}
			}

			//// conditionally set secret values
			//if storeType.PasswordOptions.StoreRequired {
			//	log.Debug().Str("storePassword", hashSecretValue(store.Password.Value)).
			//		Msg("setting store password")
			//
			//	//csvData[store.Id]["Password"] = parseSecretField(store.Password) // todo: find parseSecretField
			//	csvData[store.Id]["Password"] = store.Password.Value
			//}
			//// add ServerUsername and ServerPassword Properties if required for type
			//if storeType.ServerRequired {
			//	log.Debug().Interface("store.ServerUsername", store.Properties["ServerUsername"]).
			//		Str("store.Password", hashSecretValue(store.Password.Value)).
			//		Msg("setting store.ServerUsername")
			//	//csvData[store.Id]["Properties.ServerUsername"] = parseSecretField(store.Properties["ServerUsername"]) // todo: find parseSecretField
			//	//csvData[store.Id]["Properties.ServerPassword"] = parseSecretField(store.Properties["ServerPassword"]) // todo: find parseSecretField
			//	csvData[store.Id]["Properties.ServerUsername"] = store.Properties["ServerUsername"]
			//	csvData[store.Id]["Properties.ServerPassword"] = store.Properties["ServerPassword"]
			//}
		}

		// write csv file header row
		var filePath string
		if outpath != "" {
			filePath = outpath
		} else {
			filePath = fmt.Sprintf("export_stores_%d.%s", &typeID, "csv")
		}
		log.Debug().Str("filePath", filePath).Msg("Writing export file")

		var csvContent [][]string
		headerRow := make([]string, len(csvHeaders))

		log.Debug().Msg("Writing header row")
		for k, v := range csvHeaders {
			headerRow[k] = v
		}
		log.Trace().Interface("row", headerRow).Send()
		csvContent = append(csvContent, headerRow)
		index := 1

		log.Debug().Msg("Writing data rows")
		for _, data := range csvData {
			log.Debug().Int("index", index).Msg("processing data row")
			row := make([]string, len(csvHeaders)) // reset row
			for i, header := range csvHeaders {
				log.Trace().Int("index", i).
					Str("header", header).
					Msg("processing header")
				if data[header] != nil {
					if s, ok := data[header].(string); ok {
						log.Trace().Str("s", s).
							Msg("setting row value")
						row[i] = s
					} else {
						log.Trace().Interface("data[header]", data[header]).
							Msg("marshalling data[header]")
						strData, _ := json.Marshal(data[header])
						row[i] = string(strData)
						log.Trace().Int("index", i).
							Str("row[i]", row[i]).
							Msg("setting row value")
					}
				}
			}
			log.Debug().Msg("appending row to csvContent")
			csvContent = append(csvContent, row)
			index++
			log.Debug().Msg("row appended to csvContent")
		}

		writeCsvFile(filePath, csvContent)

		fmt.Printf("\nStores exported for store type with id %d written to %s\n", typeID, filePath)

		return nil
	},
}

func getHeadersForStoreType(id interface{}, kfClient api.Client) (int64, string, map[int]string) {
	csvHeaders := make(map[int]string)

	//check if interface is a slice of interfaces
	if _, ok := id.([]interface{}); ok {
		id = id.([]interface{})[0]
		log.Debug().Interface("id", id).Msg("id is a slice of interfaces, setting id to first element")
	}

	storeType, err := kfClient.GetCertificateStoreType(id)
	if err != nil {
		log.Printf("Error: %s", err)
		fmt.Printf("Error: %s\n", err)
		panic("error retrieving store type")
	}
	output, jErr := json.Marshal(storeType)
	if jErr != nil {
		log.Printf("Error: %s", jErr)
	}

	dec := json.NewDecoder(bytes.NewReader(output))
	dec.UseNumber()

	jsonParsedObj, _ := gabs.ParseJSONDecoder(dec)

	// iterate through properties and determine header positions
	properties, _ := jsonParsedObj.S("Properties").Children()
	offset := 0

	for idx, name := range bulkStoreImportCSVHeader {
		if name == "Properties" {
			propIdx := idx
			for pIdx, property := range properties {
				loc := propIdx + pIdx
				pName := "Properties." + property.S("Name").Data().(string)
				csvHeaders[loc] = pName
			}
			offset = len(properties) - 1
		} else {
			csvHeaders[idx+offset] = name
		}
	}
	// add Password field if flag was set
	if storeType.PasswordOptions.StoreRequired {
		csvHeaders[len(csvHeaders)] = "Password"
	}
	intId, _ := jsonParsedObj.S("StoreType").Data().(json.Number).Int64()
	shortName, snOk := jsonParsedObj.S("ShortName").Data().(string)
	if !snOk {
		log.Printf("Error: %s", "unable to retrieve store type id or short name")
		fmt.Printf("Error: %s\n", "unable to retrieve store type id or short name")
		shortName = ""
	}
	return intId, shortName, csvHeaders
}

func getRequiredProperties(id interface{}, kfClient api.Client) (int64, []string) {

	storeType, err := kfClient.GetCertificateStoreType(id)
	if err != nil {
		log.Printf("Error: %s", err)
		fmt.Printf("Error: %s\n", err)
		panic("error retrieving store type")
	}

	output, jErr := json.Marshal(storeType)
	if jErr != nil {
		log.Printf("Error: %s", jErr)
	}

	dec := json.NewDecoder(bytes.NewReader(output))
	dec.UseNumber()

	//fmt.Printf("\n %s \n", output)
	jsonParsedObj, _ := gabs.ParseJSONDecoder(dec)

	//iterate through properties and determine header positions

	properties, _ := jsonParsedObj.S("Properties").Children()
	reqProps := make([]string, 0)
	for _, prop := range properties {
		if prop.S("Required").Data() == true {
			name := prop.S("Name")
			reqProps = append(reqProps, name.Data().(string))
		}
	}
	intId, _ := jsonParsedObj.S("StoreType").Data().(json.Number).Int64()

	return intId, reqProps
}

func unmarshalPropertiesString(properties string) map[string]interface{} {
	if properties != "" {
		// First, unmarshal JSON properties string to []interface{}
		var tempInterface interface{}
		if err := json.Unmarshal([]byte(properties), &tempInterface); err != nil {
			return make(map[string]interface{})
		}
		// Then, iterate through each key:value pair and serialize into map[string]string
		newMap := make(map[string]interface{})
		for key, value := range tempInterface.(map[string]interface{}) {
			newMap[key] = value
		}
		return newMap
	}

	return make(map[string]interface{})
}

//func parseSecretField(secretField interface{}) interface{} {
//	var secret api.StorePasswordConfig
//	secretByte, errors := json.Marshal(secretField)
//	if errors != nil {
//		log.Printf("Error in Marshalling: %s", errors)
//		fmt.Printf("Error in Marshalling: %s\n", errors)
//		panic("error marshalling secret field as StorePasswordConfig")
//	}
//
//	errors = json.Unmarshal(secretByte, &secret)
//	if errors != nil {
//		log.Printf("Error in Unmarshalling: %s", errors)
//		fmt.Printf("Error in Unmarshalling: %s\n", errors)
//		panic("error unmarshalling secret field as StorePasswordConfig")
//	}
//
//	if secret.IsManaged {
//		params := make(map[string]string)
//		for _, p := range *secret.ProviderTypeParameterValues {
//			params[*p.ProviderTypeParam.Name] = *p.Value
//		}
//		return map[string]interface{}{
//			"Provider":   secret.ProviderId,
//			"Parameters": params,
//		}
//	} else {
//		if secret.Value != "" {
//			return map[string]string{
//				"SecretValue": secret.Value,
//			}
//		} else {
//			return map[string]*string{
//				"SecretValue": nil,
//			}
//		}
//	}
//}

func getJsonForRequest(headerRow []string, row []string) *gabs.Container {
	log.Debug().Msgf("Getting JSON for request")
	reqJson := gabs.New()
	for hIdx, header := range headerRow {
		log.Debug().Msgf("Processing header '%s'", header)
		if strings.ToUpper(row[hIdx]) == "TRUE" {
			reqJson.Set(true, strings.Split(header, ".")...)
		} else if strings.ToUpper(row[hIdx]) == "FALSE" {
			reqJson.Set(false, strings.Split(header, ".")...)
		} else if row[hIdx] != "" {
			tryInt, errors := strconv.Atoi(row[hIdx])
			if errors == nil {
				reqJson.Set(tryInt, strings.Split(header, ".")...)
			} else {
				var obj map[string]interface{}
				errors = json.Unmarshal([]byte(row[hIdx]), &obj)
				if errors == nil {
					reqJson.Set(obj, strings.Split(header, ".")...)
				} else {
					reqJson.Set(row[hIdx], strings.Split(header, ".")...)
				}
			}
		}
	}
	//fmt.Printf("[DEBUG] get JSON for create store request: %s", reqJson.String())
	log.Debug().Msgf("JSON for create store request: %s", reqJson.String())
	return reqJson
}

func writeCsvFile(outpath string, rows [][]string) error {
	log.Debug().Msgf("Writing CSV file '%s'", outpath)
	csvFile, err := os.Create(outpath)
	if err != nil {
		//log.Fatal("Cannot create file", err)
		log.Error().Err(err).Msgf("Cannot create file '%s'", outpath)
		return err
	}
	csvWriter := csv.NewWriter(csvFile)

	//fmt.Println()
	for _, v := range rows {
		//fmt.Println()
		wErr := csvWriter.Write(v)
		if wErr != nil {
			//fmt.Printf("%s", wErr)
			outputError(wErr, false, "")
			//log.Printf("[ERROR] Error writing row to CSV: %s", wErr)
			log.Error().Err(wErr).Msgf("Error writing row to CSV: %v", v)
		}
		csvWriter.Flush()
	}

	ioErr := csvFile.Close()
	if ioErr != nil {
		//fmt.Println(ioErr)
		outputError(ioErr, false, "")
		//log.Printf("[ERROR] Error closing file: %s", ioErr)
		log.Error().Err(ioErr).Msgf("Error closing file")
		return ioErr
	}
	return nil
}

func validateStoreTypeInputs(storeTypeID int, storeTypeName string, outputFormat string) (interface{}, error) {
	log.Debug().Int("storeTypeId", storeTypeID).
		Str("storeTypeName", storeTypeName).
		Msg("Validating store type inputs")
	var st interface{}
	// Check inputs
	switch {
	case storeTypeID < 0 && storeTypeName == "":
		noIdentifierError := fmt.Errorf("'store-type-id' or 'store-type-name' must be provided")
		log.Error().Err(noIdentifierError).Send()
		return "", noIdentifierError
	case storeTypeID >= 0 && storeTypeName != "":
		conflictingIdentifiersError := fmt.Errorf("'store-type-id' and 'store-type-name' are mutually exclusive")
		log.Error().Err(conflictingIdentifiersError).Send()
		return "", conflictingIdentifiersError
	case storeTypeID >= 0:
		st = storeTypeID
	case storeTypeName != "":
		st = storeTypeName
	default:
		log.Error().Err(InvalidInputError).Send()
		return "", InvalidInputError
	}
	return st, nil
}

func init() {

	var (
		storeTypeName string
		storeTypeId   int
		outPath       string
		file          string
		resultsPath   string
	)

	storesCmd.AddCommand(importStoresCmd)
	storesCmd.AddCommand(storesExportCmd)
	importStoresCmd.AddCommand(storesCreateImportTemplateCmd)
	importStoresCmd.AddCommand(storesCreateFromCSVCmd)

	storesCreateImportTemplateCmd.Flags().StringVarP(&storeTypeName, "store-type-name", "n", "", "The name of the cert store type for the template.  Use if store-type-id is unknown.")
	storesCreateImportTemplateCmd.Flags().IntVarP(&storeTypeId, "store-type-id", "i", -1, "The ID of the cert store type for the template.")
	storesCreateImportTemplateCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
		"Path and name of the template file to generate.. If not specified, the file will be written to the current directory.")
	storesCreateImportTemplateCmd.MarkFlagsMutuallyExclusive("store-type-name", "store-type-id")

	storesCreateFromCSVCmd.Flags().StringVarP(&storeTypeName, "store-type-name", "n", "", "The name of the cert store type.  Use if store-type-id is unknown.")
	storesCreateFromCSVCmd.Flags().IntVarP(&storeTypeId, "store-type-id", "i", -1, "The ID of the cert store type for the stores.")
	storesCreateFromCSVCmd.Flags().StringVarP(&file, "file", "f", "", "CSV file containing cert stores to create.")
	storesCreateFromCSVCmd.MarkFlagRequired("file")
	storesCreateFromCSVCmd.Flags().BoolP("dry-run", "d", false, "Do not import, just check for necessary fields.")
	storesCreateFromCSVCmd.Flags().StringVarP(&resultsPath, "results-path", "o", "", "CSV file containing cert stores to create. defaults to <imported file name>_results.csv")

	storesExportCmd.Flags().StringVarP(&storeTypeName, "store-type-name", "n", "", "The name of the cert store type for the template.  Use if store-type-id is unknown.")
	storesExportCmd.Flags().IntVarP(&storeTypeId, "store-type-id", "i", -1, "The ID of the cert store type for the template.")
	storesExportCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
		"Path and name of the template file to generate.. If not specified, the file will be written to the current directory.")
	storesExportCmd.MarkFlagsMutuallyExclusive("store-type-name", "store-type-id")

}
