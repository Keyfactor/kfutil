// Package cmd Copyright 2023 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

// Flag enums

// End enums

// Helpers
func buildStoreTypePropertiesInterface(properties []interface{}) ([]api.StoreTypePropertyDefinition, error) {
	var output []api.StoreTypePropertyDefinition

	for _, prop := range properties {
		log.Printf("Prop: %v", prop)
		p := prop.(map[string]interface{})
		output = append(output, api.StoreTypePropertyDefinition{
			Name:         p["Name"].(string),
			DisplayName:  p["DisplayName"].(string),
			Type:         p["Type"].(string),
			DependsOn:    p["DependsOn"].(string),
			DefaultValue: p["DefaultValue"],
			Required:     p["Required"].(bool),
		})
	}

	return output, nil
}

// End helpers

// storeTypesCmd represents the storeTypes command
var storeTypesCmd = &cobra.Command{
	Use:   "store-types",
	Short: "Keyfactor certificate store types APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate store types.`,
}

// storesTypesListCmd represents the list command
var storesTypesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate store types.",
	Long:  `List certificate store types.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("hostname")
		kfcUsername, _ := cmd.Flags().GetString("username")
		kfcPassword, _ := cmd.Flags().GetString("password")
		kfcDomain, _ := cmd.Flags().GetString("domain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
		storeTypes, err := kfClient.ListCertificateStoreTypes()
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(storeTypes)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

var storesTypeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific store type by either name or ID.",
	Long:  `Get a specific store type by either name or ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		genericFormat, _ := cmd.Flags().GetBool("generic")
		outputFormat, _ := cmd.Flags().GetString("format")
		gitRef, _ := cmd.Flags().GetString("git-ref")
		kfcHostName, _ := cmd.Flags().GetString("hostname")
		kfcUsername, _ := cmd.Flags().GetString("username")
		kfcPassword, _ := cmd.Flags().GetString("password")
		kfcDomain, _ := cmd.Flags().GetString("domain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		if gitRef == "" {
			gitRef = "main"
		}
		isExperimental := false
		outputType := "full"

		if genericFormat {
			outputType = "generic"
		}

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		id, _ := cmd.Flags().GetInt("id")
		name, _ := cmd.Flags().GetString("name")
		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
		var st interface{}
		// Check inputs
		if id < 0 && name == "" {
			validStoreTypes := getValidStoreTypes("", gitRef)
			prompt := &survey.Select{
				Message: "Choose an option:",
				Options: validStoreTypes,
			}
			var selected string
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				fmt.Println(err)
			}
			st = selected
			return
		} else if id >= 0 && name != "" {
			log.Printf("Error: ID and Name are mutually exclusive.")
			fmt.Printf("Error: ID and Name are mutually exclusive.\n")
			return
		} else if id >= 0 {
			st = id
		} else if name != "" {
			st = name
		} else {
			log.Printf("Error: Invalid input.")
			fmt.Printf("Error: Invalid input.\n")
			return
		}

		storeTypes, err := kfClient.GetCertificateStoreType(st)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := formatStoreTypeOutput(storeTypes, outputFormat, outputType)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

func formatStoreTypeOutput(storeType *api.CertificateStoreType, outputFormat string, outputType string) (string, error) {
	var sOut interface{}
	sOut = storeType
	if outputType == "generic" {
		// Convert to api.GenericCertificateStoreType
		var genericProperties []api.StoreTypePropertyDefinitionGeneric
		for _, prop := range *storeType.Properties {
			genericProp := api.StoreTypePropertyDefinitionGeneric{
				Name:         prop.Name,
				DisplayName:  prop.DisplayName,
				Type:         prop.Type,
				DependsOn:    prop.DependsOn,
				DefaultValue: prop.DefaultValue,
				Required:     prop.Required,
			}
			genericProperties = append(genericProperties, genericProp)
		}

		var genericEntryParameters []api.EntryParameterGeneric
		for _, param := range *storeType.EntryParameters {
			genericParam := api.EntryParameterGeneric{
				Name:         param.Name,
				DisplayName:  param.DisplayName,
				Type:         param.Type,
				RequiredWhen: param.RequiredWhen,
				DependsOn:    param.DependsOn,
				DefaultValue: param.DefaultValue,
				Options:      param.Options,
			}
			genericEntryParameters = append(genericEntryParameters, genericParam)
		}

		// Check if entry parameters are empty and if they aren't then set jobProperties to empty list
		//jobProperties := storeType.JobProperties
		//if len(genericEntryParameters) > 0 {
		//	log.Println("[WARN] Entry parameters are not empty, setting jobProperties to empty list to prevent 'Only the job properties or entry parameters fields can be set, not both.'")
		//	jobProperties = &[]string{}
		//}

		genericStoreType := api.CertificateStoreTypeGeneric{
			Name:                storeType.Name,
			ShortName:           storeType.ShortName,
			Capability:          storeType.Capability,
			SupportedOperations: storeType.SupportedOperations,
			Properties:          &genericProperties,
			EntryParameters:     &genericEntryParameters,
			PasswordOptions:     storeType.PasswordOptions,
			//StorePathType:       storeType.StorePathType,
			StorePathValue:    storeType.StorePathValue,
			PrivateKeyAllowed: storeType.PrivateKeyAllowed,
			//JobProperties:      jobProperties,
			ServerRequired:     storeType.ServerRequired,
			PowerShell:         storeType.PowerShell,
			BlueprintAllowed:   storeType.BlueprintAllowed,
			CustomAliasAllowed: storeType.CustomAliasAllowed,
		}
		sOut = genericStoreType
	}

	if outputFormat == "json" {
		output, jErr := json.MarshalIndent(sOut, "", "  ")
		if jErr != nil {
			log.Printf("Error: %s", jErr)
			return "", jErr
		}
		return fmt.Sprintf("%s", output), nil
	} else if outputFormat == "yaml" || outputFormat == "yml" {
		output, jErr := yaml.Marshal(sOut)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
			return "", jErr
		}
		return fmt.Sprintf("%s", output), nil
	} else {
		return "", fmt.Errorf("invalid output format: %s", outputFormat)
	}
}

var storesTypeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new certificate store type in Keyfactor.",
	Long:  `Create a new certificate store type in Keyfactor.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		gitRef, _ := cmd.Flags().GetString("git-ref")
		kfcHostName, _ := cmd.Flags().GetString("hostname")
		kfcUsername, _ := cmd.Flags().GetString("username")
		kfcPassword, _ := cmd.Flags().GetString("password")
		kfcDomain, _ := cmd.Flags().GetString("domain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		if gitRef == "" {
			gitRef = "main"
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		//Check if store type is valid
		validStoreTypes := getValidStoreTypes("", gitRef)
		storeType, _ := cmd.Flags().GetString("name")
		listTypes, _ := cmd.Flags().GetBool("list")
		storeTypeConfigFile, _ := cmd.Flags().GetString("from-file")

		// if gitRef is null or empty, then set it to "master"
		if gitRef == "" {
			gitRef = "main"
		}

		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		storeTypeIsValid := false

		if listTypes {
			fmt.Println("Valid store types:")
			for _, st := range validStoreTypes {
				fmt.Printf("\t%s\n", st)
			}
			fmt.Println("Use these values with the --name flag.")
			return
		}

		if storeTypeConfigFile != "" {
			createdStore, err := createStoreFromFile(storeTypeConfigFile, kfClient)
			if err != nil {
				fmt.Printf("Failed to create store type from file \"%s\"", err)
				return
			}

			fmt.Printf("Created store type called \"%s\"\n", createdStore.Name)
			return
		}

		if storeType == "" {
			prompt := &survey.Select{
				Message: "Choose an option:",
				Options: validStoreTypes,
			}
			var selected string
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				fmt.Println(err)
			}
			storeType = selected
		}
		for _, v := range validStoreTypes {
			if strings.EqualFold(v, strings.ToUpper(storeType)) {
				log.Printf("[DEBUG] Valid store type: %s", storeType)
				storeTypeIsValid = true
				break
			}
		}
		if !storeTypeIsValid {
			fmt.Printf("Error: Invalid store type: %s\nValid types are:\n", storeType)
			for _, st := range validStoreTypes {
				fmt.Println(fmt.Sprintf("\t%s", st))
			}
			log.Fatalf("Error: Invalid store type: %s", storeType)
		} else {
			//kfClient, _ := initClient(storeTypeConfigFile, profile, noPrompt, authConfig,false) //TODO: why is this here?
			storeTypeConfig, stErr := readStoreTypesConfig("", gitRef)
			if stErr != nil {
				fmt.Printf("Error: %s", stErr)
				log.Fatalf("Error: %s", stErr)
			}
			log.Printf("[DEBUG] Store type config: %v", storeTypeConfig[storeType])
			storeTypeInterface := storeTypeConfig[storeType].(map[string]interface{})

			//convert storeTypeInterface to json string
			storeTypeJson, _ := json.Marshal(storeTypeInterface)

			//convert storeTypeJson to api.StoreType
			var storeTypeObj api.CertificateStoreType
			convErr := json.Unmarshal(storeTypeJson, &storeTypeObj)
			if convErr != nil {
				fmt.Printf("Error: %s", convErr)
				log.Printf("Error: %s", convErr)
				return
			}

			log.Printf("[DEBUG] Create request: %v", storeTypeObj)
			createResp, err := kfClient.CreateStoreType(&storeTypeObj)
			if err != nil {
				fmt.Printf("Error creating store type: %s", err)
				log.Printf("[ERROR] creating store type : %s", err)
				return
			}
			log.Printf("[DEBUG] Create response: %v", createResp)
			fmt.Printf("Certificate store type %s created with ID: %d", storeType, createResp.StoreType)
		}
	},
}

func createStoreFromFile(filename string, kfClient *api.Client) (*api.CertificateStoreType, error) {

	// Read the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// Compile JSON contents to a api.CertificateStoreType struct
	var storeType api.CertificateStoreType
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&storeType)
	if err != nil {
		return nil, err
	}

	// Use the Keyfactor client to create the store type
	createResp, err := kfClient.CreateStoreType(&storeType)
	if err != nil {
		return nil, err
	}
	return createResp, nil
}

var storesTypeUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a certificate store type in Keyfactor.",
	Long:  `Update a certificate store type in Keyfactor.`,

	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		//configFile, _ := cmd.Flags().GetString("config")
		//noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		//profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		fmt.Println("update called")

		//_, _ = initClient(configFile, profile, noPrompt, false)

	},
}

var storesTypeDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a specific store type by name or ID.",
	Long:  `Delete a specific store type by name or ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		storeType, _ := cmd.Flags().GetString("name")
		gitRef, _ := cmd.Flags().GetString("git-ref")
		kfcHostName, _ := cmd.Flags().GetString("hostname")
		kfcUsername, _ := cmd.Flags().GetString("username")
		kfcPassword, _ := cmd.Flags().GetString("password")
		kfcDomain, _ := cmd.Flags().GetString("domain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		if gitRef == "" {
			gitRef = "main"
		}
		isExperimental := false

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an experimental feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		id, _ := cmd.Flags().GetInt("id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
		var st interface{}

		var validStoreTypes []string
		if id < 0 && storeType == "" {
			validStoreTypesResp, vstErr := kfClient.ListCertificateStoreTypes()
			if vstErr != nil {
				fmt.Println(vstErr)
				validStoreTypes = getValidStoreTypes("", gitRef)
			} else {
				for _, v := range *validStoreTypesResp {
					validStoreTypes = append(validStoreTypes, v.ShortName)
				}
			}
			prompt := &survey.Select{
				Message: "Choose an option:",
				Options: validStoreTypes,
			}
			var selected string
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				fmt.Println(err)
			}
			st = selected
		} else if id >= 0 && storeType != "" {
			log.Printf("Error: ID and Name are mutually exclusive.")
			fmt.Printf("Error: ID and Name are mutually exclusive.\n")
			return
		} else if id >= 0 {
			st = id
		} else if storeType != "" {
			st = storeType
		} else {
			log.Printf("Error: Invalid input.")
			fmt.Printf("Error: Invalid input.\n")
			return
		}

		log.Printf("Deleting certificate store type with ID: %d", id)
		storeTypeResponse, err := kfClient.GetCertificateStoreType(st)
		log.Printf("storeTypeResponse: %v", storeTypeResponse)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}

		if storeTypeResponse.StoreType >= 0 {
			log.Printf("Certificate store type with ID: %d found", storeTypeResponse.StoreType)
			id = storeTypeResponse.StoreType
		}

		if dryRun {
			fmt.Printf("dry run delete called on certificate store type with ID: %d", id)
		} else {
			log.Printf("Calling API to delete certificate store type with ID: %d", id)
			d, err := kfClient.DeleteCertificateStoreType(id)
			if err != nil {
				log.Printf("Error: %s", err)
				fmt.Printf("%s\n", err)
				return
			}
			log.Printf("Certificate store type %v deleted", d)
			fmt.Printf("Certificate store type %v deleted", st)
		}
	},
}

