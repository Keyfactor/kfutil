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
	"path"
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
	flagConfigFile      string // Path to a Command configuration file. Default `$HOME/.keyfactor/command_config.json`
	flagProfile         string // Profile to use in the Command configuration file. Default `default`
	flagProviderType    string // Type of auth provider to use for authentication. (only azid is supported)
	flagProviderProfile string // The flagProfile to use in the Command configuration file after it has been loaded from the auth provider.
	//providerConfig  string
	flagNoPrompt     bool   // Do not prompt for missing credentials
	flagEnableExp    bool   // Enable experimental features
	flagEnableDebug  bool   // Enable debug logging
	flagHostName     string // Hostname for Keyfactor Command instance
	flagUsername     string // Username for Keyfactor Command Basic authentication
	flagPassword     string // Password for Keyfactor Command Basic authentication
	flagDomain       string // Domain for Keyfactor Command Basic authentication
	flagClientId     string // Client ID for Keyfactor Command OAuth2 authentication
	flagClientSecret string // Client Secret for Keyfactor Command OAuth2 authentication
	flagTokenUrl     string // Token URL for Keyfactor Command OAuth2 authentication
	flagAccessToken  string // Access Token for Keyfactor Command OAuth2 authentication
	flagAPIPath      string // API Path for Keyfactor Command instance. Default `/KeyfactorAPI`
	flagLogInsecure  bool   // Log sensitive information in plaintext
	flagOutputFormat string // Output format for the command. Default `json`
	flagOffline      bool   // Do not reach out to GitHub for latest versions of PAM and Store types
	flagSkipVerify   bool   // Skip SSL verification when communicating to Keyfactor Command
)

// hashSecretValue hashes the secret value using bcrypt
func hashSecretValue(secretValue string) string {
	log.Debug().Msg(fmt.Sprintf("%s hashSecretValue()", DebugFuncEnter))
	if secretValue == "" {
		return secretValue
	}
	if !flagLogInsecure {
		log.Debug().Msg(fmt.Sprintf("%s hashSecretValue()", DebugFuncExit))
		return "*****************************"
	}
	log.Debug().Msg("insecure logging enabled, attempting to hash secret value")
	cost := 12
	log.Debug().Int("cost", cost).Msg(fmt.Sprintf("%s bcrypt.GenerateFromPassword()", DebugFuncCall))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(secretValue), cost)
	log.Debug().Msg(fmt.Sprintf("%s bcrypt.GenerateFromPassword()", DebugFuncReturn))
	if err != nil {
		log.Error().Err(err).Msg("unable to hash secret value")
		log.Debug().Msg(fmt.Sprintf("%s hashSecretValue()", DebugFuncExit))
		return "*****************************"
	}
	log.Debug().Str("hashedPassword", string(hashedPassword)).Msg(fmt.Sprintf("%s hashSecretValue()", DebugFuncExit))
	return string(hashedPassword)
}

// getServerConfigFromFile reads the configuration file and returns the server configuration
func getServerConfigFromFile(configFile string, profile string) (*auth_providers.Server, error) {
	log.Debug().
		Str("flagConfigFile", configFile).
		Str("flagProfile", profile).
		Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncEnter))

	var commandConfig *auth_providers.Config
	var serverConfig auth_providers.Server

	if profile == "" {
		if flagProfile != "" {
			log.Trace().
				Str("flagProfile", flagProfile).
				Msg("using flagProfile")
			profile = flagProfile
		} else {
			log.Trace().Msg("using default profile")
			profile = auth_providers.DefaultConfigProfile
		}
	}
	if configFile == "" {
		if flagConfigFile != "" {
			log.Trace().
				Str("flagConfigFile", flagConfigFile).
				Msg("using flagConfigFile")
			configFile = flagConfigFile
		} else {
			log.Trace().Msg("using default config file")
			homeDir, hErr := os.UserHomeDir()
			if hErr != nil {
				log.Error().Err(hErr).Msg("unable to get user home directory using current directory to locate config file")
				homeDir = "."
			}
			configFile = path.Join(homeDir, auth_providers.DefaultConfigFilePath)
		}
	}
	var cfgReadErr error
	if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
		log.Debug().Msg(fmt.Sprintf("%s auth_providers.ReadConfigFromYAML()", DebugFuncCall))
		commandConfig, cfgReadErr = auth_providers.ReadConfigFromYAML(flagConfigFile)
		log.Debug().Msg(fmt.Sprintf("%s auth_providers.ReadConfigFromYAML()", DebugFuncReturn))

	} else {
		log.Debug().Msg(fmt.Sprintf("%s auth_providers.ReadConfigFromJSON()", DebugFuncCall))
		commandConfig, cfgReadErr = auth_providers.ReadConfigFromJSON(configFile)
		log.Debug().Msg(fmt.Sprintf("%s auth_providers.ReadConfigFromJSON()", DebugFuncReturn))
	}

	if cfgReadErr != nil {
		log.Error().Err(cfgReadErr).Msg("unable to read config file")
		log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncExit))
		return nil, fmt.Errorf("unable to read config file: %s", cfgReadErr)
	}

	// check if the flagProfile exists in the config file
	var ok bool
	if serverConfig, ok = commandConfig.Servers[profile]; !ok {
		log.Error().Str("profile", profile).Msg("invalid profile")
		log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncExit))
		return nil, fmt.Errorf("invalid profile: %s", profile)
	}

	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncExit))
	return &serverConfig, nil
}

