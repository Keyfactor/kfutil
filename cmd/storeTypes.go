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
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed store_types.json
var EmbeddedStoreTypesJSON []byte

var storeTypesCmd = &cobra.Command{
	Use:   "store-types",
	Short: "Keyfactor certificate store types APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate store types.`,
}

var storesTypesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate store types.",
	Long:  `List certificate store types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// expEnabled checks
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
		storeTypes, err := kfClient.ListCertificateStoreTypes()
		if err != nil {

			log.Error().Err(err).Msg("unable to list certificate store types")
			return err
		}
		output, jErr := json.Marshal(storeTypes)
		if jErr != nil {

			log.Error().Err(jErr).Msg("unable to marshal certificate store types to JSON")
			return jErr
		}
		outputResult(output, outputFormat)
		return nil
	},
}

var storesTypeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new certificate store type in Keyfactor.",
	Long:  `Create a new certificate store type in Keyfactor.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		gitRef, _ := cmd.Flags().GetString(FlagGitRef)
		creatAll, _ := cmd.Flags().GetBool("all")
		validStoreTypes := getValidStoreTypes("", gitRef)
		storeType, _ := cmd.Flags().GetString("name")
		listTypes, _ := cmd.Flags().GetBool("list")
		storeTypeConfigFile, _ := cmd.Flags().GetString("from-file")

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
		if gitRef == "" {
			gitRef = "main"
		}
		storeTypeIsValid := false

		log.Debug().Str("storeType", storeType).
			Bool("listTypes", listTypes).
			Str("storeTypeConfigFile", storeTypeConfigFile).
			Bool("creatAll", creatAll).
			Str("gitRef", gitRef).
			Strs("validStoreTypes", validStoreTypes).
			Msg("create command flags")

		if listTypes {
			fmt.Println("Available store types:")
			sort.Strings(validStoreTypes)
			for _, st := range validStoreTypes {
				fmt.Printf("\t%s\n", st)
			}
			fmt.Println("Use these values with the --name flag.")
			return nil
		}

		if storeTypeConfigFile != "" {
			createdStore, err := createStoreFromFile(storeTypeConfigFile, kfClient)
			if err != nil {
				fmt.Printf("Failed to create store type from file \"%s\"", err)
				return err
			}

			fmt.Printf("Created store type called \"%s\"\n", createdStore.Name)
			return nil
		}

		if storeType == "" && !creatAll {
			prompt := &survey.Select{
				Message: "Choose an option:",
				Options: validStoreTypes,
			}
			var selected string
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				fmt.Println(err)
				return err
			}
			storeType = selected
		}
		for _, v := range validStoreTypes {
			if strings.EqualFold(v, strings.ToUpper(storeType)) || creatAll {
				log.Debug().Str("storeType", storeType).Msg("Store type is valid")
				storeTypeIsValid = true
				break
			}
		}
		if !storeTypeIsValid {
			log.Error().
				Str("storeType", storeType).
				Bool("isValid", storeTypeIsValid).
				Msg("Invalid store type")
			fmt.Printf("ERROR: Invalid store type: %s\nValid types are:\n", storeType)
			for _, st := range validStoreTypes {
				fmt.Println(fmt.Sprintf("\t%s", st))
			}
			log.Error().Msg(fmt.Sprintf("Invalid store type: %s", storeType))
			return fmt.Errorf("invalid store type: %s", storeType)
		}
		var typesToCreate []string
		if !creatAll {
			typesToCreate = []string{storeType}
		} else {
			typesToCreate = validStoreTypes
		}
		storeTypeConfig, stErr := readStoreTypesConfig("", gitRef)
		if stErr != nil {
			log.Error().Err(stErr).Send()
			return stErr
		}
		var createErrors []error
		for _, st := range typesToCreate {
			log.Trace().Msgf("Store type config: %v", storeTypeConfig[st])
			storeTypeInterface := storeTypeConfig[st].(map[string]interface{})
			storeTypeJSON, _ := json.Marshal(storeTypeInterface)

			var storeTypeObj api.CertificateStoreType
			convErr := json.Unmarshal(storeTypeJSON, &storeTypeObj)
			if convErr != nil {
				log.Error().Err(convErr).Msg("unable to convert store type config to JSON")
				createErrors = append(createErrors, fmt.Errorf("%v: %s", st, convErr.Error()))
				continue
			}

			log.Trace().Msgf("Store type object: %v", storeTypeObj)
			createResp, err := kfClient.CreateStoreType(&storeTypeObj)
			if err != nil {
				log.Error().Err(err).Msg("unable to create store type")
				createErrors = append(createErrors, fmt.Errorf("%v: %s", st, err.Error()))
				continue
			}
			log.Trace().Msgf("Create response: %v", createResp)
			fmt.Println(fmt.Sprintf("Certificate store type %s created with ID: %d", st, createResp.StoreType))
		}

		if len(createErrors) > 0 {
			errStr := "while creating store types:\n"
			for _, e := range createErrors {
				errStr += fmt.Sprintf("%s\n", e)
			}
			return fmt.Errorf(errStr)
		}

		return nil
	},
}

var storesTypeDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a specific store type by name or ID.",
	Long:  `Delete a specific store type by name or ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		storeType, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetInt("id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		deleteAll, _ := cmd.Flags().GetBool("all")
		gitRef, _ := cmd.Flags().GetString(FlagGitRef)

		// Debug + expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		log.Debug().Str("storeType", storeType).
			Int("id", id).Bool("dryRun", dryRun).
			Bool("deleteAll", deleteAll).
			Msg("delete command flags")

		// Authenticate
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		if gitRef == "" {
			gitRef = "main"
		}
		kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)

		var validStoreTypes []string
		var removeStoreTypes []interface{}
		if id < 0 && storeType == "" {
			validStoreTypesResp, vstErr := kfClient.ListCertificateStoreTypes()
			if vstErr != nil {
				log.Error().Err(vstErr).Msg("unable to list certificate store types")
				validStoreTypes = getValidStoreTypes("", gitRef)
			} else {
				for _, v := range *validStoreTypesResp {
					validStoreTypes = append(validStoreTypes, v.ShortName)
					removeStoreTypes = append(removeStoreTypes, v.ShortName)
				}
			}
			if !deleteAll {
				log.Info().Msg("No store type specified, prompting user to select one")
				prompt := &survey.Select{
					Message: "Choose a store type to delete:",
					Options: validStoreTypes,
				}
				var selected string
				err := survey.AskOne(prompt, &selected)
				if err != nil {
					log.Error().Err(err).Msg("user select prompt failed")
					fmt.Println(err)
				}
				log.Info().Str("storeType", selected).Msg("User selected store type")
				removeStoreTypes = []interface{}{selected}
			}
		} else if id >= 0 && storeType != "" {
			log.Error().Err(InvalidInputError).Send()
			return fmt.Errorf("ID and Name are mutually exclusive")
		} else if id >= 0 {
			removeStoreTypes = []interface{}{id}
		} else if storeType != "" {
			removeStoreTypes = []interface{}{storeType}
		} else {
			log.Error().Err(InvalidInputError).Send()
			return InvalidInputError
		}

		var removalErrors []error
		for _, st := range removeStoreTypes {
			log.Info().Str("storeType", fmt.Sprintf("%v", st)).
				Msg("Deleting certificate store type")

			log.Debug().Msg("Checking if store type exists before deleting")
			storeTypeResponse, err := kfClient.GetCertificateStoreType(st)
			log.Trace().Interface("storeTypeResponse", storeTypeResponse).Msg("GetCertificateStoreType response")
			if err != nil {
				log.Error().Err(err).Msg("unable to get certificate store type")
				removalErrors = append(removalErrors, fmt.Errorf("%v: %s", st, err.Error()))
				continue
			}

			if storeTypeResponse.StoreType >= 0 {
				log.Debug().Msg("Certificate store type found")
				id = storeTypeResponse.StoreType
			}

			if dryRun {
				outputResult(
					fmt.Sprintf("dry run delete called on certificate store type (%v) with ID: %d", st, id),
					outputFormat,
				)
			} else {
				log.Debug().Interface("storeType", st).
					Int("id", id).
					Msg("Calling API to delete certificate store type")
				d, err := kfClient.DeleteCertificateStoreType(id)
				log.Trace().Interface("deleteResponse", d).Msg("DeleteCertificateStoreType response")
				if err != nil {
					log.Error().Err(err).
						Interface("storeType", st).
						Int("id", id).
						Msg("unable to delete certificate store type")
					removalErrors = append(removalErrors, fmt.Errorf("%v: %s", st, err.Error()))
					continue
				}
				log.Info().Interface("storeType", st).
					Int("id", id).
					Msg("Certificate store type deleted")
				outputResult(fmt.Sprintf("Certificate store type (%v) deleted", st), outputFormat)
			}
		}
		if len(removalErrors) > 0 {
			errStr := "while deleting store types:\n"
			for _, e := range removalErrors {
				errStr += fmt.Sprintf("%s\n", e)
			}
			return fmt.Errorf(errStr)
		}
		return nil
	},
}

var fetchStoreTypesCmd = &cobra.Command{
	Use:   "templates-fetch",
	Short: "Fetches store type templates from Keyfactor's Github.",
	Long:  `Fetches store type templates from Keyfactor's Github.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		gitRef, _ := cmd.Flags().GetString(FlagGitRef)

		// Debug + expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		if gitRef == "" {
			gitRef = "main"
		}
		templates, err := readStoreTypesConfig("", gitRef)
		if err != nil {
			log.Error().Err(err).Send()
			return err
		}
		output, jErr := json.Marshal(templates)
		if jErr != nil {
			log.Error().Err(jErr).Msg("unable to marshal store types to JSON")
			return jErr
		}
		fmt.Println(string(output))
		return nil
	},
}

