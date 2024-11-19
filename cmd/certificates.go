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

	"github.com/Keyfactor/keyfactor-go-client/v3/api"

	"github.com/spf13/cobra"
)

// certificatesCmd represents the certificates command
var certificatesCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Keyfactor Command certificate APIs and utilities.",
	Long:  `A collections of APIs and utilities for interacting with Keyfactor certificates.`,
	Run: func(cmd *cobra.Command, args []string) {
		expEnabled, _ := cmd.Flags().GetBool("exp")
		isExperimental := true

		_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}
		fmt.Println("NOT IMPLEMENTED: certificates called")
	},
}

func init() {
	//RootCmd.AddCommand(certificatesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// certificatesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// certificatesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func certToString(response *api.GetCertificateResponse) string {
	sansString := ""
	for _, san := range response.SubjectAltNameElements {
		sansString += fmt.Sprintf("%s,", san.Value)
	}
	if len(sansString) > 0 {
		sansString = sansString[:len(sansString)-1]
	}
	return fmt.Sprintf(
		"DN=(%s),SANs=(%s),TP=(%s),ID=(%d)",
		response.IssuedDN,
		sansString,
		response.Thumbprint,
		response.Id,
	)
}
