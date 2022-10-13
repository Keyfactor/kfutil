/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client/api"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

// storeTypesCmd represents the storeTypes command
var storeTypesCmd = &cobra.Command{
	Use:   "store-types",
	Short: "Keyfactor certificate store types APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate store types.`,
}

// storesTypesListCmd represents the list command
var storesTypesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate store types.",
	Long:  `List certificate store types.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(ioutil.Discard)
		kfClient, _ := initClient()
		storeTypes, err := kfClient.ListCertificateStoreTypes()
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(storeTypes)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

var storesTypeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific store type by either name or ID.",
	Long:  `Get a specific store type by either name or ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(ioutil.Discard)
		id, _ := cmd.Flags().GetInt("id")
		name, _ := cmd.Flags().GetString("name")
		kfClient, _ := initClient()
		var st interface{}
		// Check inputs
		if id < 0 && name == "" {
			log.Printf("Error: ID must be a positive integer.")
			fmt.Printf("Error: ID must be a positive integer.\n")
			return
		} else if id >= 0 && name != "" {
			log.Printf("Error: ID and Name are mutually exclusive.")
			fmt.Printf("Error: ID and Name are mutually exclusive.\n")
			return
		} else if id >= 0 {
			st = id
		} else if name != "" {
			st = name
		} else {
			log.Printf("Error: Invalid input.")
			fmt.Printf("Error: Invalid input.\n")
			return
		}

		storeTypes, err := kfClient.GetCertificateStoreType(st)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(storeTypes)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Printf("%s", output)
	},
}

var storesTypeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new certificate store type in Keyfactor.",
	Long:  `Create a new certificate store type in Keyfactor.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")
	},
}

var storesTypeUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a certificate store type in Keyfactor.",
	Long:  `Update a certificate store type in Keyfactor.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(ioutil.Discard)
		fmt.Println("update called")
	},
}

var storesTypeDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a specific store type by ID.",
	Long:  `Delete a specific store type by ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(ioutil.Discard)
		id, _ := cmd.Flags().GetInt("id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		kfClient, _ := initClient()
		_, err := kfClient.GetCertificateStoreType(id)
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		if dryRun {
			fmt.Printf("dry run delete called on certificate store type with ID: %d", id)
		} else {
			d, err := kfClient.DeleteCertificateStoreType(id)
			if err != nil {
				log.Printf("Error: %s", err)
				fmt.Printf("Error: %s\n", err)
				return
			}
			fmt.Printf("Certificate store type with ID: %d deleted", d.ID)
		}
	},
}

var fetchStoreTypes = &cobra.Command{
	Use:   "templates-fetch",
	Short: "Fetches store type templates from Keyfactor's Github.",
	Long:  `Fetches store type templates from Keyfactor's Github.`,
	Run: func(cmd *cobra.Command, args []string) {
		templates, err := readStoreTypesConfig()
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(templates)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Println(string(output))
	},
}

var generateStoreTypeTemplate = &cobra.Command{
	Use:   "templates-generate",
	Short: "Generates either a JSON or CSV template file for certificate store type bulk operations.",
	Long:  `Generates either a JSON or CSV template file for certificate store type bulk operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		templates, err := readStoreTypesConfig()
		if err != nil {
			log.Printf("Error: %s", err)
			fmt.Printf("Error: %s\n", err)
			return
		}
		output, jErr := json.Marshal(templates)
		if jErr != nil {
			log.Printf("Error: %s", jErr)
		}
		fmt.Println(string(output))
	},
}

func readStoreTypesConfig() (map[string]api.CertificateStoreType, error) {
	content, err := ioutil.ReadFile("./store_types.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var payload map[string]api.CertificateStoreType
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Printf("Error during Unmarshal(): %s", err)
		return nil, err
	}
	return payload, nil
}

func init() {
	rootCmd.AddCommand(storeTypesCmd)

	// GET store type templates
	storeTypesCmd.AddCommand(fetchStoreTypes)

	// LIST command
	storeTypesCmd.AddCommand(storesTypesListCmd)

	// GET commands
	storeTypesCmd.AddCommand(storesTypeGetCmd)
	var storeTypeID int
	var storeTypeName string
	var dryRun bool
	storesTypeGetCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to get.")
	storesTypeGetCmd.Flags().StringVarP(&storeTypeName, "name", "n", "", "Name of the certificate store type to get.")
	storesTypeGetCmd.MarkFlagsMutuallyExclusive("id", "name")

	// CREATE command
	storeTypesCmd.AddCommand(storesTypeCreateCmd)

	// UPDATE command
	storeTypesCmd.AddCommand(storesTypeUpdateCmd)

	// DELETE command
	storeTypesCmd.AddCommand(storesTypeDeleteCmd)
	storesTypeDeleteCmd.Flags().IntVarP(&storeTypeID, "id", "i", -1, "ID of the certificate store type to get.")
	storesTypeDeleteCmd.Flags().BoolVarP(&dryRun, "dry-run", "t", false, "Specifies whether to perform a dry run.")
	storesTypeDeleteCmd.MarkFlagRequired("id")

}