// getServerConfigFromEnv reads the environment variables and returns the server configuration.
// NOTE: global CLI flags WILL OVERRIDE the environment variables if set.
func getServerConfigFromEnv() (*auth_providers.Server, error) {
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncEnter))

	oAuthNoParamsConfig := &auth_providers.CommandConfigOauth{}
	basicAuthNoParamsConfig := &auth_providers.CommandAuthConfigBasic{}

	username, _ := os.LookupEnv(auth_providers.EnvKeyfactorUsername)
	if flagUsername != "" {
		log.Debug().
			Str("flagUsername", flagUsername).
			Str("envUsername", username).
			Msg("using flagUsername")
		username = flagUsername
	}

	password, _ := os.LookupEnv(auth_providers.EnvKeyfactorPassword)
	if flagPassword != "" {
		log.Debug().
			Str("flagPassword", hashSecretValue(flagPassword)).
			Str("envPassword", hashSecretValue(password)).
			Msg("using flagPassword")
		password = flagPassword
	}
	domain, _ := os.LookupEnv(auth_providers.EnvKeyfactorDomain)
	if flagDomain != "" {
		log.Debug().
			Str("flagDomain", flagDomain).
			Str("envDomain", domain).
			Msg("using flagDomain")
		domain = flagDomain
	}
	hostname, _ := os.LookupEnv(auth_providers.EnvKeyfactorHostName)
	if flagHostName != "" {
		log.Debug().
			Str("flagHostName", flagHostName).
			Str("envHostName", hostname).
			Msg("using flagHostName")
		hostname = flagHostName
	}
	apiPath, _ := os.LookupEnv(auth_providers.EnvKeyfactorAPIPath)
	if flagAPIPath != "" {
		log.Debug().
			Str("flagAPIPath", flagAPIPath).
			Str("envAPIPath", apiPath).
			Msg("using flagAPIPath")
		apiPath = flagAPIPath
	}
	clientId, _ := os.LookupEnv(auth_providers.EnvKeyfactorClientID)
	if flagClientId != "" {
		log.Debug().
			Str("flagClientId", flagClientId).
			Str("envClientId", clientId).
			Msg("using flagClientId")
		clientId = flagClientId
	}
	clientSecret, _ := os.LookupEnv(auth_providers.EnvKeyfactorClientSecret)
	if flagClientSecret != "" {
		log.Debug().
			Str("flagClientSecret", hashSecretValue(flagClientSecret)).
			Str("envClientSecret", hashSecretValue(clientSecret)).
			Msg("using flagClientSecret")
		clientSecret = flagClientSecret
	}

	accessToken, _ := os.LookupEnv(auth_providers.EnvKeyfactorAccessToken)
	if flagAccessToken != "" {
		log.Debug().
			Str("flagAccessToken", flagAccessToken).
			Str("envAccessToken", accessToken).
			Msg("using flagAccessToken")
		accessToken = flagAccessToken
	}

	tokenUrl, _ := os.LookupEnv(auth_providers.EnvKeyfactorAuthTokenURL)
	if flagTokenUrl != "" {
		log.Debug().
			Str("flagTokenUrl", flagTokenUrl).
			Str("envTokenUrl", tokenUrl).
			Msg("using flagTokenUrl")
		tokenUrl = flagTokenUrl
	}
	skipVerify, _ := os.LookupEnv(auth_providers.EnvKeyfactorSkipVerify)
	if flagSkipVerify {
		log.Debug().
			Bool("flagSkipVerify", flagSkipVerify).
			Str("envSkipVerify", skipVerify).
			Msg("using flagSkipVerify")
		skipVerify = "true"
	}
	var skipVerifyBool bool

	isBasicAuth := username != "" && password != ""
	isOAuth := (clientId != "" && clientSecret != "" && tokenUrl != "") || accessToken != ""

	if skipVerify != "" {
		//convert to bool
		skipVerify = strings.ToLower(skipVerify)
		skipVerifyBool = skipVerify == "true" || skipVerify == "1" || skipVerify == "yes" || skipVerify == "y" || skipVerify == "t"
		log.Debug().Bool("skipVerifyBool", skipVerifyBool).Msg("skipVerifyBool")
	}

	if isBasicAuth {
		log.Debug().
			Str("username", username).
			Str("password", hashSecretValue(password)).
			Str("domain", domain).
			Str("hostname", hostname).
			Str("apiPath", apiPath).
			Bool("skipVerify", skipVerifyBool).
			Msg(fmt.Sprintf("%s basicAuthNoParamsConfig.Authenticate()", DebugFuncCall))
		_ = basicAuthNoParamsConfig.WithCommandHostName(hostname).
			WithCommandAPIPath(apiPath).
			WithSkipVerify(skipVerifyBool)

		bErr := basicAuthNoParamsConfig.
			WithUsername(username).
			WithPassword(password).
			WithDomain(domain).
			Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s basicAuthNoParamsConfig.Authenticate()", DebugFuncReturn))
		if bErr != nil {
			log.Error().Err(bErr).Msg("unable to authenticate with provided credentials")
			log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncExit))
			return nil, bErr
		}
		log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncExit))
		return basicAuthNoParamsConfig.GetServerConfig(), nil
	} else if isOAuth {
		log.Debug().
			Str("clientId", clientId).
			Str("clientSecret", hashSecretValue(clientSecret)).
			Str("tokenUrl", tokenUrl).
			Str("accessToken", accessToken).
			Str("hostname", hostname).
			Str("apiPath", apiPath).
			Bool("skipVerify", skipVerifyBool).
			Msg(fmt.Sprintf("%s oAuthNoParamsConfig.Authenticate()", DebugFuncCall))
		_ = oAuthNoParamsConfig.CommandAuthConfig.WithCommandHostName(hostname).
			WithCommandAPIPath(apiPath).
			WithSkipVerify(skipVerifyBool)
		oErr := oAuthNoParamsConfig.
			WithClientId(clientId).
			WithClientSecret(clientSecret).
			WithTokenUrl(tokenUrl).
			WithAccessToken(accessToken).
			Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s oAuthNoParamsConfig.Authenticate()", DebugFuncReturn))
		if oErr != nil {
			log.Error().Err(oErr).Msg("unable to authenticate with provided credentials")
			log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncExit))
			return nil, oErr
		}

		log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncExit))
		return oAuthNoParamsConfig.GetServerConfig(), nil
	}

	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncExit))
	return nil, fmt.Errorf("incomplete environment variable configuration")
}

