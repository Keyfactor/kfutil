// Package cmd Copyright 2023 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/spf13/cobra"
	"log"
)

// inventoryCmd represents the inventory command
var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Commands related to certificate store inventory management",
	Long:  `Commands related to certificate store inventory management`,
}

var inventoryClearCmd = &cobra.Command{
	Use:                    "clear",
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Clears the certificate store inventory of ALL certificates.",
	GroupID:                "",
	Long:                   `Clears the certificate store inventory of ALL certificates.`,
	Example:                "",
	ValidArgs:              nil,
	ValidArgsFunction:      nil,
	Args:                   nil,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Annotations:            nil,
	Version:                "",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		storeID, _ := cmd.Flags().GetStringSlice("sid")
		machineName, _ := cmd.Flags().GetStringSlice("client")
		storeType, _ := cmd.Flags().GetStringSlice("store-type")
		containerType, _ := cmd.Flags().GetStringSlice("container")
		allStores, _ := cmd.Flags().GetBool("all")

		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		if storeID == nil && machineName == nil && storeType == nil && containerType == nil && !allStores {
			fmt.Println("You must specify at least one of the following options: --sid, --client, --store-type, --container, --all")
			return
		}

		sIdMap := make(map[string]bool)
		for _, sId := range storeID {
			sIdMap[sId] = true
		}
		mNameMap := make(map[string]bool)
		for _, mName := range machineName {
			mNameMap[mName] = true
		}
		sTypeMap := make(map[string]bool)
		for _, sType := range storeType {
			sTypeMap[sType] = true
		}
		cTypeMap := make(map[string]bool)
		for _, cType := range containerType {
			cTypeMap[cType] = true
		}
		var filteredStores []api.GetCertificateStoreResponse

		sTypeLookup := make(map[string]bool)
		if !allStores {
			params := make(map[string]interface{})
			allStoresResponse, _ := kfClient.ListCertificateStores(&params) //nil
			for _, store := range *allStoresResponse {
				sTypeName, stErr := kfClient.GetCertificateStoreTypeById(store.CertStoreType)
				sTypeLookup[sTypeName.ShortName] = true
				if stErr != nil {
					fmt.Printf("Error getting store type name for store type id %d: %s\n", store.CertStoreType, stErr)
					log.Fatal(stErr)
				}
				if sIdMap[store.Id] || mNameMap[store.ClientMachine] || sTypeMap[sTypeName.ShortName] || cTypeMap[store.ContainerName] {
					filteredStores = append(filteredStores, store)
				}
			}
		} else {
			params := make(map[string]interface{})
			allStoresResp, fErr := kfClient.ListCertificateStores(&params)
			if fErr != nil {
				fmt.Printf("Error listing certificate stores: %s\n", fErr)
				log.Fatal(fErr)
			}
			filteredStores = *allStoresResp
		}

		for _, store := range filteredStores {

			sInvs, iErr := kfClient.GetCertStoreInventory(store.Id) //TODO: This is a placeholder for the actual API call
			if iErr != nil {
				fmt.Printf("Error unable to get inventory from certificate store %s. %s\n", store.Id, iErr)
				log.Printf("[ERROR]  %s", iErr)
			}
			schedule := &api.InventorySchedule{
				Immediate: boolToPointer(true),
			}

			if !force {
				fmt.Printf("This will clear the inventory of ALL certificates in the store %s:%s. Are you sure you sure?! Press 'y' to continue? (y/n) ", store.ClientMachine, store.StorePath)
				var answer string
				fmt.Scanln(&answer)
				if answer != "y" {
					fmt.Println("Aborting")
					return
				}
			}

			for _, inv := range *sInvs {
				certs := inv.Certificates
				for _, cert := range certs {
					st := api.CertificateStore{ //TODO: This conversion is a bit weird to have to do.  Should be able to pass the store directly.
						CertificateStoreId: store.Id,
						Alias:              cert.Thumbprint,
						Overwrite:          true,
						EntryPassword:      nil,
						PfxPassword:        "",
						IncludePrivateKey:  true,
					}
					var stores []api.CertificateStore
					stores = append(stores, st)
					removeReq := api.RemoveCertificateFromStore{
						CertificateId:     cert.Id,
						CertificateStores: &stores,
						InventorySchedule: schedule,
					}
					if !dryRun {
						_, err := kfClient.RemoveCertificateFromStores(&removeReq)
						if err != nil {
							fmt.Printf("Error removing certificate %s(%d) from store %s: %s\n", cert.IssuedDN, cert.Id, st.CertificateStoreId, err)
							log.Printf("[ERROR] %s", err)
							continue
						}
					} else {
						fmt.Printf("Dry run: Would have removed certificate %s(%d) from store %s\n", cert.IssuedDN, cert.Id, st.CertificateStoreId)
					}
				}

			}
			fmt.Println("Inventory cleared")
		}
	},
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	CompletionOptions:          cobra.CompletionOptions{},
	TraverseChildren:           false,
	Hidden:                     true,
	SilenceErrors:              false,
	SilenceUsage:               true,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         true,
	SuggestionsMinimumDistance: 0,
}

var inventoryAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds one or more certificates to one or more certificate store inventories.",
	Long: `Adds one or more certificates to one or more certificate store inventories. The certificate(s) to add can be
specified by thumbprint, Keyfactor command certificate ID, or subject name. The store(s) to add the certificate(s) to can be
specified by Keyfactor command store ID, client machine name, store type, or container type. At least one or more stores
and one or more certificates must be specified. If multiple stores and/or certificates are specified, the command will
attempt to add all the certificate(s) meeting the specified criteria to all stores meeting the specified criteria.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		storeIDs, _ := cmd.Flags().GetStringSlice("sid")
		thumbprints, _ := cmd.Flags().GetStringSlice("thumbprint")
		certIDs, _ := cmd.Flags().GetStringSlice("cid")
		subjects, _ := cmd.Flags().GetStringSlice("cn")
		machineNames, _ := cmd.Flags().GetStringSlice("client")
		storeTypes, _ := cmd.Flags().GetStringSlice("store-type")
		containerType, _ := cmd.Flags().GetStringSlice("container")
		allStores, _ := cmd.Flags().GetBool("all-stores")

		if !allStores && (len(storeIDs) == 0 && len(machineNames) == 0 && len(storeTypes) == 0 && len(containerType) == 0) {
			fmt.Println("At least one store parameter must be specified: [sid, client, store-type, container]. Or specify --all-stores.")
			log.Fatalf("At least one store must be specified")
		}

		if len(thumbprints) == 0 && len(certIDs) == 0 && len(subjects) == 0 {
			fmt.Println("At least one certificate parameter must be specified. [thumbprint, cid, cn]")
			log.Fatalf("At least one certificate must be specified")
		}

		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		if storeIDs == nil && machineNames == nil && storeTypes == nil && containerType == nil && !allStores {
			fmt.Println("You must specify at least one of the following options: --sid, --client, --store-type, --container, --all")
			return
		}

		sIdMap := make(map[string]bool)
		for _, sId := range storeIDs {
			sIdMap[sId] = true
		}
		mNameMap := make(map[string]bool)
		for _, mName := range machineNames {
			mNameMap[mName] = true
		}
		sTypeMap := make(map[string]bool)
		for _, sType := range storeTypes {
			sTypeMap[sType] = true
		}
		cTypeMap := make(map[string]bool)
		for _, cType := range containerType {
			cTypeMap[cType] = true
		}
		var filteredStores []api.GetCertificateStoreResponse
		var filteredCerts []api.GetCertificateResponse

		for _, cn := range subjects {
			cert, err := kfClient.ListCertificates(map[string]string{
				"subject": cn,
			})
			if err != nil {
				fmt.Printf("Unable to find certificate with subject: %s\n", cn)
				continue
			}
			filteredCerts = append(filteredCerts, cert...)
		}
		for _, thumbprint := range thumbprints {
			cert, err := kfClient.ListCertificates(map[string]string{
				"thumbprint": thumbprint,
			})
			if err != nil {
				fmt.Printf("Unable to find certificate with thumbprint: %s\n", thumbprint)
				continue
			}
			filteredCerts = append(filteredCerts, cert...)
		}
		for _, certID := range certIDs {
			cert, err := kfClient.ListCertificates(map[string]string{
				"id": certID,
			})
			if err != nil {
				fmt.Printf("Unable to find certificate with ID: %s\n", certID)
				continue
			}
			filteredCerts = append(filteredCerts, cert...)
		}

		sTypeLookup := make(map[string]bool)
		if !allStores {
			params := make(map[string]interface{})
			allStoresResponse, _ := kfClient.ListCertificateStores(&params)
			for _, store := range *allStoresResponse {
				sTypeName, stErr := kfClient.GetCertificateStoreTypeById(store.CertStoreType)
				sTypeLookup[sTypeName.ShortName] = true
				if stErr != nil {
					fmt.Printf("Error getting store type name for store type id %d: %s", store.CertStoreType, stErr)
				}
				if sIdMap[store.Id] || mNameMap[store.ClientMachine] || sTypeMap[sTypeName.ShortName] || cTypeMap[store.ContainerName] {
					filteredStores = append(filteredStores, store)
				}
			}
		} else {
			params := make(map[string]interface{})
			allStoresResp, fErr := kfClient.ListCertificateStores(&params)
			if fErr != nil {
				fmt.Printf("Error getting listing certificate stores: %s", fErr)
				log.Fatal(fErr)
			}
			filteredStores = *allStoresResp
		}

		for _, store := range filteredStores {
			schedule := &api.InventorySchedule{
				Immediate: boolToPointer(true),
			}
			for _, cert := range filteredCerts {
				st := api.CertificateStore{ //TODO: This conversion is weird. Should be able to use the store directly.
					CertificateStoreId: store.Id,
					Alias:              cert.Thumbprint,
					Overwrite:          true,
					EntryPassword:      nil,
					PfxPassword:        "",
					IncludePrivateKey:  true,
				}
				var stores []api.CertificateStore
				stores = append(stores, st)
				addReq := api.AddCertificateToStore{
					CertificateId:     cert.Id,
					CertificateStores: &stores,
					InventorySchedule: schedule,
				}
				if !dryRun {
					if !force {
						fmt.Printf("This will add the certificate %s(%d) to certificate store %s%s's inventory. Are you sure you shouldPass to continue? (y/n) ", cert.IssuedCN, cert.Id, store.ClientMachine, store.StorePath)
						var answer string
						fmt.Scanln(&answer)
						if answer != "y" {
							fmt.Println("Aborting")
							return
						}
					}
					_, err := kfClient.AddCertificateToStores(&addReq)
					if err != nil {
						fmt.Printf("Error adding certificate %s(%d) to store %s: %s\n", cert.IssuedCN, cert.Id, st.CertificateStoreId, err)
						log.Printf("[ERROR]  %s", err)
						continue
					}
				} else {
					fmt.Printf("Dry run: Would have added certificate %s(%d) from store %s", cert.IssuedDN, cert.Id, st.CertificateStoreId)
				}
			}

		}
		fmt.Println("Inventory updated successfully")
	},
}

var inventoryRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a certificate from the certificate store inventory.",
	Long:  `Removes a certificate from the certificate store inventory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		storeIDs, _ := cmd.Flags().GetStringSlice("sid")
		thumbprints, _ := cmd.Flags().GetStringSlice("thumbprint")
		certIDs, _ := cmd.Flags().GetStringSlice("cid")
		subjects, _ := cmd.Flags().GetStringSlice("cn")
		machineNames, _ := cmd.Flags().GetStringSlice("client")
		storeTypes, _ := cmd.Flags().GetStringSlice("store-type")
		containerType, _ := cmd.Flags().GetStringSlice("container")
		allStores, _ := cmd.Flags().GetBool("all-stores")

		if !allStores && (len(storeIDs) == 0 && len(machineNames) == 0 && len(storeTypes) == 0 && len(containerType) == 0) {
			fmt.Println("At least one store parameter must be specified: [sid, client, store-type, container]. Or specify --all-stores.")
			log.Fatalf("At least one store must be specified")
		}

		if len(thumbprints) == 0 && len(certIDs) == 0 && len(subjects) == 0 {
			fmt.Println("At least one certificate parameter must be specified. [thumbprint, cid, cn]")
			log.Fatalf("At least one certificate must be specified")
		}

		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		if storeIDs == nil && machineNames == nil && storeTypes == nil && containerType == nil && !allStores {
			fmt.Println("You must specify at least one of the following options: --sid, --client, --store-type, --container, --all")
			return
		}

		sIdMap := make(map[string]bool)
		for _, sId := range storeIDs {
			sIdMap[sId] = true
		}
		mNameMap := make(map[string]bool)
		for _, mName := range machineNames {
			mNameMap[mName] = true
		}
		sTypeMap := make(map[string]bool)
		for _, sType := range storeTypes {
			sTypeMap[sType] = true
		}
		cTypeMap := make(map[string]bool)
		for _, cType := range containerType {
			cTypeMap[cType] = true
		}
		var filteredStores []api.GetCertificateStoreResponse
		var filteredCerts []api.GetCertificateResponse

		for _, cn := range subjects {
			cert, err := kfClient.ListCertificates(map[string]string{
				"subject": cn,
			})
			if err != nil {
				fmt.Printf("Unable to find certificate with subject: %s\n", cn)
				continue
			}
			filteredCerts = append(filteredCerts, cert...)
		}
		for _, thumbprint := range thumbprints {
			cert, err := kfClient.ListCertificates(map[string]string{
				"thumbprint": thumbprint,
			})
			if err != nil {
				fmt.Printf("Unable to find certificate with thumbprint: %s\n", thumbprint)
				continue
			}
			filteredCerts = append(filteredCerts, cert...)
		}
		for _, certID := range certIDs {
			cert, err := kfClient.ListCertificates(map[string]string{
				"id": certID,
			})
			if err != nil {
				fmt.Printf("Unable to find certificate with ID: %s\n", certID)
				continue
			}
			filteredCerts = append(filteredCerts, cert...)
		}

		sTypeLookup := make(map[string]bool)
		if !allStores {
			params := make(map[string]interface{})
			allStoresResponse, _ := kfClient.ListCertificateStores(&params)
			for _, store := range *allStoresResponse {
				sTypeName, stErr := kfClient.GetCertificateStoreTypeById(store.CertStoreType)
				sTypeLookup[sTypeName.ShortName] = true
				if stErr != nil {
					fmt.Printf("Error getting store type name for store type id %d: %s\n", store.CertStoreType, stErr)
					log.Fatal(stErr)
				}
				if sIdMap[store.Id] || mNameMap[store.ClientMachine] || sTypeMap[sTypeName.ShortName] || cTypeMap[store.ContainerName] {
					filteredStores = append(filteredStores, store)
				}
			}
		} else {
			params := make(map[string]interface{})
			allStoresResp, fErr := kfClient.ListCertificateStores(&params)
			if fErr != nil {
				fmt.Printf("Error listing certificate stores: %s\n", fErr)
				log.Fatal(fErr)
			}
			filteredStores = *allStoresResp
		}

		for _, store := range filteredStores {
			schedule := &api.InventorySchedule{
				Immediate: boolToPointer(true),
			}
			for _, cert := range filteredCerts {
				st := api.CertificateStore{ //TODO: This conversion is weird. Should be able to use the store directly.
					CertificateStoreId: store.Id,
					Alias:              cert.Thumbprint,
					Overwrite:          true,
					EntryPassword:      nil,
					PfxPassword:        "",
					IncludePrivateKey:  true,
				}
				var stores []api.CertificateStore
				stores = append(stores, st)
				removeReq := api.RemoveCertificateFromStore{
					CertificateId:     cert.Id,
					CertificateStores: &stores,
					InventorySchedule: schedule,
				}
				if !dryRun {
					if !force {
						fmt.Printf("This will remove the certificate %s from certificate store %s%s's inventory. Are you sure you shouldPass to continue? (y/n) ", certToString(&cert), store.ClientMachine, store.StorePath)
						var answer string
						fmt.Scanln(&answer)
						if answer != "y" {
							fmt.Println("Aborting")
							return
						}
					}
					_, err := kfClient.RemoveCertificateFromStores(&removeReq)
					if err != nil {
						fmt.Printf("Error removing certificate %s to store %s: %s\n", certToString(&cert), st.CertificateStoreId, err)
						log.Printf("[ERROR] %s", err)
						continue
					}
				} else {
					fmt.Printf("Dry run: Would have removed certificate %s from store %s\n", certToString(&cert), st.CertificateStoreId)
				}
			}

		}
		fmt.Println("Inventory updated successfully")
	},
}

