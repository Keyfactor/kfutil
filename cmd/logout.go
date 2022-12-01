// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"fmt"
	"io"
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
		log.SetOutput(io.Discard)
		err := os.Remove(fmt.Sprintf("%s/.keyfactor/%s", os.Getenv("HOME"), DefaultConfigFileName))
		if err != nil {
			fmt.Println("Error removing config file: ", err)
			log.Fatal("[ERROR] removing config file: ", err)
		}
		fmt.Println("Logged out successfully!")
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