func overrideServerConfigFile(conf *auth_providers.Server) *auth_providers.Server {
	log.Debug().Msg(fmt.Sprintf("%s overrideServerConfigFile()", DebugFuncEnter))
	if flagHostName != "" {
		log.Trace().
			Str("flagHostName", flagHostName).
			Str("conf.Host", conf.Host).
			Msg("overriding hostname")
		conf.Host = flagHostName
	}
	if flagAPIPath != "" {
		log.Trace().
			Str("flagAPIPath", flagAPIPath).
			Str("conf.APIPath", conf.APIPath).
			Msg("overriding API path")
		conf.APIPath = flagAPIPath
	}
	if flagSkipVerify {
		log.Trace().
			Bool("flagSkipVerify", flagSkipVerify).
			Bool("conf.SkipTLSVerify", conf.SkipTLSVerify).
			Msg("overriding skip verify")
		conf.SkipTLSVerify = true
	}
	if flagUsername != "" {
		log.Trace().
			Str("flagUsername", flagUsername).
			Str("conf.Username", conf.Username).
			Msg("overriding username")
		conf.Username = flagUsername
	}
	if flagPassword != "" {
		log.Trace().
			Str("flagPassword", hashSecretValue(flagPassword)).
			Str("conf.Password", hashSecretValue(conf.Password)).
			Msg("overriding password")
		conf.Password = flagPassword
	}
	if flagDomain != "" {
		log.Trace().
			Str("flagDomain", flagDomain).
			Str("conf.Domain", conf.Domain).
			Msg("overriding domain")
		conf.Domain = flagDomain
	}
	if flagClientId != "" {
		log.Trace().
			Str("flagClientId", flagClientId).
			Str("conf.ClientID", conf.ClientID).
			Msg("overriding client ID")
		conf.ClientID = flagClientId
	}
	if flagClientSecret != "" {
		log.Trace().
			Str("flagClientSecret", hashSecretValue(flagClientSecret)).
			Str("conf.ClientSecret", hashSecretValue(conf.ClientSecret)).
			Msg("overriding client secret")
		conf.ClientSecret = flagClientSecret
	}
	if flagTokenUrl != "" {
		log.Trace().
			Str("flagTokenUrl", flagTokenUrl).
			Str("conf.OAuthTokenUrl", conf.OAuthTokenUrl).
			Msg("overriding token URL")
		conf.OAuthTokenUrl = flagTokenUrl
	}
	if flagAccessToken != "" {
		log.Trace().
			Str("flagAccessToken", flagAccessToken).
			Str("conf.AccessToken", conf.AccessToken).
			Msg("overriding access token")
		conf.AccessToken = flagAccessToken
	}
	if flagProviderType != "" {
		log.Trace().
			Str("flagProviderType", flagProviderType).
			Str("conf.AuthProvider.Type", conf.AuthProvider.Type).
			Msg("overriding provider type")
		conf.AuthProvider.Type = flagProviderType
	}
	log.Debug().Msg(fmt.Sprintf("%s overrideServerConfigFile()", DebugFuncExit))
	return conf
}

