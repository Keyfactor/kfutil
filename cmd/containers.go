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

// containersCmd represents the containers command
var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Keyfactor certificate store container API and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate store containers.`,
}

var containersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create certificate store container.",
	Long:  `Create certificate store container.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(flagEnableExp, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an flagEnableExp feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}
		fmt.Println("Create store containers not implemented.")
	},
}

var containersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get certificate store container by ID or name.",
	Long:  `Get certificate store container by ID or name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags
		id := cmd.Flag("id").Value.String()

		isExperimental := true
		debugErr := warnExperimentalFeature(flagEnableExp, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		kfClient, _ := initClient(false)

		agents, aErr := kfClient.GetStoreContainer(id)
		if aErr != nil {
			fmt.Printf("Error, unable to get container %s. %s\n", id, aErr)
			log.Fatalf("Error: %s", aErr)
		}
		output, jErr := json.Marshal(agents)
		if jErr != nil {
			fmt.Printf("Error invalid API response from Keyfactor. %s\n", jErr)
			log.Fatalf("[ERROR]: %s", jErr)
		}
		fmt.Printf("%s", output)
		return nil
	},
}

var containersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update certificate store container by ID or name.",
	Long:  `Update certificate store container by ID or name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags

		isExperimental := true
		informDebug(flagEnableDebug)
		debugErr := warnExperimentalFeature(flagEnableExp, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Authenticate
		//

		// CLI Logic
		return fmt.Errorf("update store containers not implemented")
	},
}

var containersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete certificate store container by ID or name.",
	Long:  `Delete certificate store container by ID or name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags

		// Debug + flagEnableExp checks
		isExperimental := true
		informDebug(flagEnableDebug)
		debugErr := warnExperimentalFeature(flagEnableExp, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Authenticate
		//
		//kfClient, _ := initClient(flagConfigFile, flagProfile, flagProviderType, flagProviderProfile, flagNoPrompt, authConfig, false)

		// CLI Logic
		return fmt.Errorf("delete store containers not implemented")
	},
}

var containersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate store containers.",
	Long:  `List certificate store containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		// Specific flags

		// Debug + flagEnableExp checks
		isExperimental := true
		informDebug(flagEnableDebug)
		debugErr := warnExperimentalFeature(flagEnableExp, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		// Authenticate
		//
		kfClient, _ := initClient(false)

		// CLI Logic
		agents, aErr := kfClient.GetStoreContainers()
		if aErr != nil {
			fmt.Printf("Error, unable to list store containers. %s\n", aErr)
			log.Fatalf("Error: %s", aErr)
		}
		output, jErr := json.Marshal(agents)
		if jErr != nil {
			fmt.Printf("Error invalid API response from Keyfactor. %s\n", jErr)
			log.Fatalf("[ERROR]: %s", jErr)
		}
		fmt.Printf("%s", output)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(containersCmd)
	// LIST containers command
	containersCmd.AddCommand(containersListCmd)
	// GET containers command
	containersCmd.AddCommand(containersGetCmd)
	containersGetCmd.Flags().StringP("id", "i", "", "ID or name of the cert store container.")
	containersGetCmd.MarkFlagRequired("id")
	// CREATE containers command
	//containersCmd.AddCommand(containersCreateCmd)
	// UPDATE containers command
	//containersCmd.AddCommand(containersUpdateCmd)
	// DELETE containers command
	//containersCmd.AddCommand(containersDeleteCmd)
	// Utility functions
}
