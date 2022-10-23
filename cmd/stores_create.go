/*
Copyright � 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Jeffail/gabs/v2"
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

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a file with certificate store parameters and create them in keyfactor.",
	Long:  `Tools for generating import templates and importing certificate stores`,
}

// storesCreateCmd is the action for importing a csv file for bulk creating stores
// var storesCreateCmd = &cobra.Command{
// 	Use:   "create -file [-out]",
// 	Short: "Create certificate stores",
// 	Long:  `Certificate stores: Will parse a CSV and attempt to create a certificate store for each row with the provided parameters.`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		var failures []string
// 		var storeTypeFields []string
// 		kfClient, _ := initClient()
// 		storesFile, _ := cmd.Flags().GetString("file")
// 		outputPath, _ := cmd.Flags().GetString("out")
// 		dryRun, _ := cmd.Flags().GetBool("dry-run")
// 		storeTypeId, _ := cmd.Flags().GetInt32("store-type-id")

// 		// get storetype from first.

// 		//foreach row attempt to create the store
// 		//track errors and output into errors.csv file.

// 		//addRootsFile, _ := cmd.Flags().GetString("add-certs")
// 		//removeRootsFile, _ := cmd.Flags().GetString("remove-certs")

// 		log.Printf("[DEBUG] storesFile: %s", storesFile)
// 		log.Printf("[DEBUG] output path: %s", outputPath)
// 		log.Printf("[DEBUG] dryRun: %t", dryRun)
// 		log.Printf("[DEBUG] storeTypeId: %t", storeTypeId)

// 		// Read in the stores CSV
// 		csvFile, _ := os.Open(storesFile)
// 		reader := csv.NewReader(bufio.NewReader(csvFile))
// 		rowIndex := 0
// 		headerMap := make(map[int]string)
// 		propertiesMap := make(map[string]interface{}) //a map keyed by property name, containing maps for that property

// 		storeEntries, _ := reader.ReadAll()
// 		var stores = make(map[string]StoreCSVEntry)
// 		row, err := reader.Read()
// 		//first row is header

// 		if err == io.EOF || err != nil {
// 			panic(err) // or handle it another way
// 		}

// 		for i, v := range row {
// 			headerMap[v] = i
// 		}

// 		for {
// 			row, err := reader.Read()
// 			if err == io.EOF {
// 				break
// 			} else if err != nil {
// 				panic(err) // or handle it another way
// 			}
// 			reqBody := getJsonForRequest(headerMap, row)
// 			// use the `row` here
// 		}

// 		for _, entry := range storeEntries {
// 			if rowIndex == 0 {
// 				// Add mapping: Column/property name --> record index
// 				for i, v := range entry {
// 					headerMap[v] = i
// 				}

// 				keys := make([]string, len(headerMap))

// 				i := 0

// 				for k := range headerMap {
// 					propertyPath := strings.Split(keys[i], propertyDelimiter) //dynamic json fields parsed from "Field:Property";

// 					if len(propertyPath) > 1 {
// 						subMap := make(map[string]map[string]string)
// 						for j := len(propertyPath); j > 0; j-- { //allow arbitrary number of sub-properties
// 							propertiesMap[propertyPath[j-1]][property[j]]
// 						}
// 						headerMap[0] = subMap
// 						propertiesMap[property[0]][property[1]] = k //property[0] is property name, property[1] is sub-property name.
// 					} else {
// 						keys[i] = k
// 					}
// 					i++
// 				}

// 				rowIndex++
// 				continue
// 			}
// 			if rowIndex == 1 {
// 				// first row, so check the store type value and make sure the headers match.  if not, abort.
// 				for k := range headerMap {

// 				}
// 			}

// 			//apiResp, err := kfClient.GetCertificateStoreByID(entry[0])

// 			// if err != nil {
// 			// 	//log.Fatalf("Error getting cert store: %s", err)
// 			// 	log.Printf("[ERROR] Error getting cert store: %s", err)
// 			// 	lookupFailures = append(lookupFailures, strings.Join(entry, ","))
// 			// 	continue
// 			// }

// 			log.Printf("[DEBUG] Store: %s", apiResp)
// 		}
// 		storesJson, _ := json.Marshal(stores)
// 		fmt.Println(string(storesJson))

// 		// Read in the add addCerts CSV
// 		var addCerts = make(map[string]RotCert)
// 		if addRootsFile != "" {
// 			addCerts, err := readCertsFile(addRootsFile)
// 			if err != nil {
// 				log.Fatalf("Error reading addCerts file: %s", err)
// 			}
// 			addCertsJson, _ := json.Marshal(addCerts)
// 			fmt.Printf("[DEBUG] add certs JSON: %s", string(addCertsJson))
// 			fmt.Println("add rot called")
// 		} else {
// 			log.Printf("[DEBUG] No addCerts file specified")
// 			log.Printf("[DEBUG] No addCerts = %s", addCerts)
// 		}

// 		// Read in the remove removeCerts CSV
// 		var removeCerts = make(map[string]RotCert)
// 		if removeRootsFile != "" {
// 			removeCerts, err := readCertsFile(removeRootsFile)
// 			if err != nil {
// 				log.Fatalf("Error reading removeCerts file: %s", err)
// 			}
// 			removeCertsJson, _ := json.Marshal(removeCerts)
// 			fmt.Println(string(removeCertsJson))
// 			fmt.Println("remove rot called")
// 		} else {
// 			log.Printf("[DEBUG] No removeCerts file specified")
// 			log.Printf("[DEBUG] No removeCerts = %s", removeCerts)
// 		}
// 	},
// }

// func getJsonForRequest(headerMap map[string]int, row []string) *gabs.Container {
// 	jsonObj := gabs.New()

// 	for k, v := range headerMap { // k is the string header name, v is the index.
// 		jsonObj.SetP(row[v], k) //this expects properties and sub-properties to be in the header formatted like <propertyname>.<subpropertyname>.<sub-subpropertyname>.<...>
// 	}

// 	fmt.Printf("[DEBUG] get JSON for create store request: %s", jsonObj.String())

// 	return jsonObj
// }

// func readStoresFile(certsFilePath string) (map[string]RotCert, error) {
// 	// Read in the cert CSV
// 	csvFile, _ := os.Open(certsFilePath)
// 	reader := csv.NewReader(bufio.NewReader(csvFile))
// 	certEntries, _ := reader.ReadAll()
// 	//var certs = make(map[string]RotCert)
// 	for _, entry := range certEntries {
// 		switch entry[0] {
// 		case "CertId", "thumbprint", "id", "certId", "Thumbprint":
// 			continue // Skip header
// 		}

// 		certs[entry[0]] = RotCert{
// 			ThumbPrint: entry[0],
// 		}
// 		// Get certificate context
// 		//args := &api.GetCertificateContextArgs{
// 		//	IncludeMetadata:  boolToPointer(true),
// 		//	IncludeLocations: boolToPointer(true),
// 		//	CollectionId:     nil,
// 		//	Id:               certificateIdInt,
// 		//}
// 		//cResp, err := r.p.client.GetCertificateContext(args)
// 	}
// 	return certs, nil
// }

var storesCreateTemplateCmd = &cobra.Command{
	Use:   "generate-template --storetype",
	Short: "For generating a CSV template with headers for bulk store creation.",
	Long: `kfutil stores generate-template creates a csv file containing headers for a specific cert store type.
			Store-type-name OR store-type-id is required.
			Store type IDs can be found by running the "store-types" command.`,
	Run: func(cmd *cobra.Command, args []string) {
		kfClient, _ := initClient()
		storeTypeName, _ := cmd.Flags().GetString("store-type-name")
		storeTypeId, _ := cmd.Flags().GetInt("store-type-id")
		outpath, _ := cmd.Flags().GetString("outpath")

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
		csvHeaders := make(map[int]string)

		storeType, err := kfClient.GetCertificateStoreType(st)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(storeType)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}

		//fmt.Printf("\n %s \n", output)
		jsonParsedObj, _ := gabs.ParseJSON(output)

		// iterate through properties and determine header positions

		properties := jsonParsedObj.S("Properties").Children()
		offset := 0

		for idx, name := range baseFieldNames {
			if name == "Properties" {
				propIdx := idx
				for pIdx, property := range properties {
					loc := propIdx + pIdx
					pName := "Properties." + property.S("Name").Data().(string)
					// fmt.Printf("\n Setting property %s to have index %d", pName, loc)
					csvHeaders[loc] = pName
				}
				offset = len(properties) - 1
			} else {
				csvHeaders[idx+offset] = name
			}
		}

		// write csv file header row

		var filePath string
		if outpath != "" {
			filePath = outpath
		} else {
			filePath = fmt.Sprintf("%s_template.%s", "createstores", "csv")
		}
		csvFile, err := os.Create(filePath)
		if err != nil {
			log.Fatal("Cannot create file", err)
		}
		csvWriter := csv.NewWriter(csvFile)

		row := make([]string, len(csvHeaders))

		for k, v := range csvHeaders {
			row[k] = v
		}

		wErr := csvWriter.Write(row)
		if wErr != nil {
			fmt.Printf("%s", wErr)
			log.Printf("[ERROR] Error writing row to CSV: %s", wErr)
		}

		csvWriter.Flush()
		ioErr := csvFile.Close()
		if ioErr != nil {
			fmt.Println(ioErr)
			log.Printf("[ERROR] Error closing audit file: %s", ioErr)
		}
		fmt.Printf("Audit report written to %s\n", csvFile.Name())

	}}

//command initialization

var (
	storeTypeName string
	storeTypeId   int
	outPath       string
)

func init() {
	storesCmd.AddCommand(importCmd)
	importCmd.AddCommand(storesCreateTemplateCmd)
	storesCreateTemplateCmd.Flags().StringVarP(&storeTypeName, "store-type-name", "n", "", "The name of the cert store type for the template.  Use if store-type-id is unknown.")

	storesCreateTemplateCmd.Flags().IntVarP(&storeTypeId, "store-type-id", "i", -1, "The ID of the cert store type for the template.")
	storesCreateTemplateCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
		"Path and name of the template file to generate.. If not specified, the file will be written to the current directory.")
	storesCreateTemplateCmd.MarkFlagsMutuallyExclusive("store-type-name", "store-type-id")

	//storesCreateCmd.Flags().StringVarP(&stores, "file", "f", "", "CSV file containing cert stores to create.")
	//storesCreateCmd.MarkFlagRequired("file")
	//storesCreateCmd.Flags().StringVarP(&stores, "--check-format-only", "-c", "", "Check for existence of necessary header fields.")
	//storesCreateCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	//storesCreateCmd.MarkFlagRequired("certs")
}
