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
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Removes the credentials file '$HOME/.keyfactor/command_config.json'.",
	Long:  `Removes the credentials file '$HOME/.keyfactor/command_config.json'.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("flagEnableDebug")
		//flagConfigFile, _ := cmd.Flags().GetString("config")
		//flagNoPrompt, _ := cmd.Flags().GetBool("no-prompt")
		//flagProfile, _ := cmd.Flags().GetString("flagProfile")

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		err := os.Remove(fmt.Sprintf("%s/.keyfactor/%s", os.Getenv("HOME"), DefaultConfigFileName))
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Config file does not exist.")
				fmt.Println("Logged out successfully!")
				return
			}
			fmt.Println("Error removing config file: ", err)
			//log.Fatal("[ERROR] removing config file: ", err)
		}
		fmt.Println("Logged out successfully!")
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