var inventoryShowCmd = &cobra.Command{
	Use:                    "show",
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Show the inventory of a certificate store.",
	GroupID:                "",
	Long:                   `Show the inventory of a certificate store.`,
	Example:                "",
	ValidArgs:              nil,
	ValidArgsFunction:      nil,
	Args:                   nil,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Annotations:            nil,
	Version:                "",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debugFlag")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")
		expEnabled, _ := cmd.Flags().GetBool("exp")
		kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
		kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
		kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
		kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
		kfcAPIPath, _ := cmd.Flags().GetString("api-path")
		authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
		isExperimental := true

		_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
		if expErr != nil {
			fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
			log.Fatalf("[ERROR]: %s", expErr)
		}

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)
		storeIDs, _ := cmd.Flags().GetStringSlice("sid")
		clientMachineNames, _ := cmd.Flags().GetStringSlice("client")
		storeTypes, _ := cmd.Flags().GetStringSlice("store-type")
		containers, _ := cmd.Flags().GetStringSlice("container")

		kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)

		if len(storeIDs) == 0 && len(clientMachineNames) == 0 && len(storeTypes) == 0 && len(containers) == 0 {
			fmt.Println("No filters specified. Unable to show inventory. Please specify at least one filter: [--sid, --client, --store-type, --container]")
			return
		}

		params := map[string]interface{}{
			"ClientMachine": []string{},
			"ContainerId":   []int{},
			"Storepath":     []string{},
			"CertStoreType": []string{},
			"AgentId":       []string{},
			"ContainerName": []string{},
			"Id":            []string{},
		}
		for _, cm := range clientMachineNames {
			params["ClientMachine"] = append(params["ClientMachine"].([]string), cm)
		}
		for _, st := range storeTypes {
			params["CertStoreType"] = append(params["CertStoreType"].([]string), st)
		}
		for _, c := range containers {
			params["ContainerName"] = append(params["ContainerName"].([]string), c)
		}
		for _, s := range storeIDs {
			params["Id"] = append(params["Id"].([]string), s)
		}
		//params := make(map[string]interface{})
		stResp, err := kfClient.ListCertificateStores(&params)
		if err != nil {
			fmt.Println("Error, unable to list certificate stores. ", err)
			log.Printf("[ERROR] Unable to list certificate stores: %s\n", err)
			return
		}

		lkup := make(map[string]interface{})
		var output []map[string]interface{}
		for _, cStore := range *stResp {
			inv, err := kfClient.GetCertStoreInventory(cStore.Id)
			if err != nil {
				fmt.Printf("Error, unable to retrieve certificate store inventory from %v: %s\n", cStore, err)
				log.Printf("[ERROR]  %s", err)
			}
			invData := make(map[string]interface{})
			invData["StoreId"] = cStore.Id
			invData["Storepath"] = cStore.StorePath
			invData["StoreType"] = cStore.CertStoreType
			invData["ContainerName"] = cStore.ContainerName
			invData["ClientMachine"] = cStore.ClientMachine
			invData["Inventory"] = inv
			if _, ok := lkup[cStore.Id]; !ok {
				output = append(output, invData)
			}
			lkup[cStore.Id] = invData
		}
		// return JSON response
		json, jsErr := json.Marshal(lkup)
		if jsErr != nil {
			fmt.Printf("Error, unable to format JSON: %s\n", jsErr)
			log.Println("[ERROR] ", jsErr)
			return
		}
		fmt.Println(string(json))
	},
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
	CompletionOptions:          cobra.CompletionOptions{},
	TraverseChildren:           false,
	Hidden:                     false,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
}

