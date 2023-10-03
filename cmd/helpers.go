package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func boolToPointer(b bool) *bool {
	return &b
}

func checkDebug(v bool) bool {
	envDebug := os.Getenv("KFUTIL_DEBUG")
	envValue, _ := strconv.ParseBool(envDebug)
	switch {
	case (envValue && !v) || (envValue && v):
		//log.SetOutput(os.Stdout)
		return envValue
	case v:
		//log.SetOutput(os.Stdout)
		return v
	default:
		//log.SetOutput(io.Discard)
		return v
	}
}

func csvRemoveLastColumn(inputFile, outputFile string) error {
	// Open the input CSV file
	input, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	// Create a CSV reader for the input file
	reader := csv.NewReader(input)

	// Open the output CSV file
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Create a CSV writer for the output file
	writer := csv.NewWriter(output)

	// Read and process each row
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		// Remove the last column (i.e., the last element in the record slice)
		if len(record) > 0 {
			record = record[:len(record)-1]
		}

		// Write the modified row to the output CSV file
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	// Flush any buffered data to the output file
	writer.Flush()

	// Check for errors during the writing process
	return writer.Error()
}

func csvToMap(filename string) ([]map[string]string, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader
	csvReader := csv.NewReader(file)

	// Read the CSV header to get column names
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	// Initialize a slice to store the maps
	var data []map[string]string

	// Read and process each row
	for {
		// Read a row
		row, err := csvReader.Read()
		if err == nil {
			// Create a map for the row data
			rowMap := make(map[string]string)

			// Populate the map with data from the row
			for i, column := range header {
				rowMap[column] = row[i]
			}

			// Append the map to the data slice
			data = append(data, rowMap)
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}

	return data, nil
}

func findMatchingFiles(pattern string) ([]string, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Use the filepath package to create a glob pattern
	globPattern := filepath.Join(currentDir, pattern)

	// Use the filepath package to perform the glob operation
	matchingFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, err
	}

	return matchingFiles, nil
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func informDebug(debugFlag bool) {
	debugModeEnabled := checkDebug(debugFlag)
	zerolog.SetGlobalLevel(zerolog.Disabled)
	if debugModeEnabled {
		//zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func initClient(flagConfig string, flagProfile string, flagAuthProviderType string, flagAuthProviderProfile string, noPrompt bool, authConfig *api.AuthConfig, saveConfig bool) (*api.Client, error) {
	log.Debug().Msg("Enter initClient()")
	var clientAuth api.AuthConfig
	var commandConfig ConfigurationFile

	log.Debug().Msg("call: authEnvVars()")
	commandConfig, _ = authEnvVars(flagConfig, flagProfile, saveConfig)

	// check if commandConfig is empty
	if commandConfig.Servers == nil || len(commandConfig.Servers) == 0 {
		log.Debug().Msg("commandConfig is empty")
		if flagConfig != "" || !validConfigFileEntry(commandConfig, flagProfile) {
			log.Debug().
				Str("flagConfig", flagConfig).
				Str("flagProfile", flagProfile).
				Bool("noPrompt", noPrompt).
				Bool("saveConfig", saveConfig).
				Msg("call: authConfigFile()")
			commandConfig, _ = authConfigFile(flagConfig, flagProfile, "", noPrompt, saveConfig)
			log.Debug().Msg("complete: authConfigFile()")
		}
	} else {
		log.Debug().Msg("commandConfig is not empty and is valid")
		authProviderProfile, _ := os.LookupEnv("KUTIL_AUTH_PROVIDER_PROFILE")
		log.Debug().Str("authProviderProfile", authProviderProfile).Send()
		if authProviderProfile != "" {
			flagProfile = authProviderProfile
		} else if flagAuthProviderProfile != "" {
			flagProfile = flagAuthProviderProfile
		}
	}
	log.Debug().Str("flagProfile", flagProfile).Send()

	if flagProfile == "" {
		flagProfile = "default"
	}

	//Params from authConfig take precedence over everything else
	if authConfig != nil {
		// replace commandConfig with authConfig params that aren't null or empty
		log.Debug().Str("flagProfile", flagProfile).Msg("Loading profile from authConfig")
		configEntry := commandConfig.Servers[flagProfile]
		if authConfig.Hostname != "" {
			log.Debug().Str("authConfig.Hostname", authConfig.Hostname).
				Str("configEntry.Hostname", configEntry.Hostname).
				Str("flagProfile", flagProfile).
				Msg("Config file profile file hostname is set")
			configEntry.Hostname = authConfig.Hostname
		}
		if authConfig.Username != "" {
			log.Debug().Str("authConfig.Username", authConfig.Username).
				Str("configEntry.Username", configEntry.Username).
				Str("flagProfile", flagProfile).
				Msg("Config file profile file username is set")
			configEntry.Username = authConfig.Username
		}
		if authConfig.Password != "" {
			log.Debug().Str("authConfig.Password", hashSecretValue(authConfig.Password)).
				Str("configEntry.Password", hashSecretValue(configEntry.Password)).
				Str("flagProfile", flagProfile).
				Msg("Config file profile file password is set")
			configEntry.Password = authConfig.Password
		}
		if authConfig.Domain != "" {
			log.Debug().Str("authConfig.Domain", authConfig.Domain).
				Str("configEntry.Domain", configEntry.Domain).
				Str("flagProfile", flagProfile).
				Msg("Config file profile file domain is set")
			configEntry.Domain = authConfig.Domain
		} else if authConfig.Username != "" {
			log.Debug().Str("authConfig.Username", authConfig.Username).
				Str("configEntry.Username", configEntry.Username).
				Str("flagProfile", flagProfile).
				Msg("Attempting to get domain from username")
			tDomain := getDomainFromUsername(authConfig.Username)
			if tDomain != "" {
				log.Debug().Str("configEntry.Domain", tDomain).
					Msg("domain set from username")
				configEntry.Domain = tDomain
			}
		}
		if authConfig.APIPath != "" && configEntry.APIPath == "" {
			log.Debug().Str("authConfig.APIPath", authConfig.APIPath).
				Str("configEntry.APIPath", configEntry.APIPath).
				Str("flagProfile", flagProfile).
				Msg("Config file profile file APIPath is set")
			configEntry.APIPath = authConfig.APIPath
		}
		log.Debug().Str("flagProfile", flagProfile).Msg("Setting configEntry")
		commandConfig.Servers[flagProfile] = configEntry
	}

	if !validConfigFileEntry(commandConfig, flagProfile) {
		if !noPrompt {
			// Auth user interactively
			authConfigEntry := commandConfig.Servers[flagProfile]
			commandConfig, _ = authInteractive(authConfigEntry.Hostname, authConfigEntry.Username, authConfigEntry.Password, authConfigEntry.Domain, authConfigEntry.APIPath, flagProfile, false, false, flagConfig)
		} else {
			//log.Fatalf("[ERROR] auth config profile: %s", flagProfile)
			log.Error().Str("flagProfile", flagProfile).Msg("invalid auth config profile")
			return nil, fmt.Errorf("invalid auth config profile: %s", flagProfile)
		}
	}

	clientAuth.Username = commandConfig.Servers[flagProfile].Username
	clientAuth.Password = commandConfig.Servers[flagProfile].Password
	clientAuth.Domain = commandConfig.Servers[flagProfile].Domain
	clientAuth.Hostname = commandConfig.Servers[flagProfile].Hostname
	clientAuth.APIPath = commandConfig.Servers[flagProfile].APIPath

	log.Debug().Str("clientAuth.Username", clientAuth.Username).
		Str("clientAuth.Password", hashSecretValue(clientAuth.Password)).
		Str("clientAuth.Domain", clientAuth.Domain).
		Str("clientAuth.Hostname", clientAuth.Hostname).
		Str("clientAuth.APIPath", clientAuth.APIPath).
		Msg("Client authentication params")

	log.Debug().Msg("call: api.NewKeyfactorClient()")
	c, err := api.NewKeyfactorClient(&clientAuth)
	log.Debug().Msg("complete: api.NewKeyfactorClient()")

	if err != nil {
		//fmt.Printf("Error connecting to Keyfactor: %s\n", err)
		outputError(err, true, "text")
		//log.Fatalf("[ERROR] creating Keyfactor client: %s", err)
		return nil, fmt.Errorf("unable to create Keyfactor Command client: %s", err)
	}
	log.Info().Msg("Keyfactor Command client created")
	return c, nil
}

func initGenClient(flagConfig string, flagProfile string, noPrompt bool, authConfig *api.AuthConfig, saveConfig bool) (*keyfactor.APIClient, error) {
	var commandConfig ConfigurationFile

	commandConfig, _ = authEnvVars(flagConfig, "", saveConfig)

	if flagConfig != "" || !validConfigFileEntry(commandConfig, flagProfile) {
		commandConfig, _ = authConfigFile(flagConfig, flagProfile, "", noPrompt, saveConfig)
	}

	if flagProfile == "" {
		flagProfile = "default"
	}

	//Params from authConfig take precedence over everything else
	if authConfig != nil {
		// replace commandConfig with authConfig params that aren't null or empty
		configEntry := commandConfig.Servers[flagProfile]
		if authConfig.Hostname != "" {
			configEntry.Hostname = authConfig.Hostname
		}
		if authConfig.Username != "" {
			configEntry.Username = authConfig.Username
		}
		if authConfig.Password != "" {
			configEntry.Password = authConfig.Password
		}
		if authConfig.Domain != "" {
			configEntry.Domain = authConfig.Domain
		} else if authConfig.Username != "" {
			tDomain := getDomainFromUsername(authConfig.Username)
			if tDomain != "" {
				configEntry.Domain = tDomain
			}
		}
		if authConfig.APIPath != "" {
			configEntry.APIPath = authConfig.APIPath
		}
		commandConfig.Servers[flagProfile] = configEntry
	}

	if !validConfigFileEntry(commandConfig, flagProfile) {
		if !noPrompt {
			// Auth user interactively
			authConfigEntry := commandConfig.Servers[flagProfile]
			commandConfig, _ = authInteractive(authConfigEntry.Hostname, authConfigEntry.Username, authConfigEntry.Password, authConfigEntry.Domain, authConfigEntry.APIPath, flagProfile, false, false, flagConfig)
		} else {
			//log.Fatalf("[ERROR] auth config profile: %s", flagProfile)
			log.Error().Str("flagProfile", flagProfile).Msg("invalid auth config profile")
			return nil, fmt.Errorf("auth config profile: %s", flagProfile)
		}
	}

	sdkClientConfig := make(map[string]string)
	sdkClientConfig["host"] = commandConfig.Servers[flagProfile].Hostname
	sdkClientConfig["username"] = commandConfig.Servers[flagProfile].Username
	sdkClientConfig["password"] = commandConfig.Servers[flagProfile].Password
	sdkClientConfig["domain"] = commandConfig.Servers[flagProfile].Domain

	configuration := keyfactor.NewConfiguration(sdkClientConfig)
	c := keyfactor.NewAPIClient(configuration)
	return c, nil
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

func intToPointer(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func isExperimentalFeatureEnabled(expFlag bool, isExperimental bool) (bool, error) {
	envExp := os.Getenv("KFUTIL_EXP")
	envValue, _ := strconv.ParseBool(envExp)
	if envValue {
		return envValue, nil
	}
	if isExperimental && !expFlag {
		return false, fmt.Errorf("experimental features are not enabled. To enable experimental features, use the `--exp` flag or set the `KFUTIL_EXP` environment variable to true")
	}
	return envValue, nil
}

func generateRandomNumberString(length int) string {
	// Generate a random number string of the specified length
	var numberString string
	for i := 0; i < length; i++ {
		numberString += fmt.Sprintf("%d", rand.Intn(10))
	}
	return numberString
}

func loadJSONFile(filename string) (map[string]interface{}, error) {
	// Read the JSON file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a map to unmarshal the JSON data into
	var result map[string]interface{}

	// Create a decoder to decode the JSON data from the file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func logGlobals() {

	if !logInsecure {
		log.Debug().Str("configFile", configFile).
			Str("profile", profile).
			Str("providerType", providerType).
			Str("providerProfile", providerProfile).
			Str("providerConfig", providerConfig).
			Bool("noPrompt", noPrompt).
			Bool("expEnabled", expEnabled).
			Bool("debugFlag", debugFlag).
			Str("kfcUsername", kfcUsername).
			Str("kfcHostName", kfcHostName).
			Str("kfcPassword", hashSecretValue(kfcPassword)).
			Str("kfcDomain", kfcDomain).
			Str("kfcAPIPath", kfcAPIPath).
			Msg("Global Flags")
	} else {
		log.Debug().Str("configFile", configFile).
			Str("profile", profile).
			Str("providerType", providerType).
			Str("providerProfile", providerProfile).
			Str("providerConfig", providerConfig).
			Bool("noPrompt", noPrompt).
			Bool("expEnabled", expEnabled).
			Bool("debugFlag", debugFlag).
			Str("kfcUsername", kfcUsername).
			Str("kfcHostName", kfcHostName).
			Str("kfcPassword", kfcPassword).
			Str("kfcDomain", kfcDomain).
			Str("kfcAPIPath", kfcAPIPath).
			Msg("Global Flags")
	}

}

func mapToCSV(data []map[string]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header using keys from the first map
	var header []string
	if len(data) > 0 {
		for key := range data[0] {
			header = append(header, key)
		}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	// Write map data to CSV
	for _, row := range data {
		var record []string
		for _, key := range header {
			record = append(record, row[key])
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func outputError(err error, isFatal bool, format string) {
	if isFatal {
		if format == "json" {
			fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err))
		} else {
			fmt.Errorf(fmt.Sprintf("Fatal error: %s", err))
		}
	}
	if format == "json" {
		fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err))
	} else {
		fmt.Println(fmt.Sprintf("Error: %s", err))
	}
}

func outputResult(result interface{}, format string) {
	if format == "json" {
		fmt.Println(result)
	} else {
		fmt.Println(fmt.Sprintf("%s", result))
	}
}

func readCSVHeader(filename string) ([]string, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader
	csvReader := csv.NewReader(file)

	// Read the header row
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	return header, nil
}

func stringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func storeTypeIdentifierFlagCheck(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("store-type-name") && !cmd.Flags().Changed("store-type-id") {
		inputErr := fmt.Errorf("'store-type-id' or 'store-type-name' must be provided")
		cmd.Usage()
		return inputErr
	}
	return nil
}

func warnExperimentalFeature(expEnabled bool, isExperimental bool) error {
	_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
	if expErr != nil {
		//fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
		log.Error().Err(expErr)
		return expErr
	}
	return nil
}

func writeJSONFile(filename string, data interface{}) error {
	// Open the JSON file for writing
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create an encoder to encode the data to JSON and write it to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func returnHttpErr(resp *http.Response, err error) error {
	if resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Err(err).Str("httpResponseCode", resp.Status).
			Str("httpResponseBody", string(body)).
			Msg("unable to create PAM provider")

		// parse response body into map[string]interface{}
		var responseMap map[string]interface{}
		if jErr := json.Unmarshal(body, &responseMap); jErr != nil {
			// check of the response body is plaintext
			if string(body) != "" {
				errMsg := fmt.Errorf("%s- %s", resp.Status, string(body))
				return errMsg
			}
			log.Error().Err(jErr).Msg("unable to parse response body")
			return jErr
		}

		errMsg := fmt.Errorf("%s- %s", resp.Status, responseMap["Message"].(string))

		return errMsg
	}
	log.Error().Err(err).Str("httpResponseCode", resp.Status).
		Msg("unable to create PAM provider")
	return err
}
