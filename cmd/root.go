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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	stdlog "log"
	"os"
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
