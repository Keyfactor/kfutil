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
	"os/signal"
	"strings"
)

const DefaultConfigFileName = "command_config.json"
const DefaultConfigurationFileName = "kfcmd_config.json"

type ConfigurationFile struct {
	Servers map[string]ConfigurationFileEntry `json:"servers"`
}

type ConfigurationFileEntry struct {
	Hostname string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
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

		authenticated := false
		var (
			authConfigErr error
			authEnvErr    error
		)

		if profile == "" && configFile == "" {
			// Check for environment variables
			log.Println("[DEBUG] Checking for environment variables for kfutil configuration data.")
			authEnvErr = authenticate(false, true, configFile, noPrompt, profile)
			if authEnvErr == nil {
				authenticated = true
			} else {
				log.Println("[WARN] Unable to authenticate with environment variables.", authEnvErr)
				log.Println("[DEBUG] Attempting to authenticate via config 'default' profile.")
				authConfigFileErr = authenticate(true, false, configFile, noPrompt, profile)
				if authConfigFileErr == nil {
					authenticated = true
				} else {
					log.Println("[WARN] Unable to authenticate with config file.", authConfigFileErr)
				}
				//if noPrompt {
				//	fmt.Println("Login failed, environment variables not set. Please review https://github.com/Keyfactor/kfutil#environmental-variables for more information.")
				//	log.Fatal("[FATAL] Unable to authenticate with environment variables.")
				//}
			}
		} else if profile != "" {
			if configFile == "" {
				configFile = DefaultConfigFileName
			}
			log.Println("Checking for config file: ", configFile)
			authConfigFileErr := authenticate(true, false, configFile, noPrompt, profile)
			if authConfigFileErr == nil {
				authenticated = true
			} else {
				log.Println("[WARN] Unable to authenticate with config file.", authConfigFileErr)
			}
		} else {
			log.Println("[INFO] Checking for config file: ", configFile)
			authConfigFileErr := authenticate(true, false, configFile, noPrompt, profile)
			if authConfigFileErr == nil {
				authenticated = true
			} else {
				log.Println("[WARN] Unable to authenticate with config file.", authConfigFileErr)
			}
		}

		if !authenticated {
			fmt.Println("Login failed.")
			log.Fatal("[FATAL] Unable to authenticate")
		}
		fmt.Println(fmt.Sprintf("Login successful!"))
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

func getDomainFromUsername(username string) string {
	if strings.Contains(username, "@") {
		return strings.Split(username, "@")[1]
	} else if strings.Contains(username, "\\") {
		return strings.Split(username, "\\")[0]
	}
	return ""
}

func authenticate(fromConfig bool, fromEnv bool, configFile string, noPrompt bool, profile string) error {

	commandConfig := ConfigurationFile{}

	if fromConfig {
		commandConfig, _ = authConfigFile(configFile, noPrompt, profile)
	} else if fromEnv {
		commandConfig, _ = authEnv(noPrompt, profile)
	}

	if profile == "" {
		profile = "default"
	}
	authConfig := api.AuthConfig{
		Hostname: commandConfig.Servers[profile].Hostname,
		Username: commandConfig.Servers[profile].Username,
		Password: commandConfig.Servers[profile].Password,
		Domain:   commandConfig.Servers[profile].Domain,
	}
	// Since there's no login command in the API, we'll just try to get the list of CAs
	_, kfcErr := api.NewKeyfactorClient(&authConfig)
	if kfcErr != nil {
		log.Println("[ERROR] initializing Keyfactor client: ", kfcErr)
		return fmt.Errorf("unable to initialize Keyfactor client. %s. Please check your configuration and try again", kfcErr)
	}

	return nil
}

func authEnv(noPrompt bool, profile string) (ConfigurationFile, error) {
	configuration, err := loadConfigurationFromEnv(noPrompt, profile)
	if err != nil {
		log.Println("[ERROR] Error loading configuration from environment: ", err)
		return ConfigurationFile{}, err
	}
	if len(configuration.Servers) == 0 {
		log.Println("[ERROR] No servers found in configuration")
		return ConfigurationFile{}, fmt.Errorf("no servers found in configuration")
	}
	return configuration, nil
}

func loadConfigurationFromEnv(noPrompt bool, profile string) (ConfigurationFile, error) {
	// Get the Keyfactor Command URL
	envHostName, hostSet := os.LookupEnv("KEYFACTOR_HOSTNAME")
	envUserName, userSet := os.LookupEnv("KEYFACTOR_USERNAME")
	envPassword, passSet := os.LookupEnv("KEYFACTOR_PASSWORD")
	envDomain, domainSet := os.LookupEnv("KEYFACTOR_DOMAIN")
	if !hostSet || !userSet || !passSet {
		return ConfigurationFile{}, fmt.Errorf("missing environment variables") //TODO: Add more details
	}
	if !domainSet {
		envDomain = getDomainFromUsername(envUserName)
		if envDomain != "" {
			domainSet = true
		}
	}

	if !validConfig(envHostName, envUserName, envPassword, envDomain) {
		return ConfigurationFile{}, fmt.Errorf("invalid configuration") //TODO: Add more details
	}

	output := ConfigurationFile{
		Servers: map[string]ConfigurationFileEntry{
			"default": {
				Hostname: envHostName,
				Username: envUserName,
				Password: envPassword,
				Domain:   envDomain,
			},
		},
	}
	return output, nil
}

func authConfigFile(configFile string, noPrompt bool, profile string) (ConfigurationFile, error) {
	var configurationFile ConfigurationFile

	envHostName, hostSet := os.LookupEnv("KEYFACTOR_HOSTNAME")
	envUserName, userSet := os.LookupEnv("KEYFACTOR_USERNAME")
	envPassword, passSet := os.LookupEnv("KEYFACTOR_PASSWORD")
	envDomain, domainSet := os.LookupEnv("KEYFACTOR_DOMAIN")

	log.Println("[DEBUG] Using profile: ", profile)
	userHomeDir, hErr := os.UserHomeDir()
	if configFile == "" {
		// Set up home directory config
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
			}
		}

		configurationFile, _ = loadConfigurationFile(fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName), true)
		configFile = fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName)
	} else {
		// Load config from specified file
		configurationFile, _ = loadConfigurationFile(configFile, false)
		return configurationFile, nil
	}

	if configurationFile.Servers == nil {
		configurationFile.Servers = make(map[string]ConfigurationFileEntry)
	}

	skipEnvConfig := true

	if profile == "" || profile == "default" {
		skipEnvConfig = false
	}

	if skipEnvConfig {
		log.Println("[INFO] Using profile: ", profile)
		// Check if profile exists in config file
		configProfile, profileExists := configurationFile.Servers[profile]
		if !profileExists {
			log.Println("[WARN] Profile not found in config file.")
			fmt.Println("Profile '" + profile + "' not found in config file.")
		} else {
			log.Println("[INFO] Using kfutil config profile: ", profile)
			envHostName = configProfile.Hostname
			envUserName = configProfile.Username
			envPassword = configProfile.Password
			envDomain = configProfile.Domain
			hostSet = true
			userSet = true
			passSet = true
			domainSet = true
		}
	}

	if !hostSet {
		log.Println("[INFO] Hostname not set. Please set the KEYFACTOR_HOSTNAME environment variable.")
	}
	var host string
	if noPrompt {
		if !hostSet && !userSet && !passSet && !domainSet && envHostName == "" && envUserName == "" && envPassword == "" && envDomain == "" {
			fmt.Println("No login information provided. Please set the KEYFACTOR_HOSTNAME, KEYFACTOR_USERNAME, KEYFACTOR_PASSWORD and KEYFACTOR_DOMAIN environment variables.")
			log.Fatal("[FATAL] No login information provided. Please use interactive auth or set the KEYFACTOR_HOSTNAME, KEYFACTOR_USERNAME, KEYFACTOR_PASSWORD and KEYFACTOR_DOMAIN environment variables.")
			return ConfigurationFile{}, fmt.Errorf("no login information provided")
		}
		if !hostSet {
			if len(envHostName) == 0 {

			}
			host = envHostName
			//ehErr := os.Setenv("KEYFACTOR_HOSTNAME", host)
			//if ehErr != nil {
			//	fmt.Println("Error setting hostname: ", ehErr)
			//	log.Fatal("[ERROR] setting hostname: ", ehErr)
			//}
		}

	} else {
		fmt.Printf("Enter Keyfactor Command host URL [%s]: \n", envHostName)
		_, phErr := fmt.Scanln(&host)
		if phErr != nil {
			if phErr.Error() != "unexpected newline" {
				fmt.Println("Error getting hostname: ", phErr)
				log.Println("[ERROR] getting hostname: ", phErr)
			}
		}
		if len(host) == 0 {
			host = envHostName
		}
	}

	// Get the username

	if !userSet {
		log.Println("[INFO] Username not set. Please set the KEYFACTOR_USERNAME environment variable.")
	}
	var username string
	if noPrompt {
		if len(envUserName) == 0 {
			envUserName = configurationFile.Servers[profile].Username
		}
		username = envUserName
	} else {
		fmt.Printf("Enter your Keyfactor Command username [%s]: \n", envUserName)
		_, puErr := fmt.Scanln(&username)
		if puErr != nil {
			if puErr.Error() != "unexpected newline" {
				fmt.Println("Error getting username: ", puErr)
				log.Println("[ERROR] getting username: ", puErr)
			}
		}
	}
	if len(username) == 0 {
		if len(envUserName) == 0 {
			envUserName = configurationFile.Servers[profile].Username
		}
		username = envUserName
	}
	euErr := os.Setenv("KEYFACTOR_USERNAME", username)
	if euErr != nil {
		fmt.Println("Error setting username: ", euErr)
		log.Fatal("[ERROR] setting username: ", euErr)
	}

	// Get the password.

	if !passSet {
		log.Println("[INFO] Password not set. Please set the KEYFACTOR_PASSWORD environment variable.")
	}
	var p string
	if noPrompt {
		if len(envPassword) == 0 {
			envPassword = configurationFile.Servers[profile].Password
		}
		p = envPassword
	} else {
		if len(envPassword) > 0 {
			p = getPassword("password: [<from env KEYFACTOR_PASSWORD>]")
		} else {
			p = getPassword("password: ")
		}

		if len(p) == 0 {
			p = envPassword
		}
	}
	epErr := os.Setenv("KEYFACTOR_PASSWORD", p)
	if epErr != nil {
		fmt.Println("Error setting password: ", epErr)
		log.Fatal("[ERROR] setting password: ", epErr)
	}

	// Get AD domain if not provided in the username or config file

	var domain string
	var userDomain string
	var configDomain string
	if !domainSet {
		if strings.Contains(username, "@") {
			userDomain = strings.Split(username, "@")[1]
		} else if strings.Contains(username, "\\") {
			userDomain = strings.Split(username, "\\")[0]
		} else {
			configDomain = configurationFile.Servers[profile].Domain
		}
	}
	if noPrompt {
		//fmt.Println("Using domain: ", envDomain)
		if len(envDomain) == 0 && len(userDomain) == 0 && len(configDomain) == 0 {

			fmt.Println("Domain not set and unable to be inferred. Please set the KEYFACTOR_DOMAIN environment variable.")
			log.Fatal("[FATAL] Domain not set. Please set the KEYFACTOR_DOMAIN environment variable.")
			return ConfigurationFile{}, fmt.Errorf("domain not set and unable to be inferred")
		}
		if len(configDomain) == 0 {
			if len(userDomain) > 0 {
				log.Println("[INFO] Domain not set. Using domain from username ", userDomain)
				domain = userDomain
			} else if len(envDomain) > 0 {
				log.Println("[INFO] Domain not set. Using domain from environment variable KEYFACTOR_DOMAIN", envDomain)
				domain = envDomain
			}
		} else {
			log.Println("[WARN] KEYFACTOR_DOMAIN environment variable not set. Using domain from config file ", configDomain)
			domain = configDomain
		}
	} else {
		fmt.Printf("Enter your Keyfactor Command AD domain [%s]: \n", envDomain)
		_, sdErr := fmt.Scanln(&domain)
		if sdErr != nil {
			if sdErr.Error() != "unexpected newline" {
				fmt.Println("Error getting domain: ", sdErr)
				log.Println("[ERROR] getting domain: ", sdErr)
			}

		}
		if len(domain) == 0 {
			domain = envDomain
		}
	}
	edErr := os.Setenv("KEYFACTOR_DOMAIN", domain)
	if edErr != nil {
		fmt.Println("Error setting domain: ", edErr)
		log.Fatal("[ERROR] setting domain: ", edErr)
	}

	authConfig := api.AuthConfig{
		Hostname: host,
		Username: username,
		Password: p,
		Domain:   domain,
	}
	// Since there's no login command in the API, we'll just try to get the list of CAs
	_, kfcErr := api.NewKeyfactorClient(&authConfig)
	if kfcErr != nil {
		log.Println("[ERROR] initializing Keyfactor client: ", kfcErr)
		return ConfigurationFile{}, fmt.Errorf("unable to initialize Keyfactor client. %s. Please check your configuration and try again", kfcErr)
	}

	configuration := ConfigurationFileEntry{
		Hostname: host,
		Username: username,
		Password: p,
		Domain:   domain,
	}

	//confErr := createConfigFile(config, configFile)
	configurationErr := createOrUpdateConfigurationFile(configuration, profile, configFile)
	if configurationErr != nil {
		log.Fatal("[FATAL] Login failed due to an issue with the configuration file: ", configurationErr)
	}
	//if confErr != nil {
	//	log.Fatal("[FATAL] Login failed due ot an issue with the config file: ", confErr)
	//}

	return configurationFile, nil
}

