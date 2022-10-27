/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func initClient() (*api.Client, error) {
	var clientAuth api.AuthConfig
	clientAuth.Username = os.Getenv("KEYFACTOR_USERNAME")
	log.Printf("[DEBUG] Username: %s", clientAuth.Username)
	clientAuth.Password = os.Getenv("KEYFACTOR_PASSWORD")
	log.Printf("[DEBUG] Password: %s", clientAuth.Password)
	clientAuth.Domain = os.Getenv("KEYFACTOR_DOMAIN")
	log.Printf("[DEBUG] Domain: %s", clientAuth.Domain)
	clientAuth.Hostname = os.Getenv("KEYFACTOR_HOSTNAME")
	log.Printf("[DEBUG] Hostname: %s", clientAuth.Hostname)

	c, err := api.NewKeyfactorClient(&clientAuth)

	if err != nil {
		log.Fatalf("Error creating Keyfactor client: %s", err)
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
