/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List the status of Keyfactor services.",
	Long:  `Returns a list of all API endpoints.`,
	Run: func(cmd *cobra.Command, args []string) {
		//log.SetOutput(ioutil.Discard)
		//kfClient, _ := initClient()
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
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
