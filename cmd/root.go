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
	"github.com/Keyfactor/keyfactor-go-client-sdk/v2/api/keyfactor"
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

// hashSecretValue hashes the secret value using bcrypt
func hashSecretValue(secretValue string) string {
	log.Debug().Msg("Enter hashSecretValue()")
	if secretValue == "" {
		return secretValue
	}
	if !logInsecure {
		return "*****************************"
	}
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

// getServerConfigFromFile reads the configuration file and returns the server configuration
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

// getServerConfigFromEnv reads the environment variables and returns the server configuration
func getServerConfigFromEnv() (*auth_providers.Server, error) {
	log.Debug().Msg("Enter getServerConfigFromEnv()")

	oAuthNoParamsConfig := &auth_providers.CommandConfigOauth{}
	basicAuthNoParamsConfig := &auth_providers.CommandAuthConfigBasic{}

	username, uOk := os.LookupEnv(auth_providers.EnvKeyfactorUsername)
	password, pOk := os.LookupEnv(auth_providers.EnvKeyfactorPassword)
	domain, dOk := os.LookupEnv(auth_providers.EnvKeyfactorDomain)
	hostname, hOk := os.LookupEnv(auth_providers.EnvKeyfactorHostName)
	apiPath, aOk := os.LookupEnv(auth_providers.EnvKeyfactorAPIPath)
	clientId, cOk := os.LookupEnv(auth_providers.EnvKeyfactorClientID)
	clientSecret, csOk := os.LookupEnv(auth_providers.EnvKeyfactorClientSecret)
	tokenUrl, tOk := os.LookupEnv(auth_providers.EnvKeyfactorAuthTokenURL)
	skipVerify, svOk := os.LookupEnv(auth_providers.EnvKeyfactorSkipVerify)
	var skipVerifyBool bool

	isBasicAuth := uOk && pOk
	isOAuth := cOk && csOk && tOk

	if svOk {
		//convert to bool
		skipVerify = strings.ToLower(skipVerify)
		skipVerifyBool = skipVerify == "true" || skipVerify == "1" || skipVerify == "yes" || skipVerify == "y" || skipVerify == "t"
		log.Debug().Bool("skipVerifyBool", skipVerifyBool).Msg("skipVerifyBool")
	}
	if dOk {
		log.Debug().Str("domain", domain).Msg("domain found in environment")
	}
	if hOk {
		log.Debug().Str("hostname", hostname).Msg("hostname found in environment")
	}
	if aOk {
		log.Debug().Str("apiPath", apiPath).Msg("apiPath found in environment")
	}

	if isBasicAuth {
		log.Debug().
			Str("username", username).
			Str("password", hashSecretValue(password)).
			Str("domain", domain).
			Str("hostname", hostname).
			Str("apiPath", apiPath).
			Bool("skipVerify", skipVerifyBool).
			Msg("call: basicAuthNoParamsConfig.Authenticate()")
		basicAuthNoParamsConfig.WithCommandHostName(hostname).
			WithCommandAPIPath(apiPath).
			WithSkipVerify(skipVerifyBool)

		bErr := basicAuthNoParamsConfig.
			WithUsername(username).
			WithPassword(password).
			WithDomain(domain).
			Authenticate()
		log.Debug().Msg("complete: basicAuthNoParamsConfig.Authenticate()")
		if bErr != nil {
			log.Error().Err(bErr).Msg("unable to authenticate with provided credentials")
			return nil, bErr
		}
		log.Debug().Msg("return: getServerConfigFromEnv()")
		return basicAuthNoParamsConfig.GetServerConfig(), nil
	} else if isOAuth {
		log.Debug().
			Str("clientId", clientId).
			Str("clientSecret", hashSecretValue(clientSecret)).
			Str("tokenUrl", tokenUrl).
			Str("hostname", hostname).
			Str("apiPath", apiPath).
			Bool("skipVerify", skipVerifyBool).
			Msg("call: oAuthNoParamsConfig.Authenticate()")
		_ = oAuthNoParamsConfig.CommandAuthConfig.WithCommandHostName(hostname).
			WithCommandAPIPath(apiPath).
			WithSkipVerify(skipVerifyBool)
		oErr := oAuthNoParamsConfig.Authenticate()
		log.Debug().Msg("complete: oAuthNoParamsConfig.Authenticate()")
		if oErr != nil {
			log.Error().Err(oErr).Msg("unable to authenticate with provided credentials")
			return nil, oErr
		}

		log.Debug().Msg("return: getServerConfigFromEnv()")
		return oAuthNoParamsConfig.GetServerConfig(), nil

	}

	log.Error().Msg("unable to authenticate with provided credentials")
	return nil, fmt.Errorf("incomplete environment variable configuration")

}

// authViaConfigFile authenticates using the configuration file
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
		if conf.AuthProvider.Type != "" {
			switch conf.AuthProvider.Type {
			case "azid", "azure", "az", "akv":
				return authViaProvider(cfgFile, cfgProfile)
			}
		}
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

// authSdkViaConfigFile authenticates using the configuration file
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
		if conf.AuthProvider.Type != "" {
			switch conf.AuthProvider.Type {
			case "azid", "azure", "az", "akv":
				log.Debug().
					Str("providerType", conf.AuthProvider.Type).
					Str("providerProfile", conf.AuthProvider.Profile).
					Str("cfgFile", cfgFile).
					Str("cfgProfile", cfgProfile).
					Msg("call: authSdkViaProvider()")
				return authSdkViaProvider(cfgFile, cfgProfile)
			}
		}
		log.Debug().Msg("call: keyfactor.NewAPIClient()")
		c, cErr = keyfactor.NewAPIClient(conf)
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

// authViaEnvVars authenticates using the environment variables
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

// authSdkViaEnvVars authenticates using the environment variables
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
		c, cErr = keyfactor.NewAPIClient(conf)
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

// authViaProvider authenticates using the provider
func authViaProvider(cfgFile string, cfgProfile string) (*api.Client, error) {
	log.Debug().
		Str("providerType", providerType).
		Str("providerProfile", providerProfile).
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Msg("enter: authViaProvider()")
	var (
		c    *api.Client
		cErr error
	)

	log.Debug().Msg("call: getServerConfigFromFile()")
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg("complete: getServerConfigFromFile()")
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via provider")
		return nil, err
	}

	if providerType == "" {
		providerType = conf.AuthProvider.Type
	}

	if providerType == "azid" || providerType == "azure" {
		azConfig := &auth_providers.ConfigProviderAzureKeyVault{}
		secretName, sOk := os.LookupEnv(auth_providers.EnvAzureSecretName)
		vaultName, vOk := os.LookupEnv(auth_providers.EnvAzureVaultName)
		if !sOk {
			secretName, sOk = conf.AuthProvider.Parameters["secret_name"].(string)
		}
		if !vOk {
			vaultName, vOk = conf.AuthProvider.Parameters["vault_name"].(string)
		}
		aErr := azConfig.
			WithSecretName(secretName).
			WithVaultName(vaultName).
			Authenticate()
		if aErr != nil {
			log.Error().Err(aErr).Msg("unable to authenticate via provider")
			return nil, aErr
		}
		cfg, cfgErr := azConfig.LoadConfigFromAzureKeyVault()
		if cfgErr != nil {
			log.Error().Err(cfgErr).Msg("unable to load config from Azure Key Vault")
			return nil, cfgErr
		}
		log.Debug().Msg("call: api.NewKeyfactorClient()")
		serverConfig, serOk := cfg.Servers[providerProfile]
		if !serOk {
			log.Error().Str("profile", providerProfile).Msg("invalid profile")
			return nil, fmt.Errorf("invalid profile: %s", providerProfile)
		}
		c, cErr = api.NewKeyfactorClient(&serverConfig, nil)
		log.Debug().Msg("complete: api.NewKeyfactorClient()")
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			return nil, cErr
		}
		log.Debug().Msg("call: c.AuthClient.Authenticate()")
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg("complete: c.AuthClient.Authenticate()")
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via provider")
			return nil, authErr
		}
		return c, nil
	}
	log.Error().Str("providerType", providerType).Msg("unsupported provider type")
	return nil, fmt.Errorf("unsupported provider type: %s", providerType)
}

