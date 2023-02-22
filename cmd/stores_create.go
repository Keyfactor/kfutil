// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
)

type NewStoreCSVEntry struct {
	Id                string `json:"Id"`
	CertStoreType     string `json:"CertStoreType"`
	ClientMachine     string `json:"ClientMachine"`
	Storepath         string `json:"StorePath"`
	Properties        string `json:"Properties"`
	Approved          bool   `json:"Approved"`
	CreateIfMissing   bool   `json:"CreateIfMissing"`
	AgentId           string `json:"AgentId"`
	InventorySchedule string `json:"InventorySchedule"`
}

var (
	baseFieldNames = []string{
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

//storesCreateCmd is the action for importing a csv file for bulk creating stores

var storesCreateCmd = &cobra.Command{
	Use:   "create --file <file name to import> --store-type-id <store type id> --store-type-name <store type name> --results-path <filepath for results> --dry-run <check fields only>",
	Short: "Create certificate stores",
	Long: `Certificate stores: Will parse a CSV and attempt to create a certificate store for each row with the provided parameters.
store-type-name OR store-type-id is required.
file is the path to the file to be imported.
resultspath is where the import results will be written to.`,
	Run: func(cmd *cobra.Command, args []string) {
		kfClient, _ := initClient()
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeId, _ := cmd.Flags().GetInt("store-type-id")
		filePath, _ := cmd.Flags().GetString("file")
		outPath, _ := cmd.Flags().GetString("results-path")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		var st interface{}

		// Check inputs
		if storeTypeId < 0 && storeTypeName == "" {
			log.Printf("Error: ID must be a positive integer.")
			fmt.Printf("Error: ID must be a positive integer.\n")
			return
		} else if storeTypeId >= 0 && storeTypeName != "" {
			log.Printf("Error: ID and Name are mutually exclusive.")
			fmt.Printf("Error: ID and Name are mutually exclusive.\n")
			return
		} else if storeTypeId >= 0 {
			st = storeTypeId
		} else if storeTypeName != "" {
			st = storeTypeName
		} else {
			log.Printf("Error: Invalid input.")
			fmt.Printf("Error: Invalid input.\n")
			return
		}

		if outPath == "" {
			outPath = strings.Split(filePath, ".")[0] + "_results.csv"
		}

		log.Printf("[DEBUG] storesFile: %s", filePath)
		log.Printf("[DEBUG] output path: %s", outPath)
		log.Printf("[DEBUG] dryRun: %t", dryRun)
		log.Printf("[DEBUG] storeTypeId: %d", storeTypeId)

		// get file headers
		csvFile, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %s", err)
			log.Fatalf("Error opening CSV file: %s", err)
		}

		inFile, cErr := csv.NewReader(csvFile).ReadAll()
		if cErr != nil {
			log.Fatalf("Error reading CSV file: %s", cErr)
		}

		// check for minimum necessary required fields for creating certificate stores

		intId, reqPropertiesForStoreType := getRequiredProperties(st, *kfClient)

		// if not present in header, throw error.
		headerRow := inFile[0]

		missingFields := make([]string, 0)

		//check fields
		for _, reqField := range reqPropertiesForStoreType {
			exists := false
			for _, headerField := range headerRow {
				if strings.EqualFold(headerField, "Properties."+reqField) {
					exists = true
					continue
				}
			}
			if !exists {
				missingFields = append(missingFields, reqField)
			}
		}

		if len(missingFields) > 0 {
			fmt.Printf("Missing Required Fields in headers: %v", missingFields)
			log.Fatalf("Missing Required Fields in headers: %v", missingFields)
			return
		}

		//foreach row attempt to create the store

		//track errors
		resultsMap := make(map[int]string)
		originalMap := make(map[int][]string)
		errorCount := 0

		for idx, row := range inFile {
			originalMap[idx] = row

			if idx == 0 {
				continue
			}
			reqJson := getJsonForRequest(headerRow, row)
			reqJson.Set(intId, "CertStoreType")

			var createStoreReqParameters api.CreateStoreFctArgs
			props := unmarshalPropertiesString(reqJson.S("Properties").String())
			reqJson.Delete("Properties")
			mJson, _ := reqJson.MarshalJSON()
			conversionError := json.Unmarshal(mJson, &createStoreReqParameters)

			if conversionError != nil {
				fmt.Printf("Unable to convert the json into the request parameters object.  %s", conversionError.Error())
				return
			}

			createStoreReqParameters.Properties = props

			//make request.
			res, err := kfClient.CreateStore(&createStoreReqParameters)

			if err != nil {
				resultsMap[idx] = err.Error()
				errorCount++
			} else {
				resultsMap[idx] = fmt.Sprintf("Success.  CertStoreId = %s", res.Id)
			}
		}

		for oIdx, oRow := range originalMap {
			extendedRow := append(oRow, resultsMap[oIdx])
			originalMap[oIdx] = extendedRow
		}
		totalRows := len(resultsMap)
		totalSuccess := totalRows - errorCount

		writeCsvFile(outPath, originalMap)
		fmt.Printf("\n%d records processed.", totalRows)
		if totalSuccess > 0 {
			fmt.Printf("\n%d certificate stores successfully created.", totalSuccess)
		}
		if errorCount > 0 {
			fmt.Printf("\n%d rows had errors.", errorCount)
		}
		fmt.Printf("\nImport results written to %s\n\n", outPath)
	}}

func getJsonForRequest(headerRow []string, row []string) *gabs.Container {
	reqJson := gabs.New()

	for hIdx, header := range headerRow {

		if strings.ToUpper(row[hIdx]) == "TRUE" {
			reqJson.Set(true, strings.Split(header, ".")...)
		} else if strings.ToUpper(row[hIdx]) == "FALSE" {
			reqJson.Set(false, strings.Split(header, ".")...)
		} else if row[hIdx] != "" {
			reqJson.Set(row[hIdx], strings.Split(header, ".")...)
		}
	}
	//fmt.Printf("[DEBUG] get JSON for create store request: %s", reqJson.String())
	return reqJson
}

func writeCsvFile(outpath string, rows map[int][]string) {

	csvFile, err := os.Create(outpath)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	csvWriter := csv.NewWriter(csvFile)

	fmt.Println()

	for _, v := range rows {
		fmt.Println()
		wErr := csvWriter.Write(v)
		if wErr != nil {
			fmt.Printf("%s", wErr)
			log.Printf("[ERROR] Error writing row to CSV: %s", wErr)
		}
		csvWriter.Flush()
	}

	ioErr := csvFile.Close()
	if ioErr != nil {
		fmt.Println(ioErr)
		log.Printf("[ERROR] Error closing file: %s", ioErr)
	}
}

var storesCreateTemplateCmd = &cobra.Command{
	Use:   "generate-template --store-type-id <store type id> --store-type-name <store-type-name> --outpath <output file path>",
	Short: "For generating a CSV template with headers for bulk store creation.",
	Long: `kfutil stores generate-template creates a csv file containing headers for a specific cert store type.
store-type-name OR store-type-id is required.
outpath is the path the template should be written to.
Store type IDs can be found by running the "store-types" command.`,
	Run: func(cmd *cobra.Command, args []string) {
		kfClient, _ := initClient()
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeId, _ := cmd.Flags().GetInt("store-type-id")
		outpath, _ := cmd.Flags().GetString("outpath")

		//fmt.Printf("beginning store type id check.. id = %d, name = %s", storeTypeId, storeTypeName)

		var st interface{}
		// Check inputs
		if storeTypeId < 0 && storeTypeName == "" {
			log.Printf("Error: ID must be a positive integer.")
			fmt.Printf("Error: ID must be a positive integer.\n")
			return
		} else if storeTypeId >= 0 && storeTypeName != "" {
			log.Printf("Error: ID and Name are mutually exclusive.")
			fmt.Printf("Error: ID and Name are mutually exclusive.\n")
			return
		} else if storeTypeId >= 0 {
			st = storeTypeId
		} else if storeTypeName != "" {
			st = storeTypeName
		} else {
			log.Printf("Error: Invalid input.")
			fmt.Printf("Error: Invalid input.\n")
			return
		}

		// get storetype for the list of properties
		intId, csvHeaders := getHeadersForStoreType(st, *kfClient)

		// write csv file header row
		var filePath string
		if outpath != "" {
			filePath = outpath
		} else {
			filePath = fmt.Sprintf("%s_template_%d.%s", "createstores", intId, "csv")
		}

		csvContent := make(map[int][]string)

		row := make([]string, len(csvHeaders))

		for k, v := range csvHeaders {
			row[k] = v
		}
		csvContent[0] = row

		writeCsvFile(filePath, csvContent)

		fmt.Printf("\nTemplate file for store type with id %d written to %s\n", intId, filePath)
	}}

func getHeadersForStoreType(id interface{}, kfClient api.Client) (int64, map[int]string) {
	csvHeaders := make(map[int]string)

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
	properties := jsonParsedObj.S("Properties").Children()
	offset := 0

	for idx, name := range baseFieldNames {
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
	intId, _ := jsonParsedObj.S("StoreType").Data().(json.Number).Int64()
	return intId, csvHeaders
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

	properties := jsonParsedObj.S("Properties").Children()
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

func unmarshalPropertiesString(properties string) map[string]string {
	if properties != "" {
		// First, unmarshal JSON properties string to []interface{}
		var tempInterface interface{}
		if err := json.Unmarshal([]byte(properties), &tempInterface); err != nil {
			return make(map[string]string)
		}
		// Then, iterate through each key:value pair and serialize into map[string]string
		newMap := make(map[string]string)
		for key, value := range tempInterface.(map[string]interface{}) {
			newMap[key] = value.(string)
		}
		return newMap
	}

	return make(map[string]string)
}

//command initialization

var (
	storeTypeName string
	storeTypeId   int
	outPath       string
	file          string
	resultsPath   string
)

func init() {
	storesCmd.AddCommand(importStoresCmd)
	importStoresCmd.AddCommand(storesCreateTemplateCmd)
	importStoresCmd.AddCommand(storesCreateCmd)

	storesCreateTemplateCmd.Flags().StringVarP(&storeTypeName, "store-type-name", "n", "", "The name of the cert store type for the template.  Use if store-type-id is unknown.")
	storesCreateTemplateCmd.Flags().IntVarP(&storeTypeId, "store-type-id", "i", -1, "The ID of the cert store type for the template.")
	storesCreateTemplateCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
		"Path and name of the template file to generate.. If not specified, the file will be written to the current directory.")
	storesCreateTemplateCmd.MarkFlagsMutuallyExclusive("store-type-name", "store-type-id")

	storesCreateCmd.Flags().StringVarP(&storeTypeName, "store-type-name", "n", "", "The name of the cert store type.  Use if store-type-id is unknown.")
	storesCreateCmd.Flags().IntVarP(&storeTypeId, "store-type-id", "i", -1, "The ID of the cert store type for the stores.")
	storesCreateCmd.Flags().StringVarP(&file, "file", "f", "", "CSV file containing cert stores to create.")
	storesCreateCmd.MarkFlagRequired("file")
	storesCreateCmd.Flags().BoolP("dry-run", "d", false, "Do not import, just check for necessary fields.")
	storesCreateCmd.Flags().StringVarP(&resultsPath, "results-path", "o", "", "CSV file containing cert stores to create. defaults to <imported file name>_results.csv")
}
