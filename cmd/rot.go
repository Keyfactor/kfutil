// Package cmd /*
package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
)

type templateType string
type StoreCSVEntry struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Machine     string          `json:"address"`
	Path        string          `json:"path"`
	Thumbprints map[string]bool `json:"thumbprints,omitempty"`
	Serials     map[string]bool `json:"serials,omitempty"`
	Ids         map[int]bool    `json:"ids,omitempty"`
}
type ROTCert struct {
	ID         int                        `json:"id,omitempty"`
	ThumbPrint string                     `json:"thumbprint,omitempty"`
	CN         string                     `json:"cn,omitempty"`
	Locations  []api.CertificateLocations `json:"locations,omitempty"`
}
type ROTAction struct {
	StoreID    string `json:"store_id,omitempty"`
	StoreType  string `json:"store_type,omitempty"`
	StorePath  string `json:"store_path,omitempty"`
	Thumbprint string `json:"thumbprint,omitempty"`
	CertID     int    `json:"cert_id,omitempty"`
	Add        bool   `json:"add,omitempty"`
	Remove     bool   `json:"remove,omitempty"`
}

const (
	tTypeCerts   templateType = "certs"
	tTypeStores  templateType = "stores"
	tTypeActions templateType = "actions"
)

var (
	AuditHeader = []string{"Thumbprint", "CertID", "StoreID", "StoreType", "Machine", "Path", "AddCert", "RemoveCert", "Deployed"}
	StoreHeader = []string{"StoreID", "StoreType", "StoreMachine", "StorePath"}
	CertHeader  = []string{"Thumbprint"}
)

// String is used both by fmt.Print and by Cobra in help text
func (e *templateType) String() string {
	return string(*e)
}

// Set must have pointer receiver, so it doesn't change the value of a copy
func (e *templateType) Set(v string) error {
	switch v {
	case "certs", "stores", "actions":
		*e = templateType(v)
		return nil
	default:
		return errors.New(`must be one of "certs", "stores", or "actions"`)
	}
}

// Type is only used in help text
func (e *templateType) Type() string {
	return "string"
}

func templateTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"certs\tGenerates template CSV for certificate input to be used w/ `--add-certs` or `--remove-certs`",
		"stores\tGenerates template CSV for certificate input to be used w/ `--stores`",
		"actions\tGenerates template CSV for certificate input to be used w/ `--actions`",
	}, cobra.ShellCompDirectiveDefault
}