func createOrUpdateConfigurationFile(cfgFile ConfigurationFileEntry, configName string, configPath string) error {
	log.Println("[INFO] Creating or updating configuration file")
	log.Println("[DEBUG] configuration file path: ", configPath)

	if len(configName) == 0 {
		configName = "default"
		log.Println("[WARN] no config name provided. Using '" + configName + "'.")
	}

	existingConfig, exsErr := loadConfigurationFile(configPath, true)
	if exsErr != nil {
		log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", configName))
		existingConfig.Servers = make(map[string]ConfigurationFileEntry)
		existingConfig.Servers[configName] = cfgFile
	} else if len(existingConfig.Servers) > 0 {
		// check if the config name already exists
		if _, ok := existingConfig.Servers[configName]; ok {
			log.Println(fmt.Sprintf("[WARN] config name '%s' already exists. Overwriting existing config.", configName))
			log.Println(fmt.Sprintf("[DEBUG] existing config: %v", existingConfig.Servers[configName]))
			log.Println(fmt.Sprintf("[DEBUG] new config: %v", cfgFile))
			// print out the diff between the two configs
			diff := cmp.Diff(existingConfig.Servers[configName], cfgFile)
			if len(diff) == 0 {
				log.Println("[DEBUG] no configuration changes detected")
				return nil
			}
			log.Println(fmt.Sprintf("[DEBUG] diff: %s", diff))
		} else {
			log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", configName))
		}
		existingConfig.Servers[configName] = cfgFile
	} else {
		log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", configName))
		existingConfig.Servers = make(map[string]ConfigurationFileEntry)
		existingConfig.Servers[configName] = cfgFile
	}

	log.Println("[DEBUG] kfcfg entry: ", cfgFile)

	f, fErr := os.OpenFile(fmt.Sprintf("%s", configPath), os.O_CREATE|os.O_RDWR, 0600)
	defer f.Close()
	if fErr != nil {
		msg := fmt.Sprintf("Error creating command configuration file %s: %s", configPath, fErr)
		fmt.Println(msg)
		log.Println("[ERROR] creating command configuration file: ", fErr)
		return fErr
	}
	encoder := json.NewEncoder(f)
	enErr := encoder.Encode(&existingConfig)
	if enErr != nil {
		fmt.Println("Unable to read kfcfg file due to invalid format. ", enErr)
		log.Println("[ERROR] encoding kfcfg file: ", enErr)
		return enErr
	}
	return nil
}

