// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
)

const DefaultConfigFileName = "command_config.json"

type ConfigurationFile struct {
	Servers map[string]ConfigurationFileEntry `json:"servers"`
}

type ConfigurationFileEntry struct {
	Hostname     string       `json:"host"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	Domain       string       `json:"domain"`
	APIPath      string       `json:"api_path"`
	AuthProvider AuthProvider `json:"auth_provider"`
}

type AuthProvider struct {
	Type       string      `json:"type"`
	Profile    string      `json:"profile"`
	Parameters interface{} `json:"parameters"`
}

type AuthProviderAzureIdParams struct {
	SecretName     string `json:"secret_name"`
	AzureVaultName string `json:"vault_name"`
	TenantId       string `json:"tenant_id;omitempty"`
	SubscriptionId string `json:"subscription_id;omitempty"`
	ResourceGroup  string `json:"resource_group;omitempty"`
}

var validAuthProviders = []string{"azure-id", "azure-cli", "azid", "azcli"}
var loginCmd = &cobra.Command{
	Use:        "login",
	Aliases:    nil,
	SuggestFor: nil,
	Short:      "User interactive login to Keyfactor. Stores the credentials in the config file '$HOME/.keyfactor/command_config.json'.",
	GroupID:    "",
	Long: `Will prompt the user for a username and password and then attempt to login to Keyfactor.
You can provide the --config flag to specify a config file to use. If not provided, the default
config file will be used. The default config file is located at $HOME/.keyfactor/command_config.json.
To prevent the prompt for username and password, use the --no-prompt flag. If this flag is provided then
the CLI will default to using the environment variables: KEYFACTOR_HOSTNAME, KEYFACTOR_USERNAME, 
KEYFACTOR_PASSWORD and KEYFACTOR_DOMAIN.

WARNING: The username and password will be stored in the config file in plain text at: 
'$HOME/.keyfactor/command_config.json.'
`,
	Example:                "",
	ValidArgs:              nil,
	ValidArgsFunction:      nil,
	Args:                   nil,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Annotations:            nil,
	Version:                "",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)

		var (
			authConfigFileErrs []error
			authConfig         ConfigurationFile
			authErr            error
		)

		if profile == "" && configFile == "" {
			profile = "default"
			// Check for environment variables
			var authEnvErr []error

			log.Println("[DEBUG] Checking for environment variables for kfutil configuration data.")
			if noPrompt {
				// First try to auth with environment variables
				authConfig, authEnvErr = authEnvVars(configFile, profile, true) // always save config file is login is called
				if authEnvErr != nil {
					// Print out the error messages
					for _, err := range authEnvErr {
						log.Println(err)
					}
				}
				if !validConfigFileEntry(authConfig, profile) {
					// Attempt to auth with config file
					log.Println("[DEBUG] Attempting to authenticate via config 'default' profile.")
					authConfig, authEnvErr = authConfigFile(configFile, profile, "", noPrompt, true) // always save config file is login is called
					if authEnvErr != nil {
						// Print out the error messages
						for _, err := range authEnvErr {
							log.Println(err)
						}
					}
					if !validConfigFileEntry(authConfig, profile) {
						errMsg := fmt.Sprintf("Unable to authenticate with environment variables or config file. Please review setup.")
						log.Fatal(errMsg)
						return
					}
				}
			} else {
				// Try user interactive login
				authConfig, _ = authEnvVars(configFile, profile, false) // Silently load via env what you can
				if !validConfigFileEntry(authConfig, profile) || !noPrompt {
					existingAuth := authConfig.Servers[profile]
					authConfig, authErr = authInteractive(existingAuth.Hostname, existingAuth.Username, existingAuth.Password, existingAuth.Domain, existingAuth.APIPath, profile, !noPrompt, true, configFile)
					if authErr != nil {
						log.Fatal(authErr)
						return
					}

				}
			}
			fmt.Println(fmt.Sprintf("Login successful!"))
		} else if configFile != "" || profile != "" {
			// Attempt to auth with config file
			authConfig, authConfigFileErrs = authConfigFile(configFile, profile, "", noPrompt, true) // always save config file is login is called
			if authConfigFileErrs != nil {
				// Print out the error messages
				for _, err := range authConfigFileErrs {
					log.Println(err)
				}
			}
			if !validConfigFileEntry(authConfig, profile) && !noPrompt {
				//Attempt to auth with user interactive login
				authEntry := authConfig.Servers[profile]
				authConfig, authErr = authInteractive(authEntry.Hostname, authEntry.Username, authEntry.Password, authEntry.Domain, authEntry.APIPath, profile, false, true, configFile)
				if authErr != nil {
					log.Println(authErr)
					fmt.Println("Login failed!")
				}
			} else {
				fmt.Println(fmt.Sprintf("Login successful!"))
			}
		}
	},
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	CompletionOptions:          cobra.CompletionOptions{},
	TraverseChildren:           false,
	Hidden:                     true,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
}

func init() {
	RootCmd.AddCommand(loginCmd)
}

func validConfig(hostname string, username string, password string, domain string) bool {
	if hostname == "" || username == "" || password == "" {
		return false
	}
	if domain == "" && (!strings.Contains(username, "@") || !strings.Contains(username, "\\")) {
		return false
	}
	return true
}

func validConfigFileEntry(configFile ConfigurationFile, profile string) bool {
	if profile == "" {
		profile = "default"
	}
	if configFile.Servers[profile].Hostname == "" || configFile.Servers[profile].Username == "" || configFile.Servers[profile].Password == "" {
		return false
	}
	if configFile.Servers[profile].Domain == "" && (!strings.Contains(configFile.Servers[profile].Username, "@") || !strings.Contains(configFile.Servers[profile].Username, "\\")) {
		return false
	}
	return true
}

func getDomainFromUsername(username string) string {
	if strings.Contains(username, "@") {
		return strings.Split(username, "@")[1]
	} else if strings.Contains(username, "\\") {
		return strings.Split(username, "\\")[0]
	}
	return ""
}

func createConfigFile(hostname string, username string, password string, domain string, apiPath string, profileName string) ConfigurationFile {
	output := ConfigurationFile{
		Servers: map[string]ConfigurationFileEntry{
			profileName: {
				Hostname: hostname,
				Username: username,
				Password: password,
				Domain:   domain,
				APIPath:  apiPath,
			},
		},
	}
	return output
}

func createAuthConfig(hostname string, username string, password string, domain string, apiPath string) api.AuthConfig {
	output := api.AuthConfig{
		Hostname: hostname,
		Username: username,
		Password: password,
		Domain:   domain,
		APIPath:  apiPath,
	}
	return output
}

func createAuthConfigFromConfigFile(configFileEntry ConfigurationFileEntry) api.AuthConfig {
	output := api.AuthConfig{
		Hostname: configFileEntry.Hostname,
		Username: configFileEntry.Username,
		Password: configFileEntry.Password,
		Domain:   configFileEntry.Domain,
		APIPath:  configFileEntry.APIPath,
	}
	return output
}

func promptForInteractiveParameter(parameterName string, defaultValue string) string {
	var input string
	fmt.Printf("Enter %s [%s]: \n", parameterName, defaultValue)
	_, err := fmt.Scanln(&input)
	handleInteractiveError(err, parameterName)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptForInteractivePassword(parameterName string, defaultValue string) string {
	passwordFill := ""
	if defaultValue != "" {
		passwordFill = "********"
	}
	//log.Println("[DEBUG] password: " + defaultValue)

	fmt.Printf("Enter %s [%s]: \n", parameterName, passwordFill)
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	// check if bytePassword is empty if so the return the default value
	if len(bytePassword) == 0 {
		return defaultValue
	}
	password := string(bytePassword)
	fmt.Println("")
	return password
}

func handleInteractiveError(err error, parameterName string) {
	if err != nil {
		if err.Error() != "unexpected newline" {
			errMsg := fmt.Sprintf("getting %s: %v", parameterName, err)
			log.Println(fmt.Sprintf("[ERROR] %s", errMsg))
		}
	}
}

func saveConfigFile(configFile ConfigurationFile, configPath string, profileName string) (ConfigurationFile, error) {
	if profileName == "" {
		profileName = "default"
	}

	configurationErr := createOrUpdateConfigurationFile(configFile.Servers[profileName], profileName, configPath)
	if configurationErr != nil {
		log.Fatal("[ERROR] Unable to save configuration file to disk: ", configurationErr)
		return configFile, configurationErr
	}
	loadedConfig, loadErr := loadConfigurationFile(configPath, true)
	if loadErr != nil {
		log.Fatal("[ERROR] Unable to load configuration file after save: ", loadErr)
		return configFile, loadErr
	}
	return loadedConfig, nil
}

func authInteractive(hostname string, username string, password string, domain string, apiPath string, profileName string, forcePrompt bool, saveConfig bool, configPath string) (ConfigurationFile, error) {
	if hostname == "" || forcePrompt {
		hostname = promptForInteractiveParameter("Keyfactor Command hostname", hostname)
	}
	if username == "" || forcePrompt {
		username = promptForInteractiveParameter("Keyfactor Command username", username)
	}
	if password == "" || forcePrompt {
		password = promptForInteractivePassword("Keyfactor Command password", password)
	}
	if domain == "" || forcePrompt {
		domain = getDomainFromUsername(username)
		if domain == "" {
			domain = promptForInteractiveParameter("Keyfactor Command AD domain", domain)
		}
	}
	if apiPath == "" || forcePrompt {
		apiPath = promptForInteractiveParameter("Keyfactor Command API path", apiPath)
	}

	if profileName == "" {
		profileName = "default"
	}

	confFile := createConfigFile(hostname, username, password, domain, apiPath, profileName)

	if saveConfig {
		savedConfigFile, saveErr := saveConfigFile(confFile, configPath, profileName)
		if saveErr != nil {
			log.Println("[ERROR] Unable to save configuration file to disk: ", saveErr)
			return confFile, saveErr
		}
		return savedConfigFile, nil
	}
	return confFile, nil
}

func prepHomeDir() (string, error) {
	// Set up home directory config
	userHomeDir, hErr := os.UserHomeDir()

	if hErr != nil {
		log.Println("[ERROR] Error getting user home directory: ", hErr)
		fmt.Println("Error getting user home directory: ", hErr)
		fmt.Println("Using current directory to write config file '" + DefaultConfigFileName + "'")
		userHomeDir = "." // Default to current directory
	} else {
		userHomeDir = fmt.Sprintf("%s/.keyfactor", userHomeDir)
	}
	log.Println("[DEBUG] Configuration directory: ", userHomeDir)
	_, err := os.Stat(userHomeDir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(userHomeDir, 0700)
		if errDir != nil {
			fmt.Println("Unable to create login config file. ", errDir)
			log.Printf("[ERROR] creating directory: %s", errDir)
			return userHomeDir, errDir
		}
	}
	return userHomeDir, hErr
}

func loadConfigFileData(profileName string, configPath string, noPrompt bool, configurationFile ConfigurationFile) (string, string, string, string, string) {
	envConfig, _ := authEnvVars(configPath, profileName, false) //Load env vars first
	log.Println("[INFO] Using profileName: ", profileName)
	// Check if profileName exists in config file
	configProfile, profileExists := configurationFile.Servers[profileName]
	if !profileExists {
		log.Println("[WARN] Profile not found in config file.")
		if noPrompt {
			return envConfig.Servers[profileName].Hostname, envConfig.Servers[profileName].Username, envConfig.Servers[profileName].Password, envConfig.Servers[profileName].Domain, envConfig.Servers[profileName].APIPath
		}
	} else {
		log.Println("[INFO] Using kfutil config profileName: ", profileName)
		hostName := configProfile.Hostname
		userName := configProfile.Username
		password := configProfile.Password
		domain := configProfile.Domain
		apiPath := configProfile.APIPath
		authProvider := configProfile.AuthProvider

		if authProvider.Type != "" && authProvider.Parameters != nil {
			// cALL authViaProviderParams
			authConfig, authErr := authViaProviderParams(&authProvider)
			if authErr != nil {
				log.Println("[ERROR] Unable to authenticate via provider: ", authErr)
				return "", "", "", "", ""
			}

			// Check if authProvider profile is set
			if authProvider.Profile == "" {
				authProvider.Profile = "default"
			}

			hostName = authConfig.Servers[authProvider.Profile].Hostname
			userName = authConfig.Servers[authProvider.Profile].Username
			password = authConfig.Servers[authProvider.Profile].Password
			domain = authConfig.Servers[authProvider.Profile].Domain
			apiPath = authConfig.Servers[authProvider.Profile].APIPath

			return hostName, userName, password, domain, apiPath
		}

		if hostName == "" && envConfig.Servers[profileName].Hostname != "" {
			hostName = envConfig.Servers[profileName].Hostname
		}
		if userName == "" && envConfig.Servers[profileName].Username != "" {
			userName = envConfig.Servers[profileName].Username
		}
		if password == "" && envConfig.Servers[profileName].Password != "" {
			password = envConfig.Servers[profileName].Password
		}
		if domain == "" && envConfig.Servers[profileName].Domain != "" {
			domain = envConfig.Servers[profileName].Domain
		}
		if apiPath == "" && envConfig.Servers[profileName].APIPath != "" {
			apiPath = envConfig.Servers[profileName].APIPath
		}

		return hostName, userName, password, domain, apiPath
	}
	return "", "", "", "", ""
}

func authViaProviderParams(providerConfig *AuthProvider) (ConfigurationFile, error) {

	providerType := providerConfig.Type
	// First check if provider type and provider config are not empty
	if providerType == "" || providerConfig == nil {
		return ConfigurationFile{}, fmt.Errorf("provider type and provider config cannot be empty")
	}

	// Check if auth provider is valid
	if !validAuthProvider(providerType) {
		return ConfigurationFile{}, fmt.Errorf("invalid auth provider type '%s'. Valid auth providers are: %v", providerType, validAuthProviders)
	}

	// Check if provider type matches requested provider type
	switch providerType {
	case "azure-id", "azid", "az-id", "azureid":
		//load provider config params into AuthProviderAzureIdParams struct
		log.Println("[DEBUG] authenticating via azure-id provider")
		var providerParams AuthProviderAzureIdParams
		paramsJson, _ := json.Marshal(providerConfig.Parameters)
		jsonErr := json.Unmarshal(paramsJson, &providerParams)
		if jsonErr != nil {
			log.Println("[ERROR] unable to unmarshal providerParams: ", jsonErr)
			return ConfigurationFile{}, jsonErr
		}

		// Check if required params are set
		if providerParams.SecretName == "" || providerParams.AzureVaultName == "" {
			return ConfigurationFile{}, fmt.Errorf("provider params secret_name and vault_name are required")
		}
		log.Println("[DEBUG] providerParams: ", providerParams)
		return providerParams.authenticate()
	case "az-cli", "azcli", "azure-cli", "azurecli":
		log.Println("[DEBUG] authenticating via azure-cli provider")
		break
	default:
		log.Println("[ERROR] invalid auth provider type")
		break
	}
	return ConfigurationFile{}, fmt.Errorf("invalid auth provider type '%s'. Valid auth providers are: %v", providerType, validAuthProviders)
}

func validAuthProvider(providerType string) bool {
	if providerType == "" {
		return true // default to username/password
	}
	for _, validProvider := range validAuthProviders {
		if validProvider == providerType {
			return true
		}
	}
	return false
}

func authConfigFile(configPath string, profileName string, authProviderProfile string, noPrompt bool, saveConfig bool) (ConfigurationFile, []error) {
	var configurationFile ConfigurationFile
	var (
		hostName string
		userName string
		password string
		domain   string
		apiPath  string

		//hostSet    bool
		//userSet    bool
		//passSet    bool
		//domainSet  bool
		//apiPathSet bool
		//profileSet bool
	)

	log.Println("[DEBUG] Using profileName: ", profileName)

	if configPath == "" {
		userHomeDir, _ := prepHomeDir()
		configPath = fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName)
	}
	configurationFile, _ = loadConfigurationFile(configPath, noPrompt)

	if configurationFile.Servers == nil {
		configurationFile.Servers = make(map[string]ConfigurationFileEntry)
	}
	if profileName == "" {
		profileName = "default"
	}

	hostName, userName, password, domain, apiPath = loadConfigFileData(profileName, configPath, noPrompt, configurationFile)

	confFile := createConfigFile(hostName, userName, password, domain, apiPath, profileName)
	if saveConfig {
		savedConfigFile, saveErr := saveConfigFile(confFile, configPath, profileName)
		if saveErr != nil {
			return confFile, []error{saveErr}
		}
		return savedConfigFile, nil
	}

	configurationFile.Servers[profileName] = confFile.Servers[profileName]
	return configurationFile, nil
}

func authEnvProvider(authProvider *AuthProvider, configProfile string) (ConfigurationFile, []error) {
	log.Println(fmt.Sprintf("[INFO] authenticating with auth provider '%s' params from environment variables", authProvider.Type))

	if configProfile == "" {
		configProfile = "default"
	}
	// attempt to cast authProvider.Parameters to string
	authProviderParams, ok := authProvider.Parameters.(string)
	if !ok {
		log.Println("[ERROR] unable to cast authProvider.Parameters to string")
		return ConfigurationFile{}, []error{fmt.Errorf("invalid configuration, unable to cast authProvider.Parameters to string")}
	}

	if strings.HasPrefix(authProviderParams, "{") && strings.HasSuffix(authProviderParams, "}") {
		// authProviderParams is a json string
		log.Println("[DEBUG] authProviderParams is a json string")
		var providerParams interface{}
		log.Println("[DEBUG] converting authProviderParams to unescaped json")
		jsonErr := json.Unmarshal([]byte(authProviderParams), &providerParams)
		if jsonErr != nil {
			log.Println("[ERROR] unable to unmarshal authProviderParams: ", jsonErr)
			return ConfigurationFile{}, []error{jsonErr}
		}
		authProvider.Parameters = providerParams
	} else {
		// attempt to read as json file path
		log.Println("[DEBUG] authProviderParams is a json file path")
		var providerParams interface{}
		var providerConfigFile ConfigurationFile
		var authProviderConfig AuthProvider
		log.Println("[DEBUG] opening authProviderParams file ", authProviderParams)

		jsonFile, jsonFileErr := os.Open(authProviderParams)
		if jsonFileErr != nil {
			log.Println("[ERROR] unable to open authProviderParams file: ", jsonFileErr)
			return ConfigurationFile{}, []error{jsonFileErr}
		}
		defer jsonFile.Close()
		log.Println(fmt.Sprintf("[DEBUG] reading authProviderParams file %s as bytes", authProviderParams))
		jsonBytes, jsonBytesErr := os.ReadFile(authProviderParams)
		if jsonBytesErr != nil {
			log.Println("[ERROR] unable to read authProviderParams file: ", jsonBytesErr)
			return ConfigurationFile{}, []error{jsonBytesErr}
		}
		log.Println("[DEBUG] converting authProviderParams to unescaped json")
		jsonErr := json.Unmarshal(jsonBytes, &providerParams)
		if jsonErr != nil {
			log.Println("[ERROR] unable to unmarshal authProviderParams: ", jsonErr)
			return ConfigurationFile{}, []error{jsonErr}
		}

		//Check if provider params is a configuration file
		log.Println("[DEBUG] checking if authProviderParams is a configuration file")
		jsonErr = json.Unmarshal(jsonBytes, &providerConfigFile)
		if jsonErr == nil && providerConfigFile.Servers != nil {
			// lookup params based on configProfile
			log.Println("[DEBUG] authProviderParams is a configuration file")
			// check to see if profile exists in config file
			if _, isConfigFile := providerConfigFile.Servers[configProfile]; isConfigFile {
				log.Println(fmt.Sprintf("[DEBUG] profile '%s' found in authProviderParams file", configProfile))
				providerParams = providerConfigFile.Servers[configProfile]
				// check if providerParams is a ConfigurationFileEntry
				if _, isConfigFileEntry := providerParams.(ConfigurationFileEntry); !isConfigFileEntry {
					log.Println("[ERROR] unable to cast providerParams to ConfigurationFileEntry")
					return ConfigurationFile{}, []error{fmt.Errorf("invalid configuration, unable to cast providerParams to ConfigurationFileEntry")}
				}
				// set providerParams to ConfigurationFileEntry.AuthProvider.Parameters
				providerParams = providerConfigFile.Servers[configProfile].AuthProvider.Parameters
			} else {
				log.Println(fmt.Sprintf("[DEBUG] profile '%s' not found in authProviderParams file", configProfile))
				return ConfigurationFile{}, []error{fmt.Errorf("profile '%s' not found in authProviderParams file", configProfile)}
			}
		} else {
			//check if provider params is an AuthProvider
			log.Println("[DEBUG] checking if authProviderParams is an AuthProvider")
			//conver providerParams to unescaped json
			log.Println("[DEBUG] converting authProviderParams to unescaped json")

			//check if providerParams is a map[string]interface{}
			if _, isMap := providerParams.(map[string]interface{}); isMap {
				//check if 'auth_provider' key exists and if it does convert to json bytes
				if _, isAuthProvider := providerParams.(map[string]interface{})["auth_provider"]; isAuthProvider {
					log.Println("[DEBUG] authProviderParams is a map[string]interface{}")
					log.Println("[DEBUG] converting authProviderParams to unescaped json")
					jsonBytes, jsonBytesErr = json.Marshal(providerParams.(map[string]interface{})["auth_provider"])
					if jsonBytesErr != nil {
						log.Println("[ERROR] unable to marshal authProviderParams: ", jsonBytesErr)
						return ConfigurationFile{}, []error{jsonBytesErr}
					}
				}
			}

			jsonErr = json.Unmarshal(jsonBytes, &authProviderConfig)
			if jsonErr == nil && authProviderConfig.Type != "" && authProviderConfig.Parameters != nil {
				log.Println("[DEBUG] authProviderParams is an AuthProvider")
				providerParams = authProviderConfig.Parameters
			}
		}
		authProvider.Parameters = providerParams
	}
	log.Println("[INFO] Attempting to fetch kfutil creds from auth provider ", authProvider)
	configFile, authErr := authViaProviderParams(authProvider)
	if authErr != nil {
		log.Println("[ERROR] Unable to authenticate via provider: ", authErr)
		return ConfigurationFile{}, []error{authErr}
	}
	log.Println("[INFO] Successfully retrieved kfutil creds via auth provider")
	return configFile, nil
}

func authEnvVars(configPath string, profileName string, saveConfig bool) (ConfigurationFile, []error) {
	hostname, hostSet := os.LookupEnv("KEYFACTOR_HOSTNAME")
	username, userSet := os.LookupEnv("KEYFACTOR_USERNAME")
	password, passSet := os.LookupEnv("KEYFACTOR_PASSWORD")
	domain, domainSet := os.LookupEnv("KEYFACTOR_DOMAIN")
	apiPath, apiPathSet := os.LookupEnv("KEYFACTOR_API_PATH")
	envProfileName, _ := os.LookupEnv("KFUTIL_PROFILE")
	authProviderType, _ := os.LookupEnv("KFUTIL_AUTH_PROVIDER_TYPE")
	authProviderProfile, _ := os.LookupEnv("KUTIL_AUTH_PROVIDER_PROFILE")
	authProviderParams, _ := os.LookupEnv("KFUTIL_AUTH_PROVIDER_PARAMS") // this is a json string or a json file path

	if authProviderType != "" || authProviderParams != "" {
		if authProviderParams == "" {
			authProviderParams = fmt.Sprintf("%s/.keyfactor/%s", os.Getenv("HOME"), DefaultConfigFileName)
		}
		if authProviderProfile == "" {
			authProviderProfile = "default"
		}
		authProvider := AuthProvider{
			Type:       authProviderType,
			Profile:    authProviderProfile,
			Parameters: authProviderParams,
		}
		//check if authProviderParams is a json string or a json file path
		return authEnvProvider(&authProvider, profileName)
	}

	if profileName == "" && envProfileName != "" {
		profileName = envProfileName
	} else if profileName == "" {
		profileName = "default"
	}

	log.Printf("KEYFACTOR_HOSTNAME: %s\n", hostname)
	log.Printf("KEYFACTOR_USERNAME: %s\n", username)
	log.Printf("KEYFACTOR_DOMAIN: %s\n", domain)

	if domain == "" && username != "" {
		domain = getDomainFromUsername(username)
	}

	var outputErr []error
	if !hostSet {
		outputErr = append(outputErr, fmt.Errorf("KEYFACTOR_HOSTNAME environment variable not set. Please set the KEYFACTOR_HOSTNAME environment variable"))
	}
	if !userSet {
		outputErr = append(outputErr, fmt.Errorf("KEYFACTOR_USERNAME environment variable not set. Please set the KEYFACTOR_USERNAME environment variable"))
	}
	if !passSet {
		outputErr = append(outputErr, fmt.Errorf("KEYFACTOR_PASSWORD environment variable not set. Please set the KEYFACTOR_PASSWORD environment variable"))
	}
	if !domainSet {
		outputErr = append(outputErr, fmt.Errorf("KEYFACTOR_DOMAIN environment variable not set. Please set the KEYFACTOR_DOMAIN environment variable"))
	}
	if !apiPathSet {
		apiPath = defaultAPIPath
		apiPathSet = true
	}

	if !hostSet && !userSet && !passSet && !domainSet {
		return ConfigurationFile{}, outputErr
	}

	confFile := createConfigFile(hostname, username, password, domain, apiPath, profileName)

	if len(outputErr) > 0 {
		return confFile, outputErr
	}

	if saveConfig {
		savedConfigFile, saveErr := saveConfigFile(confFile, configPath, profileName)
		if saveErr != nil {
			return confFile, []error{saveErr}
		}
		return savedConfigFile, nil
	}
	return confFile, nil
}

func createOrUpdateConfigurationFile(cfgFile ConfigurationFileEntry, profile string, configPath string) error {
	log.Println("[INFO] Creating or updating configuration file")
	log.Println("[DEBUG] configuration file path: ", configPath)

	if len(profile) == 0 {
		profile = "default"
		log.Println("[WARN] no config name provided. Using '" + profile + "'.")
	}
	//check if configPath exists
	if configPath == "" {
		defaultDir, _ := os.UserHomeDir()
		configPath = path.Join(defaultDir, ".keyfactor", DefaultConfigFileName)
		log.Println("[WARN] no config path provided. Using '" + configPath + "'.")
	}
	confFileExists, fileErr := os.Stat(configPath)
	if fileErr != nil {
		log.Println("[WARN] ", fileErr)
	}

	existingConfig, _ := loadConfigurationFile(configPath, true)
	if len(existingConfig.Servers) > 0 {
		// check if the config name already exists
		if _, ok := existingConfig.Servers[profile]; ok {
			log.Println(fmt.Sprintf("[WARN] config name '%s' already exists. Overwriting existing config.", profile))
			log.Println(fmt.Sprintf("[DEBUG] existing config: %v", existingConfig.Servers[profile]))
			log.Println(fmt.Sprintf("[DEBUG] new config: %v", cfgFile))
			// print out the diff between the two configs
			diff := cmp.Diff(existingConfig.Servers[profile], cfgFile)
			if len(diff) == 0 && confFileExists != nil {
				log.Println("[DEBUG] no configuration changes detected")
				return nil
			}
			log.Println(fmt.Sprintf("[DEBUG] diff: %s", diff))
		}
		existingConfig.Servers[profile] = cfgFile
	} else {
		log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", profile))
		existingConfig.Servers = make(map[string]ConfigurationFileEntry)
		existingConfig.Servers[profile] = cfgFile
	}

	log.Println("[DEBUG] kfcfg entry: ", cfgFile)

	f, fErr := os.OpenFile(fmt.Sprintf("%s", configPath), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	defer f.Close()
	if fErr != nil {
		msg := fmt.Sprintf("Error creating command configuration file %s: %s", configPath, fErr)
		fmt.Println(msg)
		log.Println("[ERROR] creating command configuration file: ", fErr)
		return fErr
	}

	// convert existingConfig to json
	jsonData, jsErr := json.MarshalIndent(existingConfig, "", "  ")
	if jsErr != nil {
		fmt.Println("Unable to read kfcfg file due to invalid format. ", jsErr)
		log.Println("[ERROR] marshalling kfcfg file: ", jsErr)
		return jsErr
	}
	_, enErr := f.Write(jsonData)
	if enErr != nil {
		fmt.Println("Unable to read kfcfg file due to invalid format. ", enErr)
		log.Println("[ERROR] encoding kfcfg file: ", enErr)
		return enErr
	}
	return nil
}

func loadConfigurationFile(filePath string, silent bool) (ConfigurationFile, error) {

	//data := ConfigurationFile{Servers: make(map[string]ConfigurationFileEntry)}
	data := ConfigurationFile{}
	if filePath == "" {
		defaultDir, _ := os.UserHomeDir()
		filePath = path.Join(defaultDir, ".keyfactor", DefaultConfigFileName)
	}

	// attempt to make the directory if it doesn't exist
	dirPath := path.Dir(filePath)
	if _, dirErr := os.Stat(dirPath); os.IsNotExist(dirErr) {
		log.Println("[DEBUG] config directory does not exist, creating: ", dirPath)
		err := os.MkdirAll(dirPath, 0700)
		if err != nil {
			log.Println("[ERROR] creating config directory: ", err)
			return data, err
		}
		return data, nil // return empty data since the directory didn't exist the file won't exist
	}

	// check if file exists
	if _, fileErr := os.Stat(filePath); os.IsNotExist(fileErr) {
		log.Println("[DEBUG] config file does not exist: ", filePath)
		return data, nil // return empty data since the file doesn't exist
	}

	f, rFErr := os.ReadFile(filePath)
	if rFErr != nil {
		if !silent {
			fmt.Println(fmt.Sprintf("Unable to read config file '%s'.", rFErr))
			log.Fatal("[FATAL] Error reading config file: ", rFErr)
		}
		return data, rFErr
	}

	// Try to unmarshal as a single entry first
	var singleEntry ConfigurationFileEntry
	sjErr := json.Unmarshal(f, &singleEntry)
	if sjErr != nil {
		log.Println(fmt.Sprintf("[DEBUG] config file '%s' is a not single entry, will attempt to parse as v1 config file", filePath))
	} else if (singleEntry != ConfigurationFileEntry{}) {
		// if we successfully unmarshalled a single entry, add it to the map as the default entry
		log.Println(fmt.Sprintf("[DEBUG] config file '%s' is a single entry, adding to map", filePath))
		data.Servers = make(map[string]ConfigurationFileEntry)
		data.Servers["default"] = singleEntry
		return data, nil
	}

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		log.Println("[ERROR] decoding config file: ", jErr)
		return data, jErr
	}

	return data, nil
}

func createAuthConfigFromParams(hostname string, username string, password string, domain string, apiPath string) *api.AuthConfig {
	output := api.AuthConfig{
		Hostname: hostname,
		Username: username,
		Password: password,
		Domain:   domain,
		APIPath:  apiPath,
	}
	return &output
}
