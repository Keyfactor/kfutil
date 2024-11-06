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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path"
	"strings"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:        "login",
	Aliases:    nil,
	SuggestFor: nil,
	Short:      "User interactive login to Keyfactor. Stores the credentials in the config file '$HOME/.keyfactor/command_config.json'.",
	GroupID:    "",
	Long: `Will prompt the user for a Username and Password and then attempt to login to Keyfactor.
You can provide the --config flag to specify a config file to use. If not provided, the default
config file will be used. The default config file is located at $HOME/.keyfactor/command_config.json.
To prevent the prompt for Username and Password, use the --no-prompt flag. If this flag is provided then
the CLI will default to using the environment variables. 

For more information on the environment variables review the docs: https://github.com/Keyfactor/kfutil/tree/main?tab=readme-ov-file#environmental-variables 

WARNING: This will write the environmental credentials to disk and will be stored in the config file in plain text at: 
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
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg("Running login command")

		cmd.SilenceUsage = true
		// expEnabled checks
		isExperimental := false
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}
		stdlog.SetOutput(io.Discard)
		informDebug(debugFlag)
		logGlobals()

		var authType string
		var (
			isValidConfig bool
			kfcOAuth      *auth_providers.CommandConfigOauth
			kfcBasicAuth  *auth_providers.CommandAuthConfigBasic
		)

		log.Debug().Msg("calling getEnvConfig()")
		envConfig, envErr := getServerConfigFromEnv()
		if envErr == nil {
			log.Debug().Msg("getEnvConfig() returned")
			log.Info().
				Str("host", envConfig.Host).
				Str("authType", envConfig.AuthType).
				Msg("Login successful via environment variables")
			outputResult(fmt.Sprintf("Login successful via environment variables to %s", envConfig.Host), outputFormat)
			if profile == "" {
				profile = "default"
			}
			if configFile == "" {
				userHomeDir, hErr := prepHomeDir()
				if hErr != nil {
					log.Error().Err(hErr)
					return hErr
				}
				configFile = path.Join(userHomeDir, DefaultConfigFileName)
			}
			envConfigFile := auth_providers.Config{
				Servers: map[string]auth_providers.Server{},
			}
			envConfigFile.Servers[profile] = *envConfig
			wcErr := writeConfigFile(&envConfigFile, configFile)
			if wcErr != nil {
				return wcErr
			}
			return nil
		}

		log.Error().Err(envErr).Msg("Unable to authenticate via environment variables")

		if profile == "" {
			profile = "default"
		}
		if configFile == "" {
			userHomeDir, hErr := prepHomeDir()
			if hErr != nil {
				log.Error().Err(hErr)
				return hErr
			}
			configFile = path.Join(userHomeDir, DefaultConfigFileName)
		}

		log.Debug().
			Str("configFile", configFile).
			Str("profile", profile).
			Msg("call: auth_providers.ReadConfigFromJSON()")
		aConfig, aErr := auth_providers.ReadConfigFromJSON(configFile)
		if aErr != nil {
			log.Error().Err(aErr)
			//return aErr
		}
		log.Debug().Msg("auth_providers.ReadConfigFromJSON() returned")

		var outputServer *auth_providers.Server

		// Attempt to read existing configuration file
		if aConfig != nil {
			serverConfig, serverExists := aConfig.Servers[profile]
			if serverExists {
				// validate the config and prompt for missing values
				authType = serverConfig.GetAuthType()
				switch authType {
				case "oauth":
					oauthConfig, oErr := serverConfig.GetOAuthClientConfig()
					if oErr != nil {
						log.Error().Err(oErr)
					}
					if oauthConfig == nil {
						log.Error().Msg("OAuth configuration is empty")
						break
					}
					vErr := oauthConfig.ValidateAuthConfig()
					if vErr == nil {
						isValidConfig = true
					} else {
						log.Error().
							Err(vErr).
							Msg("invalid OAuth configuration")
						//break
					}
					outputServer = oauthConfig.GetServerConfig()
					kfcOAuth = oauthConfig
				case "basic":
					basicConfig, bErr := serverConfig.GetBasicAuthClientConfig()
					if bErr != nil {
						log.Error().Err(bErr)
					}
					if basicConfig == nil {
						log.Error().Msg("Basic Auth configuration is empty")
						break
					}
					vErr := basicConfig.ValidateAuthConfig()
					if vErr == nil {
						isValidConfig = true
					} else {
						log.Error().
							Err(vErr).
							Msg("invalid Basic Auth configuration")
						//break
					}
					outputServer = basicConfig.GetServerConfig()
					kfcBasicAuth = basicConfig
				default:
					log.Error().
						Str("authType", authType).
						Str("profile", profile).
						Str("configFile", configFile).
						Msg("unable to determine auth type from configuration")
				}
			}
		}

		if !noPrompt {
			log.Debug().Msg("prompting for interactive login")
			iConfig, iErr := authInteractive(outputServer, profile, !noPrompt, true, configFile)
			if iErr != nil {
				log.Error().Err(iErr)
				return iErr
			}
			iServer, iServerExists := iConfig.Servers[profile]
			if iServerExists {
				authType = iServer.GetAuthType()
				switch authType {
				case "oauth":
					kfcOAuth, _ = iServer.GetOAuthClientConfig()
					outputServer = kfcOAuth.GetServerConfig()
					oErr := kfcOAuth.ValidateAuthConfig()
					if oErr == nil {
						isValidConfig = true
					} else {
						log.Error().Err(oErr)
					}
				case "basic":
					kfcBasicAuth, _ = iServer.GetBasicAuthClientConfig()
					outputServer = kfcBasicAuth.GetServerConfig()
					bErr := kfcBasicAuth.ValidateAuthConfig()
					if bErr == nil {
						isValidConfig = true
					} else {
						log.Error().Err(bErr)
					}
				default:
					log.Error().Msg("unable to determine auth type from interactive configuration")
				}
			}
		}

		if !isValidConfig {
			log.Debug().Msg("prompting for interactive login")
			return fmt.Errorf("unable to determine valid configuration")
		}

		if authType == "oauth" {
			log.Debug().Msg("attempting to authenticate via OAuth")
			aErr := kfcOAuth.Authenticate()
			if aErr != nil {
				log.Error().Err(aErr)
				return aErr
			}
		} else if authType == "basic" {
			log.Debug().Msg("attempting to authenticate via Basic Auth")
			aErr := kfcBasicAuth.Authenticate()
			if aErr != nil {
				log.Error().Err(aErr)
				//outputError(aErr, true, outputFormat)
				return aErr
			}
		}

		log.Info().
			Str("profile", profile).
			Str("configFile", configFile).
			Str("host", outputServer.Host).
			Str("authType", authType).
			Msg("Login successful")
		outputResult(fmt.Sprintf("Login successful to %s", outputServer.Host), outputFormat)
		return nil
	},
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	CompletionOptions:          cobra.CompletionOptions{},
	TraverseChildren:           false,
	Hidden:                     false,
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

func writeConfigFile(configFile *auth_providers.Config, configPath string) error {
	existingConfig, exErr := auth_providers.ReadConfigFromJSON(configPath)
	if exErr != nil {
		log.Error().Err(exErr)
		wErr := auth_providers.WriteConfigToJSON(configPath, configFile)
		if wErr != nil {
			log.Error().Err(wErr)
			return wErr
		}
		log.Info().Str("configPath", configPath).Msg("Configuration file written")
		return nil
	}

	// Compare the existing config with the new config
	if cmp.Equal(existingConfig, configFile) {
		log.Info().Msg("Configuration file unchanged")
		return nil
	}

	// Merge the existing config with the new config
	mergedConfig, mErr := auth_providers.MergeConfigFromFile(configPath, configFile)
	if mErr != nil {
		log.Error().Err(mErr)
		return mErr
	}
	wErr := auth_providers.WriteConfigToJSON(configPath, mergedConfig)
	if wErr != nil {
		log.Error().Err(wErr)
		return wErr
	}
	log.Info().Str("configPath", configPath).Msg("Configuration file updated")
	return nil

}

func getDomainFromUsername(username string) string {
	if strings.Contains(username, "@") {
		return strings.Split(username, "@")[1]
	} else if strings.Contains(username, "\\") {
		return strings.Split(username, "\\")[0]
	}
	return ""
}

func createConfigFile(
	hostname string,
	username string,
	password string,
	domain string,
	apiPath string,
	profileName string,
) ConfigurationFile {
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

func promptForInteractiveParameter(parameterName string, defaultValue string) string {
	var input string
	fmt.Printf("Enter %s [%s]: \n", parameterName, defaultValue)
	_, err := fmt.Scanln(&input)
	_ = handleInteractiveError(err, parameterName)
	if input == "" {
		log.Debug().
			Str("parameterName", parameterName).
			Str("defaultValue", defaultValue).
			Msg("using default value")
		return defaultValue
	}
	log.Debug().
		Str("parameterName", parameterName).
		Str("input", input).
		Msg("using input value")
	return input
}

func promptForInteractivePassword(parameterName string, defaultValue string) string {
	passwordFill := ""
	if defaultValue != "" {
		passwordFill = "********"
	}

	fmt.Printf("Enter %s [%s]: ", parameterName, passwordFill)

	var password string

	// Check if we're in a terminal environment
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Terminal mode: read password securely
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println("") // for newline after password input
		if err != nil {
			fmt.Println("\nError reading password:", err)
			return defaultValue
		}
		password = string(bytePassword)
	} else {
		// Non-terminal mode: read password as plain text
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nError reading password:", err)
			return defaultValue
		}
		password = input
	}

	// Trim newline and check if password is empty; if so, return default
	password = password[:len(password)-1]
	if password == "" {
		return defaultValue
	}
	return password
}

func handleInteractiveError(err error, parameterName string) error {
	if err != nil {
		if err.Error() != "unexpected newline" {
			log.Error().Err(err)
			outputError(fmt.Errorf("error handling parameter '%s'=%v", parameterName, err), false, "")
			//log.Println(fmt.Sprintf("[ERROR] %s", errMsg))
			return err
		}
	}
	return nil
}

func saveConfigFile(configFile ConfigurationFile, configPath string, profileName string) (ConfigurationFile, error) {
	if profileName == "" {
		profileName = "default"
	}

	configurationErr := createOrUpdateConfigurationFile(configFile.Servers[profileName], profileName, configPath)
	if configurationErr != nil {
		//log.Fatal("[ERROR] Unable to save configuration file to disk: ", configurationErr)
		log.Error().Err(configurationErr)
		return configFile, configurationErr
	}
	loadedConfig, loadErr := loadConfigurationFile(configPath, true)
	if loadErr != nil {
		//log.Fatal("[ERROR] Unable to load configuration file after save: ", loadErr)
		log.Error().Err(loadErr)
		return configFile, loadErr
	}
	return loadedConfig, nil
}

func authInteractive(
	serverConf *auth_providers.Server,
	profileName string,
	forcePrompt bool,
	saveConfig bool,
	configPath string,
) (auth_providers.Config, error) {
	if serverConf == nil {
		serverConf = &auth_providers.Server{}
	}

	if serverConf.Host == "" || forcePrompt {
		serverConf.Host = promptForInteractiveParameter("Keyfactor Command HostName", serverConf.Host)
	}
	if serverConf.AuthType == "" || forcePrompt {
		for {
			serverConf.AuthType = promptForInteractiveParameter(
				"Keyfactor Command AuthType [basic,oauth]",
				serverConf.AuthType,
			)
			if serverConf.AuthType == "oauth" || serverConf.AuthType == "basic" {
				break
			} else {
				fmt.Println("Invalid auth type. Valid auth types are: oauth, basic")
			}
		}
	}
	if serverConf.AuthType == "basic" {
		if serverConf.Username == "" || forcePrompt {
			serverConf.Username = promptForInteractiveParameter("Keyfactor Command Username", serverConf.Username)
		}
		if serverConf.Password == "" || forcePrompt {
			serverConf.Password = promptForInteractivePassword("Keyfactor Command Password", serverConf.Password)
		}
		if serverConf.Domain == "" || forcePrompt {
			userDomain := getDomainFromUsername(serverConf.Username)
			if userDomain == "" {
				serverConf.Domain = promptForInteractiveParameter("Keyfactor Command AD Domain", serverConf.Domain)
			} else {
				serverConf.Domain = userDomain
			}
		}
	} else if serverConf.AuthType == "oauth" {
		if serverConf.ClientID == "" || forcePrompt {
			serverConf.ClientID = promptForInteractiveParameter(
				"Keyfactor Command OAuth Client ID",
				serverConf.ClientID,
			)
		}
		if serverConf.ClientSecret == "" || forcePrompt {
			serverConf.ClientSecret = promptForInteractivePassword(
				"Keyfactor Command OAuth Client Secret",
				serverConf.ClientSecret,
			)
		}
		if serverConf.OAuthTokenUrl == "" || forcePrompt {
			serverConf.OAuthTokenUrl = promptForInteractiveParameter(
				"Keyfactor Command OAuth Token URL",
				serverConf.OAuthTokenUrl,
			)
		}
	}

	if serverConf.APIPath == "" || forcePrompt {
		serverConf.APIPath = promptForInteractiveParameter("Keyfactor Command API path", serverConf.APIPath)
	}

	if serverConf.CACertPath == "" || forcePrompt {
		serverConf.CACertPath = promptForInteractiveParameter("Keyfactor Command CA Cert Path", serverConf.CACertPath)
	}

	if profileName == "" {
		profileName = "default"
	}
	if configPath == "" {
		userHomeDir, hErr := prepHomeDir()
		if hErr != nil {
			//log.Println("[ERROR] Unable to create home directory: ", hErr)
			log.Error().Err(hErr)
			return auth_providers.Config{}, hErr
		}
		configPath = path.Join(userHomeDir, DefaultConfigFileName)
	}

	confFile := auth_providers.Config{
		Servers: map[string]auth_providers.Server{},
	}
	confFile.Servers[profileName] = *serverConf

	if saveConfig {
		saveErr := writeConfigFile(&confFile, configPath)
		if saveErr != nil {
			//log.Println("[ERROR] Unable to save configuration file to disk: ", saveErr)
			log.Error().Err(saveErr)
			return confFile, saveErr
		}
	}
	return confFile, nil
}

func prepHomeDir() (string, error) {
	log.Debug().Msg("prepHomeDir() called")
	// Set up home directory config
	userHomeDir, hErr := os.UserHomeDir()

	if hErr != nil {
		//log.Println("[ERROR] Error getting user home directory: ", hErr)
		log.Error().Err(hErr)
		//fmt.Println("Error getting user home directory: ", hErr)
		outputError(fmt.Errorf("error getting user home directory: %v", hErr), false, outputFormat)
		fmt.Println("Using current directory to write config file '" + DefaultConfigFileName + "'")
		userHomeDir = "." // Default to current directory
	} else {
		userHomeDir = fmt.Sprintf("%s/.keyfactor", userHomeDir)
	}
	//log.Println("[DEBUG] Configuration directory: ", userHomeDir)
	log.Debug().Str("userHomeDir", userHomeDir).Msg("Configuration directory")
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

func loadConfigFileData(
	profileName string,
	configPath string,
	noPrompt bool,
	configurationFile ConfigurationFile,
) (string, string, string, string, string) {
	log.Debug().Str("profileName", profileName).
		Str("configPath", configPath).
		Bool("noPrompt", noPrompt).
		Msg("loadConfigFileData() called")

	log.Debug().Msg("calling authEnvVars()")
	envConfig, _ := authEnvVars(configPath, profileName, false) //Load env vars first
	log.Debug().Msg("authEnvVars() returned")

	// Check if profileName exists in config file
	log.Debug().Str("profileName", profileName).Msg("checking if profileName exists in config file")
	configProfile, profileExists := configurationFile.Servers[profileName]
	if !profileExists {
		log.Error().
			Str("profileName", profileName).
			Msg("profileName does not exist in config file")
		if noPrompt { // TODO: If profile doesn't exist how does this work?
			log.Error().Msg("noPrompt is set, unable to prompt for missing profileName")
			return envConfig.Servers[profileName].Hostname, envConfig.Servers[profileName].Username, envConfig.Servers[profileName].Password, envConfig.Servers[profileName].Domain, envConfig.Servers[profileName].APIPath
		}
	} else {
		//log.Println("[INFO] Using kfutil config profileName: ", profileName)
		log.Info().Str("profileName", profileName).Msg("Using kfutil config profileName")
		hostName := configProfile.Hostname
		userName := configProfile.Username
		password := configProfile.Password
		domain := configProfile.Domain
		apiPath := configProfile.APIPath
		authProvider := configProfile.AuthProvider
		log.Debug().Str("hostName", hostName).
			Str("userName", userName).
			Str("password", hashSecretValue(password)).
			Str("domain", domain).
			Str("apiPath", apiPath).
			Str("authProvider", fmt.Sprintf("%+v", authProvider)).
			Msg("configProfile values")

		if authProvider.Type != "" && authProvider.Parameters != nil {
			// cALL authViaProviderParams
			log.Info().Str("authProvider.Type", authProvider.Type).
				Msg("Authenticating via authProvider")
			log.Debug().Msg("calling authViaProviderParams()")
			authConfig, authErr := authViaProviderParams(&authProvider)
			if authErr != nil {
				//log.Println("[ERROR] Unable to authenticate via provider: ", authErr)
				log.Error().Err(authErr).Msg("Unable to authenticate via provider")
				return "", "", "", "", ""
			}

			// Check if authProvider profile is set
			if authProvider.Profile == "" {
				authProvider.Profile = "default"
				log.Info().Msg("Using default authProvider profile")
			}

			hostName = authConfig.Servers[authProvider.Profile].Hostname
			userName = authConfig.Servers[authProvider.Profile].Username
			password = authConfig.Servers[authProvider.Profile].Password
			domain = authConfig.Servers[authProvider.Profile].Domain
			apiPath = authConfig.Servers[authProvider.Profile].APIPath

			log.Debug().Str("hostName", hostName).
				Str("userName", userName).
				Str("password", hashSecretValue(password)).
				Str("domain", domain).
				Str("apiPath", apiPath).
				Msg("authConfig values")

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

		log.Debug().Str("hostName", hostName).
			Str("userName", userName).
			Str("password", hashSecretValue(password)).
			Str("domain", domain).
			Str("apiPath", apiPath).
			Msg("configProfile values")

		return hostName, userName, password, domain, apiPath
	}
	log.Error().Msg("Unable to load config file data")
	return "", "", "", "", ""
}

func authViaProvider() (*api.Client, error) {
	var clientAuth api.AuthConfig
	var commandConfig auth_providers.Config
	if providerType != "" {
		log.Info().Str("providerType", providerType).Msg("attempting to auth via auth provider")
		var providerConfig AuthProvider
		if providerProfile == "" {
			log.Info().Str(
				"providerProfile",
				providerProfile,
			).Msg("auth provider profile not set, defaulting to 'default'")
			providerProfile = "default"
		}

		providerConfig = AuthProvider{
			Type:       providerType,
			Profile:    providerProfile,
			Parameters: nil,
		}

		if configFile == "" {
			homeDir, hdErr := os.UserHomeDir()
			if hdErr != nil {
				homeDir, hdErr = os.Getwd()
				if hdErr != nil {
					homeDir = "." // Default to current directory
				}
			}
			configFile = path.Join(homeDir, ".keyfactor", DefaultConfigFileName)
		}

		// Load config file
		log.Debug().Str("configFile", configFile).Msg("configFile is set, loading config file")
		log.Debug().Msg("calling loadConfigurationFile()")
		configurationFile, cErr := loadConfigurationFile(configFile, true)
		log.Debug().Msg("loadConfigurationFile() returned")
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to load provider config file")
			return nil, cErr
		}
		// look for profile in config file
		log.Debug().Str("profile", profile).
			Str("providerProfile", providerProfile).
			Msg("checking if providerProfile exists in config file")

		providerConfigEntry, providerProfileExists := configurationFile.Servers[providerProfile]
		if !providerProfileExists {
			log.Error().Str("providerProfile", providerProfile).Msg("providerProfile does not exist in config file")
			return nil, fmt.Errorf("providerProfile '%s' does not exist in config file", providerProfile)
		}
		params := providerConfigEntry.AuthProvider.Parameters
		if params == nil {
			log.Error().Msg("providerProfile parameters are empty")
			return nil, fmt.Errorf("providerProfile '%s' parameters are empty", providerProfile)
		}
		providerConfig.Parameters = params

		log.Debug().Str("providerConfig.Type", providerConfig.Type).
			Msg("call: authViaProviderParams()")
		pvConfig, pErr := authViaProviderParams(&providerConfig)
		log.Debug().Msg("returned: authViaProviderParams()")
		if pErr != nil {
			log.Error().Err(pErr).
				Str("providerConfig.Type", providerConfig.Type).
				Str("providerConfig.Profile", providerConfig.Profile).
				Msg("unable to auth via provider")
			return nil, pErr
		}
		log.Trace().Interface("pvConfig", pvConfig).Send()

		//commandConfig = pvConfig //TODO: Handle this ab#55467
		clientConfig := clientAuth.GetServerConfig()
		clientConfig.Username = commandConfig.Servers[providerProfile].Username
		clientConfig.Password = commandConfig.Servers[providerProfile].Password
		clientConfig.Domain = commandConfig.Servers[providerProfile].Domain
		clientConfig.Host = commandConfig.Servers[providerProfile].Host
		clientConfig.ClientID = commandConfig.Servers[providerProfile].ClientID
		clientConfig.ClientSecret = commandConfig.Servers[providerProfile].ClientSecret
		clientConfig.OAuthTokenUrl = commandConfig.Servers[providerProfile].OAuthTokenUrl
		clientConfig.APIPath = commandConfig.Servers[providerProfile].APIPath

		log.Debug().
			Str("clientAuth.Username", clientConfig.Username).
			Str("clientAuth.Password", hashSecretValue(clientConfig.Password)).
			Str("clientAuth.Domain", clientConfig.Domain).
			Str("clientAuth.ClientID", clientConfig.ClientID).
			Str("clientAuth.ClientSecret", hashSecretValue(clientConfig.ClientSecret)).
			Str("clientAuth.Hostname", clientConfig.Host).
			Str("clientAuth.APIPath", clientConfig.APIPath).
			Msg("Client authentication params")

		log.Debug().Msg("call: api.NewKeyfactorClient()")
		c, err := api.NewKeyfactorClient(clientConfig, nil)
		log.Debug().Msg("complete: api.NewKeyfactorClient()")

		if err != nil {
			//fmt.Printf("Error connecting to Keyfactor: %s\n", err)
			outputError(err, true, "text")
			//log.Fatalf("[ERROR] creating Keyfactor client: %s", err)
			return nil, fmt.Errorf("unable to create Keyfactor Command client: %s", err)
		}
		log.Info().Msg("Keyfactor Command client created")
		log.Debug().Str("flagAuthProvider", providerType).
			Str("providerProfile", providerProfile).
			Msg("returning from provider auth")
		return c, nil
	}
	return nil, fmt.Errorf("unable to auth via provider, providerType is empty")
}

func authViaProviderParams(providerConfig *AuthProvider) (ConfigurationFile, error) {

	pt := providerConfig.Type
	// First check if provider type and provider config are not empty
	if pt == "" || providerConfig == nil {
		return ConfigurationFile{}, fmt.Errorf("provider type and provider config cannot be empty")
	}

	// Check if auth provider is valid
	if !validAuthProvider(pt) {
		return ConfigurationFile{}, fmt.Errorf(
			"invalid auth provider type '%s'. Valid auth providers are: %v",
			pt,
			ValidAuthProviders,
		)
	}

	// Check if provider type matches requested provider type
	switch pt {
	case "azure-id", "azid", "az-id", "azureid":
		//load provider config params into AuthProviderAzureIdParams struct
		log.Debug().Msg("authenticating via azure-id provider")
		var providerParams AuthProviderAzureIDParams
		log.Debug().Msg("marshalling providerConfig.Parameters")
		paramsJson, _ := json.Marshal(providerConfig.Parameters)
		log.Debug().Msg("unmarshalling providerParams")
		jsonErr := json.Unmarshal(paramsJson, &providerParams)
		if jsonErr != nil {
			log.Error().Err(jsonErr).Msg("unable to unmarshal providerParams")
			return ConfigurationFile{}, jsonErr
		}

		// Check if required params are set
		if providerParams.SecretName == "" || providerParams.AzureVaultName == "" {
			return ConfigurationFile{}, fmt.Errorf("provider params secret_name and vault_name are required")
		}
		log.Debug().Msgf("providerParams: %+v", providerParams)
		return providerParams.authenticate()
	case "az-cli", "azcli", "azure-cli", "azurecli":
		log.Debug().Msg("authenticating via azure-cli provider")
		break
	default:
		//log.Println("[ERROR] invalid auth provider type")
		log.Error().Msg("invalid auth provider type")
		break
	}
	return ConfigurationFile{}, fmt.Errorf(
		"invalid auth provider type '%s'. Valid auth providers are: %v",
		pt,
		ValidAuthProviders,
	)
}

func validAuthProvider(providerType string) bool {
	log.Debug().Str("providerType", providerType).Msg("validAuthProvider() called")
	if providerType == "" {
		return true // default to Username/Password
	}
	for _, validProvider := range ValidAuthProviders {
		if validProvider == providerType {
			log.Debug().Str("providerType", providerType).Msg("is valid auth provider type")
			return true
		}
	}
	log.Error().Str("providerType", providerType).Msg("is not valid auth provider type")
	return false
}

func authConfigFile(
	configPath string,
	profileName string,
	authProviderProfile string,
	noPrompt bool,
	saveConfig bool,
) (ConfigurationFile, []error) {
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

	//log.Println("[DEBUG] Using profileName: ", profileName)
	log.Debug().Str("profileName", profileName).
		Msg("Using profileName")

	if configPath == "" {
		log.Debug().Msg("configPath is empty, setting to default")
		log.Debug().Msg("calling prepHomeDir()")
		userHomeDir, _ := prepHomeDir()
		log.Debug().Msg("prepHomeDir() returned")

		configPath = fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName)
		log.Debug().Str("configPath", configPath).Msg("configPath set")
	}
	configurationFile, _ = loadConfigurationFile(configPath, noPrompt)

	if configurationFile.Servers == nil {
		configurationFile.Servers = make(map[string]ConfigurationFileEntry)
	}
	if profileName == "" {
		log.Debug().Msg("profileName is empty, setting to default")
		profileName = "default"
	}

	log.Debug().Msg("calling loadConfigFileData()")
	hostName, userName, password, domain, apiPath = loadConfigFileData(
		profileName,
		configPath,
		noPrompt,
		configurationFile,
	)
	log.Debug().Msg("loadConfigFileData() returned")

	log.Debug().Str("hostName", hostName).
		Str("userName", userName).
		Str("password", hashSecretValue(password)).
		Str("domain", domain).
		Str("apiPath", apiPath).
		Msg("loadConfigFileData() values")

	log.Debug().Msg("calling createConfigFile()")
	confFile := createConfigFile(hostName, userName, password, domain, apiPath, profileName)
	log.Debug().Msg("createConfigFile() returned")

	if saveConfig {
		log.Info().Str("configPath", configPath).
			Str("profileName", profileName).
			Msg("Saving configuration file")
		log.Debug().Msg("calling saveConfigFile()")
		savedConfigFile, saveErr := saveConfigFile(confFile, configPath, profileName)
		log.Debug().Msg("saveConfigFile() returned")
		if saveErr != nil {
			log.Error().Err(saveErr)
			return confFile, []error{saveErr}
		}
		log.Info().Str("configPath", configPath).
			Str("profileName", profileName).
			Msg("Configuration file saved")

		log.Debug().Msg("returning savedConfigFile")
		return savedConfigFile, nil
	}

	configurationFile.Servers[profileName] = confFile.Servers[profileName]
	return configurationFile, nil
}

func authEnvProvider(authProvider *AuthProvider, configProfile string) (ConfigurationFile, []error) {
	//log.Println(fmt.Sprintf("[INFO] authenticating with auth provider '%s' params from environment variables", authProvider.Type))
	log.Info().Str(
		"authProvider.Type",
		authProvider.Type,
	).Msg("authenticating with auth provider params from environment variables")

	if configProfile == "" {
		log.Debug().Msg("configProfile is empty, setting to default")
		configProfile = "default"
	}
	// attempt to cast authProvider.Parameters to string
	authProviderParams, ok := authProvider.Parameters.(string)
	if !ok {
		//log.Println("[ERROR] unable to cast authProvider.Parameters to string")
		log.Error().Msg("unable to cast authProvider.Parameters to string")
		return ConfigurationFile{}, []error{fmt.Errorf("invalid configuration, unable to cast authProvider.Parameters to string")}
	}

	if strings.HasPrefix(authProviderParams, "{") && strings.HasSuffix(authProviderParams, "}") {
		// authProviderParams is a json string
		//log.Println("[DEBUG] authProviderParams is a json string")
		log.Debug().Msg("authProviderParams is a json string")
		var providerParams interface{}
		//log.Println("[DEBUG] converting authProviderParams to unescaped json")
		log.Debug().Msg("converting authProviderParams to unescaped json")
		jsonErr := json.Unmarshal([]byte(authProviderParams), &providerParams)
		if jsonErr != nil {
			//log.Println("[ERROR] unable to unmarshal authProviderParams: ", jsonErr)
			log.Error().Err(jsonErr).Msg("unable to unmarshal authProviderParams")
			return ConfigurationFile{}, []error{jsonErr}
		}
		authProvider.Parameters = providerParams
	} else {
		// attempt to read as json file path
		//log.Println("[DEBUG] authProviderParams is a json file path")
		log.Debug().Msg("authProviderParams is a json file path")
		var providerParams interface{}
		var providerConfigFile ConfigurationFile
		var authProviderConfig AuthProvider
		//log.Println("[DEBUG] opening authProviderParams file ", authProviderParams)
		log.Debug().Str("authProviderParams", authProviderParams).Msg("opening authProviderParams file")

		jsonFile, jsonFileErr := os.Open(authProviderParams)
		if jsonFileErr != nil {
			//log.Println("[ERROR] unable to open authProviderParams file: ", jsonFileErr)
			log.Error().Err(jsonFileErr).Msg("unable to open authProviderParams file")
			return ConfigurationFile{}, []error{jsonFileErr}
		}
		defer jsonFile.Close()
		//log.Println(fmt.Sprintf("[DEBUG] reading authProviderParams file %s as bytes", authProviderParams))
		log.Debug().Str("authProviderParams", authProviderParams).Msg("reading authProviderParams file as bytes")
		jsonBytes, jsonBytesErr := os.ReadFile(authProviderParams)
		if jsonBytesErr != nil {
			//log.Println("[ERROR] unable to read authProviderParams file: ", jsonBytesErr)
			log.Error().Err(jsonBytesErr).Msg("unable to read authProviderParams file")
			return ConfigurationFile{}, []error{jsonBytesErr}
		}
		//log.Println("[DEBUG] converting authProviderParams to unescaped json")
		log.Debug().Msg("converting authProviderParams to unescaped json")
		jsonErr := json.Unmarshal(jsonBytes, &providerParams)
		if jsonErr != nil {
			//log.Println("[ERROR] unable to unmarshal authProviderParams: ", jsonErr)
			log.Error().Err(jsonErr).Msg("unable to unmarshal authProviderParams")
			return ConfigurationFile{}, []error{jsonErr}
		}

		//Check if provider params is a configuration file
		//log.Println("[DEBUG] checking if authProviderParams is a configuration file")
		log.Debug().Msg("checking if authProviderParams is a configuration file")
		jsonErr = json.Unmarshal(jsonBytes, &providerConfigFile)
		if jsonErr == nil && providerConfigFile.Servers != nil {
			// lookup params based on configProfile
			//log.Println("[DEBUG] authProviderParams is a configuration file")
			log.Debug().Msg("authProviderParams is a configuration file")
			// check to see if profile exists in config file
			if _, isConfigFile := providerConfigFile.Servers[configProfile]; isConfigFile {
				//log.Println(fmt.Sprintf("[DEBUG] profile '%s' found in authProviderParams file", configProfile))
				log.Debug().Str("configProfile", configProfile).Msg("profile found in authProviderParams file")
				providerParams = providerConfigFile.Servers[configProfile]
				// check if providerParams is a ConfigurationFileEntry
				if _, isConfigFileEntry := providerParams.(ConfigurationFileEntry); !isConfigFileEntry {
					//log.Println("[ERROR] unable to cast providerParams to ConfigurationFileEntry")
					log.Error().Msg("unable to cast providerParams to ConfigurationFileEntry")
					return ConfigurationFile{}, []error{fmt.Errorf("invalid configuration, unable to cast providerParams to ConfigurationFileEntry")}
				}
				// set providerParams to ConfigurationFileEntry.AuthProvider.Parameters
				providerParams = providerConfigFile.Servers[configProfile].AuthProvider.Parameters
			} else {
				//log.Println(fmt.Sprintf("[DEBUG] profile '%s' not found in authProviderParams file", configProfile))
				log.Debug().Str("configProfile", configProfile).Msg("profile not found in authProviderParams file")
				return ConfigurationFile{}, []error{
					fmt.Errorf(
						"profile '%s' not found in authProviderParams file",
						configProfile,
					),
				}
			}
		} else {
			//check if provider params is an AuthProvider
			//log.Println("[DEBUG] checking if authProviderParams is an AuthProvider")
			log.Debug().Msg("checking if authProviderParams is an AuthProvider")

			//log.Println("[DEBUG] converting authProviderParams to unescaped json")
			log.Debug().Msg("converting authProviderParams to unescaped json")

			//check if providerParams is a map[string]interface{}
			if _, isMap := providerParams.(map[string]interface{}); isMap {
				//check if 'auth_provider' key exists and if it does convert to json bytes
				log.Debug().Msg("authProviderParams is a map")
				if _, isAuthProvider := providerParams.(map[string]interface{})["auth_provider"]; isAuthProvider {
					//log.Println("[DEBUG] authProviderParams is a map[string]interface{}")
					//log.Println("[DEBUG] converting authProviderParams to unescaped json")
					log.Debug().Msg("authProviderParams is a map[string]interface{}")
					log.Debug().Msg("converting authProviderParams to unescaped json")
					jsonBytes, jsonBytesErr = json.Marshal(providerParams.(map[string]interface{})["auth_provider"])
					if jsonBytesErr != nil {
						//log.Println("[ERROR] unable to marshal authProviderParams: ", jsonBytesErr)
						log.Error().Err(jsonBytesErr).Msg("unable to marshal authProviderParams")
						return ConfigurationFile{}, []error{jsonBytesErr}
					}
				}
			}

			jsonErr = json.Unmarshal(jsonBytes, &authProviderConfig)
			if jsonErr == nil && authProviderConfig.Type != "" && authProviderConfig.Parameters != nil {
				//log.Println("[DEBUG] authProviderParams is an AuthProvider")
				log.Debug().Msg("authProviderParams is an AuthProvider")
				providerParams = authProviderConfig.Parameters
			}
		}
		authProvider.Parameters = providerParams
	}
	//log.Println("[INFO] Attempting to fetch kfutil creds from auth provider ", authProvider)
	log.Info().Str(
		"authProvider",
		fmt.Sprintf("%+v", authProvider),
	).Msg("Attempting to fetch kfutil creds from auth provider")
	configFile, authErr := authViaProviderParams(authProvider)
	if authErr != nil {
		//log.Println("[ERROR] Unable to authenticate via provider: ", authErr)
		log.Error().Err(authErr).Msg("Unable to authenticate via provider")
		return ConfigurationFile{}, []error{authErr}
	}
	//log.Println("[INFO] Successfully retrieved kfutil creds via auth provider")
	log.Info().Msg("Successfully retrieved kfutil creds via auth provider")
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
		outputErr = append(
			outputErr,
			fmt.Errorf("KEYFACTOR_HOSTNAME environment variable not set. Please set the KEYFACTOR_HOSTNAME environment variable"),
		)
	}
	if !userSet {
		outputErr = append(
			outputErr,
			fmt.Errorf("KEYFACTOR_USERNAME environment variable not set. Please set the KEYFACTOR_USERNAME environment variable"),
		)
	}
	if !passSet {
		outputErr = append(
			outputErr,
			fmt.Errorf("KEYFACTOR_PASSWORD environment variable not set. Please set the KEYFACTOR_PASSWORD environment variable"),
		)
	}
	if !domainSet {
		outputErr = append(
			outputErr,
			fmt.Errorf("KEYFACTOR_DOMAIN environment variable not set. Please set the KEYFACTOR_DOMAIN environment variable"),
		)
	}
	if !apiPathSet {
		apiPath = DefaultAPIPath
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
	//log.Println("[INFO] Creating or updating configuration file")
	log.Info().Str("configPath", configPath).
		Str("profile", profile).
		Msg("Creating or updating configuration file")
	//log.Println("[DEBUG] configuration file path: ", configPath)

	if len(profile) == 0 {
		log.Debug().Msg("profile is empty, setting to default")
		profile = "default"
	}
	//check if configPath exists
	if configPath == "" {
		defaultDir, _ := os.UserHomeDir()
		configPath = path.Join(defaultDir, ".keyfactor", DefaultConfigFileName)
		//log.Println("[WARN] no config path provided. Using '" + configPath + "'.")
		log.Debug().Str("configPath", configPath).Msg("no config path provided using default")
	}
	confFileExists, fileErr := os.Stat(configPath)
	if fileErr != nil {
		//log.Println("[WARN] ", fileErr)
		log.Error().Err(fileErr).Msg("error checking if config file exists")
	}

	existingConfig, _ := loadConfigurationFile(configPath, true)
	if len(existingConfig.Servers) > 0 {
		// check if the config name already exists
		if _, ok := existingConfig.Servers[profile]; ok {
			//log.Println(fmt.Sprintf("[WARN] config name '%s' already exists. Overwriting existing config.", profile))
			log.Info().
				Str("profile", profile).
				Msg("config profile already exists, overwriting existing config")
			//log.Println(fmt.Sprintf("[DEBUG] existing config: %v", existingConfig.Servers[profile]))
			log.Debug().Str("profile", profile).
				Str("existingConfig", fmt.Sprintf("%+v", existingConfig.Servers[profile])).
				Msg("existing config")
			//log.Println(fmt.Sprintf("[DEBUG] new config: %v", cfgFile))
			// print out the diff between the two configs
			diff := cmp.Diff(existingConfig.Servers[profile], cfgFile)
			if len(diff) == 0 && confFileExists != nil {
				//log.Println("[DEBUG] no configuration changes detected")
				log.Debug().Msg("no configuration changes detected")
				return nil
			}
			//log.Println(fmt.Sprintf("[DEBUG] diff: %s", diff))
			log.Debug().Str("diff", diff).Msg("config diff")
		}
		existingConfig.Servers[profile] = cfgFile
	} else {
		//log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", profile))
		log.Info().Str("profile", profile).Msg("adding new profile")
		existingConfig.Servers = make(map[string]ConfigurationFileEntry)
		existingConfig.Servers[profile] = cfgFile
	}

	//log.Println("[DEBUG] kfcfg entry: ", cfgFile)
	log.Debug().Str("cfgFile", fmt.Sprintf("%+v", cfgFile)).Msg("kfcfg entry")

	f, fErr := os.OpenFile(fmt.Sprintf("%s", configPath), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	defer f.Close()
	if fErr != nil {
		msg := fmt.Errorf("unable to create command configuration file %s: %s", configPath, fErr)
		outputError(msg, false, outputFormat)
		log.Error().Err(fErr).Msg("unable to create command configuration file")
		return fErr
	}

	// convert existingConfig to json
	jsonData, jsErr := json.MarshalIndent(existingConfig, "", "  ")
	if jsErr != nil {
		//fmt.Println("Unable to read kfcfg file due to invalid format. ", jsErr)
		outputError(jsErr, false, outputFormat)
		//log.Println("[ERROR] marshalling kfcfg file: ", jsErr)
		log.Error().Err(jsErr).Msg("marshalling command config file")
		return jsErr
	}
	_, enErr := f.Write(jsonData)
	if enErr != nil {
		//fmt.Println("Unable to read kfcfg file due to invalid format. ", enErr)
		outputError(enErr, false, outputFormat)
		//log.Println("[ERROR] encoding kfcfg file: ", enErr)
		log.Error().Err(enErr).Msg("encoding command config file")
		return enErr
	}
	return nil
}

func loadConfigurationFile(filePath string, silent bool) (ConfigurationFile, error) {
	log.Debug().Str("filePath", filePath).Msg("loadConfigurationFile() called")
	//data := ConfigurationFile{Servers: make(map[string]ConfigurationFileEntry)}
	data := ConfigurationFile{}
	if filePath == "" {
		log.Debug().Msg("filePath is empty, setting to default")
		defaultDir, _ := os.UserHomeDir()
		filePath = path.Join(defaultDir, ".keyfactor", DefaultConfigFileName)
	}
	log.Debug().Str("filePath", filePath).Msg("filePath set")

	// attempt to make the directory if it doesn't exist
	dirPath := path.Dir(filePath)
	if _, dirErr := os.Stat(dirPath); os.IsNotExist(dirErr) {
		//log.Println("[DEBUG] config directory does not exist, creating: ", dirPath)
		log.Debug().Str("dirPath", dirPath).Msg("config directory does not exist, creating")
		err := os.MkdirAll(dirPath, 0700)
		if err != nil {
			//log.Println("[ERROR] creating config directory: ", err)
			log.Error().Err(err).Msg("creating config directory")
			return data, err
		}
		return data, nil // return empty data since the directory didn't exist the file won't exist
	}

	// check if file exists
	if _, fileErr := os.Stat(filePath); os.IsNotExist(fileErr) {
		//log.Println("[DEBUG] config file does not exist: ", filePath)
		log.Debug().Str("filePath", filePath).Msg("config file does not exist")
		return data, nil // return empty data since the file doesn't exist
	}

	f, rFErr := os.ReadFile(filePath)
	if rFErr != nil {
		if !silent {
			//fmt.Println(fmt.Sprintf("Unable to read config file '%s'.", rFErr))
			outputError(rFErr, true, outputFormat)
			//log.Fatal("[FATAL] Error reading config file: ", rFErr)
			log.Error().Err(rFErr).Msg("error reading config file")
		}
		return data, rFErr
	}

	// Try to unmarshal as a single entry first
	var singleEntry ConfigurationFileEntry
	sjErr := json.Unmarshal(f, &singleEntry)
	if sjErr != nil {
		//log.Println(fmt.Sprintf("[DEBUG] config file '%s' is a not single entry, will attempt to parse as v1 config file", filePath))
		log.Debug().Str(
			"filePath",
			filePath,
		).Msg("config file is not a single entry, will attempt to parse as v1 config file")
	} else if (singleEntry != ConfigurationFileEntry{}) {
		// if we successfully unmarshalled a single entry, add it to the map as the default entry
		//log.Println(fmt.Sprintf("[DEBUG] config file '%s' is a single entry, adding to map", filePath))
		log.Debug().Str("filePath", filePath).Msg("config file is a single entry, adding to map")
		data.Servers = make(map[string]ConfigurationFileEntry)
		data.Servers["default"] = singleEntry
		return data, nil
	}

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		//log.Println("[ERROR] decoding config file: ", jErr)
		log.Error().Err(jErr).Msg("decoding config file")
		return data, jErr
	}

	return data, nil
}