func generateAuditReport(addCerts map[string]string, removeCerts map[string]string, stores map[string]StoreCSVEntry, kfClient *api.Client) ([][]string, map[string]ROTAction, error) {
	log.Println("[DEBUG] generateAuditReport called")
	var (
		data [][]string
	)

	data = append(data, AuditHeader)
	csvFile, fErr := os.Create("rot_audit.csv")
	if fErr != nil {
		fmt.Printf("%s", fErr)
		log.Fatalf("[ERROR] Error creating audit file: %s", fErr)
	}
	csvWriter := csv.NewWriter(csvFile)
	cErr := csvWriter.Write(AuditHeader)
	if cErr != nil {
		fmt.Printf("%s", cErr)
		log.Fatalf("[ERROR] Error writing audit header: %s", cErr)
	}
	actions := make(map[string]ROTAction)

	for _, cert := range addCerts {
		certLookupReq := api.GetCertificateContextArgs{
			IncludeMetadata:  boolToPointer(true),
			IncludeLocations: boolToPointer(true),
			CollectionId:     nil,
			Thumbprint:       cert,
			Id:               0,
		}
		certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
		if err != nil {
			log.Printf("[ERROR] Error looking up cert: %s\n%v", cert, err)
			continue
		}
		certID := certLookup.Id
		certIDStr := strconv.Itoa(certID)
		for _, store := range stores {
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is already in the store do nothing
				row := []string{cert, certIDStr, store.ID, store.Type, store.Machine, store.Path, "false", "false", "true"}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					fmt.Printf("%s", wErr)
					log.Printf("[ERROR] Error writing audit row: %s", wErr)
				}
			} else {
				// Cert is not deployed to this store and will need to be added
				row := []string{cert, certIDStr, store.ID, store.Type, store.Machine, store.Path, "true", "false", "false"}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					fmt.Printf("%s", wErr)
					log.Printf("[ERROR] Error writing audit row: %s", wErr)
				}
				actions[cert] = ROTAction{
					Thumbprint: cert,
					CertID:     certID,
					StoreID:    store.ID,
					StoreType:  store.Type,
					StorePath:  store.Path,
					Add:        true,
					Remove:     false,
				}
			}
		}
	}
	for _, cert := range removeCerts {
		certLookupReq := api.GetCertificateContextArgs{
			IncludeMetadata:  boolToPointer(true),
			IncludeLocations: boolToPointer(true),
			CollectionId:     nil,
			Thumbprint:       cert,
			Id:               0,
		}
		certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
		if err != nil {
			log.Printf("[ERROR] Error looking up cert: %s", err)
			continue
		}
		certID := certLookup.Id
		certIDStr := strconv.Itoa(certID)
		for _, store := range stores {
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is deployed to this store and will need to be removed
				row := []string{cert, certIDStr, store.ID, store.Type, store.Machine, store.Path, "false", "true", "true"}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					fmt.Printf("%s", wErr)
					log.Printf("[ERROR] Error writing row to CSV: %s", wErr)
				}
				actions[cert] = ROTAction{
					Thumbprint: cert,
					StoreID:    store.ID,
					StoreType:  store.Type,
					Add:        false,
					Remove:     true,
				}
			} else {
				// Cert is not deployed to this store do nothing
				row := []string{cert, certIDStr, store.ID, store.Type, store.Machine, store.Path, "false", "false", "false"}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					fmt.Printf("%s", wErr)
					log.Printf("[ERROR] Error writing row to CSV: %s", wErr)
				}
			}
		}
	}
	csvWriter.Flush()
	ioErr := csvFile.Close()
	if ioErr != nil {
		fmt.Println(ioErr)
		log.Printf("[ERROR] Error closing audit file: %s", ioErr)
	}
	fmt.Println("Audit report written to rot_audit.csv")
	return data, actions, nil
}

func reconcileRoots(actions map[string]ROTAction, kfClient *api.Client, dryRun bool) error {
	log.Printf("[DEBUG] Reconciling roots")
	if len(actions) == 0 {
		log.Printf("[INFO] No actions to take, roots are up-to-date.")
		return nil
	}
	for thumbprint, action := range actions {
		if action.Add {
			log.Printf("[INFO] Adding cert %s to store %s(%s)", thumbprint, action.StoreID, action.StorePath)
			if !dryRun {
				cStore := api.CertificateStore{
					CertificateStoreId: action.StoreID,
					Overwrite:          true,
				}
				var stores []api.CertificateStore
				stores = append(stores, cStore)
				schedule := &api.InventorySchedule{
					Immediate: boolToPointer(true),
				}
				addReq := api.AddCertificateToStore{
					CertificateId:     action.CertID,
					CertificateStores: &stores,
					InventorySchedule: schedule,
				}
				_, err := kfClient.AddCertificateToStores(&addReq)
				if err != nil {
					log.Fatalf("[ERROR] Error adding cert to store: %s", err)
				}
			} else {
				log.Printf("[INFO] DRY RUN: Would have added cert %s from store %s", thumbprint, action.StoreID)
			}
		} else if action.Remove {
			if !dryRun {
				log.Printf("[INFO] Removing cert from store %s", action.StoreID)
				cStore := api.CertificateStore{
					CertificateStoreId: action.StoreID,
				}
				var stores []api.CertificateStore
				stores = append(stores, cStore)
				schedule := &api.InventorySchedule{
					Immediate: boolToPointer(true),
				}
				removeReq := api.RemoveCertificateFromStore{
					CertificateId:     action.CertID,
					Alias:             fmt.Sprintf("KeyfactorAdd%d", action.CertID),
					CertificateStores: &stores,
					InventorySchedule: schedule,
				}
				_, err := kfClient.RemoveCertificateFromStores(&removeReq)
				if err != nil {
					log.Fatalf("[ERROR] Error removing cert from store: %s", err)
				}
			} else {
				log.Printf("[INFO] DRY RUN: Would have removed cert %s from store %s", thumbprint, action.StoreID)
			}
		}
	}
	return nil
}