func getPassword(prompt string) string {
	// Get the initial state of the terminal.
	initialTermState, e1 := terminal.GetState(int(os.Stdin.Fd()))
	if e1 != nil {
		panic(e1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = terminal.Restore(int(os.Stdin.Fd()), initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Print(prompt)
	p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println("")
	if err != nil {
		panic(err)
	}

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	// Return the password as a string.
	return string(p)
}

func loadConfigFile(path string, silent bool) (map[string]string, error) {
	data := make(map[string]string) // todo: make this a struct and support multiple configs in a single file

	f, rFErr := os.ReadFile(path)
	if rFErr != nil {
		if !silent {
			fmt.Println(fmt.Sprintf("Unable to read config file '%s'.", rFErr))
			log.Fatal("[FATAL] Error reading config file: ", rFErr)
		}
		return nil, rFErr
	}

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		log.Println("[ERROR] decoding config file: ", jErr)
	}

	return data, nil
}

func loadConfigurationFile(path string, silent bool) (ConfigurationFile, error) {

	//data := ConfigurationFile{Servers: make(map[string]ConfigurationFileEntry)}
	data := ConfigurationFile{}
	f, rFErr := os.ReadFile(path)
	if rFErr != nil {
		if !silent {
			fmt.Println(fmt.Sprintf("Unable to read config file '%s'.", rFErr))
			log.Fatal("[FATAL] Error reading config file: ", rFErr)
		}
		return data, rFErr
	}

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		log.Println("[ERROR] decoding config file: ", jErr)
	}

	return data, nil
}
