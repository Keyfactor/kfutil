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
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strings"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"golang.org/x/crypto/bcrypt"
)

var (
	configFile      string
	profile         string
	providerType    string
	providerProfile string
	//providerConfig  string
	noPrompt        bool
	expEnabled      bool
	debugFlag       bool
	kfcUsername     string
	kfcHostName     string
	kfcPassword     string
	kfcDomain       string
	kfcClientId     string
	kfcClientSecret string
	kfcTokenUrl     string
	kfcAPIPath      string
	logInsecure     bool
	outputFormat    string
	offline         bool
)

func hashSecretValue(secretValue string) string {
	log.Debug().Msg("Enter hashSecretValue()")
	if logInsecure {
		return secretValue
	}
	log.Trace().Str("secretValue", secretValue).Send()
	cost := 12
	log.Debug().Int("cost", cost).Msg("call: bcrypt.GenerateFromPassword()")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(secretValue), cost)
	log.Debug().Msg("returned: bcrypt.GenerateFromPassword()")
	if err != nil {
		log.Error().Err(err).Msg("unable to hash secret value")
		return "*****************************"
	}
	log.Debug().Str("hashedPassword", string(hashedPassword)).Msg("return: hashSecretValue()")
	return string(hashedPassword)
}

func getServerConfigFromFile(configFile string, profile string) (*auth_providers.Server, error) {
	var commandConfig *auth_providers.Config
	var serverConfig auth_providers.Server

	log.Debug().
		Str("configFile", configFile).
		Str("profile", profile).
		Msg("configFile or profile is not empty attempting to authenticate via config file")
	if profile == "" {
		profile = "default"
	}
	if configFile == "" {
		homeDir, _ := os.UserHomeDir()
		configFile = fmt.Sprintf("%s/%s", homeDir, auth_providers.DefaultConfigFilePath)
	}
	var cfgReadErr error
	if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
		log.Debug().Msg("call: auth_providers.ReadConfigFromYAML()")
		//commandConfig, cfgReadErr = auth_providers.ReadConfigFromYAML(configFile)
		commandConfig, cfgReadErr = auth_providers.ReadConfigFromJSON(configFile)
	} else {
		log.Debug().Msg("call: auth_providers.ReadConfigFromJSON()")
		commandConfig, cfgReadErr = auth_providers.ReadConfigFromJSON(configFile)
	}

	if cfgReadErr != nil {
		log.Error().Err(cfgReadErr).Msg("unable to read config file")
		return nil, fmt.Errorf("unable to read config file: %s", cfgReadErr)
	}

	// check if the profile exists in the config file
	var ok bool
	if serverConfig, ok = commandConfig.Servers[profile]; !ok {
		log.Error().Str("profile", profile).Msg("invalid profile")
		return nil, fmt.Errorf("invalid profile: %s", profile)
	}

	log.Debug().Msg("return: getServerConfigFromFile()")
	return &serverConfig, nil
}

func getServerConfigFromEnv() (*auth_providers.Server, error) {
	log.Debug().Msg("Enter getServerConfigFromEnv()")

	oAuthNoParamsConfig := &auth_providers.CommandConfigOauth{}
	basicAuthNoParamsConfig := &auth_providers.CommandAuthConfigBasic{}

	log.Debug().Msg("call: basicAuthNoParamsConfig.Authenticate()")
	bErr := basicAuthNoParamsConfig.Authenticate()
	log.Debug().Msg("complete: basicAuthNoParamsConfig.Authenticate()")
	if bErr == nil {
		log.Debug().Msg("return: getServerConfigFromEnv()")
		return basicAuthNoParamsConfig.GetServerConfig(), nil
	}

	oErr := oAuthNoParamsConfig.Authenticate()
	if oErr == nil {
		log.Debug().Msg("return: getServerConfigFromEnv()")
		return oAuthNoParamsConfig.GetServerConfig(), nil
	}

	log.Error().Msg("unable to authenticate with provided credentials")
	if bErr != nil {
		return nil, bErr
	}
	return nil, oErr

}

