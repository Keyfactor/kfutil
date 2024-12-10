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
	"fmt"
	"io"
	stdlog "log"
	"os"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Removes the credentials file '$HOME/.keyfactor/command_config.json'.",
	Long:  `Removes the credentials file '$HOME/.keyfactor/command_config.json'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg("Running logout command")
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

		var configFilePath string
		if configFile == "" {
			// check if environment variables for config file is set
			if os.Getenv(auth_providers.EnvKeyfactorConfigFile) != "" {
				configFilePath = os.Getenv(auth_providers.EnvKeyfactorConfigFile)
			} else {
				userHomeDir, err := os.UserHomeDir()
				if err != nil {
					userHomeDir, err = os.Getwd()
					if err != nil {
						userHomeDir = "."
					}
				}
				configFilePath = fmt.Sprintf("%s/%s", userHomeDir, auth_providers.DefaultConfigFilePath)
			}
		} else {
			configFilePath = configFile
		}

		// Remove environment variables
		log.Info().Msg("Running logout command for environment variables")
		envLogout()

		log.Info().
			Str("configFilePath", configFilePath).
			Msg("Attempting to removing config file")
		err := os.Remove(configFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error().
					Err(err).
					Msg("Config file does not exist, unable to logout.")
				fmt.Println("Config file does not exist, unable to logout.")
				return err
			}
			log.Error().
				Err(err).
				Msg("unable to remove config file, logout failed")
			fmt.Println("Error removing config file: ", err)
			return err
		}
		log.Info().
			Str("configFilePath", configFilePath).
			Msg("Config file removed successfully")
		fmt.Println("Logged out successfully!")
		return nil
	},
}

func envLogout() {
	log.Debug().Msg("Running logout command for environment variables")

	log.Debug().Msg("Unsetting base environment variables")

	log.Trace().Str("EnvKeyfactorHostName", auth_providers.EnvKeyfactorHostName).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorHostName)

	log.Trace().Str("EnvKeyfactorPort", auth_providers.EnvKeyfactorPort).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorPort)

	log.Trace().Str("EnvKeyfactorAPIPath", auth_providers.EnvKeyfactorAPIPath).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAPIPath)

	log.Trace().Str("EnvAuthCACert", auth_providers.EnvAuthCACert).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvAuthCACert)

	log.Trace().Str("EnvKeyfactorSkipVerify", auth_providers.EnvKeyfactorSkipVerify).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorSkipVerify)

	log.Trace().Str("EnvKeyfactorClientTimeout", auth_providers.EnvKeyfactorClientTimeout).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorClientTimeout)

	log.Debug().Msg("Unsetting kfutil environment variables")
	log.Trace().Str("EnvKeyfactorAuthProfile", auth_providers.EnvKeyfactorAuthProfile).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAuthProfile)

	log.Trace().Str("EnvKeyfactorConfigFile", auth_providers.EnvKeyfactorConfigFile).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorConfigFile)

	log.Debug().Msg("Unsetting command basic auth environment variables")
	log.Trace().Str("EnvKeyfactorUsername", auth_providers.EnvKeyfactorUsername).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorUsername)

	log.Trace().Str("EnvKeyfactorPassword", auth_providers.EnvKeyfactorPassword).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorPassword)

	log.Trace().Str("EnvKeyfactorDomain", auth_providers.EnvKeyfactorDomain).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorDomain)

	log.Debug().Msg("Unsetting command oauth2 environment variables")
	log.Trace().Str("EnvKeyfactorClientID", auth_providers.EnvKeyfactorClientID).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorClientID)

	log.Trace().Str("EnvKeyfactorClientSecret", auth_providers.EnvKeyfactorClientSecret).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorClientSecret)

	log.Trace().Str("EnvKeyfactorAccessToken", auth_providers.EnvKeyfactorAccessToken).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAccessToken)

	log.Trace().Str("EnvKeyfactorAuthTokenURL", auth_providers.EnvKeyfactorAuthTokenURL).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAuthTokenURL)

	log.Trace().Str("EnvKeyfactorAuthScopes", auth_providers.EnvKeyfactorAuthScopes).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAuthScopes)

	log.Trace().Str("EnvKeyfactorAuthAudience", auth_providers.EnvKeyfactorAuthAudience).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAuthAudience)

	log.Debug().Msg("Unsetting command azure environment variables")
	log.Trace().Str("EnvKeyfactorAuthProvider", auth_providers.EnvKeyfactorAuthProvider).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvKeyfactorAuthProvider)

	log.Trace().Str("EnvAzureSecretName", auth_providers.EnvAzureSecretName).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvAzureSecretName)

	log.Trace().Str("EnvAzureVaultName", auth_providers.EnvAzureVaultName).Msg("Unsetting")
	os.Unsetenv(auth_providers.EnvAzureVaultName)

}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
