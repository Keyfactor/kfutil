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
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strconv"
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/v3/api"
	"github.com/rs/zerolog/log"
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
	StoreID    string `json:"store_id,omitempty" mapstructure:"StoreID,omitempty"`
	StoreType  string `json:"store_type,omitempty" mapstructure:"StoreType,omitempty"`
	StorePath  string `json:"store_path,omitempty" mapstructure:"StorePath,omitempty"`
	Thumbprint string `json:"thumbprint,omitempty" mapstructure:"Thumbprint,omitempty"`
	Alias      string `json:"alias,omitempty" mapstructure:"Alias,omitempty"`
	CertID     int    `json:"cert_id,omitempty" mapstructure:"CertID,omitempty"`
	AddCert    bool   `json:"add,omitempty" mapstructure:"AddCert,omitempty"`
	RemoveCert bool   `json:"remove,omitempty"  mapstructure:"RemoveCert,omitempty"`
}

const (
	tTypeCerts               templateType = "certs"
	reconcileDefaultFileName string       = "rot_audit.csv"
)

var (
	AuditHeader = []string{
		"Thumbprint",
		"CertID",
		"SubjectName",
		"Issuer",
		"StoreID",
		"StoreType",
		"Machine",
		"Path",
		"AddCert",
		"RemoveCert",
		"Deployed",
		"AuditDate",
	}
	ReconciledAuditHeader = []string{
		"Thumbprint",
		"CertID",
		"SubjectName",
		"Issuer",
		"StoreID",
		"StoreType",
		"Machine",
		"Path",
		"AddCert",
		"RemoveCert",
		"Deployed",
		"ReconciledDate",
	}
	StoreHeader = []string{
		"StoreID",
		"StoreType",
		"StoreMachine",
		"StorePath",
		"ContainerId",
		"ContainerName",
		"LastQueriedDate",
	}
	CertHeader = []string{"Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"}
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

func generateAuditReport(
	addCerts map[string]string,
	removeCerts map[string]string,
	stores map[string]StoreCSVEntry,
	outpath string,
	kfClient *api.Client,
) ([][]string, map[string][]ROTAction, error) {
	log.Debug().Msg("Entering generateAuditReport")
	var (
		data [][]string
	)

	data = append(data, AuditHeader)
	var csvFile *os.File
	var fErr error
	if outpath == "" {
		csvFile, fErr = os.Create(reconcileDefaultFileName)
		outpath = reconcileDefaultFileName
	} else {
		csvFile, fErr = os.Create(outpath)
	}

	if fErr != nil {
		log.Error().
			Str("file", csvFile.Name()).
			Msg("Error creating audit file")
		outputError(fErr, true, outputFormat)

	}
	csvWriter := csv.NewWriter(csvFile)
	cErr := csvWriter.Write(AuditHeader)
	if cErr != nil {
		log.Error().
			Str("file", csvFile.Name()).
			Err(cErr).
			Msg("Error writing audit header")
		outputError(cErr, true, outputFormat)
	}
	actions := make(map[string][]ROTAction)

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
			fmt.Printf("[ERROR] looking up certificate %s: %s\n", cert, err)
			log.Printf("[ERROR] looking up cert: %s\n%v", cert, err)
			continue
		}
		certID := certLookup.Id
		certIDStr := strconv.Itoa(certID)
		for _, store := range stores {
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is already in the store do nothing
				row := []string{
					cert,
					certIDStr,
					certLookup.IssuedDN,
					certLookup.IssuerDN,
					store.ID,
					store.Type,
					store.Machine,
					store.Path,
					"false",
					"false",
					"true",
					getCurrentTime(""),
				}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					log.Error().
						Str("file", csvFile.Name()).
						Err(wErr).
						Msg("Error writing audit row")
					outputError(wErr, false, outputFormat)

				}
			} else {
				// Cert is not deployed to this store and will need to be added
				row := []string{
					cert,
					certIDStr,
					certLookup.IssuedDN,
					certLookup.IssuerDN,
					store.ID,
					store.Type,
					store.Machine,
					store.Path,
					"true",
					"false",
					"false",
					getCurrentTime(""),
				}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					log.Error().
						Err(wErr).
						Str("file", csvFile.Name()).
						Msg("Error writing audit row")
					outputError(wErr, false, outputFormat)

				}
				actions[cert] = append(
					actions[cert], ROTAction{
						Thumbprint: cert,
						//TODO: add Alias
						CertID:     certID,
						StoreID:    store.ID,
						StoreType:  store.Type,
						StorePath:  store.Path,
						AddCert:    true,
						RemoveCert: false,
					},
				)
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
			log.Printf("[ERROR] looking up cert: %s", err)
			continue
		}
		certID := certLookup.Id
		certIDStr := strconv.Itoa(certID)
		for _, store := range stores {
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is deployed to this store and will need to be removed
				row := []string{
					cert,
					certIDStr,
					certLookup.IssuedDN,
					certLookup.IssuerDN,
					store.ID,
					store.Type,
					store.Machine,
					store.Path,
					"false",
					"true",
					"true",
					getCurrentTime(""),
				}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					fmt.Printf("%s", wErr)
					log.Printf("[ERROR] writing row to CSV: %s", wErr)
				}
				actions[cert] = append(
					actions[cert], ROTAction{
						Thumbprint: cert,
						CertID:     certID,
						StoreID:    store.ID,
						StoreType:  store.Type,
						StorePath:  store.Path,
						AddCert:    false,
						RemoveCert: true,
					},
				)
			} else {
				// Cert is not deployed to this store do nothing
				row := []string{
					cert,
					certIDStr,
					certLookup.IssuedDN,
					certLookup.IssuerDN,
					store.ID,
					store.Type,
					store.Machine,
					store.Path,
					"false",
					"false",
					"false",
					getCurrentTime(""),
				}
				data = append(data, row)
				wErr := csvWriter.Write(row)
				if wErr != nil {
					fmt.Printf("%s", wErr)
					log.Printf("[ERROR] writing row to CSV: %s", wErr)
				}
			}
		}
	}
	csvWriter.Flush()
	ioErr := csvFile.Close()
	if ioErr != nil {
		fmt.Println(ioErr)
		log.Printf("[ERROR] closing audit file: %s", ioErr)
	}
	fmt.Printf("Audit report written to %s\n", outpath)
	return data, actions, nil
}