// authViaConfigFile authenticates using the configuration file
func authViaConfigFile(cfgFile string, cfgProfile string) (*api.Client, error) {
	log.Debug().
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncEnter))

	var (
		c    *api.Client
		cErr error
	)

	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncCall))
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncReturn))
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
				log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncCall))
				return authViaProvider(cfgFile, cfgProfile, conf.AuthProvider.Profile)
			}
		}
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncCall))
		conf = overrideServerConfigFile(conf)
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncReturn))

		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncCall))
		c, cErr = api.NewKeyfactorClient(conf, nil)
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncReturn))

		if cErr != nil {
			log.Error().
				Err(cErr).
				Msg("unable to create Keyfactor client")
			log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
			return nil, cErr
		}
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncCall))
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncReturn))
		if authErr == nil {
			log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
			return c, nil
		}

	}
	log.Error().
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Msg("unable to authenticate via config file")
	log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
	return nil, fmt.Errorf("unable to authenticate via config file '%s' using profile '%s'", cfgFile, cfgProfile)
}

// authSdkViaConfigFile authenticates using the configuration file
func authSdkViaConfigFile(cfgFile string, cfgProfile string) (*keyfactor.APIClient, error) {
	log.Debug().
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncEnter))

	var (
		c    *keyfactor.APIClient
		cErr error
	)

	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncCall))
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncReturn))
	if err != nil {
		log.Error().
			Err(err).
			Msg("unable to get server config from file")
		log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
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
					Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncCall))
				return authSdkViaProvider(cfgFile, cfgProfile, conf.AuthProvider.Profile)
			}
		}
		log.Debug().Msg(fmt.Sprintf("%s keyfactor.NewAPIClient()", DebugFuncCall))
		c, cErr = keyfactor.NewAPIClient(conf)
		log.Debug().Msg(fmt.Sprintf("%s keyfactor.NewAPIClient()", DebugFuncReturn))
		if cErr != nil {
			log.Error().
				Err(cErr).
				Msg("unable to create Keyfactor client")
			log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
			return nil, cErr
		}
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncCall))
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncReturn))
		if authErr == nil {
			log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
			return c, nil
		}
	}
	log.Error().Msg("unable to authenticate via config file")
	log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncExit))
	return nil, fmt.Errorf("unable to authenticate via config file '%s' using flagProfile '%s'", cfgFile, cfgProfile)
}

// authViaEnvVars authenticates using the environment variables
func authViaEnvVars() (*api.Client, error) {
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncEnter))
	var (
		c    *api.Client
		cErr error
	)
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncCall))
	conf, err := getServerConfigFromEnv()
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncReturn))
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via environment variables")
		log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
		return nil, err
	}
	if conf != nil {
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncCall))
		c, cErr = api.NewKeyfactorClient(conf, nil)
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncReturn))
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
			return nil, cErr
		}
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncCall))
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncReturn))
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via environment variables")
			log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
			return nil, authErr
		}
		log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
		return c, nil
	}
	log.Error().Msg("unable to authenticate via environment variables")
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
	return nil, fmt.Errorf("unable to authenticate via environment variables")
}

// authSdkViaEnvVars authenticates using the environment variables
func authSdkViaEnvVars() (*keyfactor.APIClient, error) {
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncEnter))

	var (
		c    *keyfactor.APIClient
		cErr error
	)

	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncCall))
	conf, err := getServerConfigFromEnv()
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromEnv()", DebugFuncReturn))
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via environment variables")
		log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
		return nil, err
	}
	if conf != nil {
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncCall))
		c, cErr = keyfactor.NewAPIClient(conf)
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncReturn))
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
			return nil, cErr
		}
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncCall))
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncReturn))
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via environment variables")
			log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
			return nil, authErr
		}
		log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
		return c, nil
	}
	log.Error().Msg("unable to authenticate via environment variables")
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncExit))
	return nil, fmt.Errorf("unable to authenticate via environment variables")
}

