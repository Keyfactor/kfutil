package cmd

import (
	"github.com/spf13/cobra"
	"kfutil/pkg/helm"
	"log"
)

// helmCmd represents the storeTypes command
var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Keyfactor Helm Chart Utilities",
	Long:  `Keyfactor Helm Chart Utilities used to configure charts and assist in the deployment of Keyfactor products.`,
}

var helmUoCmd = &cobra.Command{
	Use:   "uo",
	Short: "Keyfactor Helm Chart Utilities for the Containerized Universal Orchestrator",
	Long:  `Keyfactor Helm Chart Utilities used to configure charts and assist in the deployment of the Keyfactor Command Universal Orchestrator.`,
	Run:   UoValueBuilder,
}

func UoValueBuilder(cmd *cobra.Command, args []string) {
	// Global flags
	debugFlag, _ := cmd.Flags().GetBool("debug")
	noPrompt, _ := cmd.Flags().GetBool("no-prompt")
	profile, _ := cmd.Flags().GetString("profile")
	debugModeEnabled := checkDebug(debugFlag)
	log.Println("Debug mode enabled: ", debugModeEnabled)
	commandConfig, _ := authConfigFile("", profile, noPrompt, false)

	valueFile, _ := cmd.Flags().GetString("values")
	githubToken, _ := cmd.Flags().GetString("token")
	outputFile, _ := cmd.Flags().GetString("out")

	uo := helm.NewUniversalOrchestratorHelmValueBuilder(valueFile)
	if githubToken != "" {
		uo.SetGithubToken(githubToken)
	}
	if outputFile != "" {
		uo.SetOverrideFile(outputFile)
	}
	if commandConfig.Servers["profile"].Hostname != "" {
		// TODO this could panic if authConfigFile fails
		uo.SetHostname(commandConfig.Servers["profile"].Hostname)
	}

	uo.Build()
}

func init() {
	RootCmd.AddCommand(helmCmd)

	// Helm UO (Universal Orchestrator) Command
	var valuesFile, githubToken, outputFile string
	helmCmd.AddCommand(helmUoCmd)
	helmUoCmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values.yaml file to use as base for the chart")
	_ = helmUoCmd.MarkFlagRequired("values")
	helmUoCmd.Flags().StringVarP(&outputFile, "out", "o", "", "Path to output the modified values.yaml file. This file can then be used with helm install -f <file> to override the default values.")
	helmUoCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub token to use for API calls")
}
