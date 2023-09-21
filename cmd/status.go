// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List the status of Keyfactor services.",
	Long:  `Returns a list of all API endpoints.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		//debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		//configFile, _ := cmd.Flags().GetString("config")
		//noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		//profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		//kfClient, _ := initClient(configFile, profile, noPrompt)
		//status, err := kfClient.GetStatus()
		//if err != nil {
		//	log.Printf("Error: %s", err)
		//}
		//output, jErr := json.Marshal(status)
		//if jErr != nil {
		//	log.Printf("Error: %s", jErr)
		//}
		//fmt.Printf("%s", output)
		//
		fmt.Println("status called")
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
