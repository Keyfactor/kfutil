// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"time"
)

func initClient() (*api.Client, error) {
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

	if clientAuth.Username == "" || clientAuth.Password == "" || clientAuth.Hostname == "" {
		authConfigFile("", true)
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
	return c, err
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kfutil.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
