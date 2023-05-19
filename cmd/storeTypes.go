// // Package cmd Copyright 2022 Keyfactor
// // Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// // Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// // and limitations under the License.
package cmd

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/AlecAivazis/survey/v2"
//	"github.com/Keyfactor/keyfactor-go-client/api"
//	"github.com/spf13/cobra"
//	"io/ioutil"
//	"log"
//	"net/http"
//	"os"
//	"sort"
//	"strings"
//)
//
//// Flag enums
//
//// End enums
//
//// Helpers
//func buildStoreTypePropertiesInterface(properties []interface{}) ([]api.StoreTypePropertyDefinition, error) {
//	var output []api.StoreTypePropertyDefinition
//
//	for _, prop := range properties {
//		log.Printf("Prop: %v", prop)
//		p := prop.(map[string]interface{})
//		output = append(output, api.StoreTypePropertyDefinition{
//			Name:         p["Name"].(string),
//			DisplayName:  p["DisplayName"].(string),
//			Type:         p["Type"].(string),
//			DependsOn:    p["DependsOn"].(string),
//			DefaultValue: p["DefaultValue"],
//			Required:     p["Required"].(bool),
//		})
//	}
//
//	return output, nil
//}
//
//// End helpers
//
//// storeTypesCmd represents the storeTypes command
//var storeTypesCmd = &cobra.Command{
//	Use:   "store-types",
//	Short: "Keyfactor certificate store types APIs and utilities.",
//	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate store types.`,
//}
//
//// storesTypesListCmd represents the list command
//var storesTypesListCmd = &cobra.Command{
//	Use:   "list",
//	Short: "List certificate store types.",
//	Long:  `List certificate store types.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		// Global flags
//		debugFlag, _ := cmd.Flags().GetBool("debug")
//		configFile, _ := cmd.Flags().GetString("config")
//		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//		profile, _ := cmd.Flags().GetString("profile")
//		expEnabled, _ := cmd.Flags().GetBool("exp")
//		isExperimental := true
//
//		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
//		if expErr != nil {
//			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
//			log.Fatalf("[ERROR]: %s", expErr)
//		}
//
//		debugModeEnabled := checkDebug(debugFlag)
//		log.Println("Debug mode enabled: ", debugModeEnabled)
//		kfClient, _ := initClient(configFile, profile, noPrompt)
//		storeTypes, err := kfClient.ListCertificateStoreTypes()
//		if err != nil {
//			log.Printf("Error: %s", err)
//			fmt.Printf("Error: %s\n", err)
//			return
//		}
//		output, jErr := json.Marshal(storeTypes)
//		if jErr != nil {
//			log.Printf("Error: %s", jErr)
//		}
//		fmt.Printf("%s", output)
//	},
//}
//
//var storesTypeGetCmd = &cobra.Command{
//	Use:   "get",
//	Short: "Get a specific store type by either name or ID.",
//	Long:  `Get a specific store type by either name or ID.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		// Global flags
//		debugFlag, _ := cmd.Flags().GetBool("debug")
//		configFile, _ := cmd.Flags().GetString("config")
//		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//		profile, _ := cmd.Flags().GetString("profile")
//		expEnabled, _ := cmd.Flags().GetBool("exp")
//		isExperimental := true
//
//		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
//		if expErr != nil {
//			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
//			log.Fatalf("[ERROR]: %s", expErr)
//		}
//
//		debugModeEnabled := checkDebug(debugFlag)
//		log.Println("Debug mode enabled: ", debugModeEnabled)
//		id, _ := cmd.Flags().GetInt("id")
//		name, _ := cmd.Flags().GetString("name")
//		kfClient, _ := initClient(configFile, profile, noPrompt)
//		var st interface{}
//		// Check inputs
//		if id < 0 && name == "" {
//			log.Printf("Error: ID must be a positive integer.")
//			fmt.Printf("Error: ID must be a positive integer.\n")
//			return
//		} else if id >= 0 && name != "" {
//			log.Printf("Error: ID and Name are mutually exclusive.")
//			fmt.Printf("Error: ID and Name are mutually exclusive.\n")
//			return
//		} else if id >= 0 {
//			st = id
//		} else if name != "" {
//			st = name
//		} else {
//			log.Printf("Error: Invalid input.")
//			fmt.Printf("Error: Invalid input.\n")
//			return
//		}
//
//		storeTypes, err := kfClient.GetCertificateStoreType(st)
//		if err != nil {
//			log.Printf("Error: %s", err)
//			fmt.Printf("Error: %s\n", err)
//			return
//		}
//		output, jErr := json.Marshal(storeTypes)
//		if jErr != nil {
//			log.Printf("Error: %s", jErr)
//		}
//		fmt.Printf("%s", output)
//	},
//}
//
//var storesTypeCreateCmd = &cobra.Command{
//	Use:   "create",
//	Short: "Create a new certificate store type in Keyfactor.",
//	Long:  `Create a new certificate store type in Keyfactor.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		// Global flags
//		debugFlag, _ := cmd.Flags().GetBool("debug")
//		//configFile, _ := cmd.Flags().GetString("config")
//		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//		profile, _ := cmd.Flags().GetString("profile")
//
//		debugModeEnabled := checkDebug(debugFlag)
//		log.Println("Debug mode enabled: ", debugModeEnabled)
//		//Check if store type is valid
//		validStoreTypes := getValidStoreTypes("")
//		storeType, _ := cmd.Flags().GetString("name")
//		listTypes, _ := cmd.Flags().GetBool("list")
//		configFile, _ := cmd.Flags().GetString("from-file")
//
//		kfClient, _ := initClient(configFile, profile, noPrompt)
//
//		storeTypeIsValid := false
//
//		if listTypes {
//			fmt.Println("Valid store types:")
//			for _, st := range validStoreTypes {
//				fmt.Printf("\t%s\n", st)
//			}
//			fmt.Println("Use these values with the --name flag.")
//			return
//		}
//
//		if configFile != "" {
//			createdStore, err := createStoreFromFile(configFile, kfClient)
//			if err != nil {
//				fmt.Printf("Failed to create store type from file \"%s\"", err)
//				return
//			}
//
//			fmt.Printf("Created store type called \"%s\"\n", createdStore.Name)
//			return
//		}
//
//		if storeType == "" {
//			prompt := &survey.Select{
//				Message: "Choose an option:",
//				Options: validStoreTypes,
//			}
//			var selected string
//			err := survey.AskOne(prompt, &selected)
//			if err != nil {
//				fmt.Println(err)
//			}
//			storeType = selected
//		}
//		for _, v := range validStoreTypes {
//			if strings.EqualFold(v, strings.ToUpper(storeType)) {
//				log.Printf("[DEBUG] Valid store type: %s", storeType)
//				storeTypeIsValid = true
//				break
//			}
//		}
//		if !storeTypeIsValid {
//			fmt.Printf("Error: Invalid store type: %s\nValid types are:\n", storeType)
//			for _, st := range validStoreTypes {
//				fmt.Println(fmt.Sprintf("\t%s", st))
//			}
//			log.Fatalf("Error: Invalid store type: %s", storeType)
//		} else {
//			kfClient, _ := initClient(configFile, profile, noPrompt)
//			storeTypeConfig, stErr := readStoreTypesConfig("")
//			if stErr != nil {
//				fmt.Printf("Error: %s", stErr)
//				log.Fatalf("Error: %s", stErr)
//			}
//			log.Printf("[DEBUG] Store type config: %v", storeTypeConfig[storeType])
//			sConfig := storeTypeConfig[storeType].(map[string]interface{})
//			props, pErr := buildStoreTypePropertiesInterface(sConfig["Properties"].([]interface{}))
//			if pErr != nil {
//				fmt.Printf("Error: %s", pErr)
//				log.Printf("Error: %s", pErr)
//			}
//			createReq := api.CertificateStoreType{
//				Name:       storeTypeConfig[storeType].(map[string]interface{})["Name"].(string),
//				ShortName:  storeTypeConfig[storeType].(map[string]interface{})["ShortName"].(string),
//				Capability: storeTypeConfig[storeType].(map[string]interface{})["Capability"].(string),
//				SupportedOperations: &api.StoreTypeSupportedOperations{
//					Add:        storeTypeConfig[storeType].(map[string]interface{})["SupportedOperations"].(map[string]interface{})["Add"].(bool),
//					Create:     storeTypeConfig[storeType].(map[string]interface{})["SupportedOperations"].(map[string]interface{})["Create"].(bool),
//					Discovery:  storeTypeConfig[storeType].(map[string]interface{})["SupportedOperations"].(map[string]interface{})["Discovery"].(bool),
//					Enrollment: storeTypeConfig[storeType].(map[string]interface{})["SupportedOperations"].(map[string]interface{})["Enrollment"].(bool),
//					Remove:     storeTypeConfig[storeType].(map[string]interface{})["SupportedOperations"].(map[string]interface{})["Remove"].(bool),
//				},
//				Properties:      &props,
//				EntryParameters: &[]api.EntryParameter{},
//				PasswordOptions: &api.StoreTypePasswordOptions{
//					EntrySupported: storeTypeConfig[storeType].(map[string]interface{})["PasswordOptions"].(map[string]interface{})["EntrySupported"].(bool),
//					StoreRequired:  storeTypeConfig[storeType].(map[string]interface{})["PasswordOptions"].(map[string]interface{})["StoreRequired"].(bool),
//					Style:          storeTypeConfig[storeType].(map[string]interface{})["PasswordOptions"].(map[string]interface{})["Style"].(string),
//				},
//				//StorePathType:      "",
//				//StorePathValue:     "",
//				PrivateKeyAllowed:  storeTypeConfig[storeType].(map[string]interface{})["PrivateKeyAllowed"].(string),
//				JobProperties:      nil,
//				ServerRequired:     storeTypeConfig[storeType].(map[string]interface{})["ServerRequired"].(bool),
//				PowerShell:         storeTypeConfig[storeType].(map[string]interface{})["PowerShell"].(bool),
//				BlueprintAllowed:   storeTypeConfig[storeType].(map[string]interface{})["BlueprintAllowed"].(bool),
//				CustomAliasAllowed: storeTypeConfig[storeType].(map[string]interface{})["CustomAliasAllowed"].(string),
//				//ServerRegistration: 0,
//				//InventoryEndpoint:  "",
//				//InventoryJobType:   "",
//				//ManagementJobType:  "",
//				//DiscoveryJobType:   "",
//				//EnrollmentJobType:  "",
//			}
//			log.Printf("[DEBUG] Create request: %v", createReq)
//			createResp, err := kfClient.CreateStoreType(&createReq)
//			if err != nil {
//				fmt.Printf("Error creating store type: %s", err)
//				log.Printf("[ERROR] creating store type : %s", err)
//			}
//			log.Printf("[DEBUG] Create response: %v", createResp)
//		}
//	},
//}
//
//func createStoreFromFile(filename string, kfClient *api.Client) (*api.CertificateStoreType, error) {
//
//	// Read the file
//	file, err := os.Open(filename)
//	if err != nil {
//		return nil, err
//	}
//
//	// Compile JSON contents to a api.CertificateStoreType struct
//	var storeType api.CertificateStoreType
//	decoder := json.NewDecoder(file)
//	err = decoder.Decode(&storeType)
//	if err != nil {
//		return nil, err
//	}
//
//	// Use the Keyfactor client to create the store type
//	createResp, err := kfClient.CreateStoreType(&storeType)
//	if err != nil {
//		return nil, err
//	}
//	return createResp, nil
//}
//
//var storesTypeUpdateCmd = &cobra.Command{
//	Use:   "update",
//	Short: "Update a certificate store type in Keyfactor.",
//	Long:  `Update a certificate store type in Keyfactor.`,
//
//	Run: func(cmd *cobra.Command, args []string) {
//		// Global flags
//		debugFlag, _ := cmd.Flags().GetBool("debug")
//		configFile, _ := cmd.Flags().GetString("config")
//		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//		profile, _ := cmd.Flags().GetString("profile")
//		expEnabled, _ := cmd.Flags().GetBool("exp")
//		isExperimental := true
//
//		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
//		if expErr != nil {
//			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
//			log.Fatalf("[ERROR]: %s", expErr)
//		}
//
//		debugModeEnabled := checkDebug(debugFlag)
//		log.Println("Debug mode enabled: ", debugModeEnabled)
//		fmt.Println("update called")
//
//		_, _ = initClient(configFile, profile, noPrompt)
//
//	},
//}
//
//var storesTypeDeleteCmd = &cobra.Command{
//	Use:   "delete",
//	Short: "Delete a specific store type by ID.",
//	Long:  `Delete a specific store type by ID.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		// Global flags
//		debugFlag, _ := cmd.Flags().GetBool("debug")
//		configFile, _ := cmd.Flags().GetString("config")
//		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//		profile, _ := cmd.Flags().GetString("profile")
//		expEnabled, _ := cmd.Flags().GetBool("exp")
//		isExperimental := true
//
//		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
//		if expErr != nil {
//			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
//			log.Fatalf("[ERROR]: %s", expErr)
//		}
//
//		debugModeEnabled := checkDebug(debugFlag)
//		log.Println("Debug mode enabled: ", debugModeEnabled)
//		id, _ := cmd.Flags().GetInt("id")
//		dryRun, _ := cmd.Flags().GetBool("dry-run")
//		kfClient, _ := initClient(configFile, profile, noPrompt)
//		_, err := kfClient.GetCertificateStoreType(id)
//		if err != nil {
//			log.Printf("Error: %s", err)
//			fmt.Printf("Error: %s\n", err)
//			return
//		}
//		if dryRun {
//			fmt.Printf("dry run delete called on certificate store type with ID: %d", id)
//		} else {
//			d, err := kfClient.DeleteCertificateStoreType(id)
//			if err != nil {
//				log.Printf("Error: %s", err)
//				fmt.Printf("Error: %s\n", err)
//				return
//			}
//			fmt.Printf("Certificate store type with ID: %d deleted", d.ID)
//		}
//	},
//}
//
//var fetchStoreTypes = &cobra.Command{
//	Use:   "templates-fetch",
//	Short: "Fetches store type templates from Keyfactor's Github.",
//	Long:  `Fetches store type templates from Keyfactor's Github.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		templates, err := readStoreTypesConfig("")
//		if err != nil {
//			log.Printf("Error: %s", err)
//			fmt.Printf("Error: %s\n", err)
//			return
//		}
//		output, jErr := json.Marshal(templates)
//		if jErr != nil {
//			log.Printf("Error: %s", jErr)
//		}
//		fmt.Println(string(output))
//	},
//}
//
//var generateStoreTypeTemplate = &cobra.Command{
//	Use:   "templates-generate",
//	Short: "Generates either a JSON or CSV template file for certificate store type bulk operations.",
//	Long:  `Generates either a JSON or CSV template file for certificate store type bulk operations.`,
//	Run: func(cmd *cobra.Command, args []string) {
//		templates, err := readStoreTypesConfig("")
//		if err != nil {
//			log.Printf("Error: %s", err)
//			fmt.Printf("Error: %s\n", err)
//			return
//		}
//		output, jErr := json.Marshal(templates)
//		if jErr != nil {
//			log.Printf("Error: %s", jErr)
//		}
//		fmt.Println(string(output))
//	},
//}
//
//func getStoreTypesInternet() (map[string]interface{}, error) {
//	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfutil/main/store_types.json")
//	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfctl/master/storetypes/storetypes.json")
//	resp, rErr := http.Get("https://raw.githubusercontent.com/Keyfactor/kfutil/main/store_types.json")
//	if rErr != nil {
//		return nil, rErr
//	}
//	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//	var result map[string]interface{}
//	json.Unmarshal([]byte(body), &result)
//	return result, nil
//}
//
//func readStoreTypesConfig(fp string) (map[string]interface{}, error) {
//
//	sTypes, stErr := getStoreTypesInternet()
//	if stErr != nil {
//		fmt.Printf("Error: %s\n", stErr)
//	}
//
//	var content []byte
//	var err error
//	if sTypes == nil {
//		if fp == "" {
//			fp = "store_types.json"
//		}
//		content, err = os.ReadFile(fp)
//		if err != nil {
//			return nil, err
//		}
//	} else {
//		content, err = json.Marshal(sTypes)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	// Now let's unmarshall the data into `payload`
//	//var payload map[string]api.CertificateStoreType
//	var datas map[string]interface{}
//	err = json.Unmarshal(content, &datas)
//	if err != nil {
//		log.Printf("Error during Unmarshal(): %s", err)
//		return nil, err
//	}
//	return datas, nil
//}
//
//func getValidStoreTypes(fp string) []string {
//	validStoreTypes, rErr := readStoreTypesConfig(fp)
//	if rErr != nil {
//		log.Printf("Error: %s", rErr)
//		fmt.Printf("Error: %s\n", rErr)
//		return nil
//	}
//	validStoreTypesList := make([]string, 0, len(validStoreTypes))
//	for k := range validStoreTypes {
//		validStoreTypesList = append(validStoreTypesList, k)
//	}
//	sort.Strings(validStoreTypesList)
//	return validStoreTypesList
//
//}
//
//func init() {
//
//	validTypesString := strings.Join(getValidStoreTypes(""), ", ")
//	RootCmd.AddCommand(storeTypesCmd)
//
//	// GET store type templates
//	storeTypesCmd.AddCommand(fetchStoreTypes)
//
//	// LIST command
//	storeTypesCmd.AddCommand(storesTypesListCmd)
//
//	// GET commands
//	storeTypesCmd.AddCommand(storesTypeGetCmd)
//	var storeTypeID int
//	var storeTypeName string
//	var dryRun bool
//	storesTypeGetCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to get.")
//	storesTypeGetCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to get.")
//	storesTypeGetCmd.MarkFlagsMutuallyExclusive("id", "name")
//
//	// CREATE command
//	var listValidStoreTypes bool
//	var filePath string
//	storeTypesCmd.AddCommand(storesTypeCreateCmd)
//	storesTypeCreateCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Short name of the certificate store type to get. Valid choices are: "+validTypesString)
//	storesTypeCreateCmd.Flags().BoolVarP(&listValidStoreTypes, "list", "l", false, "List valid store types.")
//	storesTypeCreateCmd.Flags().StringVarP(&filePath, "from-file", "f", "", "Path to a JSON file containing certificate store type data for a single store.")
//	//storesTypeCreateCmd.MarkFlagRequired("name")
//
//	// UPDATE command
//	storeTypesCmd.AddCommand(storesTypeUpdateCmd)
//	storesTypeUpdateCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to get.")
//
//	// DELETE command
//	storeTypesCmd.AddCommand(storesTypeDeleteCmd)
//	storesTypeDeleteCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to get.")
//	storesTypeDeleteCmd.Flags().BoolVarP(&dryRun, "dry-run", "t", false, "Specifies whether to perform a dry run.")
//	storesTypeDeleteCmd.MarkFlagRequired("id")
//
//}
