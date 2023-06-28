// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

var colorRed = "\033[31m"
var colorWhite = "\033[37m"

var xKeyfactorRequestedWith = "APIClient"
var xKeyfactorApiVersion = "1"

func initClient(flagConfig string, flagProfile string, noPrompt bool) (*api.Client, error) {
	var clientAuth api.AuthConfig

	//Check if env vars are set if they are then ignore config file
	if os.Getenv("KEYFACTOR_HOSTNAME") != "" && os.Getenv("KEYFACTOR_USERNAME") != "" && os.Getenv("KEYFACTOR_PASSWORD") != "" {
		if os.Getenv("KEYFACTOR_DOMAIN") != "" {
			clientAuth.Domain = os.Getenv("KEYFACTOR_DOMAIN")
		} else {

		}
		clientAuth.Hostname = os.Getenv("KEYFACTOR_HOSTNAME")
		clientAuth.Username = os.Getenv("KEYFACTOR_USERNAME")
		clientAuth.Password = os.Getenv("KEYFACTOR_PASSWORD")
	}

	//Else check if config file exists

	commandConfig, _ := authConfigFile(flagConfig, noPrompt, flagProfile)

	if flagProfile == "" {
		flagProfile = "default"
	}
	clientAuth.Username = commandConfig.Servers[flagProfile].Username
	clientAuth.Password = commandConfig.Servers[flagProfile].Password
	clientAuth.Domain = commandConfig.Servers[flagProfile].Domain
	clientAuth.Hostname = commandConfig.Servers[flagProfile].Hostname
	clientAuth.APIPath = commandConfig.Servers[flagProfile].APIPath

	c, err := api.NewKeyfactorClient(&clientAuth)

	if err != nil {
		fmt.Printf("Error connecting to Keyfactor: %s\n", err)
		log.Fatalf("[ERROR] creating Keyfactor client: %s", err)
	}

	return c, nil
}

func initGenClient(profile string) *keyfactor.APIClient {
	configs, authErr := authConfigFile("", true, profile)
	if profile == "" {
		profile = "default"
	}
	cmdConfig := configs.Servers[profile]

	if authErr != nil {
		fmt.Printf("Error reading config file: %s\n", authErr)
		log.Fatalf("[ERROR] reading config file: %s", authErr)
	}

	sdkClientConfig := make(map[string]string)
	sdkClientConfig["host"] = cmdConfig.Hostname
	sdkClientConfig["username"] = cmdConfig.Username
	sdkClientConfig["password"] = cmdConfig.Password
	sdkClientConfig["domain"] = cmdConfig.Domain

	configuration := keyfactor.NewConfiguration(sdkClientConfig)
	c := keyfactor.NewAPIClient(configuration)
	return c
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
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	var (
		configFile   string
		profile      string
		noPrompt     bool
		experimental bool
		debug        bool
	)

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "", "", fmt.Sprintf("Full path to config file in JSON format. (default is $HOME/.keyfactor/%s)", DefaultConfigFileName))
	RootCmd.PersistentFlags().BoolVar(&noPrompt, "no-prompt", false, "Do not prompt for any user input and assume defaults or environmental variables are set.")
	RootCmd.PersistentFlags().BoolVar(&experimental, "exp", false, "Enable experimental features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging. (USE AT YOUR OWN RISK, this may log sensitive information to the console.)")
	RootCmd.PersistentFlags().StringVarP(&profile, "profile", "", "", "Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.")

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
		log.SetOutput(os.Stdout)
		return envValue
	case v:
		log.SetOutput(os.Stdout)
		return v
	default:
		log.SetOutput(io.Discard)
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
		return false, fmt.Errorf("experimental features are not enabled. To enable experimental features, use the --exp flag or set the KFUTIL_EXP environment variable to true")
	}
	return envValue, nil
}