// authSdkViaProvider authenticates using the provider
func authSdkViaProvider(cfgFile string, cfgProfile string) (*keyfactor.APIClient, error) {
	log.Debug().
		Str("providerType", providerType).
		Str("providerProfile", providerProfile).
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Msg("enter: authViaProvider()")
	var (
		c    *keyfactor.APIClient
		cErr error
	)

	log.Debug().Msg("call: getServerConfigFromFile()")
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg("complete: getServerConfigFromFile()")
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via provider")
		return nil, err
	}

	if providerType == "" {
		providerType = conf.AuthProvider.Type
	}

	if providerType == "azid" || providerType == "azure" {
		azConfig := &auth_providers.ConfigProviderAzureKeyVault{}
		secretName, sOk := os.LookupEnv(auth_providers.EnvAzureSecretName)
		vaultName, vOk := os.LookupEnv(auth_providers.EnvAzureVaultName)
		if !sOk {
			secretName, sOk = conf.AuthProvider.Parameters["secret_name"].(string)
		}
		if !vOk {
			vaultName, vOk = conf.AuthProvider.Parameters["vault_name"].(string)
		}
		aErr := azConfig.
			WithSecretName(secretName).
			WithVaultName(vaultName).
			Authenticate()
		if aErr != nil {
			log.Error().Err(aErr).Msg("unable to authenticate via provider")
			return nil, aErr
		}
		cfg, cfgErr := azConfig.LoadConfigFromAzureKeyVault()
		if cfgErr != nil {
			log.Error().Err(cfgErr).Msg("unable to load config from Azure Key Vault")
			return nil, cfgErr
		}

		serverConfig, serOk := cfg.Servers[providerProfile]
		if !serOk {
			log.Error().Str("profile", providerProfile).Msg("invalid profile")
			return nil, fmt.Errorf("invalid profile: %s", providerProfile)
		}
		log.Debug().Msg("call: keyfactor.NewAPIClient()")
		c, cErr = keyfactor.NewAPIClient(&serverConfig)
		log.Debug().Msg("complete: keyfactor.NewAPIClient()")
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			return nil, cErr
		}
		log.Debug().Msg("call: c.AuthClient.Authenticate()")
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg("complete: c.AuthClient.Authenticate()")
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via provider")
			return nil, authErr
		}
		return c, nil
	}
	log.Error().Str("providerType", providerType).Msg("unsupported provider type")
	return nil, fmt.Errorf("unsupported provider type: %s", providerType)
}