//	var generateStoreTypeTemplate = &cobra.Command{
//		Use:   "templates-generate",
//		Short: "Generates either a JSON or CSV template file for certificate store type bulk operations.",
//		Long:  `Generates either a JSON or CSV template file for certificate store type bulk operations.`,
//		RunE: func(cmd *cobra.Command, args []string) error {
//			cmd.SilenceUsage = true
//			gitRef, _ := cmd.Flags().GetString(FlagGitRef)
//			if gitRef == "" {
//				gitRef = "main"
//			}
//			templates, err := readStoreTypesConfig("", gitRef)
//			if err != nil {
//				log.Error().Err(err).Msg("unable to read store types config")
//				return err
//			}
//			output, jErr := json.Marshal(templates)
//			if jErr != nil {
//				log.Error().Err(jErr).Msg("unable to marshal store types to JSON")
//				return jErr
//			}
//			fmt.Println(string(output))
//			return nil
//		},
//	}
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

func getStoreTypesInternet(gitRef string) (map[string]interface{}, error) {
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfutil/main/store_types.json")
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfctl/master/storetypes/storetypes.json")

	resp, rErr := http.Get(
		fmt.Sprintf(
			"https://raw.githubusercontent.com/Keyfactor/kfutil/%s/store_types.json",
			gitRef,
		),
	)
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
	json.Unmarshal(body, &result)

	// convert to map
	var result2 map[string]interface{}
	result2 = make(map[string]interface{})
	for _, v := range result {
		v2 := v.(map[string]interface{})
		result2[v2["ShortName"].(string)] = v2
	}

	return result2, nil
}

