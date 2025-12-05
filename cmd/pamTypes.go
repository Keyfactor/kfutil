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
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	keyfactor "github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed pam_types.json
var EmbeddedPAMTypesJSON []byte

var pamTypesCmd = &cobra.Command{
	Use:   "pam-types",
	Short: "Keyfactor PAM types APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor PAM types.`,
}

var pamTypesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific defined PAM Provider type by ID or Name.",
	Long:  "Get a specific defined PAM Provider type by ID or Name.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false
		// Specific flags
		pamProviderTypeId, _ := cmd.Flags().GetString("id")
		pamProviderTypeName, _ := cmd.Flags().GetString("name")
		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		// Log flags
		log.Info().Str("name", pamProviderTypeName).
			Str("id", pamProviderTypeId).
			Msg("get PAM Provider Type")
		// Authenticate
		kfClient, cErr := initClient(false)
		if cErr != nil {
			return cErr
		}
		if pamProviderTypeId == "" && pamProviderTypeName == "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a PAM Provider Type `--id` or `--name` to get")
		}

		// CLI Logic
		if pamProviderTypeId == "" && pamProviderTypeName != "" {
			// Get ID from Name
			log.Debug().Str("pamProviderTypeName", pamProviderTypeName).
				Msg("call: GetPAMProviderTypeByName()")
			pamProviderType, getErr := kfClient.GetPAMProviderTypeByName(pamProviderTypeName)
			log.Debug().Msg("returned: GetPAMProviderTypeByName()")
			if getErr != nil {
				log.Error().Err(getErr).Send()
				return getErr
			}
			if pamProviderType != nil {
				output, mErr := json.Marshal(pamProviderType)
				if mErr != nil {
					log.Error().Err(mErr).Send()
					return mErr
				}
				log.Info().Str("output", string(output)).
					Msg("successfully retrieved PAM provider type")
				outputResult(output, outputFormat)
				return nil
			}
		}
		pamProviderType, getErr := kfClient.GetPAMProviderType(pamProviderTypeId)
		if getErr != nil {
			log.Error().Err(getErr).Send()
			return getErr
		}
		output, mErr := json.Marshal(pamProviderType)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().Str("output", string(output)).
			Msg("successfully retrieved PAM provider type")
		outputResult(output, outputFormat)
		return nil
	},
}

var pamTypesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Returns a list of all available PAM provider types.",
	Long:  "Returns a list of all available PAM provider types.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false

		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Log flags
		log.Info().Msg("list PAM Provider Types")

		// Authenticate
		kfClient, clientErr := initClient(false)
		if clientErr != nil {
			return clientErr
		}

		//// CLI Logic
		//log.Debug().Msg("call: PAMProviderGetPamProviderTypes()")
		//pamTypes, httpResponse, err := sdkClient.PAMProviderApi.
		//	PAMProviderGetPamProviderTypes(context.Background()).
		//	XKeyfactorRequestedWith(XKeyfactorRequestedWith).
		//	XKeyfactorApiVersion(XKeyfactorApiVersion).
		//	Execute()
		pamTypes, err := kfClient.ListPAMProviderTypes()
		log.Debug().Msg("returned: PAMProviderGetPamProviderTypes()")
		if err != nil {
			log.Error().Err(err).
				Msg("error listing PAM provider types")
			return err
		}

		log.Debug().Msg("Converting PAM Provider Types response to JSON")
		jsonString, mErr := json.Marshal(pamTypes)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return mErr
		}
		log.Info().
			Msg("successfully listed PAM provider types")
		outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamTypesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new PAM provider type.",
	Long: `Creates a new PAM Provider type, currently only supported from JSON file and from GitHub. To install from 
Github. To install from GitHub, use the --repo flag to specify the GitHub repository and optionally the branch to use. 
NOTE: the file from Github must be named integration-manifest.json and must use the same schema as 
https://github.com/Keyfactor/hashicorp-vault-pam/blob/main/integration-manifest.json. To install from a local file, use
--from-file to specify the path to the JSON file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		// Specific flags
		gitRef, _ := cmd.Flags().GetString(FlagGitRef)
		gitRepo, _ := cmd.Flags().GetString(FlagGitRepo)
		createAll, _ := cmd.Flags().GetBool("all")
		pamProviderTypeName, _ := cmd.Flags().GetString("name")
		listTypes, _ := cmd.Flags().GetBool("list")
		pamTypeConfigFile, _ := cmd.Flags().GetString(FlagFromFile)

		// Debug + expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		informDebug(debugFlag)

		validPAMTypes := getValidPAMTypes("", gitRef, gitRepo)

		// Authenticate
		kfClient, cErr := initClient(false)
		//sdkClient, cErr := initGenClient(false)
		if cErr != nil {
			return cErr
		}

		// CLI Logic
		if gitRef == "" {
			gitRef = DefaultGitRef
		}
		if gitRepo == "" {
			gitRepo = DefaultGitRepo
		}
		pamTypeIsValid := false

		// Log flags
		log.Info().Str("name", pamProviderTypeName).
			Bool("listTypes", listTypes).
			Str("storeTypeConfigFile", pamTypeConfigFile).
			Bool("createAll", createAll).
			Str("gitRef", gitRef).
			Str("gitRepo", gitRepo).
			Msg("create PAM Provider Type")

		if listTypes {
			fmt.Println("Available store types:")
			sort.Strings(validPAMTypes)
			for _, st := range validPAMTypes {
				fmt.Printf("\t%s\n", st)
			}
			fmt.Println("Use these values with the --name flag.")
			return nil
		}

		if pamTypeConfigFile != "" {
			createdStoreTypes, err := createPAMTypeFromFile(pamTypeConfigFile, kfClient)
			if err != nil {
				fmt.Printf("Failed to create store type from file \"%s\"", err)
				return err
			}

			for _, v := range createdStoreTypes {
				fmt.Printf("Created PAM type \"%s\"\n", v.Name)
			}
			return nil
		}

		if pamProviderTypeName == "" && !createAll {
			prompt := &survey.Select{
				Message: "Choose an option:",
				Options: validPAMTypes,
			}
			var selected string
			err := survey.AskOne(prompt, &selected)
			if err != nil {
				fmt.Println(err)
				return err
			}
			pamProviderTypeName = selected
		}

		for _, v := range validPAMTypes {
			if strings.EqualFold(v, strings.ToUpper(pamProviderTypeName)) || createAll {
				log.Debug().Str("pamType", pamProviderTypeName).Msg("PAM type is valid")
				pamTypeIsValid = true
				break
			}
		}
		if !pamTypeIsValid {
			log.Error().
				Str("pamType", pamProviderTypeName).
				Bool("isValid", pamTypeIsValid).
				Msg("Invalid pam type")
			fmt.Printf("ERROR: Invalid pam type: %s\nValid types are:\n", pamProviderTypeName)
			for _, st := range validPAMTypes {
				fmt.Println(fmt.Sprintf("\t%s", st))
			}
			log.Error().Msg(fmt.Sprintf("Invalid pam type: %s", pamProviderTypeName))
			return fmt.Errorf("invalid pam type: %s", pamProviderTypeName)
		}

		var typesToCreate []string
		if !createAll {
			typesToCreate = []string{pamProviderTypeName}
		} else {
			typesToCreate = validPAMTypes
		}

		pamTypeConfig, stErr := readPAMTypesConfig("", gitRef, gitRepo, offline)
		if stErr != nil {
			log.Error().Err(stErr).Send()
			return stErr
		}
		var createErrors []error

		for _, st := range typesToCreate {
			log.Trace().Msgf("PAM type config: %v", pamTypeConfig[st])
			pamTypeInterface := pamTypeConfig[st].(map[string]interface{})
			pamTypeJSON, _ := json.Marshal(pamTypeInterface)

			var pamTypeObj *keyfactor.ProviderTypeCreateRequest
			convErr := json.Unmarshal(pamTypeJSON, &pamTypeObj)
			if convErr != nil {
				log.Error().Err(convErr).Msg("unable to convert pam type config to JSON")
				createErrors = append(createErrors, fmt.Errorf("%v: %s", st, convErr.Error()))
				continue
			}

			log.Trace().Msgf("PAM type object: %v", pamTypeObj)
			createResp, err := kfClient.CreatePAMProviderType(pamTypeObj)
			if err != nil {
				log.Error().Err(err).Msg("unable to create pam type")
				createErrors = append(createErrors, fmt.Errorf("%v: %s", st, err.Error()))
				continue
			}
			log.Trace().Msgf("Create response: %v", createResp)
			log.Debug().Msg("Converting PAM Provider Type response to JSON")
			jsonString, mErr := json.Marshal(createResp)
			if mErr != nil {
				log.Error().Err(mErr).Send()
				return mErr
			}
			log.Info().Str("output", string(jsonString)).
				Msg("successfully created PAM provider type")
			//outputResult(jsonString, outputFormat)
			outputResult(fmt.Sprintf("PAM provider type %s created with ID: %s", st, createResp.Id), outputFormat)
		}

		if len(createErrors) > 0 {
			errStr := "while creating store types:\n"
			for _, e := range createErrors {
				errStr += fmt.Sprintf("%s\n", e)
			}
			return fmt.Errorf(errStr)
		}

		//var err error
		//if gitRepo != "" {
		//	// get JSON config from integration-manifest on GitHub
		//	log.Debug().
		//		Str("pamProviderTypeName", pamProviderTypeName).
		//		Str("gitRepo", gitRepo).
		//		Str("gitRef", gitRef).
		//		Msg("call: GetTypeFromInternet()")
		//	pamProviderType, err = GetTypeFromInternet(pamProviderTypeName, gitRepo, gitRef, pamProviderType)
		//	log.Debug().Msg("returned: GetTypeFromInternet()")
		//	if err != nil {
		//		log.Error().Err(err).Send()
		//		return err
		//	}
		//}
		//
		//if pamProviderTypeName != "" {
		//	pamProviderType.Name = pamProviderTypeName
		//}
		//
		//log.Info().Str("pamProviderTypeName", pamProviderType.Name).
		//	Msg("creating PAM provider type")
		//
		//log.Debug().Msg("call: PAMProviderCreatePamProviderType()")
		//createdPamProviderType, rErr := kfClient.CreatePAMProviderType(pamProviderType)
		//log.Debug().Msg("returned: PAMProviderCreatePamProviderType()")
		//if rErr != nil {
		//	log.Error().Err(rErr).Send()
		//	return rErr
		//}
		//
		//log.Debug().Msg("Converting PAM Provider Type response to JSON")
		//jsonString, mErr := json.Marshal(createdPamProviderType)
		//if mErr != nil {
		//	log.Error().Err(mErr).Send()
		//	return mErr
		//}
		//log.Info().Str("output", string(jsonString)).
		//	Msg("successfully created PAM provider type")
		//outputResult(jsonString, outputFormat)
		return nil
	},
}

var pamTypesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a defined PAM Provider type by ID or Name.",
	Long:  "Deletes a defined PAM Provider type by ID or Name.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		isExperimental := false
		// Specific flags
		pamProviderTypeId, _ := cmd.Flags().GetString("id")
		pamProviderTypeName, _ := cmd.Flags().GetString("name")
		deleteAll, _ := cmd.Flags().GetBool("all")
		// Debug + expEnabled checks
		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr

		}
		// Log flags
		log.Info().Str("name", pamProviderTypeName).
			Str("id", pamProviderTypeId).
			Msg("delete PAM Provider Type")
		// Authenticate
		kfClient, cErr := initClient(false)
		if cErr != nil {
			return cErr
		}
		// CLI Logic

		if deleteAll {
			if !noPrompt {
				confirmDelete := promptForInteractiveYesNo(
					"Are you sure you want to delete ALL PAM Provider Types? This action" +
						" cannot be undone.",
				)
				if !confirmDelete {
					log.Info().Msg("aborting delete of ALL PAM Provider Types")
					outputResult("Aborted delete of ALL PAM Provider Types", outputFormat)
					return nil
				}
			}
			log.Info().Msg("deleting ALL PAM Provider Types")
			pamProviderTypes, listErr := kfClient.ListPAMProviderTypes()
			if listErr != nil {
				log.Error().Err(listErr).Send()
				return listErr
			}
			if pamProviderTypes == nil || len(*pamProviderTypes) == 0 {
				log.Info().Msg("no PAM provider types to delete")
				outputResult("No PAM provider types to delete", outputFormat)
			}
			for _, pamProviderType := range *pamProviderTypes {
				log.Debug().Str("pamProviderTypeId", pamProviderType.Id).
					Msg("call: PAMProviderDeletePamProviderType()")
				delErr := kfClient.DeletePAMProviderType(pamProviderType.Id)
				if delErr != nil {
					log.Error().Err(delErr).Send()
					outputError(delErr, false, outputFormat)
					continue
				}
				log.Info().Str("id", pamProviderType.Id).Str("name", pamProviderType.Name).
					Msg("successfully deleted PAM provider type")
				outputResult(fmt.Sprintf("Deleted PAM provider type with ID %s", pamProviderType.Id), outputFormat)
			}
			log.Info().Msg("successfully deleted ALL PAM provider types")
			outputResult(fmt.Sprintf("Deleted ALL %d PAM provider types", len(*pamProviderTypes)), outputFormat)
			return nil
		}
		if pamProviderTypeId == "" && pamProviderTypeName == "" {
			cmd.Usage()
			return fmt.Errorf("must supply either a PAM Provider Type `--id` or `--name` to delete")
		}

		if pamProviderTypeId == "" && pamProviderTypeName != "" {
			// Get ID from Name
			log.Debug().Str("pamProviderTypeName", pamProviderTypeName).
				Msg("call: GetPAMProviderTypeByName()")
			pamProviderType, getErr := kfClient.GetPAMProviderTypeByName(pamProviderTypeName)
			log.Debug().Msg("returned: GetPAMProviderTypeByName()")
			if getErr != nil {
				log.Error().Err(getErr).Send()
				return getErr
			}
			pamProviderTypeId = pamProviderType.Id
		}

		log.Debug().Str("pamProviderTypeId", pamProviderTypeId).
			Msg("call: PAMProviderDeletePamProviderType()")
		delErr := kfClient.DeletePAMProviderType(pamProviderTypeId)
		if delErr != nil {
			log.Error().Err(delErr).Send()
			return delErr
		}

		log.Info().Str("name", pamProviderTypeName).
			Str("id", pamProviderTypeId).
			Msg("successfully deleted PAM provider type")
		outputResult(fmt.Sprintf("Deleted PAM provider type with ID %s", pamProviderTypeId), outputFormat)
		return nil
	},
}

func GetPAMTypeInternet(providerName string, repo string, branch string) (interface{}, error) {
	log.Debug().Str("providerName", providerName).
		Str("repo", repo).
		Str("branch", branch).
		Msg("entered: GetPAMTypeInternet()")

	if branch == "" {
		log.Info().Msg("branch not specified, using 'main' by default")
		branch = "main"
	}

	providerUrl := fmt.Sprintf(
		"https://raw.githubusercontent.com/Keyfactor/%s/%s/integration-manifest.json",
		repo,
		branch,
	)
	log.Debug().Str("providerUrl", providerUrl).
		Msg("Getting PAM Type from Internet")
	response, err := http.Get(providerUrl)
	if err != nil {
		log.Error().Err(err).
			Str("providerUrl", providerUrl).
			Msg("error getting PAM Type from Internet")
		return nil, err
	}
	log.Trace().Interface("httpResponse", response).
		Msg("GetPAMTypeInternet")

	//check response status code is 200
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response status: %s", response.Status)
	}

	defer response.Body.Close()

	log.Debug().Msg("Parsing PAM response")
	manifest, iErr := io.ReadAll(response.Body)
	if iErr != nil {
		log.Error().Err(iErr).
			Str("providerUrl", providerUrl).
			Msg("unable to read PAM response")
		return nil, iErr
	}
	log.Trace().Interface("manifest", manifest).Send()

	var manifestJson map[string]interface{}
	log.Debug().Msg("Converting PAM response to JSON")
	jErr := json.Unmarshal(manifest, &manifestJson)
	if jErr != nil {
		log.Error().Err(jErr).
			Str("providerUrl", providerUrl).
			Msg("invalid integration-manifest.json provided")
		return nil, jErr
	}
	log.Debug().Msg("Parsing manifest response for PAM type config")
	pamTypeJson := manifestJson["about"].(map[string]interface{})["pam"].(map[string]interface{})["pam_types"].(map[string]interface{})[providerName]
	if pamTypeJson == nil {
		// Check if only one PAM Type is defined
		pamTypeJson = manifestJson["about"].(map[string]interface{})["pam"].(map[string]interface{})["pam_types"].(map[string]interface{})
		if len(pamTypeJson.(map[string]interface{})) == 1 {
			for _, v := range pamTypeJson.(map[string]interface{}) {
				pamTypeJson = v
			}
		} else {
			return nil, fmt.Errorf("unable to find PAM type %s in manifest on %s", providerName, providerUrl)
		}
	}

	log.Trace().Interface("pamTypeJson", pamTypeJson).Send()
	log.Debug().Msg("returning: GetPAMTypeInternet()")
	return pamTypeJson, nil
}

func getPAMTypesInternet(gitRef string, repo string) (map[string]interface{}, error) {
	//resp, err := http.Get("https://raw.githubusercontent.com/keyfactor/kfutil/main/cmd/pam_types.json")

	baseUrl := "https://raw.githubusercontent.com/Keyfactor/%s/%s/%s"
	if gitRef == "" {
		gitRef = DefaultGitRef
	}
	if repo == "" {
		repo = DefaultGitRepo
	}

	var fileName string
	if repo == "kfutil" {
		fileName = "cmd/pam_types.json"
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
		sTypes := manifest.About.PAM.PAMTypes
		output := make(map[string]interface{})
		for _, st := range sTypes {
			output[st.Name] = st
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

func GetTypeFromInternet[T JSONImportableObject](providerName string, repo string, branch string, returnType *T) (
	*T,
	error,
) {
	log.Debug().Str("providerName", providerName).
		Str("repo", repo).
		Str("branch", branch).
		Msg("entered: GetTypeFromInternet()")

	log.Debug().Msg("call: GetPAMTypeInternet()")
	manifestJSON, err := GetPAMTypeInternet(providerName, repo, branch)
	log.Debug().Msg("returned: GetPAMTypeInternet()")
	if err != nil {
		log.Error().Err(err).Send()
		return new(T), err
	}

	log.Debug().Msg("Converting PAM Type from manifest to bytes")
	manifestJSONBytes, jErr := json.Marshal(manifestJSON)
	if jErr != nil {
		log.Error().Err(jErr).Send()
		return new(T), jErr
	}

	var objectFromJSON T
	log.Debug().Msg("Converting PAM Type from bytes to JSON")
	mErr := json.Unmarshal(manifestJSONBytes, &objectFromJSON)
	if mErr != nil {
		log.Error().Err(mErr).Send()
		return new(T), mErr
	}

	log.Debug().Msg("returning: GetTypeFromInternet()")
	return &objectFromJSON, nil
}

func GetTypeFromConfigFile[T JSONImportableObject](filename string, returnType *T) (*T, error) {
	log.Debug().Str("filename", filename).
		Msg("entered: GetTypeFromConfigFile()")

	log.Debug().Str("filename", filename).
		Msg("Opening PAM Type config file")
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Send()
		return new(T), err
	}

	var rawData map[string]interface{}
	decoder := json.NewDecoder(file)
	dErr := decoder.Decode(&rawData)
	if dErr != nil {
		log.Error().Err(dErr).Send()
		return new(T), dErr
	}

	if _, ok := rawData["about"]; ok {
		// If the file contains the full manifest, extract the PAM type config
		log.Debug().Msg("Parsing PAM Type config from manifest file")
		about := rawData["about"].(map[string]interface{})
		pam := about["pam"].(map[string]interface{})
		pamTypes := pam["pam_types"].(map[string]interface{})
		var pamTypeConfig interface{}
		if len(pamTypes) == 1 {
			for _, v := range pamTypes {
				pamTypeConfig = v
			}
		} else {
			return new(T), fmt.Errorf("multiple PAM types found in manifest file, please provide a file with a single PAM type definition")
		}

		log.Debug().Msg("Converting PAM Type config from manifest to bytes")
		pamTypeConfigBytes, jErr := json.Marshal(pamTypeConfig)
		if jErr != nil {
			log.Error().Err(jErr).Send()
			return new(T), jErr
		}

		var objectFromManifest T
		log.Debug().Msg("Converting PAM Type config from bytes to JSON")
		mErr := json.Unmarshal(pamTypeConfigBytes, &objectFromManifest)
		if mErr != nil {
			log.Error().Err(mErr).Send()
			return new(T), mErr
		}

		log.Debug().Msg("returning: GetTypeFromConfigFile()")
		return &objectFromManifest, nil
	}

	// Rewind file pointer to beginning
	_, sErr := file.Seek(0, io.SeekStart)
	if sErr != nil {
		log.Error().Err(sErr).Send()
		return new(T), sErr
	}

	// If the file contains only the PAM type config
	var objectFromFile T
	log.Debug().Msg("Decoding PAM Type config file")
	decoder = json.NewDecoder(file)
	dErr = decoder.Decode(&objectFromFile)
	if dErr != nil {
		log.Error().Err(dErr).Send()
		return new(T), dErr
	}

	log.Debug().Msg("returning: GetTypeFromConfigFile()")
	return &objectFromFile, nil
}

func getValidPAMTypes(fp string, gitRef string, gitRepo string) []string {
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
		Msg("Reading PAM types config.")
	validPAMTypes, rErr := readPAMTypesConfig(fp, gitRef, gitRepo, offline)
	if rErr != nil {
		log.Error().Err(rErr).Msg("unable to read PAM types")
		return nil
	}
	validPAMTypesList := make([]string, 0, len(validPAMTypes))
	for k := range validPAMTypes {
		validPAMTypesList = append(validPAMTypesList, k)
	}
	sort.Strings(validPAMTypesList)
	return validPAMTypesList
}

func readPAMTypesConfig(fp, gitRef string, gitRepo string, offline bool) (map[string]interface{}, error) {
	log.Debug().Str("file", fp).Str("gitRef", gitRef).Msg(DebugFuncEnter)

	var (
		sTypes map[string]interface{}
		stErr  error
	)
	if offline {
		log.Debug().Msg("Reading pam types config from file")
	} else {
		log.Debug().Msg("Reading pam types config from internet")
		sTypes, stErr = getPAMTypesInternet(gitRef, gitRepo)
	}

	if stErr != nil || sTypes == nil || len(sTypes) == 0 {
		log.Warn().Err(stErr).Msg("Using embedded pam-type definitions")
		var emPAMTypes []interface{}
		if err := json.Unmarshal(EmbeddedPAMTypesJSON, &emPAMTypes); err != nil {
			log.Error().Err(err).Msg("Unable to unmarshal embedded pam type definitions")
			return nil, err
		}
		sTypes, stErr = formatPAMTypes(&emPAMTypes)
		if stErr != nil {
			log.Error().Err(stErr).Msg("Unable to format pam types")
			return nil, stErr
		}
	}

	var content []byte
	var err error
	if sTypes == nil {
		if fp == "" {
			fp = DefaultPAMTypesFileName
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
		log.Error().Err(err).Msg("Unable to unmarshal pam types")
		return nil, err
	}
	return d, nil
}

func formatPAMTypes(pTypesList *[]interface{}) (map[string]interface{}, error) {
	if pTypesList == nil || len(*pTypesList) == 0 {
		return nil, fmt.Errorf("empty pam types list")
	}

	output := make(map[string]interface{})
	for _, v := range *pTypesList {
		v2 := v.(map[string]interface{})
		output[v2["Name"].(string)] = v2
	}

	return output, nil
}

func createPAMTypeFromFile(filename string, kfClient *keyfactor.Client) ([]keyfactor.ProviderTypeResponse, error) {
	// Read the file
	log.Debug().Str("filename", filename).Msg("Reading pam type from file")
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Error().
			Str("filename", filename).
			Err(err).Msg("unable to open file")
		return nil, err
	}

	var pamType keyfactor.ProviderTypeCreateRequest
	var pamTypes []keyfactor.ProviderTypeCreateRequest

	log.Debug().Msg("Decoding JSON file as single pam type")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pamType)
	if err != nil || (pamType.Name == "" || pamType.Parameters == nil) {
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
		pamTypes = manifest.About.PAM.PAMTypes
	} else {
		log.Debug().Msg("Decoded JSON file as single pam type")
		pamTypes = []keyfactor.ProviderTypeCreateRequest{pamType}
	}

	output := make([]keyfactor.ProviderTypeResponse, 0)
	for _, pt := range pamTypes {
		log.Debug().Msgf("Creating certificate pam type %s", pt.Name)
		createResp, cErr := kfClient.CreatePAMProviderType(&pt)
		if cErr != nil {
			log.Error().
				Str("pamType", pt.Name).
				Err(cErr).Msg("unable to create certificate pam type")
			return nil, cErr
		}
		if createResp == nil {
			log.Error().
				Str("pamType", pt.Name).
				Msg("nil response received when creating PAM provider type")
			return nil, fmt.Errorf("nil response received when creating PAM provider type %s", pt.Name)
		}
		output = append(output, *createResp)
		log.Trace().Msgf("Create response: %v", createResp)
		log.Info().Msgf("PAM provider type %s created with ID: %s", pt.Name, createResp.Id)
	}
	// Use the Keyfactor client to create the pam type
	log.Debug().Msg("PAM type created")
	return output, nil
}

func checkBug63171(cmdResp *http.Response, operation string) error {
	if cmdResp != nil && cmdResp.StatusCode == 200 {
		defer cmdResp.Body.Close()
		// .\Admin
		productVersion := cmdResp.Header.Get("X-Keyfactor-Product-Version")
		log.Debug().Str("productVersion", productVersion).Msg("Keyfactor Command Version")
		majorVersionStr := strings.Split(productVersion, ".")[0]
		// Try to convert to int
		majorVersion, err := strconv.Atoi(majorVersionStr)
		if err == nil && majorVersion >= 12 {
			// TODO: Pending resolution of this bug: https://dev.azure.com/Keyfactor/Engineering/_workitems/edit/63171
			errMsg := fmt.Sprintf(
				"PAM Provider %s is not supported in Keyfactor Command version 12 and later, "+
					"please use the Keyfactor Command UI to create PAM Providers", operation,
			)
			oErr := fmt.Errorf(errMsg)
			log.Error().Err(oErr).Send()
			outputError(oErr, true, outputFormat)
			return oErr
		}
	}
	return nil
}

func init() {
	var (
		filePath string
		name     string
		id       string
		repo     string
		branch   string
		all      bool
	)
	// PAM Provider Types
	RootCmd.AddCommand(pamTypesCmd)

	// PAM Provider Types Get
	pamTypesCmd.AddCommand(pamTypesGetCmd)
	pamTypesGetCmd.Flags().StringVarP(&id, "id", "i", "", "ID of the PAM Provider Type.")
	pamTypesGetCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesGetCmd.MarkFlagsMutuallyExclusive("id", "name")

	// PAM Provider Types List
	pamTypesCmd.AddCommand(pamTypesListCmd)

	// PAM Provider Types Create
	pamTypesCmd.AddCommand(pamTypesCreateCmd)
	pamTypesCreateCmd.Flags().StringVarP(
		&filePath,
		FlagFromFile,
		"f",
		"",
		"Path to a JSON file containing the PAM Type Object Data.",
	)
	pamTypesCreateCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesCreateCmd.Flags().BoolVarP(&all, "all", "a", false, "Create all PAM Provider Types.")
	pamTypesCreateCmd.Flags().StringVarP(&repo, "repo", "r", "", "Keyfactor repository name of the PAM Provider Type.")
	pamTypesCreateCmd.Flags().StringVarP(
		&branch,
		"branch",
		"b",
		"",
		"Branch name for the repository. Defaults to 'main'.",
	)

	// PAM Provider Types Delete
	pamTypesCmd.AddCommand(pamTypesDeleteCmd)
	pamTypesDeleteCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the PAM Provider Type.")
	pamTypesDeleteCmd.Flags().StringVarP(&id, "id", "i", "", "ID of the PAM Provider Type.")
	pamTypesDeleteCmd.Flags().BoolVarP(&all, "all", "a", false, "Delete all PAM Provider Types.")
	pamTypesDeleteCmd.MarkFlagsMutuallyExclusive("id", "name", "all")
}