func authViaConfigFile(cfgFile string, cfgProfile string) (*api.Client, error) {
	var (
		c    *api.Client
		cErr error
	)
	log.Debug().Msg("call: getServerConfigFromFile()")
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg("complete: getServerConfigFromFile()")
	if err != nil {
		log.Error().
			Err(err).
			Msg("unable to get server config from file")
		return nil, err
	}
	if conf != nil {
		log.Debug().Msg("call: api.NewKeyfactorClient()")
		c, cErr = api.NewKeyfactorClient(conf, nil)
		log.Debug().Msg("complete: api.NewKeyfactorClient()")
		if cErr != nil {
			log.Error().
				Err(cErr).
				Msg("unable to create Keyfactor client")
			return nil, cErr
		}
		log.Debug().Msg("call: c.AuthClient.Authenticate()")
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg("complete: c.AuthClient.Authenticate()")
		if authErr == nil {
			return c, nil
		}

	}
	log.Error().Msg("unable to authenticate via config file")
	return nil, fmt.Errorf("unable to authenticate via config file '%s' using profile '%s'", cfgFile, cfgProfile)
}
func authSdkViaConfigFile(cfgFile string, cfgProfile string) (*keyfactor.APIClient, error) {
	var (
		c    *keyfactor.APIClient
		cErr error
	)
	log.Debug().Msg("call: getServerConfigFromFile()")
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg("complete: getServerConfigFromFile()")
	if err != nil {
		log.Error().
			Err(err).
			Msg("unable to get server config from file")
		return nil, err
	}
	if conf != nil {
		log.Debug().Msg("call: keyfactor.NewAPIClient()")
		c = keyfactor.NewAPIClient(conf)
		log.Debug().Msg("complete: keyfactor.NewAPIClient()")
		if cErr != nil {
			log.Error().
				Err(cErr).
				Msg("unable to create Keyfactor client")
			return nil, cErr
		}
		log.Debug().Msg("call: c.AuthClient.Authenticate()")
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg("complete: c.AuthClient.Authenticate()")
		if authErr == nil {
			return c, nil
		}

	}
	log.Error().Msg("unable to authenticate via config file")
	return nil, fmt.Errorf("unable to authenticate via config file '%s' using profile '%s'", cfgFile, cfgProfile)
}

func authViaEnvVars() (*api.Client, error) {
	var (
		c    *api.Client
		cErr error
	)
	log.Debug().Msg("enter: authViaEnvVars()")
	log.Debug().Msg("call: getServerConfigFromEnv()")
	conf, err := getServerConfigFromEnv()
	log.Debug().Msg("complete: getServerConfigFromEnv()")
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via environment variables")
		log.Debug().Msg("return: authViaEnvVars()")
		return nil, err
	}
	if conf != nil {
		log.Debug().Msg("call: api.NewKeyfactorClient()")
		c, cErr = api.NewKeyfactorClient(conf, nil)
		log.Debug().Msg("complete: api.NewKeyfactorClient()")
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			log.Debug().Msg("return: authViaEnvVars()")
			return nil, cErr
		}
		log.Debug().Msg("call: c.AuthClient.Authenticate()")
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg("complete: c.AuthClient.Authenticate()")
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via environment variables")
			return nil, authErr
		}
		log.Debug().Msg("return: authViaEnvVars()")
		return c, nil
	}
	log.Error().Msg("unable to authenticate via environment variables")
	log.Debug().Msg("return: authViaEnvVars()")
	return nil, fmt.Errorf("unable to authenticate via environment variables")
}
func authSdkViaEnvVars() (*keyfactor.APIClient, error) {
	var (
		c    *keyfactor.APIClient
		cErr error
	)
	log.Debug().Msg("enter: authViaEnvVars()")
	log.Debug().Msg("call: getServerConfigFromEnv()")
	conf, err := getServerConfigFromEnv()
	log.Debug().Msg("complete: getServerConfigFromEnv()")
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via environment variables")
		log.Debug().Msg("return: authViaEnvVars()")
		return nil, err
	}
	if conf != nil {
		log.Debug().Msg("call: api.NewKeyfactorClient()")
		c = keyfactor.NewAPIClient(conf)
		log.Debug().Msg("complete: api.NewKeyfactorClient()")
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			log.Debug().Msg("return: authViaEnvVars()")
			return nil, cErr
		}
		log.Debug().Msg("call: c.AuthClient.Authenticate()")
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg("complete: c.AuthClient.Authenticate()")
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via environment variables")
			return nil, authErr
		}
		log.Debug().Msg("return: authViaEnvVars()")
		return c, nil
	}
	log.Error().Msg("unable to authenticate via environment variables")
	log.Debug().Msg("return: authViaEnvVars()")
	return nil, fmt.Errorf("unable to authenticate via environment variables")
}

