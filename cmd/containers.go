/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// containersCmd represents the containers command
var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Keyfactor CertificateStoreContainer API and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificate store containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containers called")
	},
}

var containersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create certificate store container.",
	Long:  `Create certificate store container.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containers create called")
	},
}

var containersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get certificate store container by ID or name.",
	Long:  `Get certificate store container by ID or name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containers get called")
	},
}

var containersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update certificate store container by ID or name.",
	Long:  `Update certificate store container by ID or name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containers update called")
	},
}

var containersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete certificate store container by ID or name.",
	Long:  `Delete certificate store container by ID or name.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containers delete called")
	},
}

var containersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificate store containers.",
	Long:  `List certificate store containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("containers list called")
	},
}

func init() {
	rootCmd.AddCommand(containersCmd)

	// LIST containers command
	containersCmd.AddCommand(containersListCmd)
	// GET containers command
	containersCmd.AddCommand(containersGetCmd)
	// CREATE containers command
	containersCmd.AddCommand(containersCreateCmd)
	// UPDATE containers command
	containersCmd.AddCommand(containersUpdateCmd)
	// DELETE containers command
	containersCmd.AddCommand(containersDeleteCmd)

	// Utility functions
}