func getValidStoreTypes(fp string, gitRef string) []string {
	log.Debug().
		Str("file", fp).
		Str("gitRef", gitRef).
		Msg(DebugFuncEnter)

	log.Debug().
		Str("file", fp).
		Str("gitRef", gitRef).
		Msg("Reading store types config.")
	validStoreTypes, rErr := readStoreTypesConfig(fp, gitRef)
	if rErr != nil {
		log.Error().Err(rErr).Msg("unable to read store types")
		return nil
	}
	validStoreTypesList := make([]string, 0, len(validStoreTypes))
	for k := range validStoreTypes {
		validStoreTypesList = append(validStoreTypesList, k)
	}
	sort.Strings(validStoreTypesList)
	return validStoreTypesList
}

func readStoreTypesConfig(fp string, gitRef string) (map[string]interface{}, error) {
	sTypes, stErr := getStoreTypesInternet(gitRef)
	if stErr != nil {
		log.Error().Err(stErr).Msg("unable to read store types from internet, attempting to reference embedded definitions")
		if err := json.Unmarshal(EmbeddedStoreTypesJSON, &sTypes); err != nil {
			log.Error().Err(err).Msg("unable to unmarshal embedded store type definitions")
			return nil, err
		}
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

	var d map[string]interface{}
	err = json.Unmarshal(content, &d)
	if err != nil {
		log.Error().Err(err).Msg("unable to unmarshal store types")
		return nil, err
	}
	return d, nil
}

func init() {
	defaultGitRef := "main"
	var gitRef string
	validTypesString := strings.Join(getValidStoreTypes("", defaultGitRef), ", ")
	RootCmd.AddCommand(storeTypesCmd)

	// GET store type templates
	storeTypesCmd.AddCommand(fetchStoreTypesCmd)
	fetchStoreTypesCmd.Flags().StringVarP(
		&gitRef,
		FlagGitRef,
		"b",
		"main",
		"The git branch or tag to reference when pulling store-types from the internet.",
	)

	// LIST command
	storeTypesCmd.AddCommand(storesTypesListCmd)

	// GET commands
	storeTypesCmd.AddCommand(CreateCmdStoreTypesGet())

	// CREATE command
	var listValidStoreTypes bool
	var filePath string
	var createAll bool
	var storeTypeName string
	var storeTypeID int
	storeTypesCmd.AddCommand(storesTypeCreateCmd)
	storesTypeCreateCmd.Flags().StringVarP(
		&storeTypeName,
		"name",
		"n",
		"",
		"Short name of the certificate store type to get. Valid choices are: "+validTypesString,
	)
	storesTypeCreateCmd.Flags().BoolVarP(&listValidStoreTypes, "list", "l", false, "List valid store types.")
	storesTypeCreateCmd.Flags().StringVarP(
		&filePath,
		"from-file",
		"f",
		"",
		"Path to a JSON file containing certificate store type data for a single store.",
	)
	storesTypeCreateCmd.Flags().StringVarP(
		&gitRef,
		FlagGitRef,
		"b",
		"main",
		"The git branch or tag to reference when pulling store-types from the internet.",
	)
	storesTypeCreateCmd.Flags().BoolVarP(&createAll, "all", "a", false, "Create all store types.")

	// UPDATE command
	// storeTypesCmd.AddCommand(storesTypeUpdateCmd)
	// storesTypeUpdateCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to get.")

	// DELETE command
	var deleteAll bool
	var dryRun bool
	storeTypesCmd.AddCommand(storesTypeDeleteCmd)
	storesTypeDeleteCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to delete.")
	storesTypeDeleteCmd.Flags().StringVarP(
		&storeTypeName,
		"name",
		"n",
		"",
		"Name of the certificate store type to delete.",
	)
	storesTypeDeleteCmd.Flags().BoolVarP(&dryRun, "dry-run", "t", false, "Specifies whether to perform a dry run.")
	storesTypeDeleteCmd.MarkFlagsMutuallyExclusive("id", "name")
	storesTypeDeleteCmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete all store types.")
}
