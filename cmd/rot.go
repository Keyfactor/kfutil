/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
)

type StoreCSVEntry struct {
	Id          string          `json:"id"`
	Type        string          `json:"type"`
	Machine     string          `json:"address"`
	Path        string          `json:"path"`
	Thumbprints map[string]bool `json:"thumbprints;omitempty"`
	Serials     map[string]bool `json:"serials;omitempty"`
	Ids         map[int]bool    `json:"ids;omitempty"`
}

type RotCert struct {
	Id         int                        `json:"id;omitempty"`
	ThumbPrint string                     `json:"thumbprint;omitempty"`
	CN         string                     `json:"cn;omitempty"`
	Locations  []api.CertificateLocations `json:"locations;omitempty"`
}

type RotAction struct {
	StoreId    string
	StoreType  string
	StorePath  string
	Thumbprint string
	CertId     int
	Add        bool
	Remove     bool
}

var rotCmd = &cobra.Command{
	Use:   "rot",
	Short: "Root of trust utility",
	Long:  `Root of trust allows you to manage your trusted roots using Keyfactor certificate stores.`,
	//Run: func(cmd *cobra.Command, args []string) {
	//	fmt.Println("stores called")
	//},
}

// rotAuditCmd represents the rot command
var rotAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Root Of Trust Audit",
	Long:  `Root Of Trust Audit: Will read and parse inputs to generate a report of certs that need to be added or removed from the "root of trust" stores.`,
	Run: func(cmd *cobra.Command, args []string) {
		var lookupFailures []string
		kfClient, _ := initClient()
		storesFile, _ := cmd.Flags().GetString("stores")
		addRootsFile, _ := cmd.Flags().GetString("add-certs")
		removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
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
			if entry[0] == "StoreId" {
				continue // Skip header
			}
			_, err := kfClient.GetCertificateStoreByID(entry[0])
			if err != nil {
				//log.Fatalf("Error getting cert store: %s", err)
				log.Printf("[ERROR] Error getting cert store: %s", err)
				lookupFailures = append(lookupFailures, strings.Join(entry, ","))
				continue
			}

			//log.Printf("[DEBUG] Store: %s", apiResp)
			inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
			if invErr != nil {
				log.Fatal("[ERROR] Error getting cert store inventory: %s", invErr)
			}
			stores[entry[0]] = StoreCSVEntry{
				Id:          entry[0],
				Type:        entry[1],
				Machine:     entry[2],
				Path:        entry[3],
				Thumbprints: inventory.Thumbprints,
				Serials:     inventory.Serials,
				Ids:         inventory.Ids,
			}

		}
		//storesJson, _ := json.Marshal(stores)

		// Read in the add addCerts CSV
		var certsToAdd = make(map[string]string)
		if addRootsFile != "" {
			certsToAdd, _ = readCertsFile(addRootsFile, kfClient)
			//if err != nil {
			//	log.Fatalf("Error reading addCerts file: %s", err)
			//}
			addCertsJson, _ := json.Marshal(certsToAdd)
			log.Printf("[DEBUG] add certs JSON: %s", string(addCertsJson))
			log.Println("[DEBUG] Add ROT called")
		} else {
			log.Printf("[DEBUG] No addCerts file specified")
			log.Printf("[DEBUG] No addCerts = %s", certsToAdd)
		}

		// Read in the remove removeCerts CSV
		var certsToRemove = make(map[string]string)
		if removeRootsFile != "" {
			certsToRemove, _ = readCertsFile(removeRootsFile, kfClient)
			//if err != nil {
			//	log.Fatalf("Error reading removeCerts file: %s", err)
			//}
			removeCertsJson, _ := json.Marshal(certsToRemove)
			fmt.Println(string(removeCertsJson))
			fmt.Println("remove rot called")
		} else {
			log.Printf("[DEBUG] No removeCerts file specified")
			log.Printf("[DEBUG] No removeCerts = %s", certsToRemove)
		}
		generateAuditReport(certsToAdd, certsToRemove, stores, kfClient)
	},
}

var rotReconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Root Of Trust",
	Long:  `Root Of Trust: Will parse a CSV and attempt to enroll a cert or set of certs into a list of cert stores.`,
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
			if entry[0] == "StoreId" {
				continue // Skip header
			}
			apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
			if err != nil {
				//log.Fatalf("Error getting cert store: %s", err)
				log.Printf("[ERROR] Error getting cert store: %s", err)
				lookupFailures = append(lookupFailures, strings.Join(entry, ","))
				continue
			}

			//log.Printf("[DEBUG] Store: %s", apiResp)
			inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
			if invErr != nil {
				log.Fatal("[ERROR] Error getting cert store inventory: %s", invErr)
			}

			if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
				log.Printf("[WARN] Store %s is not a root store", apiResp.Id)
				continue
			} else {
				log.Printf("[INFO] Store %s is a root store", apiResp.Id)
			}

			stores[entry[0]] = StoreCSVEntry{
				Id:          entry[0],
				Type:        entry[1],
				Machine:     entry[2],
				Path:        entry[3],
				Thumbprints: inventory.Thumbprints,
				Serials:     inventory.Serials,
				Ids:         inventory.Ids,
			}

		}

		// Read in the add addCerts CSV
		var certsToAdd = make(map[string]string)
		if addRootsFile != "" {
			certsToAdd, _ = readCertsFile(addRootsFile, kfClient)
			log.Printf("[DEBUG] ROT add certs called")
		} else {
			log.Printf("[DEBUG] No addCerts file specified")
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
		reconcileRoots(actions, kfClient, dryRun)
	},
}

func reconcileRoots(actions map[string]RotAction, kfClient *api.Client, dryRun bool) {
	log.Printf("[DEBUG] Reconciling roots")
	if len(actions) == 0 {
		log.Printf("[INFO] No actions to take, roots are up-to-date.")
		return
	}
	for thumbprint, action := range actions {
		if action.Add {
			log.Printf("[INFO] Adding cert %s to store %s(%s)", thumbprint, action.StoreId, action.StorePath)
			if !dryRun {
				cStore := api.CertificateStore{
					CertificateStoreId: action.StoreId,
					Overwrite:          true,
				}
				var stores []api.CertificateStore
				stores = append(stores, cStore)
				schedule := &api.InventorySchedule{
					Immediate: boolToPointer(true),
				}
				addReq := api.AddCertificateToStore{
					CertificateId:     action.CertId,
					CertificateStores: &stores,
					InventorySchedule: schedule,
				}
				_, err := kfClient.AddCertificateToStores(&addReq)
				if err != nil {
					log.Fatalf("[ERROR] Error adding cert to store: %s", err)
				}
			} else {
				log.Printf("[INFO] DRY RUN: Would have added cert %s from store %s", thumbprint, action.StoreId)
			}
		} else if action.Remove {
			if !dryRun {
				log.Printf("[INFO] Removing cert from store %s", action.StoreId)
				cStore := api.CertificateStore{
					CertificateStoreId: action.StoreId,
				}
				var stores []api.CertificateStore
				stores = append(stores, cStore)
				schedule := &api.InventorySchedule{
					Immediate: boolToPointer(true),
				}
				removeReq := api.RemoveCertificateFromStore{
					CertificateId:     action.CertId,
					Alias:             fmt.Sprintf("KeyfactorAdd%d", action.CertId),
					CertificateStores: &stores,
					InventorySchedule: schedule,
				}
				_, err := kfClient.RemoveCertificateFromStores(&removeReq)
				if err != nil {
					log.Fatalf("[ERROR] Error removing cert from store: %s", err)
				}
			} else {
				log.Printf("[INFO] DRY RUN: Would have removed cert %s from store %s", thumbprint, action.StoreId)
			}

		}
	}
}