// authViaProvider authenticates using the provider
func authViaProvider(cfgFile string, cfgProfile string, providerProfile string) (*api.Client, error) {
	log.Debug().
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Str("providerProfile", providerProfile).
		Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncEnter))

	var (
		c            *api.Client
		cErr         error
		providerType string
	)

	log.Debug().
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Str("providerProfile", providerProfile).
		Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncCall))
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncReturn))
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via provider")
		log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
		return nil, err
	}

	if providerProfile == "" {
		if flagProviderProfile != "" {
			log.Debug().
				Str("flagProviderProfile", flagProviderProfile).
				Msg("using flagProviderProfile")
			providerProfile = flagProviderProfile
		} else {
			log.Debug().Msg("using default provider profile")
			providerProfile = auth_providers.DefaultConfigProfile
		}
	}

	providerType = flagProviderType
	if providerType == "" {
		log.Debug().Msg("provider type flag not set, attempting to use config file value")
		providerType = conf.AuthProvider.Type
	}

	if providerType == "azid" || providerType == "azure" {
		log.Debug().Msg("azure auth provider selected")
		azConfig := &auth_providers.ConfigProviderAzureKeyVault{}
		secretName, sOk := os.LookupEnv(auth_providers.EnvAzureSecretName)
		log.Debug().Str("secretName", secretName).Bool("sOk", sOk).Msg("envSecretName")
		vaultName, vOk := os.LookupEnv(auth_providers.EnvAzureVaultName)
		log.Debug().Str("vaultName", vaultName).Bool("vOk", vOk).Msg("envVaultName")
		if !sOk {
			log.Debug().Msg("secret name not found in environment attempting to use config file values")
			secretName, sOk = conf.AuthProvider.Parameters["secret_name"].(string)
			log.Debug().Str("secretName", secretName).Bool("sOk", sOk).Msg("configSecretName")
		}
		if !vOk {
			log.Debug().Msg("vault name not found in environment attempting to use config file values")
			vaultName, vOk = conf.AuthProvider.Parameters["vault_name"].(string)
			log.Debug().Str("vaultName", vaultName).Bool("vOk", vOk).Msg("configVaultName")
		}

		log.Debug().Msg(fmt.Sprintf("%s azConfig.Authenticate()", DebugFuncCall))
		aErr := azConfig.
			WithSecretName(secretName).
			WithVaultName(vaultName).
			Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s azConfig.Authenticate()", DebugFuncReturn))
		if aErr != nil {
			log.Error().Err(aErr).Msg("unable to authenticate via provider")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, aErr
		}

		log.Debug().Msg(fmt.Sprintf("%s azConfig.LoadConfigFromAzureKeyVault()", DebugFuncCall))
		cfg, cfgErr := azConfig.LoadConfigFromAzureKeyVault()
		log.Debug().Msg(fmt.Sprintf("%s azConfig.LoadConfigFromAzureKeyVault()", DebugFuncReturn))
		if cfgErr != nil {
			log.Error().Err(cfgErr).Msg("unable to load config from Azure Key Vault")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, cfgErr
		}

		serverConfig, serOk := cfg.Servers[providerProfile]
		if !serOk {
			log.Error().Str("providerProfile", providerProfile).Msg("invalid provider profile")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, fmt.Errorf("invalid provider profile: %s", providerProfile)
		}

		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncCall))
		c, cErr = api.NewKeyfactorClient(&serverConfig, nil)
		log.Debug().Msg(fmt.Sprintf("%s api.NewKeyfactorClient()", DebugFuncReturn))
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, cErr
		}
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncCall))
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncReturn))
		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via provider")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, authErr
		}
		log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
		return c, nil
	}
	log.Error().Str("providerType", providerType).Msg("unsupported provider type")
	log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
	return nil, fmt.Errorf("unsupported provider type: %s", providerType)
}

