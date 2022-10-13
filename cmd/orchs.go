/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

// orchsCmd represents the orchs command
var orchsCmd = &cobra.Command{
	Use:   "orchs",
	Short: "Keyfactor agents APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor orchestrators.`,
}

var getOrchestratorCmd = &cobra.Command{
	Use:   "get",
	Short: "Get orchestrator by ID or machine/host name.",
	Long:  `Get orchestrator by ID or machine/host name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchestrator get called")
	},
}

var approveOrchestratorCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve orchestrator by ID or machine/host name.",
	Long:  `Approve orchestrator by ID or machine/host name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchestrator approve called")
	},
}

var disapproveOrchestratorCmd = &cobra.Command{
	Use:   "disapprove",
	Short: "Disapprove orchestrator by ID or machine/host name.",
	Long:  `Disapprove orchestrator by ID or machine/host name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchestrator disapprove called")
	},
}

var resetOrchestratorCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset orchestrator by ID or machine/host name.",
	Long:  `Reset orchestrator by ID or machine/host name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchestrator reset called")
	},
}

var getLogsOrchestratorCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get orchestrator logs by ID or machine/host name.",
	Long:  `Get orchestrator logs by ID or machine/host name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchestrator logs called")
	},
}

var listOrchestratorsCmd = &cobra.Command{
	Use:   "list",
	Short: "List orchestrators.",
	Long:  `Returns a JSON list of Keyfactor orchestrators.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(ioutil.Discard)
		kfClient, _ := initClient()
		agents, err := kfClient.GetAgentList()
		if err != nil {
			log.Printf("Error: %s", err)
		}
		output, jErr := json.Marshal(agents)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

func init() {
	rootCmd.AddCommand(orchsCmd)

	// LIST orchestrators command
	orchsCmd.AddCommand(listOrchestratorsCmd)
	// GET orchestrator command
	orchsCmd.AddCommand(getOrchestratorCmd)
	// CREATE orchestrator command TODO: API NOT SUPPORTED
	//orchsCmd.AddCommand(createOrchestratorCmd)
	// UPDATE orchestrator command TODO: API NOT SUPPORTED
	//orchsCmd.AddCommand(updateOrchestratorCmd)
	// DELETE orchestrator command TODO: API NOT SUPPORTED
	//orchsCmd.AddCommand(deleteOrchestratorCmd)
	// APPROVE orchestrator command
	orchsCmd.AddCommand(approveOrchestratorCmd)
	// DISAPPROVE orchestrator command
	orchsCmd.AddCommand(disapproveOrchestratorCmd)
	// RESET orchestrator command
	orchsCmd.AddCommand(resetOrchestratorCmd)
	// GET orchestrator logs command
	orchsCmd.AddCommand(getLogsOrchestratorCmd)
	// SET orchestrator auth certificate reenrollment command TODO: Not implemented
	//orchsCmd.AddCommand(setOrchestratorAuthCertReenrollCmd)
	// Utility commands
	//orchsCmd.AddCommand(downloadOrchestrator) TODO: Not implemented
}