func readCertsFile(certsFilePath string, kfclient *api.Client) (map[string]string, error) {
	// Read in the cert CSV
	csvFile, _ := os.Open(certsFilePath)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	certEntries, _ := reader.ReadAll()
	var certs = make(map[string]string)
	for _, entry := range certEntries {
		switch entry[0] {
		case "CertID", "thumbprint", "id", "CertId", "Thumbprint":
			continue // Skip header
		}
		certs[entry[0]] = entry[0]
	}
	return certs, nil
}

func isRootStore(st *api.GetStoreByIDResp, invs *[]api.CertStoreInventory, minCerts int, maxKeys int, maxLeaf int) bool {
	leafCount := 0
	keyCount := 0
	certCount := 0
	for _, inv := range *invs {
		log.Printf("[DEBUG] inv: %v", inv)
		certCount += len(inv.Certificates)

		for _, cert := range inv.Certificates {
			if cert.IssuedDN != cert.IssuerDN {
				leafCount++
			}
			if inv.Parameters["PrivateKeyEntry"] == "Yes" {
				keyCount++
			}
		}
	}
	if certCount < minCerts && minCerts >= 0 {
		log.Printf("[DEBUG] Store %s has %d certs, less than the required count of %d", st.Id, certCount, minCerts)
		return false
	}
	if leafCount > maxLeaf && maxLeaf >= 0 {
		log.Printf("[DEBUG] Store %s has too many leaf certs", st.Id)
		return false
	}

	if keyCount > maxKeys && maxKeys >= 0 {
		log.Printf("[DEBUG] Store %s has too many keys", st.Id)
		return false
	}

	return true
}

