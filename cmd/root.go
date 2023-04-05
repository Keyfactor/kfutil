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
	"github.com/Keyfactor/keyfactor-go-client/api"
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

func initClient(p string) (*api.Client, error) {
	log.SetOutput(io.Discard)
	var clientAuth api.AuthConfig
	clientAuth.Username = os.Getenv("KEYFACTOR_USERNAME")
	log.Printf("[DEBUG] Username: %s", clientAuth.Username)
	clientAuth.Password = os.Getenv("KEYFACTOR_PASSWORD")
	log.Printf("[DEBUG] Password: %s", clientAuth.Password)
	clientAuth.Domain = os.Getenv("KEYFACTOR_DOMAIN")
	log.Printf("[DEBUG] Domain: %s", clientAuth.Domain)
	clientAuth.Hostname = os.Getenv("KEYFACTOR_HOSTNAME")
	log.Printf("[DEBUG] Hostname: %s", clientAuth.Hostname)
	var profile string
	if p != "" {
		profile = p
	} else {
		profile = os.Getenv("KFUTIL_PROFILE")
	}
	if profile == "" {
		profile = "default"
	}

	if clientAuth.Username == "" || clientAuth.Password == "" || clientAuth.Hostname == "" {
		authConfigFile("", true, profile)
		clientAuth.Username = os.Getenv("KEYFACTOR_USERNAME")
		log.Printf("[DEBUG] Username: %s", clientAuth.Username)
		clientAuth.Password = os.Getenv("KEYFACTOR_PASSWORD")
		log.Printf("[DEBUG] Password: %s", clientAuth.Password)
		clientAuth.Domain = os.Getenv("KEYFACTOR_DOMAIN")
		log.Printf("[DEBUG] Domain: %s", clientAuth.Domain)
		clientAuth.Hostname = os.Getenv("KEYFACTOR_HOSTNAME")
		log.Printf("[DEBUG] Hostname: %s", clientAuth.Hostname)
	}
	c, err := api.NewKeyfactorClient(&clientAuth)

	if err != nil {
		fmt.Printf("Error connecting to Keyfactor: %s\n", err)
		log.Fatalf("[ERROR] creating Keyfactor client: %s", err)
	}

	return c, nil
}

func initGenClient() *keyfactor.APIClient {
	configuration := keyfactor.NewConfiguration()
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

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", fmt.Sprintf("Full path to config file in JSON format. (default is $HOME/.keyfactor/%s)", DefaultConfigFileName))
	RootCmd.PersistentFlags().BoolVar(&noPrompt, "no-prompt", false, "Do not prompt for any user input and assume defaults or environmental variables are set.")
	RootCmd.PersistentFlags().BoolVar(&experimental, "exp", false, "Enable experimental features. (USE AT YOUR OWN RISK, these features are not supported and may change or be removed at any time.)")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging. (USE AT YOUR OWN RISK, this may log sensitive information to the console.)")
	RootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "Use a specific profile from your config file. If not specified the config named 'default' will be used if it exists.")

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
	if (envValue && !v) || (envValue && v) {
		// If the env var is set and the flag is not, use the env var
		log.SetOutput(os.Stdout)
		return envValue
	} else if v {
		log.SetOutput(os.Stdout)
		return v
	}
	log.SetOutput(io.Discard)
	return v
}

func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