func initClient(saveConfig bool) (*api.Client, error) {
	log.Debug().
		Str("configFile", configFile).
		Str("profile", profile).
		Str("providerType", providerType).
		Str("providerProfile", providerProfile).
		Bool("noPrompt", noPrompt).
		Bool("saveConfig", saveConfig).
		Str("hostname", kfcHostName).
		Str("username", kfcUsername).
		Str("password", hashSecretValue(kfcPassword)).
		Str("domain", kfcDomain).
		Str("clientId", kfcClientId).
		Str("clientSecret", hashSecretValue(kfcClientSecret)).
		Str("apiPath", kfcAPIPath).
		Str("providerType", providerType).
		Str("providerProfile", providerProfile).
		Msg("enter: initClient()")
	var (
		authenticated bool
		c             *api.Client
		cErr          error
	)

	if providerType != "" {
		log.Debug().
			Str("providerType", providerType).
			Msg("call: authViaProvider()")
		return authViaProvider()
	}
	log.Debug().
		Msg("providerType is empty attempting to authenticate via params")

	if configFile != "" || profile != "" {
		c, cErr = authViaConfigFile(configFile, profile)
		if cErr == nil {
			log.Info().
				Str("configFile", configFile).
				Str("profile", profile).
				Msgf("Authenticated via config file %s using profile %s", configFile, profile)
			authenticated = true
		}
	}

	if !authenticated {
		log.Debug().Msg("call: authViaEnvVars()")
		c, cErr = authViaEnvVars()
		log.Debug().Msg("returned: authViaEnvVars()")
		if cErr == nil {
			log.Info().Msg("Authenticated via environment variables")
			authenticated = true
		}
	}

	if !authenticated {
		log.Error().Msg("unable to authenticate")
		if cErr != nil {
			log.Debug().Err(cErr).Msg("return: initClient()")
			return nil, cErr
		}
		log.Debug().Msg("return: initClient()")
		return nil, fmt.Errorf("unable to authenticate to Keyfactor Command")
	}

	log.Info().Msg("Keyfactor Command client created")
	return c, nil
}

func initGenClient(
	saveConfig bool,
) (*keyfactor.APIClient, error) {
	log.Debug().
		Str("configFile", configFile).
		Str("profile", profile).
		Str("providerType", providerType).
		Str("providerProfile", providerProfile).
		Bool("noPrompt", noPrompt).
		Bool("saveConfig", saveConfig).
		Str("hostname", kfcHostName).
		Str("username", kfcUsername).
		Str("password", hashSecretValue(kfcPassword)).
		Str("domain", kfcDomain).
		Str("clientId", kfcClientId).
		Str("clientSecret", hashSecretValue(kfcClientSecret)).
		Str("apiPath", kfcAPIPath).
		Str("providerType", providerType).
		Str("providerProfile", providerProfile).
		Msg("enter: initGenClient()")

	var (
		authenticated bool
		c             *keyfactor.APIClient
		cErr          error
	)

	if providerType != "" {
		log.Debug().
			Str("providerType", providerType).
			Msg("call: authViaProvider()")
		//return authViaProvider()
		return nil, fmt.Errorf("provider auth not supported using Keyfactor Command SDK")
	}
	log.Debug().
		Msg("providerType is empty attempting to authenticate via params")

	if configFile != "" || profile != "" {
		c, cErr = authSdkViaConfigFile(configFile, profile)
		if cErr == nil {
			log.Info().
				Str("configFile", configFile).
				Str("profile", profile).
				Msgf("Authenticated via config file %s using profile %s", configFile, profile)
			authenticated = true
		}
	}

	if !authenticated {
		log.Debug().Msg("call: authViaEnvVars()")
		c, cErr = authSdkViaEnvVars()
		log.Debug().Msg("returned: authViaEnvVars()")
		if cErr == nil {
			log.Info().Msg("Authenticated via environment variables")
			authenticated = true
		}
	}

	log.Info().Msg("Keyfactor Command client created")
	return c, nil
}