var (
	rotCmd = &cobra.Command{
		Use:   "rot",
		Short: "Root of trust utility",
		Long:  `Root of trust allows you to manage your trusted roots using Keyfactor certificate stores.`,
	}
	rotAuditCmd = &cobra.Command{
		Use:                    "audit",
		Aliases:                nil,
		SuggestFor:             nil,
		Short:                  "Root Of Trust Audit",
		Long:                   `Root Of Trust Audit: Will read and parse inputs to generate a report of certs that need to be added or removed from the "root of trust" stores.`,
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
			var lookupFailures []string
			kfClient, _ := initClient()
			storesFile, _ := cmd.Flags().GetString("stores")
			addRootsFile, _ := cmd.Flags().GetString("add-certs")
			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
			minCerts, _ := cmd.Flags().GetInt("min-certs")
			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
			maxKeys, _ := cmd.Flags().GetInt("max-keys")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			// Read in the stores CSV
			log.Printf("[DEBUG] storesFile: %s", storesFile)
			log.Printf("[DEBUG] addRootsFile: %s", addRootsFile)
			log.Printf("[DEBUG] removeRootsFile: %s", removeRootsFile)
			log.Printf("[DEBUG] dryRun: %t", dryRun)
			// Read in the stores CSV
			csvFile, _ := os.Open(storesFile)
			reader := csv.NewReader(bufio.NewReader(csvFile))
			storeEntries, _ := reader.ReadAll()
			var stores = make(map[string]StoreCSVEntry)
			validHeader := false
			for _, entry := range storeEntries {
				if strings.EqualFold(strings.Join(entry, ","), strings.Join(StoreHeader, ",")) {
					validHeader = true
					continue // Skip header
				}
				if !validHeader {
					fmt.Printf("[ERROR] Invalid header in stores file. Expected: %s", strings.Join(StoreHeader, ","))
					log.Fatalf("[ERROR] Stores CSV file is missing a valid header")
				}
				apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
				if err != nil {
					log.Printf("[ERROR] Error getting cert store: %s", err)
					_ = append(lookupFailures, strings.Join(entry, ","))
					continue
				}

				inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
				if invErr != nil {
					log.Printf("[ERROR] Error getting cert store inventory for: %s\n%s", entry[0], invErr)
				}

				if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
					log.Printf("[WARN] Store %s is not a root store", apiResp.Id)
					continue
				} else {
					log.Printf("[INFO] Store %s is a root store", apiResp.Id)
				}

				stores[entry[0]] = StoreCSVEntry{
					ID:          entry[0],
					Type:        entry[1],
					Machine:     entry[2],
					Path:        entry[3],
					Thumbprints: make(map[string]bool),
					Serials:     make(map[string]bool),
					Ids:         make(map[int]bool),
				}
				for _, cert := range *inventory {
					thumb := cert.Thumbprints
					for t, v := range thumb {
						stores[entry[0]].Thumbprints[t] = v
					}
					for t, v := range cert.Serials {
						stores[entry[0]].Serials[t] = v
					}
					for t, v := range cert.Ids {
						stores[entry[0]].Ids[t] = v
					}
				}

			}

			// Read in the add addCerts CSV
			var certsToAdd = make(map[string]string)
			if addRootsFile != "" {
				certsToAdd, _ = readCertsFile(addRootsFile, kfClient)
				//if err != nil {
				//	log.Fatalf("Error reading addCerts file: %s", err)
				//}
				addCertsJSON, _ := json.Marshal(certsToAdd)
				log.Printf("[DEBUG] add certs JSON: %s", string(addCertsJSON))
				log.Println("[DEBUG] Add ROT called")
			} else {
				log.Printf("[DEBUG] No addCerts file specified")
				log.Printf("[DEBUG] No addCerts = %s", certsToAdd)
			}

			// Read in the remove removeCerts CSV
			var certsToRemove = make(map[string]string)
			if removeRootsFile != "" {
				certsToRemove, rErr := readCertsFile(removeRootsFile, kfClient)
				if rErr != nil {
					fmt.Printf("Error reading removeCerts file: %s", rErr)
					log.Fatalf("Error reading removeCerts file: %s", rErr)
				}
				removeCertsJSON, _ := json.Marshal(certsToRemove)
				fmt.Println(string(removeCertsJSON))
				fmt.Println("remove rot called")
			} else {
				log.Printf("[DEBUG] No removeCerts file specified")
				log.Printf("[DEBUG] No removeCerts = %s", certsToRemove)
			}
			_, _, gErr := generateAuditReport(certsToAdd, certsToRemove, stores, kfClient)
			if gErr != nil {
				log.Fatalf("Error generating audit report: %s", gErr)
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
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}
	rotReconcileCmd = &cobra.Command{
		Use:                    "reconcile",
		Aliases:                nil,
		SuggestFor:             nil,
		Short:                  "Root Of Trust",
		Long:                   `Root Of Trust: Will parse a CSV and attempt to enroll a cert or set of certs into a list of cert stores.`,
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
			var lookupFailures []string
			kfClient, _ := initClient()
			storesFile, _ := cmd.Flags().GetString("stores")
			addRootsFile, _ := cmd.Flags().GetString("add-certs")
			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
			minCerts, _ := cmd.Flags().GetInt("min-certs")
			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
			maxKeys, _ := cmd.Flags().GetInt("max-keys")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			log.Printf("[DEBUG] storesFile: %s", storesFile)
			log.Printf("[DEBUG] addRootsFile: %s", addRootsFile)
			log.Printf("[DEBUG] removeRootsFile: %s", removeRootsFile)
			log.Printf("[DEBUG] dryRun: %t", dryRun)

			// Read in the stores CSV
			csvFile, _ := os.Open(storesFile)
			reader := csv.NewReader(bufio.NewReader(csvFile))
			storeEntries, _ := reader.ReadAll()
			var stores = make(map[string]StoreCSVEntry)
			for _, entry := range storeEntries {
				if entry[0] == "StoreID" {
					continue // Skip header
				}
				apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
				if err != nil {
					log.Printf("[ERROR] Error getting cert store: %s", err)
					lookupFailures = append(lookupFailures, strings.Join(entry, ","))
					continue
				}

				//log.Printf("[DEBUG] Store: %s", apiResp)
				inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
				if invErr != nil {
					log.Fatalf("[ERROR] Error getting cert store inventory: %s", invErr)
				}

				if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
					log.Printf("[WARN] Store %s is not a root store", apiResp.Id)
					continue
				} else {
					log.Printf("[INFO] Store %s is a root store", apiResp.Id)
				}

				stores[entry[0]] = StoreCSVEntry{
					ID:          entry[0],
					Type:        entry[1],
					Machine:     entry[2],
					Path:        entry[3],
					Thumbprints: make(map[string]bool),
					Serials:     make(map[string]bool),
					Ids:         make(map[int]bool),
				}
				for _, cert := range *inventory {
					thumb := cert.Thumbprints
					for t, v := range thumb {
						stores[entry[0]].Thumbprints[t] = v
					}
					for t, v := range cert.Serials {
						stores[entry[0]].Serials[t] = v
					}
					for t, v := range cert.Ids {
						stores[entry[0]].Ids[t] = v
					}
				}

			}

			// Read in the add addCerts CSV
			var certsToAdd = make(map[string]string)
			if addRootsFile != "" {
				certsToAdd, _ = readCertsFile(addRootsFile, kfClient)
				log.Printf("[DEBUG] ROT add certs called")
			} else {
				log.Printf("[INFO] No addCerts file specified")
			}

			// Read in the remove removeCerts CSV
			var certsToRemove = make(map[string]string)
			if removeRootsFile != "" {
				certsToRemove, _ = readCertsFile(removeRootsFile, kfClient)
				log.Printf("[DEBUG] ROT remove certs called")
			} else {
				log.Printf("[DEBUG] No removeCerts file specified")
			}
			_, actions, err := generateAuditReport(certsToAdd, certsToRemove, stores, kfClient)
			if err != nil {
				log.Fatalf("[ERROR] Error generating audit report: %s", err)
			}
			if len(actions) == 0 {
				fmt.Println("No reconciliation actions to take, root stores are up-to-date. Exiting.")
				return
			}
			rErr := reconcileRoots(actions, kfClient, dryRun)
			if rErr != nil {
				fmt.Printf("Error reconciling roots: %s", rErr)
				log.Fatalf("[ERROR] Error reconciling roots: %s", rErr)
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
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}
	rotGenStoreTemplateCmd = &cobra.Command{
		Use:                    "generate-template",
		Aliases:                nil,
		SuggestFor:             nil,
		Short:                  "For generating Root Of Trust template(s)",
		Long:                   `Root Of Trust: Will parse a CSV and attempt to enroll a cert or set of certs into a list of cert stores.`,
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

			templateType, _ := cmd.Flags().GetString("type")
			format, _ := cmd.Flags().GetString("format")
			outPath, _ := cmd.Flags().GetString("outPath")

			// Create CSV template file

			var filePath string
			if outPath != "" {
				filePath = outPath
			} else {
				filePath = fmt.Sprintf("%s_template.%s", templateType, format)
			}
			file, err := os.Create(filePath)
			if err != nil {
				log.Fatal("Cannot create file", err)
			}

			switch format {
			case "csv":
				writer := csv.NewWriter(file)
				var data []string
				switch templateType {
				case "stores":
					data = StoreHeader
				case "certs":
					data = CertHeader
				case "actions":
					data = AuditHeader
				}
				csvErr := writer.WriteAll([][]string{data})
				if csvErr != nil {
					fmt.Println(csvErr)
				}
				defer file.Close()

			case "json":
				writer := bufio.NewWriter(file)
				_, err := writer.WriteString("StoreID,StoreType,StoreMachine,StorePath")
				if err != nil {
					log.Fatal("Cannot write to file", err)
				}
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
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
	log.SetOutput(ioutil.Discard) //todo: remove this and set it global
	var (
		stores          string
		addCerts        string
		removeCerts     string
		minCertsInStore int
		maxPrivateKeys  int
		maxLeaves       int
		tType           = tTypeCerts
		outPath         string
		outputFormat    string
		actions         string
	)

	storesCmd.AddCommand(rotCmd)

	// Root of trust `audit` command
	rotCmd.AddCommand(rotAuditCmd)
	rotAuditCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotAuditCmd.Flags().StringVarP(&addCerts, "add-certs", "a", "",
		"CSV file containing cert(s) to enroll into the defined cert stores")
	rotAuditCmd.Flags().StringVarP(&removeCerts, "remove-certs", "r", "",
		"CSV file containing cert(s) to remove from the defined cert stores")
	rotAuditCmd.Flags().IntVarP(&minCertsInStore, "min-certs", "m", -1,
		"The minimum number of certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
	rotAuditCmd.Flags().IntVarP(&maxPrivateKeys, "max-keys", "x", -1,
		"The max number of private keys that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
	rotAuditCmd.Flags().IntVarP(&maxLeaves, "max-leaf-certs", "n", -1,
		"The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
	rotAuditCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotAuditCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
		"Path to write the audit report file to. If not specified, the file will be written to the current directory.")

	// Root of trust `reconcile` command
	rotCmd.AddCommand(rotReconcileCmd)
	rotReconcileCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotReconcileCmd.Flags().StringVarP(&addCerts, "add-certs", "a", "",
		"CSV file containing cert(s) to enroll into the defined cert stores")
	rotReconcileCmd.Flags().StringVarP(&removeCerts, "remove-certs", "r", "",
		"CSV file containing cert(s) to remove from the defined cert stores")
	rotReconcileCmd.Flags().StringVarP(&actions, "actions", "z", "",
		"CSV file containing reconciliation actions to perform. If this is specified, the other flags are ignored.")
	rotReconcileCmd.Flags().IntVarP(&minCertsInStore, "min-certs", "m", -1,
		"The minimum number of certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
	rotReconcileCmd.Flags().IntVarP(&maxPrivateKeys, "max-keys", "x", -1,
		"The max number of private keys that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
	rotReconcileCmd.Flags().IntVarP(&maxLeaves, "max-leaf-certs", "n", -1,
		"The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.")
	rotReconcileCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotReconcileCmd.Flags().BoolP("import-csv", "v", false, "Dry run mode")
	//rotReconcileCmd.MarkFlagsRequiredTogether("add-certs", "stores")
	//rotReconcileCmd.MarkFlagsRequiredTogether("remove-certs", "stores")
	rotReconcileCmd.MarkFlagsMutuallyExclusive("add-certs", "actions")
	rotReconcileCmd.MarkFlagsMutuallyExclusive("remove-certs", "actions")
	rotReconcileCmd.MarkFlagsMutuallyExclusive("stores", "actions")

	// Root of trust `generate` command
	rotCmd.AddCommand(rotGenStoreTemplateCmd)
	rotGenStoreTemplateCmd.Flags().StringVarP(&outPath, "outpath", "o", "",
		"Path to write the template file to. If not specified, the file will be written to the current directory.")
	rotGenStoreTemplateCmd.Flags().StringVarP(&outputFormat, "format", "f", "csv",
		"The type of template to generate. Only `csv` is supported at this time.")
	rotGenStoreTemplateCmd.Flags().Var(&tType, "type",
		`The type of template to generate. Only "certs|stores|actions" are supported at this time.`)
	rotGenStoreTemplateCmd.RegisterFlagCompletionFunc("type", templateTypeCompletion)
	rotGenStoreTemplateCmd.MarkFlagRequired("type")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rotAuditCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rotAuditCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