// initClient initializes the legacy Command API client
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
		c         *api.Client
		envCfgErr error
		cfgErr    error
	)

	if providerType != "" {
		log.Debug().
			Str("providerType", providerType).
			Msg("call: authViaProvider()")
		return authViaProvider("", "")
	}
	log.Debug().
		Msg("providerType is empty attempting to authenticate via params")

	if configFile != "" || profile != "" {
		log.Info().
			Str("configFile", configFile).
			Str("profile", profile).
			Msg("authenticating via config file")
		c, cfgErr = authViaConfigFile(configFile, profile)
		if cfgErr == nil {
			log.Info().
				Str("configFile", configFile).
				Str("profile", profile).
				Msgf("Authenticated via config file %s using profile %s", configFile, profile)
			return c, nil
		}
	}

	log.Info().Msg("authenticating via environment variables")
	log.Debug().Msg("call: authViaEnvVars()")
	c, envCfgErr = authViaEnvVars()
	log.Debug().Msg("returned: authViaEnvVars()")
	if envCfgErr == nil {
		log.Info().Msg("Authenticated via environment variables")
		return c, nil
	}

	log.Info().
		Str("configFile", DefaultConfigFileName).
		Str("profile", "default").
		Msg("implicit authenticating via config file using default profile")
	log.Debug().Msg("call: authViaConfigFile()")
	c, cfgErr = authViaConfigFile("", "")
	if cfgErr == nil {
		log.Info().
			Str("configFile", DefaultConfigFileName).
			Str("profile", "default").
			Msgf("authenticated implictly via config file '%s' using 'default' profile", DefaultConfigFileName)
		return c, nil
	}

	log.Error().
		Err(cfgErr).
		Err(envCfgErr).
		Msg("unable to authenticate to Keyfactor Command")
	log.Debug().Msg("return: initClient()")
	return nil, fmt.Errorf("unable to authenticate to Keyfactor Command")
}

// initGenClient initializes the SDK Command API client
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
		c       *keyfactor.APIClient
		envCErr error
		cfErr   error
	)

	if providerType != "" {
		log.Debug().
			Str("providerType", providerType).
			Msg("call: authSdkViaProvider()")
		return authSdkViaProvider("", "")
	}
	log.Debug().
		Msg("providerType is empty attempting to authenticate via params")

	if configFile != "" || profile != "" {
		log.Info().
			Str("configFile", configFile).
			Str("profile", profile).
			Msg("authenticating via config file")
		c, cfErr = authSdkViaConfigFile(configFile, profile)
		if cfErr == nil {
			log.Info().
				Str("configFile", configFile).
				Str("profile", profile).
				Msgf("Authenticated via config file %s using profile %s", configFile, profile)
			return c, nil
		}
	}

	log.Info().Msg("authenticating via environment variables")
	log.Debug().Msg("call: authViaEnvVars()")
	c, envCErr = authSdkViaEnvVars()
	log.Debug().Msg("returned: authViaEnvVars()")
	if envCErr == nil {
		log.Info().Msg("authenticated via environment variables")
		return c, nil
	}

	log.Info().
		Str("configFile", DefaultConfigFileName).
		Str("profile", "default").
		Msg("implicit authenticating via config file using default profile")
	log.Debug().Msg("call: authViaConfigFile()")
	c, cfErr = authSdkViaConfigFile("", "")
	if cfErr == nil {
		log.Info().
			Str("configFile", DefaultConfigFileName).
			Str("profile", "default").
			Msgf("authenticated implictly via config file '%s' using 'default' profile", DefaultConfigFileName)
		return c, nil
	}

	log.Error().
		Err(cfErr).
		Err(envCErr).
		Msg("unable to authenticate")
	return nil, fmt.Errorf("unable to authenticate to Keyfactor Command with provided credentials, please check your configuration")
}

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
	//RootCmd.PersistentFlags().BoolVar(
	//	&logInsecure,
	//	"log-insecure",
	//	false,
	//	"Log insecure API requests. (USE AT YOUR OWN RISK, this WILL log sensitive information to the console.)",
	//)
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
		&kfcClientId,
		"client-id",
		"",
		"",
		"OAuth2 client-id to use for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&kfcClientSecret,
		"client-secret",
		"",
		"",
		"OAuth2 client-secret to use for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&kfcClientId,
		"token-url",
		"",
		"",
		"OAuth2 token endpoint full URL to use for authenticating to Keyfactor Command.",
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
