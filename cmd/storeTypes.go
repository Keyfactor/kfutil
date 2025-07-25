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
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
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
		//stdlog.SetOutput(io.Discard)
		informDebug(debugFlag)

		// Authenticate
		kfClient, cErr := initClient(false)
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to authenticate")
			return cErr
		}

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
		gitRepo, _ := cmd.Flags().GetString(FlagGitRepo)
		creatAll, _ := cmd.Flags().GetBool("all")
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

		validStoreTypes := getValidStoreTypes("", gitRef, gitRepo)

		// Authenticate
		kfClient, _ := initClient(false)

		// CLI Logic
		if gitRef == "" {
			gitRef = DefaultGitRef
		}
		if gitRepo == "" {
			gitRepo = DefaultGitRepo
		}
		storeTypeIsValid := false

		log.Debug().Str("storeType", storeType).
			Bool("listTypes", listTypes).
			Str("storeTypeConfigFile", storeTypeConfigFile).
			Bool("creatAll", creatAll).
			Str("gitRef", gitRef).
			Str("gitRepo", gitRepo).
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
			createdStoreTypes, err := createStoreTypeFromFile(storeTypeConfigFile, kfClient)
			if err != nil {
				fmt.Printf("Failed to create store type from file \"%s\"", err)
				return err
			}

			for _, v := range createdStoreTypes {
				fmt.Printf("Created store type \"%s\"\n", v.Name)
			}
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
		storeTypeConfig, stErr := readStoreTypesConfig("", gitRef, gitRepo, offline)
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
		if gitRef == "" {
			gitRef = "main"
		}
		kfClient, _ := initClient(false)

		var validStoreTypes []string
		var removeStoreTypes []interface{}
		if id < 0 && storeType == "" {
			validStoreTypesResp, vstErr := kfClient.ListCertificateStoreTypes()
			if vstErr != nil {
				log.Error().Err(vstErr).Msg("unable to list certificate store types")
				validStoreTypes = getValidStoreTypes("", gitRef, DefaultGitRepo)
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
		gitRepo, _ := cmd.Flags().GetString(FlagGitRepo)

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
		templates, err := readStoreTypesConfig("", gitRef, gitRepo, offline)
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
func createStoreTypeFromFile(filename string, kfClient *api.Client) ([]api.CertificateStoreType, error) {
	// Read the file
	log.Debug().Str("filename", filename).Msg("Reading store type from file")
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Error().
			Str("filename", filename).
			Err(err).Msg("unable to open file")
		return nil, err
	}

	// Compile JSON contents to a api.CertificateStoreType struct
	var sType api.CertificateStoreType
	var sTypes []api.CertificateStoreType

	log.Debug().Msg("Decoding JSON file as single store type")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sType)
	if err != nil || (sType.ShortName == "" && sType.Capability == "") {
		log.Warn().Err(err).Msg("Unable to decode JSON file, attempting to parse an integration manifest")
		// Attempt to parse as an integration manifest
		var manifest IntegrationManifest
		log.Debug().Msg("Decoding JSON file as integration manifest")
		// Reset the file pointer
		_, err = file.Seek(0, 0)
		decoder = json.NewDecoder(file)
		mErr := decoder.Decode(&manifest)
		if mErr != nil {
			return nil, err
		}
		log.Debug().Msg("Decoded JSON file as integration manifest")
		sTypes = manifest.About.Orchestrator.StoreTypes
	} else {
		log.Debug().Msg("Decoded JSON file as single store type")
		sTypes = []api.CertificateStoreType{sType}
	}

	for _, st := range sTypes {
		log.Debug().Msgf("Creating certificate store type %s", st.Name)
		createResp, cErr := kfClient.CreateStoreType(&st)
		if cErr != nil {
			log.Error().
				Str("storeType", st.Name).
				Err(cErr).Msg("unable to create certificate store type")
			return nil, cErr
		}
		log.Info().Msgf("Certificate store type %s created with ID: %d", st.Name, createResp.StoreType)
	}
	// Use the Keyfactor client to create the store type
	log.Debug().Msg("Store type created")
	return sTypes, nil
}

func formatStoreTypes(sTypesList *[]interface{}) (map[string]interface{}, error) {

	if sTypesList == nil || len(*sTypesList) == 0 {
		return nil, fmt.Errorf("empty store types list")
	}

	output := make(map[string]interface{})
	for _, v := range *sTypesList {
		v2 := v.(map[string]interface{})
		output[v2["ShortName"].(string)] = v2
	}

	return output, nil
}

