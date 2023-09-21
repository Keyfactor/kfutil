// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	stdlog "log"
	"os"
	"strconv"
	"time"
)

var (
	configFile      string
	profile         string
	providerType    string
	providerProfile string
	providerConfig  string
	noPrompt        bool
	expEnabled      bool
	debugFlag       bool
	kfcUsername     string
	kfcHostName     string
	kfcPassword     string
	kfcDomain       string
	kfcAPIPath      string
	logInsecure     bool
	outputFormat    string
)

func hashSecretValue(secretValue string) string {
	log.Debug().Msg("Enter hashSecretValue()")
	// Create a new SHA-256 hasher
	hasher := sha256.New()

	// Write the string bytes to the hasher
	_, err := hasher.Write([]byte(secretValue))
	if err != nil {
		log.Error().Err(err)
		return "failed to hash secret value"
	}

	// Get the final hash as a byte slice
	hashBytes := hasher.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)
	log.Trace().Str("hashString", hashString).Msg("secret hash")

	log.Debug().Msg("Exit hashSecretValue()")
	return hashString
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
	sdkClientConfig["kfcUsername"] = commandConfig.Servers[flagProfile].Username
	sdkClientConfig["kfcPassword"] = commandConfig.Servers[flagProfile].Password
	sdkClientConfig["kfcDomain"] = commandConfig.Servers[flagProfile].Domain

	configuration := keyfactor.NewConfiguration(sdkClientConfig)
	c := keyfactor.NewAPIClient(configuration)
	return c, nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kfutil",
	Short: "Keyfactor CLI utilities",
	Long:  `A CLI wrapper around the Keyfactor Platform API.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	stdlog.SetOutput(io.Discard)
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	initLogger()

	defaultConfigPath := fmt.Sprintf("$HOME/.keyfactor/%s", DefaultConfigFileName)

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "", "", fmt.Sprintf("Full path to config file in JSON format. (default is $HOME/.keyfactor/%s)", DefaultConfigFileName))
	RootCmd.PersistentFlags().BoolVar(&noPrompt, "no-prompt", false, "Do not prompt for any user input and assume defaults or environmental variables are set.")
	RootCmd.PersistentFlags().BoolVar(&expEnabled, "exp", false, "Enable expEnabled features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)")
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debugFlag logging.")
	RootCmd.PersistentFlags().BoolVar(&logInsecure, "log-insecure", false, "Log insecure API requests. (USE AT YOUR OWN RISK, this WILL log sensitive information to the console.)")
	RootCmd.PersistentFlags().StringVarP(&profile, "profile", "", "", "Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.")
	RootCmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "Output format. (text/json)")

	RootCmd.PersistentFlags().StringVar(&providerType, "auth-provider-type", "", "Provider type choices: (azid/azcli)")
	// Validating the provider-type flag against the predefined choices
	RootCmd.PersistentFlags().SetAnnotation("auth-provider-type", cobra.BashCompCustom, ProviderTypeChoices)
	RootCmd.PersistentFlags().StringVarP(&providerProfile, "auth-provider-profile", "", "default", "Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.")
	RootCmd.PersistentFlags().StringVarP(&providerConfig, "auth-provider-config", "", defaultConfigPath, fmt.Sprintf("Full path to config file in JSON format. (default is $HOME/.keyfactor/%s)", DefaultConfigFileName))

	RootCmd.PersistentFlags().StringVarP(&kfcUsername, "username", "", "", "Username to use for authenticating to Keyfactor Command.")
	RootCmd.PersistentFlags().StringVarP(&kfcHostName, "hostname", "", "", "Hostname to use for authenticating to Keyfactor Command.")
	RootCmd.PersistentFlags().StringVarP(&kfcPassword, "password", "", "", "Password to use for authenticating to Keyfactor Command. WARNING: Remember to delete your console history if providing kfcPassword here in plain text.")
	RootCmd.PersistentFlags().StringVarP(&kfcDomain, "domain", "", "", "Domain to use for authenticating to Keyfactor Command.")
	RootCmd.PersistentFlags().StringVarP(&kfcAPIPath, "api-path", "", "KeyfactorAPI", "API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}

func boolToPointer(b bool) *bool {
	return &b
}

func intToPointer(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func stringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func IsExperimentalFeatureEnabled(expFlag bool, isExperimental bool) (bool, error) {
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
