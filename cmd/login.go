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
	"github.com/Keyfactor/keyfactor-go-client/api"
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
	Hostname string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
	APIPath  string `json:"api_path"`
}

// loginCmd represents the login command
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
					authConfig, authEnvErr = authConfigFile(configFile, profile, noPrompt, true) // always save config file is login is called
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
			authConfig, authConfigFileErrs = authConfigFile(configFile, profile, noPrompt, true) // always save config file is login is called
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

func authConfigFile(configPath string, profileName string, noPrompt bool, saveConfig bool) (ConfigurationFile, []error) {
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

func authEnvVars(configPath string, profileName string, saveConfig bool) (ConfigurationFile, []error) {
	hostname, hostSet := os.LookupEnv("KEYFACTOR_HOSTNAME")
	username, userSet := os.LookupEnv("KEYFACTOR_USERNAME")
	password, passSet := os.LookupEnv("KEYFACTOR_PASSWORD")
	domain, domainSet := os.LookupEnv("KEYFACTOR_DOMAIN")
	apiPath, apiPathSet := os.LookupEnv("KEYFACTOR_API_PATH")
	envProfileName, _ := os.LookupEnv("KFUTIL_PROFILE")
	if !apiPathSet {
		apiPath = defaultAPIPath
		apiPathSet = true
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
		_ = append(outputErr, fmt.Errorf("KEYFACTOR_HOSTNAME environment variable not set. Please set the KEYFACTOR_HOSTNAME environment variable"))
	}
	if !userSet {
		_ = append(outputErr, fmt.Errorf("KEYFACTOR_USERNAME environment variable not set. Please set the KEYFACTOR_USERNAME environment variable"))
	}
	if !passSet {
		_ = append(outputErr, fmt.Errorf("KEYFACTOR_PASSWORD environment variable not set. Please set the KEYFACTOR_PASSWORD environment variable"))
	}
	if !domainSet {
		_ = append(outputErr, fmt.Errorf("KEYFACTOR_DOMAIN environment variable not set. Please set the KEYFACTOR_DOMAIN environment variable"))
	}
	if !apiPathSet {
		apiPath = defaultAPIPath
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