func getStoreTypesInternet(gitRef string, repo string) (map[string]interface{}, error) {
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfutil/main/store_types.json")
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfctl/master/storetypes/storetypes.json")

	baseUrl := "https://raw.githubusercontent.com/Keyfactor/%s/%s/%s"
	if gitRef == "" {
		gitRef = DefaultGitRef
	}
	if repo == "" {
		repo = DefaultGitRepo
	}

	var fileName string
	if repo == "kfutil" {
		fileName = "store_types.json"
	} else {
		fileName = "integration-manifest.json"
	}

	escapedGitRef := url.PathEscape(gitRef)
	url := fmt.Sprintf(baseUrl, repo, escapedGitRef, fileName)
	log.Debug().
		Str("url", url).
		Msg("Getting store types from internet")

	// Define the timeout duration
	timeout := MinHttpTimeout * time.Second

	// Create a custom http.Client with the timeout
	client := &http.Client{
		Timeout: timeout,
	}
	resp, rErr := client.Get(url)
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
	jErr := json.Unmarshal(body, &result)
	if jErr != nil {
		log.Warn().Err(jErr).Msg("Unable to decode JSON file, attempting to parse an integration manifest")
		// Attempt to parse as an integration manifest
		var manifest IntegrationManifest
		log.Debug().Msg("Decoding JSON file as integration manifest")
		// Reset the file pointer

		mErr := json.Unmarshal(body, &manifest)
		if mErr != nil {
			return nil, jErr
		}
		log.Debug().Msg("Decoded JSON file as integration manifest")
		sTypes := manifest.About.Orchestrator.StoreTypes
		output := make(map[string]interface{})
		for _, st := range sTypes {
			output[st.ShortName] = st
		}
		return output, nil
	}
	output, sErr := formatStoreTypes(&result)
	if sErr != nil {
		return nil, err
	} else if output == nil {
		return nil, fmt.Errorf("unable to fetch store types from %s", url)
	}
	return output, nil

}

func getValidStoreTypes(fp string, gitRef string, gitRepo string) []string {
	log.Debug().
		Str("file", fp).
		Str("gitRef", gitRef).
		Str("gitRepo", gitRepo).
		Bool("offline", offline).
		Msg(DebugFuncEnter)

	log.Debug().
		Str("file", fp).
		Str("gitRef", gitRef).
		Str("gitRepo", gitRepo).
		Msg("Reading store types config.")
	validStoreTypes, rErr := readStoreTypesConfig(fp, gitRef, gitRepo, offline)
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

func readStoreTypesConfig(fp, gitRef string, gitRepo string, offline bool) (map[string]interface{}, error) {
	log.Debug().Str("file", fp).Str("gitRef", gitRef).Msg("Entering readStoreTypesConfig")

	var (
		sTypes map[string]interface{}
		stErr  error
	)
	if offline {
		log.Debug().Msg("Reading store types config from file")
	} else {
		log.Debug().Msg("Reading store types config from internet")
		sTypes, stErr = getStoreTypesInternet(gitRef, gitRepo)
	}

	if stErr != nil || sTypes == nil || len(sTypes) == 0 {
		log.Warn().Err(stErr).Msg("Using embedded store-type definitions")
		var emStoreTypes []interface{}
		if err := json.Unmarshal(EmbeddedStoreTypesJSON, &emStoreTypes); err != nil {
			log.Error().Err(err).Msg("Unable to unmarshal embedded store type definitions")
			return nil, err
		}
		sTypes, stErr = formatStoreTypes(&emStoreTypes)
		if stErr != nil {
			log.Error().Err(stErr).Msg("Unable to format store types")
			return nil, stErr
		}
	}

	var content []byte
	var err error
	if sTypes == nil {
		if fp == "" {
			fp = DefaultStoreTypesFileName
		}
		content, err = os.ReadFile(fp)
	} else {
		content, err = json.Marshal(sTypes)
	}
	if err != nil {
		return nil, err
	}

	var d map[string]interface{}
	if err = json.Unmarshal(content, &d); err != nil {
		log.Error().Err(err).Msg("Unable to unmarshal store types")
		return nil, err
	}
	return d, nil
}

func init() {
	offline = true    // temporarily set to true as it runs before the flag is set
	debugFlag = false // temporarily set to false as it runs before the flag is set
	var gitRef string
	var gitRepo string
	validTypesString := strings.Join(getValidStoreTypes("", DefaultGitRef, DefaultGitRepo), ", ")
	offline = false //revert this so that flag is not set to true by default
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

	fetchStoreTypesCmd.Flags().StringVarP(
		&gitRepo,
		FlagGitRepo,
		"r",
		DefaultGitRepo,
		"The repository to pull store-type definitions from.",
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
	storesTypeCreateCmd.Flags().StringVarP(
		&gitRepo,
		FlagGitRepo,
		"r",
		DefaultGitRepo,
		"The repository to pull store-types definitions from.",
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