func reconcileRoots(actions map[string][]ROTAction, kfClient *api.Client, reportFile string, dryRun bool) error {
	log.Debug().Msg("entered reconcileRoots")
	if len(actions) == 0 {
		log.Info().Msg("No actions to take, roots are up-to-date.")
		return nil
	}
	rFileName := fmt.Sprintf("%s_reconciled.csv", strings.Split(reportFile, ".csv")[0])
	csvFile, fErr := os.Create(rFileName)
	if fErr != nil {
		log.Error().
			Err(fErr).
			Str("file", rFileName).
			Msg("Error creating audit file")
		outputError(fErr, true, outputFormat)
		return fErr
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	cErr := csvWriter.Write(ReconciledAuditHeader)
	if cErr != nil {
		log.Debug().
			Str("file", csvFile.Name()).
			Err(cErr).
			Msg("Error writing audit header")
		outputError(cErr, true, outputFormat)
		return cErr
	}
	for thumbprint, action := range actions {
		log.Debug().
			Str("thumbprint", thumbprint).
			Interface("action", action).
			Msg("Processing thumbprint")
		for _, a := range action {
			if a.AddCert {
				log.Info().
					Str("thumbprint", thumbprint).
					Str("store", a.StoreID).
					Str("storePath", a.StorePath).
					Msg("Adding cert to store")
				if !dryRun {
					log.Debug().
						Msg("Not a dry run")

					cStore := api.CertificateStore{
						CertificateStoreId: a.StoreID,
						Overwrite:          true,
						Alias:              a.Thumbprint, //TODO: Support non-thumbprint alias
						//Alias: "",
					}

					var stores []api.CertificateStore
					stores = append(stores, cStore)
					schedule := &api.InventorySchedule{
						Immediate: boolToPointer(true),
					}
					addReq := api.AddCertificateToStore{
						CertificateId:     a.CertID,
						CertificateStores: &stores,
						InventorySchedule: schedule,
					}
					log.Debug().
						Str("thumbprint", thumbprint).
						Str("store", a.StoreID).
						Msg("Adding cert to store")

					addReqJSON, _ := json.Marshal(addReq)

					log.Debug().
						Str("addReqJSON", string(addReqJSON)).
						Msg("Request payload")
					_, err := kfClient.AddCertificateToStores(&addReq)
					if err != nil {
						log.Error().
							Err(err).
							Str("thumbprint", thumbprint).
							Str("store", a.StoreID).
							Str("storePath", a.StorePath).
							Msg("Error adding cert to store")
						outputError(err, true, outputFormat)
						continue
					}
				} else {
					log.Info().
						Str("thumbprint", thumbprint).
						Str("store", a.StoreID).
						Str("storePath", a.StorePath).
						Msg("This is a dry run, would have added cert to store")
				}
			} else if a.RemoveCert {
				if !dryRun {
					log.Info().
						Str("thumbprint", thumbprint).
						Str("store", a.StoreID).
						Str("storePath", a.StorePath).
						Msg("Removing cert from store")

					cStore := api.CertificateStore{
						CertificateStoreId: a.StoreID,
						Alias:              a.Thumbprint,
					}
					var stores []api.CertificateStore
					stores = append(stores, cStore)
					schedule := &api.InventorySchedule{
						Immediate: boolToPointer(true),
					}
					removeReq := api.RemoveCertificateFromStore{
						CertificateId:     a.CertID,
						CertificateStores: &stores,
						InventorySchedule: schedule,
					}
					_, err := kfClient.RemoveCertificateFromStores(&removeReq)
					if err != nil {
						fmt.Printf(
							"[ERROR] removing cert %s (ID: %d) from store %s (%s): %s\n",
							a.Thumbprint,
							a.CertID,
							a.StoreID,
							a.StorePath,
							err,
						)
						log.Error().
							Err(err).
							Str("thumbprint", thumbprint).
							Str("store", a.StoreID).
							Str("storePath", a.StorePath).
							Msg("Error removing cert from store")
					}
				} else {
					log.Info().
						Str("thumbprint", thumbprint).
						Str("store", a.StoreID).
						Str("storePath", a.StorePath).
						Msg("This is a dry run, would have removed cert from store")
				}
			}
		}
	}
	log.Debug().Msg("exiting reconcileRoots")
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

func isRootStore(
	st *api.GetCertificateStoreResponse,
	invs *[]api.CertStoreInventory,
	minCerts int,
	maxKeys int,
	maxLeaf int,
) bool {
	leafCount := 0
	keyCount := 0
	certCount := 0

	log.Debug().
		Int("minCerts", minCerts).
		Int("maxKeys", maxKeys).
		Int("maxLeaf", maxLeaf).
		Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))

	if invs == nil || len(*invs) == 0 {
		nullInvErr := fmt.Errorf("nil inventory response from Keyfactor Command for store '%s'", st.Id)
		log.Error().Err(nullInvErr).Str("store", st.Id).Msg("nil or empty inventory returned by Keyfactor Command")
		return false
	} else if st == nil {
		nullStoreErr := fmt.Errorf("nil store response from Keyfactor Command for store '%s'", st.Id)
		log.Error().Err(nullStoreErr).Str("store", st.Id).Msg("nil or empty store returned by Keyfactor Command")
		return false
	}

	for _, inv := range *invs {
		certCount += len(inv.Certificates)
		log.Debug().
			Int("certCount", certCount).
			Str("name", inv.Name).
			Msg("processing inventory")

		for _, cert := range inv.Certificates {
			if cert.IssuedDN != cert.IssuerDN {
				log.Debug().Str("dn", cert.IssuedDN).Msg("is a leaf cert")
				leafCount++
			} else {
				log.Debug().Str("dn", cert.IssuedDN).Msg("is a root cert")
			}

			//TODO: Do we need to look up if a cert has a private key? If so how does one know the private key isdeployed to the store?
			//if inv.Parameters["PrivateKeyEntry"] == "Yes" {
			//	keyCount++
			//}
		}
	}
	if certCount < minCerts && minCerts >= 0 {
		log.Debug().
			Str("store", st.Id).
			Int("certCount", certCount).
			Int("minCerts", minCerts).
			Msg("store has too few certs")

		return false
	}
	if leafCount > maxLeaf && maxLeaf >= 0 {
		log.Debug().
			Str("store", st.Id).
			Int("certCount", certCount).
			Int("minCerts", minCerts).
			Msg("store has too many leaf certs")

		return false
	}

	if keyCount > maxKeys && maxKeys >= 0 {
		log.Debug().
			Str("store", st.Id).
			Int("certCount", certCount).
			Int("minCerts", minCerts).
			Msg("store has too many keys")
		return false
	}

	log.Debug().
		Str("store", st.Id).
		Int("certCount", certCount).
		Int("minCerts", minCerts).
		Msg("store is a root store")

	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))
	return true
}