func readCertsFile(certsFilePath string, kfclient *api.Client) (map[string]string, error) {
	// Read in the cert CSV
	csvFile, _ := os.Open(certsFilePath)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	certEntries, _ := reader.ReadAll()
	var certs = make(map[string]string)
	for _, entry := range certEntries {
		switch entry[0] {
		case "CertId", "thumbprint", "id", "certId", "Thumbprint":
			continue // Skip header
		}
		certs[entry[0]] = entry[0]
	}
	return certs, nil
}

func generateAuditReport(addCerts map[string]string, removeCerts map[string]string, stores map[string]StoreCSVEntry, kfClient *api.Client) ([][]string, map[string]RotAction, error) {
	log.Println("[DEBUG] generateAuditReport called")
	var data [][]string
	header := []string{"Thumbprint", "StoreId", "StoreType", "Machine", "Path", "AddCert", "RemoveCert", "Deployed"}
	data = append(data, header)
	csvFile, _ := os.Create("rot_audit.csv")
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Write(header)
	actions := make(map[string]RotAction)

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
			log.Fatalf("[ERROR] Error looking up cert: %s", err)
			continue
		}
		certId := certLookup.Id
		certIdStr := strconv.Itoa(certId)
		for _, store := range stores {
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is already in the store do nothing
				row := []string{cert, certIdStr, certIdStr, store.Id, store.Type, store.Machine, store.Path, "false", "false", "true"}
				data = append(data, row)
				csvWriter.Write(row)
			} else {
				// Cert is not deployed to this store and will need to be added
				row := []string{cert, certIdStr, store.Id, store.Type, store.Machine, store.Path, "true", "false", "false"}
				data = append(data, row)
				csvWriter.Write(row)
				actions[cert] = RotAction{
					Thumbprint: cert,
					CertId:     certId,
					StoreId:    store.Id,
					StoreType:  store.Type,
					StorePath:  store.Path,
					Add:        true,
					Remove:     false,
				}
			}
		}
	}
	for _, cert := range removeCerts {
		for _, store := range stores {
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is deployed to this store and will need to be removed
				row := []string{cert, store.Id, store.Type, store.Machine, store.Path, "false", "true", "true"}
				data = append(data, row)
				csvWriter.Write(row)
				actions[cert] = RotAction{
					Thumbprint: cert,
					StoreId:    store.Id,
					StoreType:  store.Type,
					Add:        false,
					Remove:     true,
				}
			} else {
				// Cert is not deployed to this store do nothing
				row := []string{cert, store.Id, store.Type, store.Machine, store.Path, "false", "false", "false"}
				data = append(data, row)
				csvWriter.Write(row)
			}
		}
	}
	csvWriter.Flush()
	csvFile.Close()

	//log.Printf("[DEBUG] data: %s", data)
	return data, actions, nil
}

var rotGenStoreTemplateCmd = &cobra.Command{
	Use:   "generate-template",
	Short: "For generating Root Of Trust template(s)",
	Long:  `Root Of Trust: Will parse a CSV and attempt to enroll a cert or set of certs into a list of cert stores.`,
	Run: func(cmd *cobra.Command, args []string) {

		templateType, _ := cmd.Flags().GetString("type")
		format, _ := cmd.Flags().GetString("format")
		outpath, _ := cmd.Flags().GetString("outpath")

		// Create CSV template file

		var filePath string
		if outpath != "" {
			filePath = outpath
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
			var data = [][]string{}
			switch templateType {
			case "stores":
				data = [][]string{
					{"StoreId", "StoreType", "StoreMachine", "StorePath"},
				}
			case "certs":
				data = [][]string{
					{"Thumbprint"},
				}
			}
			csvErr := writer.WriteAll(data)
			if csvErr != nil {
				fmt.Println(csvErr)
			}
			defer file.Close()

		case "json":
			writer := bufio.NewWriter(file)
			_, err := writer.WriteString("StoreId,StoreType,StoreMachine,StorePath")
			if err != nil {
				log.Fatal("Cannot write to file", err)
			}
		}

	}}

