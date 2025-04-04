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
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// orchsCmd represents the orchs command
var orchsCmd = &cobra.Command{
	Use:     "orchs",
	Short:   "Keyfactor agents/orchestrators APIs and utilities.",
	Long:    `A collections of APIs and utilities for interacting with Keyfactor orchestrators.`,
	Aliases: OrchsAliases,
}

// getOrchestratorCmd represents the get orchestrator command
var getOrchestratorCmd = &cobra.Command{
	Use:   "get",
	Short: "Get orchestrator by machine/client name.",
	Long:  `Get orchestrator by machine/client name.`,
	Run: func(cmd *cobra.Command, args []string) {
		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		client := cmd.Flag("client").Value.String()
		kfClient, _ := initClient(false)
		agents, aErr := kfClient.GetAgent(client)
		if aErr != nil {
			fmt.Printf("Error, unable to get orchestrator %s. %s\n", client, aErr)
			log.Fatalf("Error: %s", aErr)
		}
		output, jErr := json.Marshal(agents)
		if jErr != nil {
			fmt.Println("Error invalid API response from Keyfactor.")
			log.Fatalf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

// listOrchestratorsCmd represents the list orchestrators command
var approveOrchestratorCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve orchestrator by machine/client name.",
	Long:  `Approve orchestrator by machine/client name.`,
	Run: func(cmd *cobra.Command, args []string) {
		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		client := cmd.Flag("client").Value.String()
		kfClient, cErr := initClient(false)
		if cErr != nil {
			fmt.Println("Error, unable to connect to Keyfactor.")
			log.Fatalf("Error: %s", cErr)
		}
		agents, aErr := kfClient.GetAgent(client)
		if aErr != nil {
			fmt.Printf("Error, unable to get orchestrator %s. %s\n", client, aErr)
			log.Fatalf("[ERROR]: %s", aErr)
		}
		agent := agents[0]
		_, aErr = kfClient.ApproveAgent(agent.AgentId)
		if aErr != nil {
			fmt.Printf("Error, unable to approve orchestrator %s. %s\n", client, aErr)
			log.Fatalf("[ERROR]: %s", aErr)
		}
		fmt.Printf("Orchestrator %s approved.\n", client)
	},
}

// disapproveOrchestratorCmd represents the disapprove orchestrator command
var disapproveOrchestratorCmd = &cobra.Command{
	Use:   "disapprove",
	Short: "Disapprove orchestrator by machine/client name.",
	Long:  `Disapprove orchestrator by machine/client name.`,
	Run: func(cmd *cobra.Command, args []string) {

		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		client := cmd.Flag("client").Value.String()
		kfClient, cErr := initClient(false)
		if cErr != nil {
			fmt.Println("Error, unable to connect to Keyfactor.")
			log.Fatalf("Error: %s", cErr)
		}
		agents, aErr := kfClient.GetAgent(client)
		if aErr != nil {
			fmt.Printf("Error, unable to get orchestrator %s. %s\n", client, aErr)
			log.Fatalf("[ERROR]: %s", aErr)
		}
		agent := agents[0]
		_, aErr = kfClient.DisApproveAgent(agent.AgentId)
		if aErr != nil {
			fmt.Printf("Error, unable to disapprove orchestrator %s. %s\n", client, aErr)
			log.Fatalf("[ERROR]: %s", aErr)
		}
		fmt.Printf("Orchestrator %s disapproved.\n", client)
	},
}

// resetOrchestratorCmd represents the reset orchestrator command
var resetOrchestratorCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset orchestrator by machine/client name.",
	Long:  `Reset orchestrator by machine/client name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("orchestrator reset called")
	},
}

// getLogsOrchestratorCmd represents the get logs orchestrator command
var getLogsOrchestratorCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get orchestrator logs by machine/client name.",
	Long:  `Get orchestrator logs by machine/client name.`,
	Run: func(cmd *cobra.Command, args []string) {
		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)

		client := cmd.Flag("client").Value.String()
		kfClient, cErr := initClient(false)
		if cErr != nil {
			fmt.Println("Error, unable to connect to Keyfactor.")
			log.Fatalf("Error: %s", cErr)
		}
		agents, aErr := kfClient.GetAgent(client)
		if aErr != nil {
			fmt.Printf("Error, unable to get logs for orchestrator %s. %s\n", client, aErr)
			log.Fatalf("[ERROR]: %s", aErr)
		}
		agent := agents[0]
		_, aErr = kfClient.FetchAgentLogs(agent.AgentId)
		if aErr != nil {
			fmt.Printf("Error, unable to get logs for orchestrator %s. %s\n", client, aErr)
			log.Fatalf("[ERROR]: %s", aErr)
		}
		fmt.Printf("Fetching logs from %s successful.\n", client)
	},
}

// listOrchestratorsCmd represents the list orchestrators command
var listOrchestratorsCmd = &cobra.Command{
	Use:   "list",
	Short: "List orchestrators.",
	Long:  `Returns a JSON list of Keyfactor orchestrators.`,
	Run: func(cmd *cobra.Command, args []string) {
		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		kfClient, _ := initClient(false)
		agents, aErr := kfClient.GetAgentList()
		if aErr != nil {
			fmt.Printf("Error, unable to get orchestrators list. %s\n", aErr)
			log.Fatalf("Error: %s", aErr)
		}
		output, jErr := json.Marshal(agents)
		if jErr != nil {
			fmt.Println("Error, unable to get orchestrators list.")
			log.Fatalf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

func init() {
	var client string

	RootCmd.AddCommand(orchsCmd)

	// EXT orchestrators command
	orchsCmd.AddCommand(NewCmdOrchsExt())

	// LIST orchestrators command
	orchsCmd.AddCommand(listOrchestratorsCmd)
	// GET orchestrator command
	orchsCmd.AddCommand(getOrchestratorCmd)
	getOrchestratorCmd.Flags().StringVarP(
		&client,
		"client",
		"c",
		"",
		"Get a specific orchestrator by machine or client name.",
	)
	getOrchestratorCmd.MarkFlagRequired("client")
	// CREATE orchestrator command
	//orchsCmd.AddCommand(createOrchestratorCmd)
	// UPDATE orchestrator command
	//orchsCmd.AddCommand(updateOrchestratorCmd)
	// DELETE orchestrator command
	//orchsCmd.AddCommand(deleteOrchestratorCmd)
	// APPROVE orchestrator command
	orchsCmd.AddCommand(approveOrchestratorCmd)
	approveOrchestratorCmd.Flags().StringVarP(
		&client,
		"client",
		"c",
		"",
		"Approve a specific orchestrator by machine or client name.",
	)
	approveOrchestratorCmd.MarkFlagRequired("client")
	// DISAPPROVE orchestrator command
	orchsCmd.AddCommand(disapproveOrchestratorCmd)
	disapproveOrchestratorCmd.Flags().StringVarP(
		&client,
		"client",
		"c",
		"",
		"Disapprove a specific orchestrator by machine or client name.",
	)
	disapproveOrchestratorCmd.MarkFlagRequired("client")
	// RESET orchestrator command
	orchsCmd.AddCommand(resetOrchestratorCmd)
	resetOrchestratorCmd.Flags().StringVarP(
		&client,
		"client",
		"c",
		"",
		"Reset a specific orchestrator by machine or client name.",
	)
	resetOrchestratorCmd.MarkFlagRequired("client")
	// GET orchestrator logs command
	orchsCmd.AddCommand(getLogsOrchestratorCmd)
	getLogsOrchestratorCmd.Flags().StringVarP(
		&client,
		"client",
		"c",
		"",
		"Get logs for a specific orchestrator by machine or client name.",
	)
	getLogsOrchestratorCmd.MarkFlagRequired("client")
	// SET orchestrator auth certificate reenrollment command
	//orchsCmd.AddCommand(setOrchestratorAuthCertReenrollCmd)
	// Utility commands
	//orchsCmd.AddCommand(downloadOrchestrator)
	//orchsCmd.AddCommand(installOrchestrator)
}

var (
	OrchsAliases = []string{"orchestrator", "orchestrators", "orch"}
)
