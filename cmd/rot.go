// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

//
//import (
//	"bufio"
//	"encoding/csv"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/Keyfactor/keyfactor-go-client/v2/api"
//	"github.com/spf13/cobra"
//	"log"
//	"os"
//	"strconv"
//	"strings"
//)
//
//type templateType string
//type StoreCSVEntry struct {
//	ID          string          `json:"id"`
//	Type        string          `json:"type"`
//	Machine     string          `json:"address"`
//	Path        string          `json:"path"`
//	Thumbprints map[string]bool `json:"thumbprints,omitempty"`
//	Serials     map[string]bool `json:"serials,omitempty"`
//	Ids         map[int]bool    `json:"ids,omitempty"`
//}
//type ROTCert struct {
//	ID         int                        `json:"id,omitempty"`
//	ThumbPrint string                     `json:"thumbprint,omitempty"`
//	CN         string                     `json:"cn,omitempty"`
//	Locations  []api.CertificateLocations `json:"locations,omitempty"`
//}
//type ROTAction struct {
//	StoreID    string `json:"store_id,omitempty"`
//	StoreType  string `json:"store_type,omitempty"`
//	StorePath  string `json:"store_path,omitempty"`
//	Thumbprint string `json:"thumbprint,omitempty"`
//	CertID     int    `json:"cert_id,omitempty" mapstructure:"CertID,omitempty"`
//	AddCert    bool   `json:"add,omitempty" mapstructure:"AddCert,omitempty"`
//	RemoveCert bool   `json:"remove,omitempty"  mapstructure:"RemoveCert,omitempty"`
//}
//
//const (
//	tTypeCerts               templateType = "certs"
//	reconcileDefaultFileName string       = "rot_audit.csv"
//)
//
//var (
//	AuditHeader           = []string{"Thumbprint", "CertID", "SubjectName", "Issuer", "StoreID", "StoreType", "Machine", "Path", "AddCert", "RemoveCert", "Deployed", "AuditDate"}
//	ReconciledAuditHeader = []string{"Thumbprint", "CertID", "SubjectName", "Issuer", "StoreID", "StoreType", "Machine", "Path", "AddCert", "RemoveCert", "Deployed", "ReconciledDate"}
//	StoreHeader           = []string{"StoreID", "StoreType", "StoreMachine", "StorePath", "ContainerId", "ContainerName", "LastQueriedDate"}
//	CertHeader            = []string{"Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"}
//)
//
//// String is used both by fmt.Print and by Cobra in help text
//func (e *templateType) String() string {
//	return string(*e)
//}
//
//// Set must have pointer receiver, so it doesn't change the value of a copy
//func (e *templateType) Set(v string) error {
//	switch v {
//	case "certs", "stores", "actions":
//		*e = templateType(v)
//		return nil
//	default:
//		return errors.New(`must be one of "certs", "stores", or "actions"`)
//	}
//}
//
//// Type is only used in help text
//func (e *templateType) Type() string {
//	return "string"
//}
//
//func templateTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
//	return []string{
//		"certs\tGenerates template CSV for certificate input to be used w/ `--add-certs` or `--remove-certs`",
//		"stores\tGenerates template CSV for certificate input to be used w/ `--stores`",
//		"actions\tGenerates template CSV for certificate input to be used w/ `--actions`",
//	}, cobra.ShellCompDirectiveDefault
//}
//
//func generateAuditReport(addCerts map[string]string, removeCerts map[string]string, stores map[string]StoreCSVEntry, outpath string, kfClient *api.Client) ([][]string, map[string][]ROTAction, error) {
//	log.Println("[DEBUG] generateAuditReport called")
//	var (
//		data [][]string
//	)
//
//	data = append(data, AuditHeader)
//	var csvFile *os.File
//	var fErr error
//	if outpath == "" {
//		csvFile, fErr = os.Create(reconcileDefaultFileName)
//		outpath = reconcileDefaultFileName
//	} else {
//		csvFile, fErr = os.Create(outpath)
//	}
//
//	if fErr != nil {
//		fmt.Printf("%s", fErr)
//		log.Fatalf("[ERROR] creating audit file: %s", fErr)
//	}
//	csvWriter := csv.NewWriter(csvFile)
//	cErr := csvWriter.Write(AuditHeader)
//	if cErr != nil {
//		fmt.Printf("%s", cErr)
//		log.Fatalf("[ERROR] writing audit header: %s", cErr)
//	}
//	actions := make(map[string][]ROTAction)
//
//	for _, cert := range addCerts {
//		certLookupReq := api.GetCertificateContextArgs{
//			IncludeMetadata:  boolToPointer(true),
//			IncludeLocations: boolToPointer(true),
//			CollectionId:     nil,
//			Thumbprint:       cert,
//			Id:               0,
//		}
//		certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
//		if err != nil {
//			fmt.Printf("[ERROR] looking up certificate %s: %s\n", cert, err)
//			log.Printf("[ERROR] looking up cert: %s\n%v", cert, err)
//			continue
//		}
//		certID := certLookup.Id
//		certIDStr := strconv.Itoa(certID)
//		for _, store := range stores {
//			if _, ok := store.Thumbprints[cert]; ok {
//				// Cert is already in the store do nothing
//				row := []string{cert, certIDStr, certLookup.IssuedDN, certLookup.IssuerDN, store.ID, store.Type, store.Machine, store.Path, "false", "false", "true", GetCurrentTime()}
//				data = append(data, row)
//				wErr := csvWriter.Write(row)
//				if wErr != nil {
//					fmt.Printf("[ERROR] writing audit file row: %s\n", wErr)
//					log.Printf("[ERROR] writing audit row: %s", wErr)
//				}
//			} else {
//				// Cert is not deployed to this store and will need to be added
//				row := []string{cert, certIDStr, certLookup.IssuedDN, certLookup.IssuerDN, store.ID, store.Type, store.Machine, store.Path, "true", "false", "false", GetCurrentTime()}
//				data = append(data, row)
//				wErr := csvWriter.Write(row)
//				if wErr != nil {
//					fmt.Printf("[ERROR] writing audit file row: %s\n", wErr)
//					log.Printf("[ERROR] writing audit row: %s", wErr)
//				}
//				actions[cert] = append(actions[cert], ROTAction{
//					Thumbprint: cert,
//					CertID:     certID,
//					StoreID:    store.ID,
//					StoreType:  store.Type,
//					StorePath:  store.Path,
//					AddCert:    true,
//					RemoveCert: false,
//				})
//			}
//		}
//	}
//	for _, cert := range removeCerts {
//		certLookupReq := api.GetCertificateContextArgs{
//			IncludeMetadata:  boolToPointer(true),
//			IncludeLocations: boolToPointer(true),
//			CollectionId:     nil,
//			Thumbprint:       cert,
//			Id:               0,
//		}
//		certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
//		if err != nil {
//			log.Printf("[ERROR] looking up cert: %s", err)
//			continue
//		}
//		certID := certLookup.Id
//		certIDStr := strconv.Itoa(certID)
//		for _, store := range stores {
//			if _, ok := store.Thumbprints[cert]; ok {
//				// Cert is deployed to this store and will need to be removed
//				row := []string{cert, certIDStr, certLookup.IssuedDN, certLookup.IssuerDN, store.ID, store.Type, store.Machine, store.Path, "false", "true", "true", GetCurrentTime()}
//				data = append(data, row)
//				wErr := csvWriter.Write(row)
//				if wErr != nil {
//					fmt.Printf("%s", wErr)
//					log.Printf("[ERROR] writing row to CSV: %s", wErr)
//				}
//				actions[cert] = append(actions[cert], ROTAction{
//					Thumbprint: cert,
//					CertID:     certID,
//					StoreID:    store.ID,
//					StoreType:  store.Type,
//					StorePath:  store.Path,
//					AddCert:    false,
//					RemoveCert: true,
//				})
//			} else {
//				// Cert is not deployed to this store do nothing
//				row := []string{cert, certIDStr, certLookup.IssuedDN, certLookup.IssuerDN, store.ID, store.Type, store.Machine, store.Path, "false", "false", "false", GetCurrentTime()}
//				data = append(data, row)
//				wErr := csvWriter.Write(row)
//				if wErr != nil {
//					fmt.Printf("%s", wErr)
//					log.Printf("[ERROR] writing row to CSV: %s", wErr)
//				}
//			}
//		}
//	}
//	csvWriter.Flush()
//	ioErr := csvFile.Close()
//	if ioErr != nil {
//		fmt.Println(ioErr)
//		log.Printf("[ERROR] closing audit file: %s", ioErr)
//	}
//	fmt.Printf("Audit report written to %s\n", outpath)
//	return data, actions, nil
//}
//
//func reconcileRoots(actions map[string][]ROTAction, kfClient *api.Client, reportFile string, dryRun bool) error {
//	log.Printf("[DEBUG] Reconciling roots")
//	if len(actions) == 0 {
//		log.Printf("[INFO] No actions to take, roots are up-to-date.")
//		return nil
//	}
//	rFileName := fmt.Sprintf("%s_reconciled.csv", strings.Split(reportFile, ".csv")[0])
//	csvFile, fErr := os.Create(rFileName)
//	if fErr != nil {
//		fmt.Printf("[ERROR] creating reconciled report file: %s", fErr)
//	}
//	csvWriter := csv.NewWriter(csvFile)
//	cErr := csvWriter.Write(ReconciledAuditHeader)
//	if cErr != nil {
//		fmt.Printf("%s", cErr)
//		log.Fatalf("[ERROR] writing audit header: %s", cErr)
//	}
//	for thumbprint, action := range actions {
//
//		for _, a := range action {
//			if a.AddCert {
//				log.Printf("[INFO] Adding cert %s to store %s(%s)", thumbprint, a.StoreID, a.StorePath)
//				if !dryRun {
//					cStore := api.CertificateStore{
//						CertificateStoreId: a.StoreID,
//						Overwrite:          true,
//					}
//					var stores []api.CertificateStore
//					stores = append(stores, cStore)
//					schedule := &api.InventorySchedule{
//						Immediate: boolToPointer(true),
//					}
//					addReq := api.AddCertificateToStore{
//						CertificateId:     a.CertID,
//						CertificateStores: &stores,
//						InventorySchedule: schedule,
//					}
//					log.Printf("[DEBUG] Adding cert %s to store %s", thumbprint, a.StoreID)
//					log.Printf("[TRACE] Add request: %+v", addReq)
//					addReqJSON, _ := json.Marshal(addReq)
//					log.Printf("[TRACE] Add request JSON: %s", addReqJSON)
//					_, err := kfClient.AddCertificateToStores(&addReq)
//					if err != nil {
//						fmt.Printf("[ERROR] adding cert %s (%d) to store %s (%s): %s\n", a.Thumbprint, a.CertID, a.StoreID, a.StorePath, err)
//						continue
//					}
//				} else {
//					log.Printf("[INFO] DRY RUN: Would have added cert %s from store %s", thumbprint, a.StoreID)
//				}
//			} else if a.RemoveCert {
//				if !dryRun {
//					log.Printf("[INFO] Removing cert from store %s", a.StoreID)
//					cStore := api.CertificateStore{
//						CertificateStoreId: a.StoreID,
//						Alias:              a.Thumbprint,
//					}
//					var stores []api.CertificateStore
//					stores = append(stores, cStore)
//					schedule := &api.InventorySchedule{
//						Immediate: boolToPointer(true),
//					}
//					removeReq := api.RemoveCertificateFromStore{
//						CertificateId:     a.CertID,
//						CertificateStores: &stores,
//						InventorySchedule: schedule,
//					}
//					_, err := kfClient.RemoveCertificateFromStores(&removeReq)
//					if err != nil {
//						fmt.Printf("[ERROR] removing cert %s (ID: %d) from store %s (%s): %s\n", a.Thumbprint, a.CertID, a.StoreID, a.StorePath, err)
//					}
//				} else {
//					fmt.Printf("DRY RUN: Would have removed cert %s from store %s\n", thumbprint, a.StoreID)
//					log.Printf("[INFO] DRY RUN: Would have removed cert %s from store %s", thumbprint, a.StoreID)
//				}
//			}
//		}
//	}
//	return nil
//}
//
//func readCertsFile(certsFilePath string, kfclient *api.Client) (map[string]string, error) {
//	// Read in the cert CSV
//	csvFile, _ := os.Open(certsFilePath)
//	reader := csv.NewReader(bufio.NewReader(csvFile))
//	certEntries, _ := reader.ReadAll()
//	var certs = make(map[string]string)
//	for _, entry := range certEntries {
//		switch entry[0] {
//		case "CertID", "thumbprint", "id", "CertId", "Thumbprint":
//			continue // Skip header
//		}
//		certs[entry[0]] = entry[0]
//	}
//	return certs, nil
//}
//
//func isRootStore(st *api.GetCertificateStoreResponse, invs *[]api.CertStoreInventory, minCerts int, maxKeys int, maxLeaf int) bool {
//	leafCount := 0
//	keyCount := 0
//	certCount := 0
//	for _, inv := range *invs {
//		log.Printf("[DEBUG] inv: %v", inv)
//		certCount += len(inv.Certificates)
//
//		for _, cert := range inv.Certificates {
//			if cert.IssuedDN != cert.IssuerDN {
//				leafCount++
//			}
//			if inv.Parameters["PrivateKeyEntry"] == "Yes" {
//				keyCount++
//			}
//		}
//	}
//	if certCount < minCerts && minCerts >= 0 {
//		log.Printf("[DEBUG] Store %s has %d certs, less than the required count of %d", st.Id, certCount, minCerts)
//		return false
//	}
//	if leafCount > maxLeaf && maxLeaf >= 0 {
//		log.Printf("[DEBUG] Store %s has too many leaf certs", st.Id)
//		return false
//	}
//
//	if keyCount > maxKeys && maxKeys >= 0 {
//		log.Printf("[DEBUG] Store %s has too many keys", st.Id)
//		return false
//	}
//
//	return true
//}
//
//var (
//	rotCmd = &cobra.Command{
//		Use:   "rot",
//		Short: "Root of trust utility",
//		Long: `Root of trust allows you to manage your trusted roots using Keyfactor certificate stores.
//For example if you wish to add a list of "root" certs to a list of certificate stores you would simply generate and fill
//out the template CSV file. These template files can be generated with the following commands:
//kfutil stores rot generate-template --type certs
//kfutil stores rot generate-template --type stores
//Once those files are filled out you can use the following command to add the certs to the stores:
//kfutil stores rot audit --certs-file <certs-file> --stores-file <stores-file>
//Will generate a CSV report file 'rot_audit.csv' of what actions will be taken. If those actions are correct you can run
//the following command to actually perform the actions:
//kfutil stores rot reconcile --certs-file <certs-file> --stores-file <stores-file>
//OR if you want to use the audit report file generated you can run this command:
//kfutil stores rot reconcile --import-csv <audit-file>
//`,
//	}
//	rotAuditCmd = &cobra.Command{
//		Use:                    "audit",
//		Aliases:                nil,
//		SuggestFor:             nil,
//		Short:                  "Audit generates a CSV report of what actions will be taken based on input CSV files.",
//		Long:                   `Root of Trust Audit: Will read and parse inputs to generate a report of certs that need to be added or removed from the "root of trust" stores.`,
//		Example:                "",
//		ValidArgs:              nil,
//		ValidArgsFunction:      nil,
//		Args:                   nil,
//		ArgAliases:             nil,
//		BashCompletionFunction: "",
//		Deprecated:             "",
//		Annotations:            nil,
//		Version:                "",
//		PersistentPreRun:       nil,
//		PersistentPreRunE:      nil,
//		PreRun:                 nil,
//		PreRunE:                nil,
//		Run: func(cmd *cobra.Command, args []string) {
//			// Global flags
//			debugFlag, _ := cmd.Flags().GetBool("debugFlag")
//			configFile, _ := cmd.Flags().GetString("config")
//			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//			profile, _ := cmd.Flags().GetString("profile")
//			kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
//			kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
//			kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
//			kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
//			kfcAPIPath, _ := cmd.Flags().GetString("api-path")
//			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
//
//			debugModeEnabled := checkDebug(debugFlag)
//			log.Println("Debug mode enabled: ", debugModeEnabled)
//			var lookupFailures []string
//			kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
//			storesFile, _ := cmd.Flags().GetString("stores")
//			addRootsFile, _ := cmd.Flags().GetString("add-certs")
//			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
//			minCerts, _ := cmd.Flags().GetInt("min-certs")
//			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
//			maxKeys, _ := cmd.Flags().GetInt("max-keys")
//			dryRun, _ := cmd.Flags().GetBool("dry-run")
//			outpath, _ := cmd.Flags().GetString("outpath")
//			// Read in the stores CSV
//			log.Printf("[DEBUG] storesFile: %s", storesFile)
//			log.Printf("[DEBUG] addRootsFile: %s", addRootsFile)
//			log.Printf("[DEBUG] removeRootsFile: %s", removeRootsFile)
//			log.Printf("[DEBUG] dryRun: %t", dryRun)
//			// Read in the stores CSV
//			csvFile, _ := os.Open(storesFile)
//			reader := csv.NewReader(bufio.NewReader(csvFile))
//			storeEntries, _ := reader.ReadAll()
//			var stores = make(map[string]StoreCSVEntry)
//			validHeader := false
//			for _, entry := range storeEntries {
//				if strings.EqualFold(strings.Join(entry, ","), strings.Join(StoreHeader, ",")) {
//					validHeader = true
//					continue // Skip header
//				}
//				if !validHeader {
//					fmt.Printf("[ERROR] Invalid header in stores file. Expected: %s", strings.Join(StoreHeader, ","))
//					log.Fatalf("[ERROR] Stores CSV file is missing a valid header")
//				}
//				apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
//				if err != nil {
//					log.Printf("[ERROR] getting cert store: %s", err)
//					_ = append(lookupFailures, strings.Join(entry, ","))
//					continue
//				}
//
//				inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
//				if invErr != nil {
//					log.Printf("[ERROR] getting cert store inventory for: %s\n%s", entry[0], invErr)
//				}
//
//				if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
//					fmt.Printf("Store %s is not a root store, skipping.\n", entry[0])
//					log.Printf("[WARN] Store %s is not a root store", apiResp.Id)
//					continue
//				} else {
//					log.Printf("[INFO] Store %s is a root store", apiResp.Id)
//				}
//
//				stores[entry[0]] = StoreCSVEntry{
//					ID:          entry[0],
//					Type:        entry[1],
//					Machine:     entry[2],
//					Path:        entry[3],
//					Thumbprints: make(map[string]bool),
//					Serials:     make(map[string]bool),
//					Ids:         make(map[int]bool),
//				}
//				for _, cert := range *inventory {
//					thumb := cert.Thumbprints
//					for t, v := range thumb {
//						stores[entry[0]].Thumbprints[t] = v
//					}
//					for t, v := range cert.Serials {
//						stores[entry[0]].Serials[t] = v
//					}
//					for t, v := range cert.Ids {
//						stores[entry[0]].Ids[t] = v
//					}
//				}
//
//			}
//
//			// Read in the add addCerts CSV
//			var certsToAdd = make(map[string]string)
//			if addRootsFile != "" {
//				var rcfErr error
//				certsToAdd, rcfErr = readCertsFile(addRootsFile, kfClient)
//				if rcfErr != nil {
//					fmt.Printf("[ERROR] reading certs file %s: %s", addRootsFile, rcfErr)
//					log.Fatalf("[ERROR] reading addCerts file: %s", rcfErr)
//				}
//				addCertsJSON, _ := json.Marshal(certsToAdd)
//				log.Printf("[DEBUG] add certs JSON: %s", string(addCertsJSON))
//				log.Println("[DEBUG] AddCert ROT called")
//			} else {
//				log.Printf("[DEBUG] No addCerts file specified")
//				log.Printf("[DEBUG] No addCerts = %s", certsToAdd)
//			}
//
//			// Read in the remove removeCerts CSV
//			var certsToRemove = make(map[string]string)
//			if removeRootsFile != "" {
//				var rcfErr error
//				certsToRemove, rcfErr = readCertsFile(removeRootsFile, kfClient)
//				if rcfErr != nil {
//					fmt.Printf("[ERROR] reading removeCerts file %s: %s", removeRootsFile, rcfErr)
//					log.Fatalf("[ERROR] reading removeCerts file: %s", rcfErr)
//				}
//				removeCertsJSON, _ := json.Marshal(certsToRemove)
//				log.Printf("[DEBUG] remove certs JSON: %s", string(removeCertsJSON))
//			} else {
//				log.Printf("[DEBUG] No removeCerts file specified")
//				log.Printf("[DEBUG] No removeCerts = %s", certsToRemove)
//			}
//			_, _, gErr := generateAuditReport(certsToAdd, certsToRemove, stores, outpath, kfClient)
//			if gErr != nil {
//				log.Fatalf("[ERROR] generating audit report: %s", gErr)
//			}
//		},
//		RunE:                       nil,
//		PostRun:                    nil,
//		PostRunE:                   nil,
//		PersistentPostRun:          nil,
//		PersistentPostRunE:         nil,
//		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
//		CompletionOptions:          cobra.CompletionOptions{},
//		TraverseChildren:           false,
//		Hidden:                     false,
//		SilenceErrors:              false,
//		SilenceUsage:               false,
//		DisableFlagParsing:         false,
//		DisableAutoGenTag:          false,
//		DisableFlagsInUseLine:      false,
//		DisableSuggestions:         false,
//		SuggestionsMinimumDistance: 0,
//	}
//	rotReconcileCmd = &cobra.Command{
//		Use:        "reconcile",
//		Aliases:    nil,
//		SuggestFor: nil,
//		Short:      "Reconcile either takes in or will generate an audit report and then add/remove certs as needed.",
//		Long: `Root of Trust (rot): Will parse either a combination of CSV files that define certs to
//add and/or certs to remove with a CSV of certificate stores or an audit CSV file. If an audit CSV file is provided, the
//add and remove actions defined in the audit file will be immediately executed. If a combination of CSV files are provided,
//the utility will first generate an audit report and then execute the add/remove actions defined in the audit report.`,
//		Example:                "",
//		ValidArgs:              nil,
//		ValidArgsFunction:      nil,
//		Args:                   nil,
//		ArgAliases:             nil,
//		BashCompletionFunction: "",
//		Deprecated:             "",
//		Annotations:            nil,
//		Version:                "",
//		PersistentPreRun:       nil,
//		PersistentPreRunE:      nil,
//		PreRun:                 nil,
//		PreRunE:                nil,
//		Run: func(cmd *cobra.Command, args []string) {
//			// Global flags
//			debugFlag, _ := cmd.Flags().GetBool("debugFlag")
//			configFile, _ := cmd.Flags().GetString("config")
//			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//			profile, _ := cmd.Flags().GetString("profile")
//			kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
//			kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
//			kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
//			kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
//			kfcAPIPath, _ := cmd.Flags().GetString("api-path")
//			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
//
//			debugModeEnabled := checkDebug(debugFlag)
//
//			log.Println("Debug mode enabled: ", debugModeEnabled)
//
//			var lookupFailures []string
//			kfClient, _ := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
//			storesFile, _ := cmd.Flags().GetString("stores")
//			addRootsFile, _ := cmd.Flags().GetString("add-certs")
//			isCSV, _ := cmd.Flags().GetBool("import-csv")
//			reportFile, _ := cmd.Flags().GetString("input-file")
//			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
//			minCerts, _ := cmd.Flags().GetInt("min-certs")
//			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
//			maxKeys, _ := cmd.Flags().GetInt("max-keys")
//			dryRun, _ := cmd.Flags().GetBool("dry-run")
//			outpath, _ := cmd.Flags().GetString("outpath")
//
//			log.Printf("[DEBUG] configFile: %s", configFile)
//			log.Printf("[DEBUG] storesFile: %s", storesFile)
//			log.Printf("[DEBUG] addRootsFile: %s", addRootsFile)
//			log.Printf("[DEBUG] removeRootsFile: %s", removeRootsFile)
//			log.Printf("[DEBUG] dryRun: %t", dryRun)
//
//			// Parse existing audit report
//			if isCSV && reportFile != "" {
//				log.Printf("[DEBUG] isCSV: %t", isCSV)
//				log.Printf("[DEBUG] reportFile: %s", reportFile)
//				// Read in the CSV
//				csvFile, err := os.Open(reportFile)
//				if err != nil {
//					fmt.Printf("[ERROR] opening file: %s", err)
//					log.Fatalf("[ERROR] opening CSV file: %s", err)
//				}
//				validHeader := false
//
//				aCSV := csv.NewReader(csvFile)
//				aCSV.FieldsPerRecord = -1
//				inFile, cErr := aCSV.ReadAll()
//				if cErr != nil {
//					fmt.Printf("[ERROR] reading CSV file: %s", cErr)
//					log.Fatalf("[ERROR] reading CSV file: %s", cErr)
//				}
//				actions := make(map[string][]ROTAction)
//				fieldMap := make(map[int]string)
//				for i, field := range AuditHeader {
//					fieldMap[i] = field
//				}
//				for ri, row := range inFile {
//					if strings.EqualFold(strings.Join(row, ","), strings.Join(AuditHeader, ",")) {
//						validHeader = true
//						continue // Skip header
//					}
//					if !validHeader {
//						fmt.Printf("[ERROR] Invalid header in stores file. Expected: %s", strings.Join(AuditHeader, ","))
//						log.Fatalf("[ERROR] Stores CSV file is missing a valid header")
//					}
//					action := make(map[string]interface{})
//
//					for i, field := range row {
//						fieldInt, iErr := strconv.Atoi(field)
//						if iErr != nil {
//							log.Printf("[DEBUG] Field %s is not an int", field)
//							action[fieldMap[i]] = field
//						} else {
//							action[fieldMap[i]] = fieldInt
//						}
//
//					}
//
//					addCertStr, aOk := action["AddCert"].(string)
//					if !aOk {
//						addCertStr = ""
//					}
//					addCert, acErr := strconv.ParseBool(addCertStr)
//					if acErr != nil {
//						addCert = false
//					}
//
//					removeCertStr, rOk := action["RemoveCert"].(string)
//					if !rOk {
//						removeCertStr = ""
//					}
//					removeCert, rcErr := strconv.ParseBool(removeCertStr)
//					if rcErr != nil {
//						removeCert = false
//					}
//
//					sType, sOk := action["StoreType"].(string)
//					if !sOk {
//						sType = ""
//					}
//
//					sPath, pOk := action["Path"].(string)
//					if !pOk {
//						sPath = ""
//					}
//
//					tp, tpOk := action["Thumbprint"].(string)
//					if !tpOk {
//						tp = ""
//					}
//					cid, cidOk := action["CertID"].(int)
//					if !cidOk {
//						cid = -1
//					}
//
//					if !tpOk && !cidOk {
//						fmt.Printf("[ERROR] Missing Thumbprint or CertID for row %d in report file %s", ri, reportFile)
//						log.Printf("[ERROR] Invalid action: %v", action)
//						continue
//					}
//
//					sId, sIdOk := action["StoreID"].(string)
//					if !sIdOk {
//						fmt.Printf("[ERROR] Missing StoreID for row %d in report file %s", ri, reportFile)
//						log.Printf("[ERROR] Invalid action: %v", action)
//						continue
//					}
//					if cid == -1 && tp != "" {
//						certLookupReq := api.GetCertificateContextArgs{
//							IncludeMetadata:  boolToPointer(true),
//							IncludeLocations: boolToPointer(true),
//							CollectionId:     nil,
//							Thumbprint:       tp,
//							Id:               0,
//						}
//						certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
//						if err != nil {
//							fmt.Printf("[ERROR] looking up certificate %s: %s\n", tp, err)
//							log.Printf("[ERROR] looking up cert: %s\n%v", tp, err)
//							continue
//						}
//						cid = certLookup.Id
//					}
//
//					a := ROTAction{
//						StoreID:    sId,
//						StoreType:  sType,
//						StorePath:  sPath,
//						Thumbprint: tp,
//						CertID:     cid,
//						AddCert:    addCert,
//						RemoveCert: removeCert,
//					}
//
//					actions[a.Thumbprint] = append(actions[a.Thumbprint], a)
//				}
//				if len(actions) == 0 {
//					fmt.Println("No reconciliation actions to take, root stores are up-to-date. Exiting.")
//					return
//				}
//				rErr := reconcileRoots(actions, kfClient, reportFile, dryRun)
//				if rErr != nil {
//					fmt.Printf("[ERROR] reconciling roots: %s", rErr)
//					log.Fatalf("[ERROR] reconciling roots: %s", rErr)
//				}
//				defer csvFile.Close()
//
//				orchsURL := fmt.Sprintf("https://%s/Keyfactor/Portal/AgentJobStatus/Index", kfClient.Hostname)
//
//				fmt.Println(fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL))
//			} else {
//				// Read in the stores CSV
//				csvFile, _ := os.Open(storesFile)
//				reader := csv.NewReader(bufio.NewReader(csvFile))
//				storeEntries, _ := reader.ReadAll()
//				var stores = make(map[string]StoreCSVEntry)
//				for i, entry := range storeEntries {
//					if entry[0] == "StoreID" || entry[0] == "StoreId" || i == 0 {
//						continue // Skip header
//					}
//					apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
//					if err != nil {
//						log.Printf("[ERROR] getting cert store: %s", err)
//						lookupFailures = append(lookupFailures, entry[0])
//						continue
//					}
//					inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
//					if invErr != nil {
//						log.Fatalf("[ERROR] getting cert store inventory: %s", invErr)
//					}
//
//					if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
//						log.Printf("[WARN] Store %s is not a root store", apiResp.Id)
//						continue
//					} else {
//						log.Printf("[INFO] Store %s is a root store", apiResp.Id)
//					}
//
//					stores[entry[0]] = StoreCSVEntry{
//						ID:          entry[0],
//						Type:        entry[1],
//						Machine:     entry[2],
//						Path:        entry[3],
//						Thumbprints: make(map[string]bool),
//						Serials:     make(map[string]bool),
//						Ids:         make(map[int]bool),
//					}
//					for _, cert := range *inventory {
//						thumb := cert.Thumbprints
//						for t, v := range thumb {
//							stores[entry[0]].Thumbprints[t] = v
//						}
//						for t, v := range cert.Serials {
//							stores[entry[0]].Serials[t] = v
//						}
//						for t, v := range cert.Ids {
//							stores[entry[0]].Ids[t] = v
//						}
//					}
//
//				}
//				if len(lookupFailures) > 0 {
//					fmt.Printf("[ERROR] the following stores were not found: %s", strings.Join(lookupFailures, ","))
//					log.Fatalf("[ERROR] the following stores were not found: %s", strings.Join(lookupFailures, ","))
//				}
//				if len(stores) == 0 {
//					fmt.Println("[ERROR] no root stores found. Exiting.")
//					log.Fatalf("[ERROR] No root stores found. Exiting.")
//				}
//				// Read in the add addCerts CSV
//				var certsToAdd = make(map[string]string)
//				if addRootsFile != "" {
//					certsToAdd, _ = readCertsFile(addRootsFile, kfClient)
//					log.Printf("[DEBUG] ROT add certs called")
//				} else {
//					log.Printf("[INFO] No addCerts file specified")
//				}
//
//				// Read in the remove removeCerts CSV
//				var certsToRemove = make(map[string]string)
//				if removeRootsFile != "" {
//					certsToRemove, _ = readCertsFile(removeRootsFile, kfClient)
//					log.Printf("[DEBUG] ROT remove certs called")
//				} else {
//					log.Printf("[DEBUG] No removeCerts file specified")
//				}
//				_, actions, err := generateAuditReport(certsToAdd, certsToRemove, stores, outpath, kfClient)
//				if err != nil {
//					log.Fatalf("[ERROR] generating audit report: %s", err)
//				}
//				if len(actions) == 0 {
//					fmt.Println("No reconciliation actions to take, root stores are up-to-date. Exiting.")
//					return
//				}
//				rErr := reconcileRoots(actions, kfClient, reportFile, dryRun)
//				if rErr != nil {
//					fmt.Printf("[ERROR] reconciling roots: %s", rErr)
//					log.Fatalf("[ERROR] reconciling roots: %s", rErr)
//				}
//				if lookupFailures != nil {
//					fmt.Printf("The following stores could not be found: %s", strings.Join(lookupFailures, ","))
//				}
//				orchsURL := fmt.Sprintf("https://%s/Keyfactor/Portal/AgentJobStatus/Index", kfClient.Hostname)
//
//				fmt.Println(fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL))
//			}
//
//		},
//		RunE:                       nil,
//		PostRun:                    nil,
//		PostRunE:                   nil,
//		PersistentPostRun:          nil,
//		PersistentPostRunE:         nil,
//		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
//		CompletionOptions:          cobra.CompletionOptions{},
//		TraverseChildren:           false,
//		Hidden:                     false,
//		SilenceErrors:              false,
//		SilenceUsage:               false,
//		DisableFlagParsing:         false,
//		DisableAutoGenTag:          false,
//		DisableFlagsInUseLine:      false,
//		DisableSuggestions:         false,
//		SuggestionsMinimumDistance: 0,
//	}
//	rotGenStoreTemplateCmd = &cobra.Command{
//		Use:                    "generate-template",
//		Aliases:                nil,
//		SuggestFor:             nil,
//		Short:                  "For generating Root Of Trust template(s)",
//		Long:                   `Root Of Trust: Will parse a CSV and attempt to deploy a cert or set of certs into a list of cert stores.`,
//		Example:                "",
//		ValidArgs:              nil,
//		ValidArgsFunction:      nil,
//		Args:                   nil,
//		ArgAliases:             nil,
//		BashCompletionFunction: "",
//		Deprecated:             "",
//		Annotations:            nil,
//		Version:                "",
//		PersistentPreRun:       nil,
//		PersistentPreRunE:      nil,
//		PreRun:                 nil,
//		PreRunE:                nil,
//		Run: func(cmd *cobra.Command, args []string) {
//			// Global flags
//			debugFlag, _ := cmd.Flags().GetBool("debugFlag")
//			configFile, _ := cmd.Flags().GetString("config")
//			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
//			profile, _ := cmd.Flags().GetString("profile")
//			kfcHostName, _ := cmd.Flags().GetString("kfcHostName")
//			kfcUsername, _ := cmd.Flags().GetString("kfcUsername")
//			kfcPassword, _ := cmd.Flags().GetString("kfcPassword")
//			kfcDomain, _ := cmd.Flags().GetString("kfcDomain")
//			kfcAPIPath, _ := cmd.Flags().GetString("api-path")
//			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
//
//			debugModeEnabled := checkDebug(debugFlag)
//			log.Println("Debug mode enabled: ", debugModeEnabled)
//
//			templateType, _ := cmd.Flags().GetString("type")
//			format, _ := cmd.Flags().GetString("format")
//			outPath, _ := cmd.Flags().GetString("outpath")
//			storeType, _ := cmd.Flags().GetStringSlice("store-type")
//			containerName, _ := cmd.Flags().GetStringSlice("container-name")
//			collection, _ := cmd.Flags().GetStringSlice("collection")
//			subjectName, _ := cmd.Flags().GetStringSlice("cn")
//			stID := -1
//			var storeData []api.GetCertificateStoreResponse
//			var csvStoreData [][]string
//			var csvCertData [][]string
//			var rowLookup = make(map[string]bool)
//			kfClient, cErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
//			if len(storeType) != 0 {
//				for _, s := range storeType {
//					if cErr != nil {
//						log.Fatalf("[ERROR] creating client: %s", cErr)
//					}
//					var sType *api.CertificateStoreType
//					var stErr error
//					if s == "all" {
//						sType = &api.CertificateStoreType{
//							Name:                "",
//							ShortName:           "",
//							Capability:          "",
//							StoreType:           0,
//							ImportType:          0,
//							LocalStore:          false,
//							SupportedOperations: nil,
//							Properties:          nil,
//							EntryParameters:     nil,
//							PasswordOptions:     nil,
//							StorePathType:       "",
//							StorePathValue:      "",
//							PrivateKeyAllowed:   "",
//							JobProperties:       nil,
//							ServerRequired:      false,
//							PowerShell:          false,
//							BlueprintAllowed:    false,
//							CustomAliasAllowed:  "",
//							ServerRegistration:  0,
//							InventoryEndpoint:   "",
//							InventoryJobType:    "",
//							ManagementJobType:   "",
//							DiscoveryJobType:    "",
//							EnrollmentJobType:   "",
//						}
//					} else {
//						// check if s is an int
//						sInt, err := strconv.Atoi(s)
//						if err == nil {
//							sType, stErr = kfClient.GetCertificateStoreTypeById(sInt)
//						} else {
//							sType, stErr = kfClient.GetCertificateStoreTypeByName(s)
//						}
//						if stErr != nil {
//							fmt.Printf("[ERROR] getting store type '%s'. %s\n", s, stErr)
//							continue
//						}
//						stID = sType.StoreType // This is the template type ID
//					}
//
//					if stID >= 0 || s == "all" {
//						log.Printf("[DEBUG] Store type ID: %d\n", stID)
//						params := make(map[string]interface{})
//						stores, sErr := kfClient.ListCertificateStores(&params)
//						if sErr != nil {
//							fmt.Printf("[ERROR] getting certificate stores of type '%s': %s\n", s, sErr)
//							log.Fatalf("[ERROR] getting certificate stores of type '%s': %s", s, sErr)
//						}
//						for _, store := range *stores {
//							if store.CertStoreType == stID || s == "all" {
//								storeData = append(storeData, store)
//								if !rowLookup[store.Id] {
//									lineData := []string{
//										//"StoreID", "StoreType", "StoreMachine", "StorePath", "ContainerId"
//										store.Id, fmt.Sprintf("%s", sType.ShortName), store.ClientMachine, store.StorePath, fmt.Sprintf("%d", store.ContainerId), store.ContainerName, GetCurrentTime(),
//									}
//									csvStoreData = append(csvStoreData, lineData)
//									rowLookup[store.Id] = true
//								}
//							}
//						}
//					}
//				}
//				fmt.Println("Done")
//			}
//			if len(containerName) != 0 {
//				for _, c := range containerName {
//
//					if cErr != nil {
//						log.Fatalf("[ERROR] creating client: %s", cErr)
//					}
//					cStoresResp, scErr := kfClient.GetCertificateStoreByContainerID(c)
//					if scErr != nil {
//						fmt.Printf("[ERROR] getting store container: %s\n", scErr)
//					}
//					if cStoresResp != nil {
//						for _, store := range *cStoresResp {
//							sType, stErr := kfClient.GetCertificateStoreType(store.CertStoreType)
//							if stErr != nil {
//								fmt.Printf("[ERROR] getting store type: %s\n", stErr)
//								continue
//							}
//							storeData = append(storeData, store)
//							if !rowLookup[store.Id] {
//								lineData := []string{
//									// "StoreID", "StoreType", "StoreMachine", "StorePath", "ContainerId"
//									store.Id, sType.ShortName, store.ClientMachine, store.StorePath, fmt.Sprintf("%d", store.ContainerId), store.ContainerName, GetCurrentTime(),
//								}
//								csvStoreData = append(csvStoreData, lineData)
//								rowLookup[store.Id] = true
//							}
//						}
//
//					}
//				}
//			}
//			if len(collection) != 0 {
//				for _, c := range collection {
//					if cErr != nil {
//						fmt.Println("[ERROR] connecting to Keyfactor. Please check your configuration and try again.")
//						log.Fatalf("[ERROR] creating client: %s", cErr)
//					}
//					q := make(map[string]string)
//					q["collection"] = c
//					certsResp, scErr := kfClient.ListCertificates(q)
//					if scErr != nil {
//						fmt.Printf("No certificates found in collection: %s\n", scErr)
//					}
//					if certsResp != nil {
//						for _, cert := range certsResp {
//							if !rowLookup[cert.Thumbprint] {
//								lineData := []string{
//									// "Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"
//									cert.Thumbprint, cert.IssuedCN, cert.IssuerDN, fmt.Sprintf("%d", cert.Id), fmt.Sprintf("%v", cert.Locations), GetCurrentTime(),
//								}
//								csvCertData = append(csvCertData, lineData)
//								rowLookup[cert.Thumbprint] = true
//							}
//						}
//
//					}
//				}
//			}
//			if len(subjectName) != 0 {
//				for _, s := range subjectName {
//					if cErr != nil {
//						fmt.Println("[ERROR] connecting to Keyfactor. Please check your configuration and try again.")
//						log.Fatalf("[ERROR] creating client: %s", cErr)
//					}
//					q := make(map[string]string)
//					q["subject"] = s
//					certsResp, scErr := kfClient.ListCertificates(q)
//					if scErr != nil {
//						fmt.Printf("No certificates found with CN: %s\n", scErr)
//					}
//					if certsResp != nil {
//						for _, cert := range certsResp {
//							if !rowLookup[cert.Thumbprint] {
//								locationsFormatted := ""
//								for _, loc := range cert.Locations {
//									locationsFormatted += fmt.Sprintf("%s:%s\n", loc.StoreMachine, loc.StorePath)
//								}
//								lineData := []string{
//									// "Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"
//									cert.Thumbprint, cert.IssuedCN, cert.IssuerDN, fmt.Sprintf("%d", cert.Id), locationsFormatted, GetCurrentTime(),
//								}
//								csvCertData = append(csvCertData, lineData)
//								rowLookup[cert.Thumbprint] = true
//							}
//						}
//
//					}
//				}
//			}
//			// Create CSV template file
//
//			var filePath string
//			if outPath != "" {
//				filePath = outPath
//			} else {
//				filePath = fmt.Sprintf("%s_template.%s", templateType, format)
//			}
//			file, err := os.Create(filePath)
//			if err != nil {
//				fmt.Printf("[ERROR] creating file: %s", err)
//				log.Fatal("Cannot create file", err)
//			}
//
//			switch format {
//			case "csv":
//				writer := csv.NewWriter(file)
//				var data [][]string
//				switch templateType {
//				case "stores":
//					data = append(data, StoreHeader)
//					if len(csvStoreData) != 0 {
//						data = append(data, csvStoreData...)
//					}
//				case "certs":
//					data = append(data, CertHeader)
//					if len(csvCertData) != 0 {
//						data = append(data, csvCertData...)
//					}
//				case "actions":
//					data = append(data, AuditHeader)
//				}
//				csvErr := writer.WriteAll(data)
//				if csvErr != nil {
//					fmt.Println(csvErr)
//				}
//				defer file.Close()
//
//			case "json":
//				writer := bufio.NewWriter(file)
//				_, err := writer.WriteString("StoreID,StoreType,StoreMachine,StorePath")
//				if err != nil {
//					log.Fatal("Cannot write to file", err)
//				}
//			}
//			fmt.Printf("Template file created at %s.\n", filePath)
//		},
//		RunE:                       nil,
//		PostRun:                    nil,
//		PostRunE:                   nil,
//		PersistentPostRun:          nil,
//		PersistentPostRunE:         nil,
//		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
//		CompletionOptions:          cobra.CompletionOptions{},
//		TraverseChildren:           false,
//		Hidden:                     false,
//		SilenceErrors:              false,
//		SilenceUsage:               false,
//		DisableFlagParsing:         false,
//		DisableAutoGenTag:          false,
//		DisableFlagsInUseLine:      false,
//		DisableSuggestions:         false,
//		SuggestionsMinimumDistance: 0,
//	}
//)
//
//func init() {
//	log.SetFlags(log.LstdFlags | log.Lshortfile)
//	log.SetOutput(os.Stdout)
//	var (
//		stores          string
//		addCerts        string
//		removeCerts     string
//		minCertsInStore int
//		maxPrivateKeys  int
//		maxLeaves       int
//		tType           = tTypeCerts
//		outPath         string
//		outputFormat    string
//		inputFile       string
//		storeTypes      []string
//		containerNames  []string
//		collections     []string
//		subjectNames    []string
//	)
//
//	storesCmd.AddCommand(rotCmd)
//
//	// Root of trust `audit` command
//	rotCmd.AddCommand(rotAuditCmd)
//	rotAuditCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
//	rotAuditCmd.Flags().StringVarP(&addCerts, "add-certs", "a", "",
//		"CSV file containing cert(s) to enroll into the defined cert stores")
//	rotAuditCmd.Flags().StringVarP(&removeCerts, "remove-certs", "r", "",
//		"CSV file containing cert(s) to remove from the defined cert stores")
//	rotAuditCmd.Flags().IntVarP(&minCertsInStore, "min-certs", "m", -1,
//		"The minimum number of certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
//	rotAuditCmd.Flags().IntVarP(&maxPrivateKeys, "max-keys", "k", -1,
//		"The max number of private keys that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
//	rotAuditCmd.Flags().IntVarP(&maxLeaves, "max-leaf-certs", "l", -1,
//		"The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
//	rotAuditCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
//	rotAuditCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
//		"Path to write the audit report file to. If not specified, the file will be written to the current directory.")
//
//	// Root of trust `reconcile` command
//	rotCmd.AddCommand(rotReconcileCmd)
//	rotReconcileCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
//	rotReconcileCmd.Flags().StringVarP(&addCerts, "add-certs", "a", "",
//		"CSV file containing cert(s) to enroll into the defined cert stores")
//	rotReconcileCmd.Flags().StringVarP(&removeCerts, "remove-certs", "r", "",
//		"CSV file containing cert(s) to remove from the defined cert stores")
//	rotReconcileCmd.Flags().IntVarP(&minCertsInStore, "min-certs", "m", -1,
//		"The minimum number of certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
//	rotReconcileCmd.Flags().IntVarP(&maxPrivateKeys, "max-keys", "k", -1,
//		"The max number of private keys that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
//	rotReconcileCmd.Flags().IntVarP(&maxLeaves, "max-leaf-certs", "l", -1,
//		"The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
//	rotReconcileCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
//	rotReconcileCmd.Flags().BoolP("import-csv", "v", false, "Import an audit report file in CSV format.")
//	rotReconcileCmd.Flags().StringVarP(&inputFile, "input-file", "i", reconcileDefaultFileName,
//		"Path to a file generated by 'stores rot audit' command.")
//	rotReconcileCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
//		"Path to write the audit report file to. If not specified, the file will be written to the current directory.")
//	//rotReconcileCmd.MarkFlagsRequiredTogether("add-certs", "stores")
//	//rotReconcileCmd.MarkFlagsRequiredTogether("remove-certs", "stores")
//	rotReconcileCmd.MarkFlagsMutuallyExclusive("add-certs", "import-csv")
//	rotReconcileCmd.MarkFlagsMutuallyExclusive("remove-certs", "import-csv")
//	rotReconcileCmd.MarkFlagsMutuallyExclusive("stores", "import-csv")
//
//	// Root of trust `generate` command
//	rotCmd.AddCommand(rotGenStoreTemplateCmd)
//	rotGenStoreTemplateCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
//		"Path to write the template file to. If not specified, the file will be written to the current directory.")
//	rotGenStoreTemplateCmd.Flags().StringVarP(&outputFormat, "format", "f", "csv",
//		"The type of template to generate. Only `csv` is supported at this time.")
//	rotGenStoreTemplateCmd.Flags().Var(&tType, "type",
//		`The type of template to generate. Only "certs|stores|actions" are supported at this time.`)
//	rotGenStoreTemplateCmd.Flags().StringSliceVar(&storeTypes, "store-type", []string{}, "Multi value flag. Attempt to pre-populate the stores template with the certificate stores matching specified store types. If not specified, the template will be empty.")
//	rotGenStoreTemplateCmd.Flags().StringSliceVar(&containerNames, "container-name", []string{}, "Multi value flag. Attempt to pre-populate the stores template with the certificate stores matching specified container types. If not specified, the template will be empty.")
//	rotGenStoreTemplateCmd.Flags().StringSliceVar(&subjectNames, "cn", []string{}, "Subject name(s) to pre-populate the 'certs' template with. If not specified, the template will be empty. Does not work with SANs.")
//	rotGenStoreTemplateCmd.Flags().StringSliceVar(&collections, "collection", []string{}, "Certificate collection name(s) to pre-populate the stores template with. If not specified, the template will be empty.")
//
//	rotGenStoreTemplateCmd.RegisterFlagCompletionFunc("type", templateTypeCompletion)
//	rotGenStoreTemplateCmd.MarkFlagRequired("type")
//}