func isRootStore(st *api.GetStoreByIDResp, inv *api.CertStoreInventory, minCerts int, maxKeys int, maxLeaf int) bool {
	certCount := len(inv.Certificates)
	if certCount < minCerts {
		log.Printf("[DEBUG] Store %s has %d certs, less than the required count of %d", st.Id, certCount, minCerts)
		return false
	}
	leafCount := 0
	keyCount := 0
	for _, cert := range inv.Certificates {
		if cert.IssuedDN != cert.IssuerDN {
			leafCount++
			if leafCount > maxLeaf {
				log.Printf("[DEBUG] Store %s has too many leaf certs", st.Id)
				return false
			}
		}
		if inv.Parameters["PrivateKeyEntry"] == "Yes" {
			keyCount++
			if keyCount > maxKeys {
				log.Printf("[DEBUG] Store %s has too many keys", st.Id)
				return false
			}
		}
	}

	return true
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
	var stores string
	var addCerts string
	var removeCerts string
	var minCertsInStore int
	var maxPrivateKeys int
	var maxLeaves int

	storesCmd.AddCommand(rotCmd)
	rotCmd.AddCommand(rotAuditCmd)
	rotAuditCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotAuditCmd.MarkFlagRequired("stores")
	rotAuditCmd.Flags().StringVarP(&addCerts, "add-certs", "a", "", "CSV file containing cert(s) to enroll into the defined cert stores")
	rotAuditCmd.Flags().StringVarP(&removeCerts, "remove-certs", "r", "", "CSV file containing cert(s) to remove from the defined cert stores")
	rotAuditCmd.Flags().IntVarP(&minCertsInStore, "min-certs", "m", 1, "The minimum number of certs that should be in a store to be considered a 'root' store")
	rotAuditCmd.Flags().IntVarP(&maxPrivateKeys, "max-keys", "x", 5, "The max number of private keys that should be in a store to be considered a 'root' store")
	rotAuditCmd.Flags().IntVarP(&maxLeaves, "max-leaf-certs", "n", 5, "The max number of non-root-certs that should be in a store to be considered a 'root' store")
	rotAuditCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotAuditCmd.MarkFlagRequired("certs")

	rotCmd.AddCommand(rotReconcileCmd)
	rotReconcileCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotReconcileCmd.MarkFlagRequired("stores")
	rotReconcileCmd.Flags().StringVarP(&addCerts, "add-certs", "a", "", "CSV file containing cert(s) to enroll into the defined cert stores")
	rotReconcileCmd.Flags().StringVarP(&removeCerts, "remove-certs", "r", "", "CSV file containing cert(s) to remove from the defined cert stores")
	rotReconcileCmd.Flags().IntVarP(&minCertsInStore, "min-certs", "m", 1, "The minimum number of certs that should be in a store to be considered a 'root' store")
	rotReconcileCmd.Flags().IntVarP(&maxPrivateKeys, "max-keys", "x", 5, "The max number of private keys that should be in a store to be considered a 'root' store")
	rotReconcileCmd.Flags().IntVarP(&maxLeaves, "max-leaf-certs", "n", 5, "The max number of non-root-certs that should be in a store to be considered a 'root' store")
	rotReconcileCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotReconcileCmd.MarkFlagRequired("certs")

	rotCmd.AddCommand(rotGenStoreTemplateCmd)
	rotGenStoreTemplateCmd.Flags().String("outpath", "", "Output file to write the template to")
	rotGenStoreTemplateCmd.Flags().String("format", "csv", "The type of template to generate. Only `csv` is supported at this time.")
	rotGenStoreTemplateCmd.Flags().String("type", "stores", "The type of template to generate. Only `certs|stores` are supported at this time.")
	//rotGenStoreTemplateCmd.MarkFlagRequired("type")
	//rotGenStoreTemplateCmd.MarkFlagRequired("format")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rotAuditCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rotAuditCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