func init() {
	var (
		ids          []string
		clients      []string
		thumbprints  []string
		types        []string
		containers   []string
		all          bool
		force        bool
		dryRun       bool
		cIDs         []string
		subjectNames []string
	)

	storesCmd.AddCommand(inventoryCmd)

	inventoryCmd.AddCommand(inventoryClearCmd)
	inventoryClearCmd.Flags().StringSliceVar(&ids, "sid", []string{}, "The Keyfactor Command ID of the certificate store(s) remove all inventory from.")
	inventoryClearCmd.Flags().StringSliceVar(&clients, "client", []string{}, "Remove all inventory from store(s) of specific client machine(s).")
	inventoryClearCmd.Flags().StringSliceVar(&types, "store-type", []string{}, "Remove all inventory from store(s) of specific store type(s).")
	inventoryClearCmd.Flags().StringSliceVar(&containers, "container", []string{}, "Remove all inventory from store(s) of specific container type(s).")
	inventoryClearCmd.Flags().BoolVar(&all, "all", false, "Remove all inventory from all certificate stores.")
	inventoryClearCmd.Flags().BoolVar(&force, "force", false, "Force removal of inventory without prompting for confirmation.")
	inventoryClearCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Do not remove inventory, only show what would be removed.")

	inventoryCmd.AddCommand(inventoryAddCmd)
	inventoryAddCmd.Flags().StringSliceVar(&ids, "sid", []string{}, "The Keyfactor Command ID of the certificate store(s) to add inventory to.")
	inventoryAddCmd.Flags().StringSliceVar(&clients, "client", []string{}, "Add a certificate to all stores of specific client machine(s).")
	inventoryAddCmd.Flags().StringSliceVar(&types, "store-type", []string{}, "Add a certificate to all stores of specific store type(s).")
	inventoryAddCmd.Flags().StringSliceVar(&containers, "container", []string{}, "Add a certificate to all stores of specific container type(s).")
	inventoryAddCmd.Flags().StringSliceVar(&thumbprints, "thumbprint", []string{}, "The thumbprint of the certificate(s) to add to the store(s).")
	inventoryAddCmd.Flags().StringSliceVar(&cIDs, "cid", []string{}, "The Keyfactor command certificate ID(s) of the certificate to add to the store(s).")
	inventoryAddCmd.Flags().StringSliceVar(&subjectNames, "cn", []string{}, "Subject name(s) of the certificate(s) to add to the store(s).")
	inventoryAddCmd.Flags().BoolVar(&all, "all-stores", false, "Add the certificate(s) to all certificate stores.")
	inventoryAddCmd.Flags().BoolVar(&force, "force", false, "Force addition of inventory without prompting for confirmation.")
	inventoryAddCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Do not add inventory, only show what would be added.")

	inventoryCmd.AddCommand(inventoryRemoveCmd)
	inventoryRemoveCmd.Flags().StringSliceVar(&ids, "sid", []string{}, "The Keyfactor Command ID of the certificate store(s) to remove inventory from.")
	inventoryRemoveCmd.Flags().StringSliceVar(&clients, "client", []string{}, "Remove certificate(s) from all stores of specific client machine(s).")
	inventoryRemoveCmd.Flags().StringSliceVar(&types, "store-type", []string{}, "Remove certificate(s) from all stores of specific store type(s).")
	inventoryRemoveCmd.Flags().StringSliceVar(&containers, "container", []string{}, "Remove certificate(s) from all stores of specific container type(s).")
	inventoryRemoveCmd.Flags().StringSliceVar(&thumbprints, "thumbprint", []string{}, "The thumbprint of the certificate(s) to remove from the store(s).")
	inventoryRemoveCmd.Flags().StringSliceVar(&cIDs, "cid", []string{}, "The Keyfactor command certificate ID(s) of the certificate to remove from the store(s).")
	inventoryRemoveCmd.Flags().StringSliceVar(&subjectNames, "cn", []string{}, "Subject name(s) of the certificate(s) to remove from the store(s).")
	inventoryRemoveCmd.Flags().BoolVar(&all, "all-stores", false, "Remove the certificate(s) from all certificate stores.")
	inventoryRemoveCmd.Flags().BoolVar(&force, "force", false, "Force removal of inventory without prompting for confirmation.")
	inventoryRemoveCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Do not remove inventory, only show what would be removed.")

	inventoryCmd.AddCommand(inventoryShowCmd)
	inventoryShowCmd.Flags().StringSliceVar(&ids, "sid", []string{}, "The Keyfactor Command ID of the certificate store(s) to retrieve inventory from.")
	inventoryShowCmd.Flags().StringSliceVar(&clients, "client", []string{}, "Show certificate inventories for stores of specific client machine(s).")
	inventoryShowCmd.Flags().StringSliceVar(&types, "store-type", []string{}, "Show certificate inventories for stores of specific store type(s).")
	inventoryShowCmd.Flags().StringSliceVar(&containers, "container", []string{}, "Show certificate inventories for stores of specific container type(s).")

}