// authSdkViaProvider authenticates using the provider
func authSdkViaProvider(cfgFile string, cfgProfile string, providerProfile string) (*keyfactor.APIClient, error) {
	log.Debug().
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Str("providerProfile", providerProfile).
		Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncEnter))

	var (
		c            *keyfactor.APIClient
		cErr         error
		providerType string
	)

	log.Debug().
		Str("cfgFile", cfgFile).
		Str("cfgProfile", cfgProfile).
		Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncCall))
	conf, err := getServerConfigFromFile(cfgFile, cfgProfile)
	log.Debug().Msg(fmt.Sprintf("%s getServerConfigFromFile()", DebugFuncReturn))

	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate via provider")
		log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
		return nil, err
	}

	providerType = flagProviderType
	if providerType == "" {
		log.Debug().Msg("provider type flag not set, attempting to use config file value")
		providerType = conf.AuthProvider.Type
	}

	if providerProfile == "" {
		if flagProviderProfile != "" {
			log.Debug().
				Str("flagProviderProfile", flagProviderProfile).
				Msg("using flagProviderProfile")
			providerProfile = flagProviderProfile
		} else {
			log.Debug().Msg("using default provider profile")
			providerProfile = auth_providers.DefaultConfigProfile
		}
	}

	if providerType == "azid" || providerType == "azure" {
		log.Debug().Msg("azure auth provider selected")
		azConfig := &auth_providers.ConfigProviderAzureKeyVault{}
		secretName, sOk := os.LookupEnv(auth_providers.EnvAzureSecretName)
		log.Trace().Str("secretName", secretName).Bool("sOk", sOk).Msg("envSecretName")
		vaultName, vOk := os.LookupEnv(auth_providers.EnvAzureVaultName)
		log.Trace().Str("vaultName", vaultName).Bool("vOk", vOk).Msg("envVaultName")
		if !sOk {
			log.Debug().Msg("secret name not found in environment attempting to use config file values")
			secretName, sOk = conf.AuthProvider.Parameters["secret_name"].(string)
			log.Trace().Str("secretName", secretName).Bool("sOk", sOk).Msg("configSecretName")
		}
		if !vOk {
			log.Debug().Msg("vault name not found in environment attempting to use config file values")
			vaultName, vOk = conf.AuthProvider.Parameters["vault_name"].(string)
			log.Trace().Str("vaultName", vaultName).Bool("vOk", vOk).Msg("configVaultName")
		}

		log.Debug().Msg(fmt.Sprintf("%s azConfig.Authenticate()", DebugFuncCall))
		aErr := azConfig.
			WithSecretName(secretName).
			WithVaultName(vaultName).
			Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s azConfig.Authenticate()", DebugFuncReturn))
		if aErr != nil {
			log.Error().Err(aErr).Msg("unable to authenticate via provider")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, aErr
		}

		log.Debug().Msg(fmt.Sprintf("%s azConfig.LoadConfigFromAzureKeyVault()", DebugFuncCall))
		cfg, cfgErr := azConfig.LoadConfigFromAzureKeyVault()
		log.Debug().Msg(fmt.Sprintf("%s azConfig.LoadConfigFromAzureKeyVault()", DebugFuncReturn))

		if cfgErr != nil {
			log.Error().Err(cfgErr).Msg("unable to load config from Azure Key Vault")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, cfgErr
		}

		serverConfig, serOk := cfg.Servers[providerProfile]
		if !serOk {
			log.Error().Str("providerProfile", providerProfile).Msg("invalid flagProfile")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, fmt.Errorf("invalid providerProfile: %s", providerProfile)
		}
		log.Debug().Msg(fmt.Sprintf("%s keyfactor.NewAPIClient()", DebugFuncCall))
		c, cErr = keyfactor.NewAPIClient(&serverConfig)
		log.Debug().Msg(fmt.Sprintf("%s keyfactor.NewAPIClient()", DebugFuncReturn))
		if cErr != nil {
			log.Error().Err(cErr).Msg("unable to create Keyfactor client")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, cErr
		}
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncCall))
		authErr := c.AuthClient.Authenticate()
		log.Debug().Msg(fmt.Sprintf("%s c.AuthClient.Authenticate()", DebugFuncReturn))

		if authErr != nil {
			log.Error().Err(authErr).Msg("unable to authenticate via provider")
			log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
			return nil, authErr
		}
		log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
		return c, nil
	}
	log.Error().Str("providerType", providerType).Msg("unsupported provider type")
	log.Debug().Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncExit))
	return nil, fmt.Errorf("unsupported provider type: %s", providerType)
}

