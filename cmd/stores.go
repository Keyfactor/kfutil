// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
)

// storesCmd represents the stores command
var storesCmd = &cobra.Command{
	Use:   "stores",
	Short: "Keyfactor certificate stores APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate stores.`,
	//Run: func(cmd *cobra.Command, args []string) {
	//	fmt.Println("stores called")
	//},
}

var storesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate stores.",
	Long:  `List certificate stores.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		kfClient, _ := initClient()
		params := make(map[string]interface{})
		stores, err := kfClient.ListCertificateStores(&params)

		if err != nil {
			log.Printf("Error: %s", err)
		}
		output, jErr := json.Marshal(stores)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

var storesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a certificate store by ID.",
	Long:  `Get a certificate store by ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		storeId, _ := cmd.Flags().GetString("id")
		kfClient, _ := initClient()
		stores, err := kfClient.GetCertificateStoreByID(storeId)
		if err != nil {
			log.Printf("Error: %s", err)
		}
		output, jErr := json.Marshal(stores)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

func init() {
	var storeId string
	RootCmd.AddCommand(storesCmd)
	storesCmd.AddCommand(storesListCmd)
	storesCmd.AddCommand(storesGetCmd)
	storesGetCmd.Flags().StringVarP(&storeId, "id", "i", "", "ID of the certificate store to get.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