var (
	rotCmd = &cobra.Command{
		Use:   "rot",
		Short: "Root of trust utility",
		Long: `Root of trust allows you to manage your trusted roots using Keyfactor certificate stores.
For example if you wish to add a list of "root" certs to a list of certificate stores you would simply generate and fill
out the template CSV file. These template files can be generated with the following commands:
kfutil stores rot generate-template --type certs
kfutil stores rot generate-template --type stores
Once those files are filled out you can use the following command to add the certs to the stores:
kfutil stores rot audit --certs-file <certs-file> --stores-file <stores-file>
Will generate a CSV report file 'rot_audit.csv' of what actions will be taken. If those actions are correct you can run
the following command to actually perform the actions:
kfutil stores rot reconcile --certs-file <certs-file> --stores-file <stores-file>
OR if you want to use the audit report file generated you can run this command:
kfutil stores rot reconcile --import-csv <audit-file>
`,
	}
	rotAuditCmd = &cobra.Command{
		Use:                    "audit",
		Aliases:                nil,
		SuggestFor:             nil,
		Short:                  "Audit generates a CSV report of what actions will be taken based on input CSV files.",
		Long:                   `Root of Trust Audit: Will read and parse inputs to generate a report of certs that need to be added or removed from the "root of trust" stores.`,
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Global flags
			cmd.SilenceUsage = true
			// expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			stdlog.SetOutput(io.Discard)
			informDebug(debugFlag)

			var lookupFailures []string
			// Authenticate
			kfClient, cErr := initClient(false)
			if cErr != nil {
				log.Error().Err(cErr).Msg("unable to authenticate")
				return cErr
			}

			// Local flags
			storesFile, _ := cmd.Flags().GetString("stores")
			addRootsFile, _ := cmd.Flags().GetString("add-certs")
			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
			minCerts, _ := cmd.Flags().GetInt("min-certs")
			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
			maxKeys, _ := cmd.Flags().GetInt("max-keys")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			outpath, _ := cmd.Flags().GetString("outpath")
			// Read in the stores CSV
			log.Debug().
				Str("storesFile", storesFile).
				Str("addRootsFile", addRootsFile).
				Str("removeRootsFile", removeRootsFile).
				Bool("dryRun", dryRun).
				Int("minCerts", minCerts).
				Int("maxLeaves", maxLeaves).
				Int("maxKeys", maxKeys).
				Str("outpath", outpath).
				Msg("flags")

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
					log.Error().
						Str("expectedHeader", strings.Join(StoreHeader, ",")).
						Msg("Invalid header in stores file")
				}
				apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
				if err != nil {
					log.Error().
						Err(err).
						Str("store", entry[0]).
						Msg("Error getting store from Keyfactor Command")
					_ = append(lookupFailures, strings.Join(entry, ","))
					continue
				}

				inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
				if invErr != nil {
					log.Error().Err(invErr).Str("store", entry[0]).Msg("Error getting store inventory")
					outputError(invErr, true, outputFormat)
					return invErr
				}

				if inventory == nil {
					invalidRespErr := fmt.Errorf(
						"invalid inventory response from Keyfactor Command for store '%s'",
						entry[0],
					)
					log.Error().Err(invalidRespErr).Str("store", entry[0]).Msg("invalid response")
					outputError(invalidRespErr, true, outputFormat)
					return invalidRespErr
				}
				//var inventory []api.CertStoreInventory //TODO: Update this to use SDK inventory

				if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
					outputResult(fmt.Sprintf("Store %s is not a root store, skipping.\n", entry[0]), outputFormat)
					log.Error().
						Str("store", entry[0]).
						Str("id", apiResp.Id).
						Int("type", apiResp.CertStoreType).
						Str("machine", apiResp.ClientMachine).
						Str("path", apiResp.StorePath).
						Msg("Store is not a root store")
					continue
				} else {
					outputResult(fmt.Sprintf("Store %s is a root store.\n", entry[0]), outputFormat)
					log.Info().
						Str("store", entry[0]).
						Str("id", apiResp.Id).
						Int("type", apiResp.CertStoreType).
						Str("machine", apiResp.ClientMachine).
						Str("path", apiResp.StorePath).
						Msg("Is a root store")
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

				log.Debug().Str("store", entry[0]).
					Str("id", apiResp.Id).
					Int("type", apiResp.CertStoreType).
					Str("machine", apiResp.ClientMachine).
					Str("path", apiResp.StorePath).
					Msg("Iterating store inventory")
				for _, cert := range *inventory {
					thumb := cert.Thumbprints

					log.Debug().Str("store", entry[0]).
						Str("id", apiResp.Id).
						Int("type", apiResp.CertStoreType).
						Str("machine", apiResp.ClientMachine).
						Str("path", apiResp.StorePath).
						Msg("Iterating inventory thumbprints")
					for _, v := range thumb {
						stores[entry[0]].Thumbprints[v] = true
					}
					log.Debug().Str("store", entry[0]).
						Str("id", apiResp.Id).
						Int("type", apiResp.CertStoreType).
						Str("machine", apiResp.ClientMachine).
						Str("path", apiResp.StorePath).
						Msg("Iterating inventory serial numbers")
					for _, v := range cert.Serials {
						stores[entry[0]].Serials[v] = true
					}

					log.Debug().Str("store", entry[0]).
						Str("id", apiResp.Id).
						Int("type", apiResp.CertStoreType).
						Str("machine", apiResp.ClientMachine).
						Str("path", apiResp.StorePath).
						Msg("Iterating certificate IDs")
					for _, v := range cert.Ids {
						stores[entry[0]].Ids[v] = true
					}
				}

			}

			// Read in the add addCerts CSV
			var certsToAdd = make(map[string]string)
			if addRootsFile != "" {
				var rcfErr error

				log.Debug().
					Str("addRootsFile", addRootsFile).
					Msg("Reading addCerts file")
				certsToAdd, rcfErr = readCertsFile(addRootsFile, kfClient)
				if rcfErr != nil {
					outputError(rcfErr, true, outputFormat)
					log.Error().
						Err(rcfErr).
						Msg("reading addCerts file")
					return rcfErr
				}
				addCertsJSON, _ := json.Marshal(certsToAdd)
				log.Debug().Str("addCerts", string(addCertsJSON)).Msg("addCerts")
			} else {
				log.Debug().Msg("No addCerts file specified")
			}

			// Read in the remove removeCerts CSV
			var certsToRemove = make(map[string]string)
			if removeRootsFile != "" {
				var rcfErr error
				certsToRemove, rcfErr = readCertsFile(removeRootsFile, kfClient)
				if rcfErr != nil {
					outputError(rcfErr, true, outputFormat)
					log.Error().Err(rcfErr).Msg("failed reading removeCerts file")
					return rcfErr
				}
				removeCertsJSON, _ := json.Marshal(certsToRemove)
				log.Debug().Str("removeCerts", string(removeCertsJSON)).Msg("removeCerts")
			} else {
				log.Debug().Msg("No removeCerts file specified")
			}

			log.Debug().
				Str("outpath", outpath).
				Interface("certsToAdd", certsToAdd).
				Interface("certsToRemove", certsToRemove).
				Msg("Generating audit report")
			_, _, gErr := generateAuditReport(certsToAdd, certsToRemove, stores, outpath, kfClient)
			if gErr != nil {
				outputError(gErr, true, outputFormat)
				log.Error().Err(gErr).Msg("failed generating audit report")
				return gErr
			}

			return nil
		},
		Run:                        nil,
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
		Use:        "reconcile",
		Aliases:    nil,
		SuggestFor: nil,
		Short:      "Reconcile either takes in or will generate an audit report and then add/remove certs as needed.",
		Long: `Root of Trust (rot): Will parse either a combination of CSV files that define certs to
add and/or certs to remove with a CSV of certificate stores or an audit CSV file. If an audit CSV file is provided, the
add and remove actions defined in the audit file will be immediately executed. If a combination of CSV files are provided,
the utility will first generate an audit report and then execute the add/remove actions defined in the audit report.`,
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Global flags
			cmd.SilenceUsage = true
			// expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			stdlog.SetOutput(io.Discard)
			informDebug(debugFlag)

			var lookupFailures []string
			// Authenticate
			kfClient, cErr := initClient(false)
			if cErr != nil {
				log.Error().Err(cErr).Msg("unable to authenticate")
				return cErr
			}

			// Local flags
			storesFile, _ := cmd.Flags().GetString("stores")
			addRootsFile, _ := cmd.Flags().GetString("add-certs")
			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
			minCerts, _ := cmd.Flags().GetInt("min-certs")
			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
			maxKeys, _ := cmd.Flags().GetInt("max-keys")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			outpath, _ := cmd.Flags().GetString("outpath")
			isCSV, _ := cmd.Flags().GetBool("import-csv")
			reportFile, _ := cmd.Flags().GetString("input-file")
			// Read in the stores CSV
			log.Debug().
				Str("storesFile", storesFile).
				Str("addRootsFile", addRootsFile).
				Str("removeRootsFile", removeRootsFile).
				Bool("dryRun", dryRun).
				Int("minCerts", minCerts).
				Int("maxLeaves", maxLeaves).
				Int("maxKeys", maxKeys).
				Str("outpath", outpath).
				Bool("isCSV", isCSV).
				Str("reportFile", reportFile).
				Msg("flags")

			// Parse existing audit report
			if isCSV && reportFile != "" {
				log.Debug().
					Str("reportFile", reportFile).
					Msg("reading CSV audit report")
				// Read in the CSV
				csvFile, err := os.Open(reportFile)
				if err != nil {
					outputError(err, true, outputFormat)
					log.Error().
						Err(err).
						Str("reportFile", reportFile).
						Msg("failed opening CSV file")
					return err
				}
				validHeader := false

				log.Debug().
					Str("reportFile", reportFile).
					Msg("parsing CSV audit report")
				aCSV := csv.NewReader(csvFile)
				aCSV.FieldsPerRecord = -1
				inFile, cErr := aCSV.ReadAll()
				if cErr != nil {
					log.Error().Err(cErr).Str("reportFile", reportFile).
						Msg("failed parsing CSV file")
				}
				actions := make(map[string][]ROTAction)
				fieldMap := make(map[int]string)

				log.Debug().
					Str("reportFile", reportFile).
					Msg("mapping CSV header")
				for i, field := range AuditHeader {
					fieldMap[i] = field
				}
				for ri, row := range inFile {
					if strings.EqualFold(strings.Join(row, ","), strings.Join(AuditHeader, ",")) {
						validHeader = true
						log.Debug().
							Str("reportFile", reportFile).
							Msg("skipping header")
						continue
					}
					if !validHeader {
						log.Error().
							Str("reportFile", reportFile).
							Str("expectedHeader", strings.Join(AuditHeader, ",")).
							Str("inputFileHeader", strings.Join(row, ",")).
							Msg("invalid header")
						log.Debug().
							Int("row", ri).
							Str("reportFile", reportFile).
							Msg("searching for valid header")
					}

					action := make(map[string]interface{})

					log.Debug().Int("row", ri).Msg("processing row data")
					for i, field := range row {
						fieldInt, iErr := strconv.Atoi(field)
						if iErr != nil {
							log.Debug().Int("row", ri).Str("field", field).Msg("field is not an integer")
							action[fieldMap[i]] = field
						} else {
							log.Debug().Int("row", ri).Str("field", field).Msg("field is an integer")
							action[fieldMap[i]] = fieldInt
						}
					}

					addCertStr, aOk := action["AddCert"].(string)
					if !aOk {
						addCertStr = ""
					}
					addCert, acErr := strconv.ParseBool(addCertStr)
					if acErr != nil {
						addCert = false
					}

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"RemoveCert\" col")
					removeCertStr, rOk := action["RemoveCert"].(string)
					if !rOk {
						removeCertStr = ""
					}
					removeCert, rcErr := strconv.ParseBool(removeCertStr)
					if rcErr != nil {
						removeCert = false
					}

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"StoreType\" col")
					sType, sOk := action["StoreType"].(string)
					if !sOk {
						sType = ""
					}

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"Path\" col")
					sPath, pOk := action["Path"].(string)
					if !pOk {
						sPath = ""
					}

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"Thumbprint\" col")
					tp, tpOk := action["Thumbprint"].(string)
					if !tpOk {
						tp = ""
					}

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"Alias\" col")
					alias, aliasOk := action["Alias"].(string)
					if !aliasOk {
						alias = ""
					}
					log.Debug().Str("alias", alias).Send()

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"CertID\" col")
					cid, cidOk := action["CertID"].(int)
					if !cidOk {
						cid = -1
					}

					if !tpOk && !cidOk {
						outputError(
							fmt.Errorf(
								fmt.Sprintf(
									"Missing Thumbprint or CertID for row '%d' in report file '%s'",
									ri,
									reportFile,
								),
							), false, outputFormat,
						)
						log.Error().
							Str("reportFile", reportFile).
							Int("row", ri).Msg("missing thumbprint or certID for row")
						continue
					}

					log.Debug().
						Str("reportFile", reportFile).
						Int("row", ri).
						Msg("parsing \"StoreID\" col")
					sId, sIdOk := action["StoreID"].(string)
					if !sIdOk {
						sIdErr := fmt.Errorf("missing 'StoreID' for row '%d' in report file '%s'", ri, reportFile)
						outputError(sIdErr, true, outputFormat)
						log.Error().Err(sIdErr).Str("reportFile", reportFile).
							Int("row", ri).Msg("invalid action")
						continue
					}

					if cid == -1 && tp != "" {
						log.Debug().Msg("creating lookup by thumbprint request")
						certLookupReq := api.GetCertificateContextArgs{
							IncludeMetadata:  boolToPointer(true),
							IncludeLocations: boolToPointer(true),
							CollectionId:     nil,
							Thumbprint:       tp,
							Id:               0,
						}
						certLookup, certLookupErr := kfClient.GetCertificateContext(&certLookupReq)
						if certLookupErr != nil {
							outputError(certLookupErr, true, outputFormat)
							log.Error().Err(certLookupErr).Str("thumbprint", tp).Msg("failed looking up cert")
							continue
						}
						cid = certLookup.Id
					}

					log.Debug().
						Int("row", ri).
						Str("reportFile", reportFile).
						Str("StoreID", sId).
						Str("StoreType", sType).
						Str("Path", sPath).
						Str("Thumbprint", tp).
						Str("CertID", fmt.Sprintf("%d", cid)).
						Bool("AddCert", addCert).
						Bool("RemoveCert", removeCert).
						Msg("creating ROTAction")
					a := ROTAction{
						StoreID:    sId,
						StoreType:  sType,
						StorePath:  sPath,
						Thumbprint: tp,
						CertID:     cid,
						AddCert:    addCert,
						RemoveCert: removeCert,
					}

					actions[a.Thumbprint] = append(actions[a.Thumbprint], a)
				}
				if len(actions) == 0 {
					outputResult(
						"No reconciliation actions to take, root stores are up-to-date. Exiting.",
						outputFormat,
					)
					log.Info().
						Str("reportFile", reportFile).
						Msg("No reconciliation actions to take, root stores are up-to-date. Exiting.")
					return nil
				}

				log.Debug().Msg("reconciling roots")
				rErr := reconcileRoots(actions, kfClient, reportFile, dryRun)
				if rErr != nil {
					log.Error().
						Err(rErr).
						Str("reportFile", reportFile).
						Msg("failed reconciling roots")
					return rErr
				}
				defer csvFile.Close()

				jobStatusURL := fmt.Sprintf(
					"https://%s/KeyfactorPortal/AgentJobStatus/Index",
					kfClient.AuthClient.GetServerConfig().Host,
				)

				log.Info().Str("reportFile", reportFile).
					Str("jobStatusURL", jobStatusURL).
					Msg("reconciliation complete")
				outputResult("Reconciliation complete. Job status URL: "+jobStatusURL, outputFormat)
				return nil
			} else {
				// Read in the stores CSV
				log.Debug().
					Str("storesFile", storesFile).
					Msg("opening stores CSV file")
				csvFile, csvErr := os.Open(storesFile)
				if csvErr != nil {
					outputError(csvErr, true, outputFormat)
					log.Error().
						Err(csvErr).
						Str("storesFile", storesFile).
						Msg("failed opening CSV file")
					return csvErr
				}

				defer csvFile.Close()

				log.Debug().
					Str("storesFile", storesFile).
					Msg("reading stores CSV file data")
				reader := csv.NewReader(bufio.NewReader(csvFile))
				storeEntries, stErr := reader.ReadAll()
				if stErr != nil {
					log.Error().Err(stErr).Str("storesFile", storesFile).
						Msg("failed reading CSV file")
					return stErr
				}
				if len(storeEntries) == 0 {
					noStoresErr := fmt.Errorf("no stores found in CSV file")
					outputError(noStoresErr, true, outputFormat)
					log.Error().
						Str("storesFile", storesFile).
						Err(noStoresErr).
						Send()
					return fmt.Errorf("no stores found in CSV file")
				}

				log.Debug().Str("storesFile", storesFile).
					Int("storeEntries", len(storeEntries)).
					Msg("processing stores CSV file data")
				var stores = make(map[string]StoreCSVEntry)
				for i, entry := range storeEntries {
					if entry[0] == "StoreID" || entry[0] == "StoreId" || i == 0 {
						log.Debug().Str("storesFile", storesFile).Msg("skipping file header")
						continue // Skip header
					}

					log.Debug().Str("storesFile", storesFile).
						Str("StoreID", entry[0]).
						Msg("calling GetCertificateStoreByID for store")
					apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
					if err != nil {
						log.Error().
							Err(err).Str("StoreID", entry[0]).
							Msg("unable to get certificate by ID")
						lookupFailures = append(lookupFailures, entry[0])
						continue
					}
					//inventory, invErr := kfClient.GetCertStoreInventoryV1(entry[0])
					inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
					if invErr != nil {
						outputError(invErr, true, outputFormat)
						log.Error().
							Err(invErr).
							Str("storesFile", storesFile).
							Str("StoreID", entry[0]).
							Msg("unable to get inventory")
						continue
					}

					if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
						log.Info().Str("storesFile", storesFile).
							Str("StoreID", entry[0]).
							Int("minCerts", minCerts).
							Int("maxLeaves", maxLeaves).
							Int("maxKeys", maxKeys).
							Msg("is not a root store")
						//lookupFailures = append(lookupFailures, entry[0])
						continue
					} else {
						log.Info().Str("storesFile", storesFile).
							Str("StoreID", entry[0]).
							Msg("is a root store")
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
						for _, v := range thumb {
							stores[entry[0]].Thumbprints[v] = true
						}
						for _, v := range cert.Serials {
							stores[entry[0]].Serials[v] = true
						}
						for _, v := range cert.Ids {
							stores[entry[0]].Ids[v] = true
						}
					}

				}
				if len(lookupFailures) > 0 {
					failedErr := fmt.Errorf(
						"the following stores were not found: %s",
						strings.Join(lookupFailures, ","),
					)
					outputError(failedErr, true, outputFormat)
					log.Error().
						Err(failedErr).
						Strs("lookupFailures", lookupFailures).
						Msg("failed to lookup stores")
					return failedErr
				}
				if len(stores) == 0 {
					noStoresErr := fmt.Errorf("no stores found in CSV file %s", storesFile)
					outputError(noStoresErr, true, outputFormat)
					return noStoresErr
				}
				// Read in the add addCerts CSV
				var certsToAdd = make(map[string]string)
				if addRootsFile != "" {
					log.Debug().Str("addRootsFile", addRootsFile).Msg("calling readCerts")
					certsToAdd, _ = readCertsFile(addRootsFile, kfClient)
					//TODO: Handle error here?
				} else {
					log.Info().Str("addRootsFile", addRootsFile).Msg("no certs to add to trust stores")
				}

				// Read in the remove removeCerts CSV
				var certsToRemove = make(map[string]string)
				if removeRootsFile != "" {
					log.Debug().Str("removeRootsFile", removeRootsFile).Msg("calling readCerts")
					certsToRemove, _ = readCertsFile(removeRootsFile, kfClient)
					//TODO: Handle error here?
				} else {
					log.Info().Str("removeRootsFile", removeRootsFile).Msg("no certs to remove from trust stores")
				}

				log.Debug().
					Str("storesFile", storesFile).
					Str("addRootsFile", addRootsFile).
					Interface("certsToAdd", certsToAdd).
					Str("removeRootsFile", removeRootsFile).
					Interface("certsToRemove", certsToRemove).
					Str("outpath", outpath).
					Msg("calling generateAuditReport")
				_, actions, err := generateAuditReport(certsToAdd, certsToRemove, stores, outpath, kfClient)
				if err != nil {
					outputError(err, true, outputFormat)
					log.Error().Err(err).
						Str("storesFile", storesFile).
						Str("addRootsFile", addRootsFile).
						Str("removeRootsFile", removeRootsFile).
						Str("outpath", outpath).
						Msg("failed to generate audit report")
					return err
				}
				if len(actions) == 0 {
					log.Info().Str("storesFile", storesFile).
						Str("addRootsFile", addRootsFile).
						Str("removeRootsFile", removeRootsFile).
						Str("outpath", outpath).
						Msg("no reconciliation actions to take, root stores are up-to-date")
					outputResult(
						"No reconciliation actions to take, root stores are up-to-date. Exiting.",
						outputFormat,
					)
					return nil
				}

				log.Debug().Str("reportFile", reportFile).Msg("calling reconcileRoots")
				rErr := reconcileRoots(actions, kfClient, reportFile, dryRun)
				if rErr != nil {
					outputError(rErr, true, outputFormat)
					log.Error().
						Err(rErr).
						Str("reportFile", reportFile).Msg("failed reconciling roots")
					return rErr
				}
				if lookupFailures != nil {
					lookupErr := fmt.Errorf(
						"the following stores could not be found: %s",
						strings.Join(lookupFailures, ","),
					)
					outputError(lookupErr, true, outputFormat)
					log.Error().
						Err(lookupErr).
						Strs("lookupFailures", lookupFailures).
						Msg("failed to lookup stores")
					return lookupErr
				}
				orchsURL := fmt.Sprintf(
					"https://%s/KeyfactorPortal/AgentJobStatus/Index",
					kfClient.AuthClient.GetServerConfig().Host,
				)
				log.Info().
					Str("reportFile", reportFile).
					Str("orchsURL", orchsURL).
					Msg("reconciliation complete")
				outputResult(
					fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL),
					outputFormat,
				)
				return nil
			}

		},
		Run:                        nil,
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
		Long:                   `Root Of Trust: Will parse a CSV and attempt to deploy a cert or set of certs into a list of cert stores.`,
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Global flags
			cmd.SilenceUsage = true
			// expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			stdlog.SetOutput(io.Discard)
			informDebug(debugFlag)

			// Authenticate
			kfClient, cErr := initClient(false)
			if cErr != nil {
				log.Error().Err(cErr).Msg("unable to authenticate")
				return cErr
			}

			templateType, _ := cmd.Flags().GetString("type")
			format, _ := cmd.Flags().GetString("format")
			outPath, _ := cmd.Flags().GetString("outpath")
			storeType, _ := cmd.Flags().GetStringSlice("store-type")
			containerName, _ := cmd.Flags().GetStringSlice("container-name")
			collection, _ := cmd.Flags().GetStringSlice("collection")
			subjectName, _ := cmd.Flags().GetStringSlice("cn")
			log.Debug().Str("templateType", templateType).
				Str("format", format).
				Str("outPath", outPath).
				Strs("storeType", storeType).
				Strs("containerName", containerName).
				Strs("collection", collection).
				Strs("subjectName", subjectName).
				Msg("flags")
			//if templateType == "" {
			//	return fmt.Errorf("template type must be specified")
			//}

			stID := -1
			var storeData []api.GetCertificateStoreResponse
			var csvStoreData [][]string
			var csvCertData [][]string
			var rowLookup = make(map[string]bool)
			if len(storeType) != 0 {
				for _, s := range storeType {
					log.Debug().Str("storeType", s).Msg("processing store-type")
					var sType *api.CertificateStoreType
					var stErr error
					if s == "all" {
						log.Info().Str("storeType", s).Msg("getting all stores")
						sType = &api.CertificateStoreType{
							Name:                "",
							ShortName:           "",
							Capability:          "",
							StoreType:           0,
							ImportType:          0,
							LocalStore:          false,
							SupportedOperations: nil,
							Properties:          nil,
							EntryParameters:     nil,
							PasswordOptions:     nil,
							StorePathType:       "",
							StorePathValue:      "",
							PrivateKeyAllowed:   "",
							JobProperties:       nil,
							ServerRequired:      false,
							PowerShell:          false,
							BlueprintAllowed:    false,
							CustomAliasAllowed:  "",
							ServerRegistration:  0,
							InventoryEndpoint:   "",
							InventoryJobType:    "",
							ManagementJobType:   "",
							DiscoveryJobType:    "",
							EnrollmentJobType:   "",
						}
					} else {
						// check if s is an int
						log.Debug().Str("storeType", s).Msg("checking if store-type is an int")
						sInt, err := strconv.Atoi(s)
						if err == nil {
							log.Debug().Str("storeType", s).Msg("calling GetCertificateStoreByID")
							sType, stErr = kfClient.GetCertificateStoreTypeById(sInt)
						} else {
							log.Debug().Str("storeType", s).Msg("calling GetCertificateStoreByName")
							sType, stErr = kfClient.GetCertificateStoreTypeByName(s)
						}
						if stErr != nil {
							outputError(stErr, true, format)
							log.Error().Err(stErr).Str("storeType", s).Msg("failed to get stores")
							continue
						}
						stID = sType.StoreType // This is the template type ID
					}

					if stID >= 0 || s == "all" {
						log.Debug().
							Int("stID", stID).
							Str("storeType", s).
							Msg("valid store type")
						params := make(map[string]interface{})

						log.Debug().Str("storeType", s).Msg("calling ListCertificateStores")
						stores, sErr := kfClient.ListCertificateStores(&params)
						if sErr != nil {
							log.Error().Err(sErr).Str("storeType", s).Msg("failed to get stores")
							outputError(sErr, true, format)
							return sErr
						}

						if stores == nil {
							invalidRespErr := fmt.Errorf("invalid response from Keyfactor Command when listing certificate stores")
							log.Error().Err(invalidRespErr).Str("storeType", s).Msg("stores is nil")
							outputError(invalidRespErr, true, format)
							return invalidRespErr
						}
						log.Debug().Str("storeType", s).Msg("processing stores")

						for _, store := range *stores {
							if store.CertStoreType == stID || s == "all" {
								storeData = append(storeData, store)
								if !rowLookup[store.Id] {
									lineData := []string{
										//"StoreID", "StoreType", "StoreMachine", "StorePath", "ContainerId"
										store.Id,
										fmt.Sprintf("%s", sType.ShortName),
										store.ClientMachine,
										store.StorePath,
										fmt.Sprintf("%d", store.ContainerId),
										store.ContainerName,
										getCurrentTime(""),
									}
									log.Debug().Str("storeType", s).
										Strs("lineData", lineData).
										Msg("adding line data")
									csvStoreData = append(csvStoreData, lineData)
									rowLookup[store.Id] = true
								}
							}
						}
					}
				}
				log.Info().Msg("lookups by store-type completed")
			}
			containers := len(containerName)
			if containers > 0 {
				log.Info().
					Int("containers", containers).
					Msg("processing container-names")
				for _, c := range containerName {
					cStoresResp, scErr := kfClient.GetCertificateStoreByContainerID(c)
					if scErr != nil {
						log.Error().Err(scErr).Str("containerName", c).Msg("failed to get stores by container name")
						return cErr
					}
					if cStoresResp == nil {
						invalidRespErr := fmt.Errorf(
							"invalid response from Keyfactor Command when listing stores by container name '%s'",
							c,
						)
						outputError(invalidRespErr, true, format)
						log.Error().Err(invalidRespErr).Str("containerName", c).Msg("invalid response")
						return invalidRespErr
					}
					for _, store := range *cStoresResp {
						log.Debug().
							Str("containerName", c).
							Int("storeType", store.CertStoreType).
							Msg("calling GetCertificateStoreType")
						sType, stErr := kfClient.GetCertificateStoreType(store.CertStoreType)
						if stErr != nil {
							outputError(stErr, false, format)
							log.Error().Err(stErr).Str(
								"containerName",
								c,
							).Msg("failed to get store-type by container name")
							continue
						}
						storeData = append(storeData, store)
						if !rowLookup[store.Id] {
							lineData := []string{
								// "StoreID", "StoreType", "StoreMachine", "StorePath", "ContainerId"
								store.Id,
								sType.ShortName,
								store.ClientMachine,
								store.StorePath,
								fmt.Sprintf("%d", store.ContainerId),
								store.ContainerName,
								getCurrentTime(""),
							}
							log.Debug().Str("containerName", c).
								Strs("lineData", lineData).
								Msg("adding line data")
							csvStoreData = append(csvStoreData, lineData)
							rowLookup[store.Id] = true
						}
					}
				}
				log.Info().Msg("lookups by container-name completed")
			}

			collections := len(collection)
			if collections > 0 {
				log.Info().
					Int("collections", collections).
					Msg("processing collections")
				for _, c := range collection {
					q := make(map[string]string)
					q["collection"] = c
					log.Debug().Str("collection", c).Msg("calling ListCertificateStores")
					certsResp, scErr := kfClient.ListCertificates(q)
					if scErr != nil {
						log.Error().Err(scErr).Str("collection", c).Msg("failed to list certificates by collection")
						outputError(scErr, true, format)
						return scErr
					}
					if certsResp == nil {
						invalidRespErr := fmt.Errorf(
							"invalid response from Keyfactor Command when listing certificates by collection '%s'",
							c,
						)
						outputError(invalidRespErr, true, format)
						log.Error().Err(invalidRespErr).Str("collection", c).Msg("invalid response")
						return invalidRespErr
					}
					for _, cert := range certsResp {
						if !rowLookup[cert.Thumbprint] {
							lineData := []string{
								// "Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"
								cert.Thumbprint,
								cert.IssuedCN,
								cert.IssuerDN,
								fmt.Sprintf("%d", cert.Id),
								fmt.Sprintf("%v", cert.Locations),
								getCurrentTime(""),
							}
							log.Debug().
								Str("collection", c).
								Strs("lineData", lineData).
								Msg("adding line data")
							csvCertData = append(csvCertData, lineData)
							rowLookup[cert.Thumbprint] = true
						}
					}
				}
				log.Info().
					Int("collections", collections).
					Msg("lookups by collection completed")
			}

			cns := len(subjectName)
			if cns > 0 {
				log.Info().
					Int("subjectNames", cns).
					Msg("processing subject-names")
				for _, s := range subjectName {
					q := make(map[string]string)
					q["subject"] = s

					log.Debug().Str("subjectName", s).Msg("calling ListCertificates")
					certsResp, scErr := kfClient.ListCertificates(q)
					if scErr != nil {
						log.Error().Err(scErr).Str("subjectName", s).Msg("failed to list certificates by subject name")
						outputError(scErr, true, format)
						return scErr
					}
					if certsResp == nil {
						invalidRespErr := fmt.Errorf(
							"invalid response returned from Keyfactor Command when calling ListCertificates by subject name '%s'",
							s,
						)
						log.Error().
							Err(invalidRespErr).
							Str("subjectName", s).
							Send()
						outputError(invalidRespErr, true, format)
						return invalidRespErr
					}

					log.Debug().
						Str("subjectName", s).
						Msg("processing certs to build 'thumbprint' to 'locations map'")
					for _, cert := range certsResp {
						if rowLookup[cert.Thumbprint] {
							log.Debug().Str(
								"thumbprint",
								cert.Thumbprint,
							).Msg("thumbprint already exists in lookup, skipping")
							continue
						}
						locationsFormatted := ""
						for _, loc := range cert.Locations {
							locationsFormatted += fmt.Sprintf("%s:%s\n", loc.StoreMachine, loc.StorePath)
							log.Debug().
								Str("thumbprint", cert.Thumbprint).
								Str("storePath", loc.StorePath).
								Str("formattedLocations", locationsFormatted).
								Msg("processing location")
						}
						lineData := []string{
							// "Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"
							cert.Thumbprint,
							cert.IssuedCN,
							cert.IssuerDN,
							fmt.Sprintf("%d", cert.Id),
							locationsFormatted,
							getCurrentTime(""),
						}
						log.Debug().
							Str("subjectName", s).
							Strs("lineData", lineData).
							Msg("adding line data")
						csvCertData = append(csvCertData, lineData)
						rowLookup[cert.Thumbprint] = true
					}
				}
				log.Info().Int("subjectNames", cns).Msg("lookups by subject-name completed")
			}

			var filePath string
			if outPath != "" {
				filePath = outPath
			} else {
				filePath = fmt.Sprintf("%s_template.%s", templateType, format)
			}

			log.Debug().Str("filePath", filePath).Msg("writing file")
			file, err := os.Create(filePath)
			defer file.Close()
			if err != nil {
				log.Error().Err(err).Str("filePath", filePath).Msg("failed to create file")
				outputError(err, true, format)
				return err
			}

			switch format {
			case "csv":
				log.Debug().
					Str("filePath", filePath).
					Str("templateType", templateType).
					Msg("writing csv")
				writer := csv.NewWriter(file)
				var data [][]string

				switch templateType {
				case "stores":
					log.Debug().
						Str("filePath", filePath).
						Str("templateType", templateType).
						Msg("writing stores csv")
					data = append(data, StoreHeader)
					if len(csvStoreData) != 0 {
						data = append(data, csvStoreData...)
					}
				case "certs":
					log.Debug().
						Str("filePath", filePath).
						Str("templateType", templateType).
						Msg("writing certs csv")
					data = append(data, CertHeader)
					if len(csvCertData) != 0 {
						data = append(data, csvCertData...)
					}
				case "actions":
					log.Debug().
						Str("filePath", filePath).
						Str("templateType", templateType).
						Msg("writing audit csv")
					data = append(data, AuditHeader)
				}
				csvErr := writer.WriteAll(data)
				if csvErr != nil {
					log.Error().Err(csvErr).Str("filePath", filePath).Msg("failed to write csv")
					outputError(csvErr, true, format)
					return csvErr
				}
			case "json":
				log.Debug().
					Str("filePath", filePath).
					Str("templateType", templateType).
					Msg("writing json")
				writer := bufio.NewWriter(file)
				_, err := writer.WriteString("StoreID,StoreType,StoreMachine,StorePath")
				if err != nil {
					log.Error().Err(err).Str("filePath", filePath).Msg("failed to write json")
					outputError(err, true, format)
					return err
				}
			}
			log.Info().Str("filePath", filePath).Msg("template file written")
			outputResult(fmt.Sprintf("Template generated at %s", filePath), outputFormat)
			return nil
		},
		Run:                        nil,
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
		inputFile       string
		storeTypes      []string
		containerNames  []string
		collections     []string
		subjectNames    []string
	)

	storesCmd.AddCommand(rotCmd)

	// Root of trust `audit` command
	rotCmd.AddCommand(rotAuditCmd)
	rotAuditCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotAuditCmd.Flags().StringVarP(
		&addCerts, "add-certs", "a", "",
		"CSV file containing cert(s) to enroll into the defined cert stores",
	)
	rotAuditCmd.Flags().StringVarP(
		&removeCerts, "remove-certs", "r", "",
		"CSV file containing cert(s) to remove from the defined cert stores",
	)
	rotAuditCmd.Flags().IntVarP(
		&minCertsInStore,
		"min-certs",
		"m",
		-1,
		"The minimum number of certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.",
	)
	rotAuditCmd.Flags().IntVarP(
		&maxPrivateKeys,
		"max-keys",
		"k",
		-1,
		"The max number of private keys that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.",
	)
	rotAuditCmd.Flags().IntVarP(
		&maxLeaves,
		"max-leaf-certs",
		"l",
		-1,
		"The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.",
	)
	rotAuditCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotAuditCmd.Flags().StringVarP(
		&outPath, "outpath", "o", "",
		"Path to write the audit report file to. If not specified, the file will be written to the current directory.",
	)

	// Root of trust `reconcile` command
	rotCmd.AddCommand(rotReconcileCmd)
	rotReconcileCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotReconcileCmd.Flags().StringVarP(
		&addCerts, "add-certs", "a", "",
		"CSV file containing cert(s) to enroll into the defined cert stores",
	)
	rotReconcileCmd.Flags().StringVarP(
		&removeCerts, "remove-certs", "r", "",
		"CSV file containing cert(s) to remove from the defined cert stores",
	)
	rotReconcileCmd.Flags().IntVarP(
		&minCertsInStore,
		"min-certs",
		"m",
		-1,
		"The minimum number of certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.",
	)
	rotReconcileCmd.Flags().IntVarP(
		&maxPrivateKeys,
		"max-keys",
		"k",
		-1,
		"The max number of private keys that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.",
	)
	rotReconcileCmd.Flags().IntVarP(
		&maxLeaves,
		"max-leaf-certs",
		"l",
		-1,
		"The max number of non-root-certs that should be in a store to be considered a 'root' store. If set to `-1` then all stores will be considered.",
	)
	rotReconcileCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotReconcileCmd.Flags().BoolP("import-csv", "v", false, "Import an audit report file in CSV format.")
	rotReconcileCmd.Flags().StringVarP(
		&inputFile, "input-file", "i", reconcileDefaultFileName,
		"Path to a file generated by 'stores rot audit' command.",
	)
	rotReconcileCmd.Flags().StringVarP(
		&outPath, "outpath", "o", "",
		"Path to write the audit report file to. If not specified, the file will be written to the current directory.",
	)
	//rotReconcileCmd.MarkFlagsRequiredTogether("add-certs", "stores")
	//rotReconcileCmd.MarkFlagsRequiredTogether("remove-certs", "stores")
	rotReconcileCmd.MarkFlagsMutuallyExclusive("add-certs", "import-csv")
	rotReconcileCmd.MarkFlagsMutuallyExclusive("remove-certs", "import-csv")
	rotReconcileCmd.MarkFlagsMutuallyExclusive("stores", "import-csv")

	// Root of trust `generate` command
	rotCmd.AddCommand(rotGenStoreTemplateCmd)
	rotGenStoreTemplateCmd.Flags().StringVarP(
		&outPath, "outpath", "o", "",
		"Path to write the template file to. If not specified, the file will be written to the current directory.",
	)
	rotGenStoreTemplateCmd.Flags().StringVarP(
		&outputFormat, "format", "f", "csv",
		"The type of template to generate. Only `csv` is supported at this time.",
	)
	rotGenStoreTemplateCmd.Flags().Var(
		&tType, "type",
		`The type of template to generate. Only "certs|stores|actions" are supported at this time.`,
	)
	rotGenStoreTemplateCmd.Flags().StringSliceVar(
		&storeTypes,
		"store-type",
		[]string{},
		"Multi value flag. Attempt to pre-populate the stores template with the certificate stores matching specified store types. If not specified, the template will be empty.",
	)
	rotGenStoreTemplateCmd.Flags().StringSliceVar(
		&containerNames,
		"container-name",
		[]string{},
		"Multi value flag. Attempt to pre-populate the stores template with the certificate stores matching specified container types. If not specified, the template will be empty.",
	)
	rotGenStoreTemplateCmd.Flags().StringSliceVar(
		&subjectNames,
		"cn",
		[]string{},
		"Subject name(s) to pre-populate the 'certs' template with. If not specified, the template will be empty. Does not work with SANs.",
	)
	rotGenStoreTemplateCmd.Flags().StringSliceVar(
		&collections,
		"collection",
		[]string{},
		"Certificate collection name(s) to pre-populate the stores template with. If not specified, the template will be empty.",
	)

	rotGenStoreTemplateCmd.RegisterFlagCompletionFunc("type", templateTypeCompletion)
	rotGenStoreTemplateCmd.MarkFlagRequired("type")
}