// initClient initializes the legacy Command API client
func initClient(saveConfig bool) (*api.Client, error) {
	log.Debug().
		Str("flagConfigFile", flagConfigFile).
		Str("flagProfile", flagProfile).
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Bool("flagNoPrompt", flagNoPrompt).
		Bool("saveConfig", saveConfig).
		Str("hostname", flagHostName).
		Str("username", flagUsername).
		Str("password", hashSecretValue(flagPassword)).
		Str("domain", flagDomain).
		Str("clientId", flagClientId).
		Str("clientSecret", hashSecretValue(flagClientSecret)).
		Str("tokenUrl", flagTokenUrl).
		Str("accessToken", flagAccessToken).
		Str("apiPath", flagAPIPath).
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Msg(fmt.Sprintf("%s initClient()", DebugFuncEnter))
	var (
		c         *api.Client
		envCfgErr error
		cfgErr    error
	)

	if flagProviderType != "" {
		log.Debug().
			Str("flagProviderType", flagProviderType).
			Msg(fmt.Sprintf("%s authViaProvider()", DebugFuncCall))
		return authViaProvider("", "", "")
	}

	if flagConfigFile != "" || flagProfile != "" {
		log.Info().
			Str("flagConfigFile", flagConfigFile).
			Str("flagProfile", flagProfile).
			Msg("authenticating via config file")
		log.Debug().
			Str("flagConfigFile", flagConfigFile).
			Str("flagProfile", flagProfile).
			Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncCall))
		c, cfgErr = authViaConfigFile(flagConfigFile, flagProfile)
		log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncReturn))

		if cfgErr == nil {
			log.Info().
				Str("flagConfigFile", flagConfigFile).
				Str("flagProfile", flagProfile).
				Msgf("Authenticated via config file %s using flagProfile %s", flagConfigFile, flagProfile)
			log.Debug().Msg(fmt.Sprintf("%s initClient()", DebugFuncExit))
			return c, nil
		}
	}

	log.Info().Msg("authenticating via environment variables")
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncCall))
	c, envCfgErr = authViaEnvVars()
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncReturn))
	if envCfgErr == nil {
		log.Info().Msg("Authenticated via environment variables")
		log.Debug().Msg(fmt.Sprintf("%s initClient()", DebugFuncExit))
		return c, nil
	}

	log.Info().
		Str("flagConfigFile", DefaultConfigFileName).
		Str("flagProfile", "default").
		Msg("implicit authenticating via config file using default flagProfile")
	log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncCall))
	c, cfgErr = authViaConfigFile("", "")
	log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncReturn))

	if cfgErr == nil {
		log.Info().
			Str("flagConfigFile", DefaultConfigFileName).
			Str("flagProfile", "default").
			Msgf("authenticated implictly via config file '%s' using 'default' flagProfile", DefaultConfigFileName)
		log.Debug().Msg(fmt.Sprintf("%s initClient()", DebugFuncExit))
		return c, nil
	}

	log.Error().
		Err(cfgErr).
		Err(envCfgErr).
		Msg("unable to authenticate to Keyfactor Command")
	log.Debug().Msg(fmt.Sprintf("%s initClient()", DebugFuncExit))
	return nil, fmt.Errorf("unable to authenticate to Keyfactor Command")
}

