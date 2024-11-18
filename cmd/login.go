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
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path"
	"strings"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
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
	if len(password) > 0 {
		password = password[:len(password)-1]
	}
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
		if len(serverConf.Scopes) == 0 || forcePrompt {
			scopesCsv := promptForInteractiveParameter(
				"OAuth Scopes",
				strings.Join(serverConf.Scopes, ","),
			)
			serverConf.Scopes = strings.Split(scopesCsv, ",")
		}
		if serverConf.Audience == "" || forcePrompt {
			serverConf.Audience = promptForInteractiveParameter(
				"OAuth Audience",
				serverConf.Audience,
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