var fetchStoreTypes = &cobra.Command{
	Use:   "templates-fetch",
	Short: "Fetches store type templates from Keyfactor's Github.",
	Long:  `Fetches store type templates from Keyfactor's Github.`,
	Run: func(cmd *cobra.Command, args []string) {
		gitRef, _ := cmd.Flags().GetString("git-ref")
		if gitRef == "" {
			gitRef = "main"
		}
		templates, err := readStoreTypesConfig("", gitRef)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(templates)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Println(string(output))
	},
}

var generateStoreTypeTemplate = &cobra.Command{
	Use:   "templates-generate",
	Short: "Generates either a JSON or CSV template file for certificate store type bulk operations.",
	Long:  `Generates either a JSON or CSV template file for certificate store type bulk operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		gitRef, _ := cmd.Flags().GetString("git-ref")
		if gitRef == "" {
			gitRef = "main"
		}
		templates, err := readStoreTypesConfig("", gitRef)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(templates)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Println(string(output))
	},
}

func getStoreTypesInternet(gitRef string) (map[string]interface{}, error) {
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfutil/main/store_types.json")
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfctl/master/storetypes/storetypes.json")

	resp, rErr := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/Keyfactor/kfutil/%s/store_types.json", gitRef))
	if rErr != nil {
		return nil, rErr
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// read as list of interfaces
	var result []interface{}
	json.Unmarshal([]byte(body), &result)

	// convert to map
	var result2 map[string]interface{}
	result2 = make(map[string]interface{})
	for _, v := range result {
		v2 := v.(map[string]interface{})
		result2[v2["ShortName"].(string)] = v2
	}

	return result2, nil
}

func readStoreTypesConfig(fp string, gitRef string) (map[string]interface{}, error) {

	sTypes, stErr := getStoreTypesInternet(gitRef)
	if stErr != nil {
		fmt.Printf("Error: %s\n", stErr)
	}

	var content []byte
	var err error
	if sTypes == nil {
		if fp == "" {
			fp = "store_types.json"
		}
		content, err = os.ReadFile(fp)
		if err != nil {
			return nil, err
		}
	} else {
		content, err = json.Marshal(sTypes)
		if err != nil {
			return nil, err
		}
	}

	// Now let's unmarshall the data into `payload`
	//var payload map[string]api.CertificateStoreType
	var datas map[string]interface{}
	err = json.Unmarshal(content, &datas)
	if err != nil {
		log.Printf("Error during Unmarshal(): %s", err)
		return nil, err
	}
	return datas, nil
}

func getValidStoreTypes(fp string, gitRef string) []string {
	validStoreTypes, rErr := readStoreTypesConfig(fp, gitRef)
	if rErr != nil {
		log.Printf("Error: %s", rErr)
		fmt.Printf("Error: %s\n", rErr)
		return nil
	}
	validStoreTypesList := make([]string, 0, len(validStoreTypes))
	for k := range validStoreTypes {
		validStoreTypesList = append(validStoreTypesList, k)
	}
	sort.Strings(validStoreTypesList)
	return validStoreTypesList

}

func init() {

	defaultGitRef := "main"
	var gitRef string
	validTypesString := strings.Join(getValidStoreTypes("", defaultGitRef), ", ")
	RootCmd.AddCommand(storeTypesCmd)

	// GET store type templates
	storeTypesCmd.AddCommand(fetchStoreTypes)

	// LIST command
	storeTypesCmd.AddCommand(storesTypesListCmd)

	// GET commands
	storeTypesCmd.AddCommand(storesTypeGetCmd)
	var storeTypeID int
	var storeTypeName string
	var dryRun bool
	var genericFormat bool
	var outputFormat string
	storesTypeGetCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to get.")
	storesTypeGetCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to get.")
	storesTypeGetCmd.MarkFlagsMutuallyExclusive("id", "name")
	storesTypeGetCmd.Flags().BoolVarP(&genericFormat, "generic", "g", false, "Output the store type in a generic format stripped of all fields specific to the Command instance.")
	storesTypeGetCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format. Valid choices are: 'json', 'yaml'. Default is 'json'.")
	storesTypeGetCmd.Flags().StringVarP(&gitRef, "git-ref", "b", "main", "The git branch or tag to reference when pulling store-types from the internet.")

	// CREATE command
	var listValidStoreTypes bool
	var filePath string
	storeTypesCmd.AddCommand(storesTypeCreateCmd)
	storesTypeCreateCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Short name of the certificate store type to get. Valid choices are: "+validTypesString)
	storesTypeCreateCmd.Flags().BoolVarP(&listValidStoreTypes, "list", "l", false, "List valid store types.")
	storesTypeCreateCmd.Flags().StringVarP(&filePath, "from-file", "f", "", "Path to a JSON file containing certificate store type data for a single store.")
	storesTypeCreateCmd.Flags().StringVarP(&gitRef, "git-ref", "b", "main", "The git branch or tag to reference when pulling store-types from the internet.")
	//storesTypeCreateCmd.MarkFlagRequired("name")

	// UPDATE command
	//storeTypesCmd.AddCommand(storesTypeUpdateCmd)
	//storesTypeUpdateCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to get.")

	// DELETE command
	storeTypesCmd.AddCommand(storesTypeDeleteCmd)
	storesTypeDeleteCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to delete.")
	storesTypeDeleteCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to delete.")
	storesTypeDeleteCmd.Flags().BoolVarP(&dryRun, "dry-run", "t", false, "Specifies whether to perform a dry run.")
	storesTypeDeleteCmd.MarkFlagsMutuallyExclusive("id", "name")

}