// initGenClient initializes the SDK Command API client
func initGenClient(
	saveConfig bool,
) (*keyfactor.APIClient, error) {
	log.Debug().
		Str("flagConfigFile", flagConfigFile).
		Str("flagProfile", flagProfile).
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Bool("flagNoPrompt", flagNoPrompt).
		Bool("saveConfig", saveConfig).
		Str("hostname", flagHostName).
		Str("username", flagUsername).
		Str("password", hashSecretValue(flagPassword)).
		Str("domain", flagDomain).
		Str("clientId", flagClientId).
		Str("clientSecret", hashSecretValue(flagClientSecret)).
		Str("tokenUrl", flagTokenUrl).
		Str("accessToken", flagAccessToken).
		Str("apiPath", flagAPIPath).
		Str("flagProviderType", flagProviderType).
		Str("flagProviderProfile", flagProviderProfile).
		Msg(fmt.Sprintf("%s initGenClient()", DebugFuncEnter))

	var (
		c       *keyfactor.APIClient
		envCErr error
		cfErr   error
	)

	if flagProviderType != "" {
		log.Debug().
			Str("flagProviderType", flagProviderType).
			Msg(fmt.Sprintf("%s authSdkViaProvider()", DebugFuncCall))
		return authSdkViaProvider("", "", "")
	}

	if flagConfigFile != "" || flagProfile != "" {
		log.Info().
			Str("flagConfigFile", flagConfigFile).
			Str("flagProfile", flagProfile).
			Msg("authenticating via config file")
		log.Debug().
			Str("flagConfigFile", flagConfigFile).
			Str("flagProfile", flagProfile).
			Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncCall))
		c, cfErr = authSdkViaConfigFile(flagConfigFile, flagProfile)
		log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncReturn))

		if cfErr == nil {
			log.Info().
				Str("flagConfigFile", flagConfigFile).
				Str("flagProfile", flagProfile).
				Msgf("Authenticated via config file %s using flagProfile %s", flagConfigFile, flagProfile)
			log.Debug().Msg(fmt.Sprintf("%s initGenClient()", DebugFuncExit))
			return c, nil
		}
	}

	log.Info().Msg("authenticating via environment variables")
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncCall))
	c, envCErr = authSdkViaEnvVars()
	log.Debug().Msg(fmt.Sprintf("%s authViaEnvVars()", DebugFuncReturn))

	if envCErr == nil {
		log.Info().Msg("authenticated via environment variables")
		log.Debug().Msg(fmt.Sprintf("%s initGenClient()", DebugFuncExit))
		return c, nil
	}

	log.Info().
		Str("flagConfigFile", DefaultConfigFileName).
		Str("flagProfile", "default").
		Msg("implicit authenticating via config file using default flagProfile")
	log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncCall))
	c, cfErr = authSdkViaConfigFile("", "")
	log.Debug().Msg(fmt.Sprintf("%s authViaConfigFile()", DebugFuncReturn))

	if cfErr == nil {
		log.Info().
			Str("flagConfigFile", DefaultConfigFileName).
			Str("flagProfile", "default").
			Msgf("authenticated implictly via config file '%s' using 'default' flagProfile", DefaultConfigFileName)
		log.Debug().Msg(fmt.Sprintf("%s initGenClient()", DebugFuncExit))
		return c, nil
	}

	log.Error().
		Err(cfErr).
		Err(envCErr).
		Msg("unable to authenticate")
	log.Debug().Msg(fmt.Sprintf("%s initGenClient()", DebugFuncExit))
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
		log.Debug().Msg("returned: makeDocsCmd.Run()")
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
		&flagConfigFile,
		"config",
		"",
		"",
		fmt.Sprintf("Full path to config file in JSON format. (default is %s)", defaultConfigPath),
	)
	RootCmd.PersistentFlags().BoolVar(
		&flagNoPrompt,
		"no-prompt",
		false,
		"Do not prompt for any user input and assume defaults or environmental variables are set.",
	)
	RootCmd.PersistentFlags().BoolVar(
		&flagEnableExp,
		"exp",
		false,
		"Enable flagEnableExp features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)",
	)
	RootCmd.PersistentFlags().BoolVar(
		&flagSkipVerify,
		"skip-verify",
		false,
		"Skip SSL verification when communicating to Keyfactor Command. (USE AT YOUR OWN RISK, this will disable SSL verification.)",
	)

	RootCmd.PersistentFlags().BoolVar(
		&flagOffline,
		"flagOffline",
		false,
		"Will not attempt to connect to GitHub for latest release information and resources.",
	)
	RootCmd.PersistentFlags().BoolVar(&flagEnableDebug, "debug", false, "Enable flagEnableDebug logging.")
	//RootCmd.PersistentFlags().BoolVar(
	//	&flagLogInsecure,
	//	"log-insecure",
	//	false,
	//	"Log insecure API requests. (USE AT YOUR OWN RISK, this WILL log sensitive information to the console.)",
	//)
	RootCmd.PersistentFlags().StringVarP(
		&flagProfile,
		"flagProfile",
		"",
		"",
		"Use a specific flagProfile from your config file. If not specified the config named 'default' will be used if it exists.",
	)
	RootCmd.PersistentFlags().StringVar(
		&flagOutputFormat,
		"format",
		"text",
		"How to format the CLI output. Currently only `text` is supported.",
	)

	RootCmd.PersistentFlags().StringVar(&flagProviderType, "auth-provider-type", "", "Provider type choices: (azid)")
	// Validating the provider-type flag against the predefined choices
	RootCmd.PersistentFlags().SetAnnotation("auth-provider-type", cobra.BashCompCustom, ProviderTypeChoices)
	RootCmd.PersistentFlags().StringVarP(
		&flagProviderProfile,
		"auth-provider-flagProfile",
		"",
		"default",
		"The flagProfile to use defined in the securely stored config. If not specified the config named 'default' will be used if it exists.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&flagUsername,
		"username",
		"",
		"",
		"Username to use for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&flagClientId,
		"client-id",
		"",
		"",
		"OAuth2 client-id to use for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&flagClientSecret,
		"client-secret",
		"",
		"",
		"OAuth2 client-secret to use for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&flagClientId,
		"token-url",
		"",
		"",
		"OAuth2 token endpoint full URL to use for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&flagAccessToken,
		"token",
		"",
		"",
		"OAuth2 access token for authenticating to Keyfactor Command.",
	)

	RootCmd.PersistentFlags().StringVarP(
		&flagHostName,
		"hostname",
		"",
		"",
		"Hostname to use for authenticating to Keyfactor Command.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&flagPassword,
		"password",
		"",
		"",
		"Password to use for authenticating to Keyfactor Command. WARNING: Remember to delete your console history if providing flagPassword here in plain text.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&flagDomain,
		"domain",
		"",
		"",
		"Domain to use for authenticating to Keyfactor Command.",
	)
	RootCmd.PersistentFlags().StringVarP(
		&flagAPIPath,
		"api-path",
		"",
		"KeyfactorAPI",
		"API Path to use for authenticating to Keyfactor Command. (default is KeyfactorAPI)",
	)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	RootCmd.AddCommand(makeDocsCmd)
}