//func initGenClientV1(
//	flagConfig string,
//	flagProfile string,
//	noPrompt bool,
//	authConfig *api.AuthConfig,
//	saveConfig bool,
//) (*keyfactor.APIClient, error) {
//	var commandConfig ConfigurationFile
//
//	if providerType != "" {
//		return authViaProviderGenClient()
//	}
//
//	commandConfig, _ = authEnvVars(flagConfig, "", saveConfig)
//
//	if flagConfig != "" || !validConfigFileEntry(commandConfig, flagProfile) {
//		commandConfig, _ = authConfigFile(flagConfig, flagProfile, "", noPrompt, saveConfig)
//	}
//
//	if flagProfile == "" {
//		flagProfile = "default"
//	}
//
//	//Params from authConfig take precedence over everything else
//	if authConfig != nil {
//		// replace commandConfig with authConfig params that aren't null or empty
//		configEntry := commandConfig.Servers[flagProfile]
//		if authConfig.Hostname != "" {
//			configEntry.Hostname = authConfig.Hostname
//		}
//		if authConfig.Username != "" {
//			configEntry.Username = authConfig.Username
//		}
//		if authConfig.Password != "" {
//			configEntry.Password = authConfig.Password
//		}
//		if authConfig.Domain != "" {
//			configEntry.Domain = authConfig.Domain
//		} else if authConfig.Username != "" {
//			tDomain := getDomainFromUsername(authConfig.Username)
//			if tDomain != "" {
//				configEntry.Domain = tDomain
//			}
//		}
//		if authConfig.APIPath != "" {
//			configEntry.APIPath = authConfig.APIPath
//		}
//		commandConfig.Servers[flagProfile] = configEntry
//	}
//
//	if !validConfigFileEntry(commandConfig, flagProfile) {
//		if !noPrompt {
//			// Auth user interactively
//			authConfigEntry := commandConfig.Servers[flagProfile]
//			commandConfig, _ = authInteractive(
//				authConfigEntry.Hostname,
//				authConfigEntry.Username,
//				authConfigEntry.Password,
//				authConfigEntry.Domain,
//				authConfigEntry.APIPath,
//				flagProfile,
//				false,
//				false,
//				flagConfig,
//			)
//		} else {
//			//log.Fatalf("[ERROR] auth config profile: %s", flagProfile)
//			log.Error().Str("flagProfile", flagProfile).Msg("invalid auth config profile")
//			return nil, fmt.Errorf("auth config profile: %s", flagProfile)
//		}
//	}
//
//	sdkClientConfig := make(map[string]string)
//	sdkClientConfig["host"] = commandConfig.Servers[flagProfile].Hostname
//	sdkClientConfig["username"] = commandConfig.Servers[flagProfile].Username
//	sdkClientConfig["password"] = commandConfig.Servers[flagProfile].Password
//	sdkClientConfig["domain"] = commandConfig.Servers[flagProfile].Domain
//
//	configuration := keyfactor.NewConfiguration(sdkClientConfig)
//	c := keyfactor.NewAPIClient(configuration)
//	return c, nil
//}

var makeDocsCmd = &cobra.Command{
	Use:    "makedocs",
	Short:  "Generate markdown documentation for kfutil",
	Long:   `Generate markdown documentation for kfutil.`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Msg("Enter makeDocsCmd.Run()")
		doc.GenMarkdownTree(RootCmd, "./docs")
		log.Debug().Msg("complete: makeDocsCmd.Run()")
	},
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

	RootCmd.PersistentFlags().StringVarP(
		&configFile,
		"config",
		"",
		"",
		fmt.Sprintf("Full path to config file in JSON format. (default is %s)", defaultConfigPath),
	)
	RootCmd.PersistentFlags().BoolVar(
		&noPrompt,
		"no-prompt",
		false,
		"Do not prompt for any user input and assume defaults or environmental variables are set.",
	)
	RootCmd.PersistentFlags().BoolVar(
		&expEnabled,
		"exp",
		false,
		"Enable expEnabled features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)",
	)
	RootCmd.PersistentFlags().BoolVar(
		&offline,
		"offline",
		false,
		"Will not attempt to connect to GitHub for latest release information and resources.",
	)
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debugFlag logging.")
	RootCmd.PersistentFlags().BoolVar(
		&logInsecure,
		"log-insecure",
		false,
		"Log insecure API requests. (USE AT YOUR OWN RISK, this WILL log sensitive information to the console.)",
	)
	RootCmd.PersistentFlags().StringVarP(
		&profile,
		"profile",
		"",
		"",
		"Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.",
	)
	RootCmd.PersistentFlags().StringVar(
		&outputFormat,
		"format",
		"text",
		"How to format the CLI output. Currently only `text` is supported.",
	)

	RootCmd.PersistentFlags().StringVar(&providerType, "auth-provider-type", "", "Provider type choices: (azid)")
	// Validating the provider-type flag against the predefined choices
	RootCmd.PersistentFlags().SetAnnotation("auth-provider-type", cobra.BashCompCustom, ProviderTypeChoices)
	RootCmd.PersistentFlags().StringVarP(
		&providerProfile,
		"auth-provider-profile",
		"",
		"default",
		"The profile to use defined in the securely stored config. If not specified the config named 'default' will be used if it exists.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&kfcUsername,
		"username",
		"",
		"",
		"Username to use for authenticating to Keyfactor Command.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&kfcHostName,
		"hostname",
		"",
		"",
		"Hostname to use for authenticating to Keyfactor Command.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&kfcPassword,
		"password",
		"",
		"",
		"Password to use for authenticating to Keyfactor Command. WARNING: Remember to delete your console history if providing kfcPassword here in plain text.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&kfcDomain,
		"domain",
		"",
		"",
		"Domain to use for authenticating to Keyfactor Command.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&kfcAPIPath,
		"api-path",
		"",
		"KeyfactorAPI",
		"API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI)",
	)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	RootCmd.AddCommand(makeDocsCmd)
}
