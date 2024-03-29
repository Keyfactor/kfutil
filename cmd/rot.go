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
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	sdk "github.com/Keyfactor/keyfactor-go-client-sdk/api/keyfactor"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type templateType string

const (
	tTypeCerts               templateType = "certs"
	reconcileDefaultFileName string       = "rot_audit.csv"
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
		"certsGenerates template CSV for certificate input to be used w/ `--add-certs` or `--remove-certs`",
		"storesGenerates template CSV for certificate input to be used w/ `--stores`",
		"actionsGenerates template CSV for certificate input to be used w/ `--actions`",
	}, cobra.ShellCompDirectiveDefault
}

func generateAuditReport(
	addCerts map[string]string,
	removeCerts map[string]string,
	stores map[string]StoreCSVEntry,
	outputFilePath string,
	kfClient *api.Client,
) ([][]string, map[string][]ROTAction, error) {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "generateAuditReport"))

	log.Info().Str("output_file", outputFilePath).Msg("Generating audit report")
	var (
		data [][]string
	)

	data = append(data, AuditHeader)
	var csvFile *os.File
	var fErr error
	log.Debug().Str("output_file", outputFilePath).Msg("Checking for output file")
	if outputFilePath == "" {
		log.Debug().Str("output_file", reconcileDefaultFileName).Msg("No output file specified, using default")
		csvFile, fErr = os.Create(reconcileDefaultFileName)
		outputFilePath = reconcileDefaultFileName
	} else {
		csvFile, fErr = os.Create(outputFilePath)
	}

	if fErr != nil {
		fmt.Printf("%s", fErr)
		log.Error().Err(fErr).Str("output_file", outputFilePath).Msg("Error creating output file")
	}

	log.Trace().Str("output_file", outputFilePath).Msg("Creating CSV writer")
	csvWriter := csv.NewWriter(csvFile)
	log.Debug().Str("output_file", outputFilePath).Strs("csv_header", AuditHeader).Msg("Writing header to CSV")
	cErr := csvWriter.Write(AuditHeader)
	if cErr != nil {
		log.Error().Err(cErr).Str("output_file", outputFilePath).Msg("Error writing header to CSV")
		return nil, nil, cErr
	}

	log.Trace().Str("output_file", outputFilePath).Msg("Creating actions map")
	actions := make(map[string][]ROTAction)

	var errs []error
	for tp, cId := range addCerts {
		log.Debug().Str("thumbprint", tp).
			Str("cert_id", cId).
			Msg("Looking up certificate")
		certLookupReq := api.GetCertificateContextArgs{}
		if cId != "" {
			certIdInt, cErr := strconv.Atoi(cId)
			if cErr != nil {
				log.Error().
					Err(cErr).
					Str("thumbprint", tp).
					Msg("Error converting cert ID to integer, skipping")
				errs = append(errs, cErr)
				continue
			}
			certLookupReq = api.GetCertificateContextArgs{
				IncludeMetadata:  boolToPointer(true),
				IncludeLocations: boolToPointer(true),
				CollectionId:     nil, //todo: add CollectionID support
				Thumbprint:       "",
				Id:               certIdInt,
			}
		} else {
			certLookupReq = api.GetCertificateContextArgs{
				IncludeMetadata:  boolToPointer(true),
				IncludeLocations: boolToPointer(true),
				CollectionId:     nil, //todo: add CollectionID support
				Thumbprint:       tp,
				Id:               0, //todo: should also allow KFC ID
			}
		}

		log.Debug().
			Str("thumbprint", tp).
			Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificateContext"))
		certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
		if err != nil {
			log.Error().
				Err(err).
				Str("thumbprint", tp).
				Msg("Error looking up certificate, skipping")
			errMsg := fmt.Errorf(
				"error recieved from Keyfactor Command when looking up thumbprint '%s':'%w'",
				tp,
				err,
			)
			errs = append(errs, errMsg)
			continue
		}
		certID := certLookup.Id
		certIDStr := strconv.Itoa(certID)
		log.Debug().Str("thumbprint", tp).Msg("Iterating over stores")
		for _, store := range stores {
			log.Debug().Str("thumbprint", tp).Str("store_id", store.ID).Msg("Checking if cert is deployed to store")
			if _, ok := store.Thumbprints[tp]; ok {
				// Cert is already in the store do nothing
				log.Info().Str("thumbprint", tp).Str("store_id", store.ID).Msg("Cert is already deployed to store")
				row := []string{
					//todo: this should be a toCSV field on whatever object this is
					tp,
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
				log.Trace().Str("thumbprint", tp).Strs("row", row).Msg("Appending data row")
				data = append(data, row)
				log.Trace().Str("thumbprint", tp).Strs("row", row).Msg("Writing data row to CSV")
				wErr := csvWriter.Write(row)
				if wErr != nil {
					log.Error().
						Err(wErr).
						Str("thumbprint", tp).
						Str("output_file", outputFilePath).
						Strs("row", row).
						Msg("Error writing row to CSV")
				}
			} else {
				// Cert is not deployed to this store and will need to be added
				log.Info().
					Str("thumbprint", tp).
					Str("store_id", store.ID).
					Msg("Cert is not deployed to store")
				row := []string{
					//todo: this should be a toCSV
					tp,
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
				log.Trace().
					Str("thumbprint", tp).
					Strs("row", row).
					Msg("Appending data row")
				data = append(data, row)
				log.Debug().
					Str("thumbprint", tp).
					Strs("row", row).
					Msg("Writing data row to CSV")
				wErr := csvWriter.Write(row)
				if wErr != nil {
					log.Error().
						Err(wErr).
						Str("thumbprint", tp).
						Str("output_file", outputFilePath).
						Strs("row", row).
						Msg("Error writing row to CSV")
				}
				log.Debug().
					Str("thumbprint", tp).
					Msg("Adding 'add' action to actions map")
				actions[tp] = append(
					actions[tp], ROTAction{
						Thumbprint: tp,
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
		log.Debug().Str("thumbprint", cert).Msg("Looking up certificate to remove")
		certLookupReq := api.GetCertificateContextArgs{
			IncludeMetadata:  boolToPointer(true),
			IncludeLocations: boolToPointer(true),
			CollectionId:     nil, //todo: add support for collection ID
			Thumbprint:       cert,
			Id:               0,
		}
		log.Debug().Str("thumbprint", cert).Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificateContext"))
		certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
		if err != nil {
			log.Error().
				Err(err).
				Str("thumbprint", cert).
				Msg("Error looking up certificate, unable to remove from store")
			errs = append(errs, err)
			continue
		} else if certLookup == nil {
			log.Error().
				Err(ErrKfcEmptyResponse).
				Str("thumbprint", cert).
				Msg(fmt.Sprintf("%s when looking up certificate", ErrMsgEmptyResponse))
			errs = append(errs, ErrKfcEmptyResponse)
			continue
		}

		certID := certLookup.Id
		log.Trace().
			Str("thumbprint", cert).
			Int("cert_id", certID).
			Msg("Converting cert ID to string")
		certIDStr := strconv.Itoa(certID)
		for _, store := range stores {
			storeIdentifier := fmt.Sprintf("%s/%s", store.Machine, store.Path)
			log.Debug().Str("thumbprint", cert).
				Str("store_id", store.ID).
				Str("store_name", storeIdentifier).
				Msg("Checking if cert is deployed to store")
			if _, ok := store.Thumbprints[cert]; ok {
				// Cert is deployed to this store and will need to be removed
				log.Info().
					Str("thumbprint", cert).
					Str("store_id", store.ID).
					Str("store_name", storeIdentifier).
					Msg("Cert is deployed to store")
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
				log.Trace().
					Str("thumbprint", cert).
					Strs("row", row).
					Msg("Appending data row")
				data = append(data, row)
				log.Debug().
					Str("thumbprint", cert).
					Strs("row", row).
					Msg("Writing data row to CSV")
				wErr := csvWriter.Write(row)
				if wErr != nil {
					log.Error().
						Err(wErr).
						Str("thumbprint", cert).
						Str("output_file", outputFilePath).
						Strs("row", row).
						Msg("Error writing row to CSV")
					errs = append(errs, wErr)
					//todo: continue?
				}
				log.Debug().Str("thumbprint", cert).Msg("Adding remove action to actions map")
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
				log.Info().Str("thumbprint", cert).Str(
					"store_id",
					store.ID,
				).Msg("Cert is not deployed to store, skipping")
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
				log.Trace().Str("thumbprint", cert).Strs("row", row).Msg("Appending data row")
				data = append(data, row)
				log.Debug().Str("thumbprint", cert).Strs("row", row).Msg("Writing data row to CSV")
				wErr := csvWriter.Write(row)
				if wErr != nil {
					log.Error().Err(wErr).Str("thumbprint", cert).Str("output_file", outputFilePath).Strs(
						"row",
						row,
					).Msg("Error writing row to CSV")
					errs = append(errs, wErr)
				}
			}
		}
	}
	log.Trace().
		Str("output_file", outputFilePath).
		Msg("Flushing CSV writer")
	csvWriter.Flush()
	log.Trace().
		Str("output_file", outputFilePath).
		Msg("Closing CSV file")
	ioErr := csvFile.Close()
	if ioErr != nil {
		log.Error().
			Err(ioErr).
			Str("output_file", outputFilePath).
			Msg("Error closing CSV file")
	}
	log.Info().
		Str("output_file", outputFilePath).
		Msg("Audit report written to disk successfully")
	fmt.Printf("Audit report written to %s\n", outputFilePath) //todo: remove or propagate message to CLI
	fmt.Printf(
		"Please review the report and run `kfutil stores rot reconcile --import-csv --input"+
			"-file %s` apply the changes\n", outputFilePath,
	)

	if len(errs) > 0 {
		errStr := mergeErrsToString(&errs, false)
		log.Trace().Str("output_file", outputFilePath).Str(
			"errors",
			errStr,
		).Msg("The following errors occurred while generating audit report")
		return data, actions, fmt.Errorf("the following errors occurred while generating audit report:\r\n%s", errStr)
	}
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "generateAuditReport"))
	return data, actions, nil
}

func reconcileRoots(actions map[string][]ROTAction, kfClient *api.Client, reportFile string, dryRun bool) error {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "reconcileRoots"))
	if len(actions) == 0 {
		log.Info().Msg("No actions to reconcile detected, root of trust stores are up-to-date.")
		return nil
	}
	log.Info().Msg("Reconciling root of trust stores")

	rFileName := fmt.Sprintf("%s_reconciled.csv", strings.Split(reportFile, ".csv")[0])
	log.Debug().
		Str("report_file", reportFile).
		Str("reconciled_file", rFileName).
		Msg("Creating reconciled report file")
	csvFile, fErr := os.Create(rFileName)
	if fErr != nil {
		log.Error().
			Err(fErr).
			Str("reconciled_file", rFileName).
			Msg("Error creating reconciled report file")
		return fErr
	}
	log.Trace().Str("reconciled_file", rFileName).Msg("Creating CSV writer")
	csvWriter := csv.NewWriter(csvFile)

	log.Debug().Str("reconciled_file", rFileName).Strs("csv_header", ReconciledAuditHeader).Msg("Writing header to CSV")
	cErr := csvWriter.Write(ReconciledAuditHeader)
	if cErr != nil {
		log.Error().Err(cErr).Str("reconciled_file", rFileName).Msg("Error writing header to CSV")
		return cErr
	}
	log.Info().Str("report_file", reportFile).Msg("Processing reconciliation actions")
	var errs []error
	for thumbprint, action := range actions {
		for _, a := range action {
			if a.AddCert {
				if !dryRun {
					log.Info().Str("thumbprint", thumbprint).Str("store_id", a.StoreID).Str(
						"store_path",
						a.StorePath,
					).Msg("Attempting to add cert to store")
					log.Debug().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating orchestrator 'add' job request")

					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating certificate store object")
					apiStore := api.CertificateStore{
						CertificateStoreId: a.StoreID,
						Overwrite:          true,
					}

					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating certificate store array")
					var stores []api.CertificateStore
					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Appending certificate store to array")
					stores = append(stores, apiStore)

					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating inventory 'immediate' schedule")
					schedule := &api.InventorySchedule{
						Immediate: boolToPointer(true),
					}

					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating add certificate request")
					addReq := api.AddCertificateToStore{
						CertificateId:     a.CertID,
						CertificateStores: &stores,
						InventorySchedule: schedule,
					}

					log.Trace().Str("thumbprint", thumbprint).Interface(
						"add_request",
						addReq,
					).Msg("Converting add request to JSON")
					addReqJSON, jErr := json.Marshal(addReq)
					if jErr != nil {
						log.Error().Err(jErr).Str("thumbprint", thumbprint).Msg("Error converting add request to JSON")
						errMsg := fmt.Errorf(
							"error converting add request for '%s' in stores '%v' to JSON: %s",
							thumbprint, stores, jErr,
						)
						errs = append(errs, errMsg)
						continue
					}
					log.Debug().Str("thumbprint", thumbprint).Str(
						"add_request",
						string(addReqJSON),
					).Msg(fmt.Sprintf(DebugFuncCall, "kfClient.AddCertificateToStores"))
					_, err := kfClient.AddCertificateToStores(&addReq)
					if err != nil {
						fmt.Printf(
							"[ERROR] adding cert %s (%d) to store %s (%s): %s\n",
							a.Thumbprint,
							a.CertID,
							a.StoreID,
							a.StorePath,
							err,
						)
						log.Error().Err(err).Str("thumbprint", thumbprint).Str(
							"store_id",
							a.StoreID,
						).Str("store_path", a.StorePath).Msg("unable to add cert to store")
						continue
					}
				} else {
					log.Info().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("DRY RUN: Would have added cert to store")
				}
			} else if a.RemoveCert {
				if !dryRun {
					log.Info().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Attempting to remove cert from store")
					cStore := api.CertificateStore{
						CertificateStoreId: a.StoreID,
						Alias:              a.Thumbprint,
					}
					log.Trace().Interface("store_object", cStore).Msg("Converting store to slice of single store")
					var stores []api.CertificateStore
					stores = append(stores, cStore)

					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating inventory 'immediate' schedule")
					schedule := &api.InventorySchedule{
						Immediate: boolToPointer(true),
					}

					log.Trace().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Creating remove certificate request")
					removeReq := api.RemoveCertificateFromStore{
						CertificateId:     a.CertID,
						CertificateStores: &stores,
						InventorySchedule: schedule,
					}
					log.Debug().Str("thumbprint", thumbprint).Interface(
						"remove_request",
						removeReq,
					).Msg(fmt.Sprintf(DebugFuncCall, "kfClient.RemoveCertificateFromStores"))
					_, err := kfClient.RemoveCertificateFromStores(&removeReq)
					if err != nil {
						log.Error().Err(err).Str("thumbprint", thumbprint).Str(
							"store_id",
							a.StoreID,
						).Str("store_path", a.StorePath).Msg("unable to remove cert from store")
					}
				} else {
					fmt.Printf(
						"DRY RUN: Would have removed cert %s from store %s\n", thumbprint,
						a.StoreID,
					) //todo: propagate back to CLI
					log.Info().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("DRY RUN: Would have removed cert from store")
				}
			}
		}
	}
	log.Info().Str("reconciled_file", rFileName).Msg("Reconciliation actions scheduled on Keyfactor Command")
	if len(errs) > 0 {
		errStr := mergeErrsToString(&errs, false)
		log.Trace().Str("reconciled_file", rFileName).Str(
			"errors",
			errStr,
		).Msg("The following errors occurred while reconciling actions")
		return fmt.Errorf("The following errors occurred while reconciling actions:\r\n%s", errStr)
	}
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "reconcileRoots"))
	return nil
}

func readCertsFile(certsFilePath string, kfclient *api.Client) (map[string]string, error) {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "readCertsFile"))
	// Read in the cert CSV
	log.Info().Str("certs_file", certsFilePath).Msg("Reading in certs file")
	csvFile, ioErr := os.Open(certsFilePath)
	if ioErr != nil {
		log.Error().Err(ioErr).Str("certs_file", certsFilePath).Msg("Error reading in certs file")
		return nil, ioErr
	}

	log.Trace().Str("certs_file", certsFilePath).Msg("Creating CSV reader")
	reader := csv.NewReader(bufio.NewReader(csvFile))

	certEntries, rErr := reader.ReadAll()
	if rErr != nil {
		log.Error().Err(rErr).Str("certs_file", certsFilePath).Msg("Error reading in certs file")
		return nil, rErr
	}

	//validate header
	if len(certEntries) == 0 {
		log.Error().Str("certs_file", certsFilePath).Msg("Empty CSV file")
		return nil, errors.New("empty CSV file")
	}

	log.Debug().Str("certs_file", certsFilePath).Msg("Parsing CSV data")
	var certs = make(map[string]string)
	log.Trace().Str("certs_file", certsFilePath).Msg("Iterating over CSV data")
	headerMap := make(map[string]int)
	for i, entry := range certEntries {
		if i == 0 {
			for j, h := range entry {
				headerMap[h] = j
			}
			continue
		}

		log.Trace().Strs("entry", entry).Msg("Processing row")
		switch entry[0] {
		case "CertID", "thumbprint", "id", "CertId", "Thumbprint": //todo: is there a way to do this with a var?
			log.Trace().Strs("entry", entry).Msg("Skipping header row")
			continue // Skip header
		}
		tp := entry[headerMap["Thumbprint"]]
		if tp == "" {
			log.Warn().Strs("entry", entry).Msg("Thumbprint is empty, skipping")
			continue
		}

		cId := entry[headerMap["CertID"]]
		if cId == "" {
			log.Warn().Strs("entry", entry).Msg("CertID is empty, skipping")
			continue
		}

		log.Trace().Strs("entry", entry).Msg("Adding thumbprint to map")
		certs[tp] = cId
		log.Trace().Interface("certs", certs).Msg("Cert map")
	}
	log.Info().Str("certs_file", certsFilePath).Msg("Certs file read successfully")
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "readCertsFile"))
	log.Trace().Interface("certs", certs).Msg("Returning certs map")
	return certs, nil
}

func isRootStore(
	st *api.GetCertificateStoreResponse,
	invs *[]api.CertStoreInventory,
	minCerts int,
	maxKeys int,
	maxLeaf int,
) bool {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "isRootStore"))
	leafCount := 0
	keyCount := 0
	certCount := 0

	log.Info().
		Str("store_id", st.Id).
		Msg("Checking if store is a root store")

	if invs == nil || len(*invs) == 0 {
		log.Warn().Str("store_id", st.Id).Msg("No certificates found in inventory for store")
		//log.Info().Str("store_id", st.Id).Msg("Empty store is not a root store")
		//return false
		invs = &[]api.CertStoreInventory{}
	}

	log.Debug().Str("store_id", st.Id).Msg("Iterating over inventory")
	for _, inv := range *invs {
		log.Trace().Str("store_id", st.Id).Interface("inv", inv).Msg("Processing inventory")
		certCount += len(inv.Certificates)

		if len(inv.Certificates) == 0 {
			log.Warn().Str("store_id", st.Id).Msg("No certificates found in inventory for store")
			log.Info().Str("store_id", st.Id).Msg("Empty store is not a root store")
			continue
		}

		log.Debug().Str("store_id", st.Id).Msg("Iterating over certificates in inventory")
		for _, cert := range inv.Certificates {
			log.Debug().Str("store_id", st.Id).Str("cert_thumbprint", cert.Thumbprint).Msg("Checking if cert is a leaf")
			if cert.IssuedDN != cert.IssuerDN {
				log.Debug().Str("store_id", st.Id).Str("cert_thumbprint", cert.Thumbprint).Msg("Cert is a leaf")
				leafCount++
			}

			log.Debug().Str("store_id", st.Id).Str(
				"cert_thumbprint",
				cert.Thumbprint,
			).Msg("Checking if cert has a private key")
			//if inv.Parameters["PrivateKeyEntry"] == "Yes" {
			//	log.Debug().Str("store_id", st.Id).Str(
			//		"cert_thumbprint",
			//		cert.Thumbprint,
			//	).Msg("Cert has a private key")
			//	keyCount++
			//}
		}
	}

	log.Info().Str("store_id", st.Id).
		Int("cert_count", certCount).
		Int("min_certs", minCerts).
		Msg("Checking if store meets minimum cert count")
	if certCount < minCerts && minCerts >= 0 {
		log.Info().Str("store_id", st.Id).
			Int("cert_count", certCount).
			Int("min_certs", minCerts).
			Msg("Store does not meet minimum cert count to be considered a root of trust")
		log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))
		return false
	}
	if leafCount > maxLeaf && maxLeaf >= 0 {
		log.Info().Str("store_id", st.Id).
			Int("leaf_count", leafCount).
			Int("max_leaves", maxLeaf).
			Msg("Store has too many leaf certs to be considered a root of trust")
		log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))
		return false
	}

	if keyCount > maxKeys && maxKeys >= 0 {
		log.Info().Str("store_id", st.Id).
			Int("key_count", keyCount).
			Int("max_keys", maxKeys).
			Msg("Store has too many private keys to be considered a root of trust")
		log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))
		return false
	}

	log.Info().Str("store_id", st.Id).
		Int("cert_count", certCount).
		Int("leaf_count", leafCount).
		Int("key_count", keyCount).
		Msg("Store meets criteria to be considered a root of trust")
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))
	return true
}

func findTrustStores(
	minCerts int,
	maxKeys int,
	maxLeaf int,
) []api.GetCertificateStoreResponse {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "findTrustStores"))
	var trustStores []api.GetCertificateStoreResponse
	log.Info().Msg("Finding root of trust stores")
	log.Debug().Msg("Iterating over stores")

	// fetch list of stores from Keyfactor
	stores, err := kfClient.GetCertificateStores()

	for _, st := range *stores {
		log.Trace().Str("store_id", st.Id).Msg("Processing store")
		if isRootStore(&st, invs, minCerts, maxKeys, maxLeaf) {
			log.Debug().Str("store_id", st.Id).Msg("Store is a root of trust")
			trustStores = append(trustStores, st)
		}
	}
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "findTrustStores"))
	return trustStores
}

func validateStoresInput(storesFile *string, noPrompt *bool, kfClient *api.Client) (string, error) {
	if noPrompt == nil {
		noPrompt = boolToPointer(false)
	}

	if storesFile == nil || *storesFile == "" {
		if *noPrompt {
			return "", fmt.Errorf("stores file is required, use flag `--stores` to specify 1 or more file paths")
		}
		apiOrFile := promptSelectFromAPIorFile("certificate stores")
		switch apiOrFile {
		case "API":
			selectedStores := promptSelectStores(kfClient)
			if len(selectedStores) == 0 {
				return "", errors.New("no certificate stores selected, unable to continue")
			}
			//create stores file
			storesFile = stringToPointer(fmt.Sprintf("%s", DefaultROTAuditStoresOutfilePath))
			// create file
			f, ioErr := os.Create(*storesFile)
			if ioErr != nil {
				log.Error().Err(ioErr).Str("stores_file", *storesFile).Msg("Error creating stores file")
				return "", ioErr
			}
			defer f.Close()
			// create CSV writer
			log.Debug().Str("stores_file", *storesFile).Msg("Creating CSV writer")
			writer := csv.NewWriter(f)
			defer writer.Flush()
			// write header
			log.Debug().Str("stores_file", *storesFile).Msg("Writing header to stores file")
			wErr := writer.Write(StoreHeader)
			if wErr != nil {
				log.Error().Err(wErr).Str("stores_file", *storesFile).Msg("Error writing header to stores file")
				return "", wErr
			}
			// write selected stores
			for _, store := range selectedStores {
				log.Debug().Str("store_id", store).Msg("Adding store to stores file")
				//parse ID from selection `<id>: <name>`
				storeId := strings.Split(store, ":")[1]
				//remove () and white spaces from storeId
				storeId = strings.Trim(strings.Trim(strings.Trim(storeId, " "), "("), ")")

				storeInstance := ROTStore{
					StoreID:       storeId,
					StoreType:     "",
					StoreMachine:  "",
					StorePath:     "",
					ContainerId:   "",
					ContainerName: "",
					LastQueried:   "",
				}
				storeLine := storeInstance.toCSV()

				wErr = writer.Write(strings.Split(storeLine, ","))
				if wErr != nil {
					log.Error().Err(wErr).Str(
						"stores_file",
						*storesFile,
					).Msg("Error writing store to stores file")
					continue
				}
			}
			writer.Flush()
			f.Close()
			return *storesFile, nil
		case "File":
			return promptForFilePath("Input a file path for the CSV file containing stores to audit."), nil
		}
	}
	return *storesFile, nil
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			// Specific Flags
			storesFile, _ := cmd.Flags().GetString("stores")
			addRootsFile, _ := cmd.Flags().GetString("add-certs")
			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
			minCerts, _ := cmd.Flags().GetInt("min-certs")
			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
			maxKeys, _ := cmd.Flags().GetInt("max-keys")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			outputFilePath, _ := cmd.Flags().GetString("outputFilePath")

			// Debug + expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			informDebug(debugFlag)

			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)

			var lookupFailures []string
			kfClient, cErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
			if cErr != nil {
				log.Error().Err(cErr).Msg("Error initializing Keyfactor client")
				return cErr
			}

			// validate flags
			var storesErr error
			log.Debug().Str("stores_file", storesFile).Bool("no_prompt", noPrompt).
				Msg(fmt.Sprintf(DebugFuncCall, "validateStoresInput"))
			storesFile, storesErr = validateStoresInput(&storesFile, &noPrompt, kfClient)
			if storesErr != nil {
				return storesErr
			}

			log.Debug().Str("add_file", addRootsFile).Str("remove_file", removeRootsFile).Bool("no_prompt", noPrompt).
				Msg(fmt.Sprintf(DebugFuncCall, "validateCertsInput"))
			var certsErr error
			addRootsFile, removeRootsFile, certsErr = validateCertsInput(
				addRootsFile, removeRootsFile,
				kfClient,
			)
			if certsErr != nil {
				log.Error().Err(cErr).Msg("Invalid certs input please provide certs to add or remove.")
				return cErr
			}

			log.Info().Str("stores_file", storesFile).
				Str("add_file", addRootsFile).
				Str("remove_file", removeRootsFile).
				Bool("dry_run", dryRun).
				Msg("Performing root of trust audit")

			// Read in the stores CSV
			log.Debug().Str("stores_file", storesFile).Msg("Reading in stores file")
			csvFile, ioErr := os.Open(storesFile)
			if ioErr != nil {
				log.Error().Err(ioErr).Str("stores_file", storesFile).Msg("Error reading in stores file")
				return ioErr
			}

			log.Trace().Str("stores_file", storesFile).Msg("Creating CSV reader")
			reader := csv.NewReader(bufio.NewReader(csvFile))

			log.Debug().Str("stores_file", storesFile).Msg("Reading CSV data")
			storeEntries, rErr := reader.ReadAll()
			if rErr != nil {
				log.Error().Err(rErr).Str("stores_file", storesFile).Msg("Error reading in stores file")
				return rErr
			}

			log.Debug().Str("stores_file", storesFile).Msg("Validating CSV header")
			var stores = make(map[string]StoreCSVEntry)
			validHeader := false
			for _, entry := range storeEntries {
				log.Trace().Strs("entry", entry).Msg("Processing row")
				if strings.EqualFold(strings.Join(entry, ","), strings.Join(StoreHeader, ",")) {
					validHeader = true
					continue // Skip header
				}
				if !validHeader {
					log.Error().
						Strs("header", entry).
						Strs("expected_header", StoreHeader).
						Msg("Invalid header in stores file")
					return fmt.Errorf("invalid header in stores file please use '%s'", strings.Join(StoreHeader, ","))
				}

				log.Debug().Strs("entry", entry).
					Str("store_id", entry[0]).
					Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificateStoreByID"))
				apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
				if err != nil {
					log.Error().Err(err).Str("store_id", entry[0]).Msg("Error getting cert store")
					lookupFailures = append(lookupFailures, strings.Join(entry, ","))
					continue
				}

				log.Debug().Str("store_id", entry[0]).
					Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertStoreInventoryV1"))
				inventory, invErr := kfClient.GetCertStoreInventory(entry[0])
				if invErr != nil {
					log.Error().Err(invErr).Str("store_id", entry[0]).Msg("Error getting cert store inventory")
					lookupFailures = append(lookupFailures, strings.Join(entry, ","))
					continue
				} else if inventory == nil {
					log.Error().Str(
						"store_id",
						entry[0],
					).Msg("No inventory response returned for store from Keyfactor Command")
					lookupFailures = append(lookupFailures, strings.Join(entry, ","))
					continue
				}

				if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
					fmt.Printf(
						"Store %s is not a root store, skipping.\n",
						entry[0],
					) //todo: support for output formatting
					log.Warn().Str("store_id", entry[0]).Msg("Store is not considered a root of trust store")
					continue
				}

				log.Info().Str("store_id", entry[0]).Msg("Store is considered a root of trust store")

				log.Trace().Str("store_id", entry[0]).Msg("Creating store entry")

				stores[entry[0]] = StoreCSVEntry{
					ID:          entry[0],
					Type:        entry[1],
					Machine:     entry[2],
					Path:        entry[3],
					Thumbprints: make(map[string]bool),
					Serials:     make(map[string]bool),
					Ids:         make(map[int]bool),
				}

				log.Debug().Str("store_id", entry[0]).Msg("Iterating over inventory")
				for _, cert := range *inventory {
					log.Trace().Str("store_id", entry[0]).Interface("cert", cert).Msg("Processing inventory")
					thumb := cert.Thumbprints
					trcMsg := "Adding cert to store"
					for t, v := range thumb {
						//log.Trace().Str("store_id", entry[0]).Str("thumbprint", t).Msg(trcMsg)
						//stores[entry[0]].Thumbprints[t] = v
						log.Trace().Str("store_id", entry[0]).
							Int("thumbprint", t).
							Str("value", v).
							Msg(trcMsg)
						stores[entry[0]].Thumbprints[v] = true
					}
					for t, v := range cert.Serials {
						log.Trace().Str("store_id", entry[0]).
							Int("serial", t).
							Str("value", v).
							Msg(trcMsg)
						//stores[entry[0]].Serials[t] = v
						//stores[entry[0]].Serials[v] = t
						stores[entry[0]].Serials[v] = true
					}
					for t, v := range cert.Ids {
						log.Trace().Str("store_id", entry[0]).
							Int("cert_id", t).
							Int("value", v).
							Msg(trcMsg)
						//stores[entry[0]].Ids[t] = v
						stores[entry[0]].Ids[v] = true
					}
				}
				log.Trace().Strs("entry", entry).Msg("Row processed")
			}

			if len(lookupFailures) > 0 {
				log.Error().Strs("lookup_failures", lookupFailures).Msg("The following stores could not be looked up")
				return fmt.Errorf(
					"the following stores could not be found on Keyfactor Command:\n%s\nThese errors MUST be resolved"+
						" in order to proceed", strings.Join(
						lookupFailures, ","+
							"\r\n",
					),
				)
			}

			// Read in the add addCerts CSV
			var certsToAdd = make(map[string]string)

			if addRootsFile == "" {
				log.Debug().Msg("No addCerts file specified")
			} else {
				log.Info().Str("add_certs_file", addRootsFile).Msg("Reading certs to add file")
				var rcfErr error
				log.Debug().Str("add_certs_file", addRootsFile).Msg(fmt.Sprintf(DebugFuncCall, "readCertsFile"))
				certsToAdd, rcfErr = readCertsFile(addRootsFile, kfClient)
				if rcfErr != nil {
					log.Error().Err(rcfErr).Str("add_certs_file", addRootsFile).Msg("Error reading certs to add file")
					return rcfErr
				}

				log.Debug().Str("add_certs_file", addRootsFile).Msg("Creating JSON of certs to add")
				addCertsJSON, jErr := json.Marshal(certsToAdd)
				if jErr != nil {
					log.Error().Err(jErr).Str(
						"add_certs_file",
						addRootsFile,
					).Msg("Error converting certs to add to JSON")
					return jErr
				}
				log.Trace().Str("add_certs_file", addRootsFile).
					Str("add_certs_json", string(addCertsJSON)).
					Msg("Certs to add file read successfully")
			}

			// Read in the remove removeCerts CSV
			var certsToRemove = make(map[string]string)
			if removeRootsFile == "" {
				log.Info().Msg("No removeCerts file specified")
			} else {
				log.Info().Str("remove_certs_file", removeRootsFile).Msg("Reading certs to remove file")
				var rcfErr error
				log.Debug().Str("remove_certs_file", removeRootsFile).Msg(fmt.Sprintf(DebugFuncCall, "readCertsFile"))
				certsToRemove, rcfErr = readCertsFile(removeRootsFile, kfClient)
				if rcfErr != nil {
					log.Error().Err(rcfErr).Str(
						"remove_certs_file",
						removeRootsFile,
					).Msg("Error reading certs to remove file")
				}

				removeCertsJSON, jErr := json.Marshal(certsToRemove)
				if jErr != nil {
					log.Error().Err(jErr).Str(
						"remove_certs_file",
						removeRootsFile,
					).Msg("Error converting certs to remove to JSON")
					return jErr
				}
				log.Trace().Str("remove_certs_file", removeRootsFile).
					Str("remove_certs_json", string(removeCertsJSON)).
					Msg("Certs to remove file read successfully")
			}

			log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "generateAuditReport"))
			_, _, gErr := generateAuditReport(certsToAdd, certsToRemove, stores, outputFilePath, kfClient)
			if gErr != nil {
				log.Error().Err(gErr).Msg("Error generating audit report")
				return gErr
			}

			log.Info().
				Str("outputFilePath", outputFilePath).
				Msg("Audit report generated successfully")
			log.Debug().
				Msg(fmt.Sprintf(DebugFuncExit, "generateAuditReport"))
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
			cmd.SilenceUsage = true

			// Specific Flags
			storesFile, _ := cmd.Flags().GetString("stores")
			addRootsFile, _ := cmd.Flags().GetString("add-certs")
			isCSV, _ := cmd.Flags().GetBool("import-csv")
			reportFile, _ := cmd.Flags().GetString("input-file")
			removeRootsFile, _ := cmd.Flags().GetString("remove-certs")
			minCerts, _ := cmd.Flags().GetInt("min-certs")
			maxLeaves, _ := cmd.Flags().GetInt("max-leaf-certs")
			maxKeys, _ := cmd.Flags().GetInt("max-keys")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			outputFilePath, _ := cmd.Flags().GetString("outputFilePath")

			// Debug + expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			informDebug(debugFlag)

			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)

			kfClient, clErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
			if clErr != nil {
				log.Error().Err(clErr).Msg("Error initializing Keyfactor client")
				return clErr
			}

			log.Info().Str("stores_file", storesFile).
				Str("add_file", addRootsFile).
				Str("remove_file", removeRootsFile).
				Bool("dry_run", dryRun).
				Msg("Performing root of trust reconciliation")

			// Parse existing audit report
			if isCSV && reportFile != "" {
				err := processCSVReportFile(reportFile, kfClient, dryRun)
				if err != nil {
					log.Error().Err(err).Msg("Error processing audit report")
					return err
				}
				return nil
			} else {
				log.Debug().
					Str("stores_file", storesFile).
					Str("add_file", addRootsFile).
					Str("remove_file", removeRootsFile).
					Str("report_file", reportFile).
					Bool("dry_run", dryRun).
					Msg(fmt.Sprintf(DebugFuncCall, "processFromStoresAndCertFiles"))
				err := processFromStoresAndCertFiles(
					storesFile,
					addRootsFile,
					removeRootsFile,
					reportFile,
					outputFilePath,
					minCerts,
					maxLeaves,
					maxKeys,
					kfClient,
					dryRun,
				)
				if err != nil {
					log.Error().Err(err).Msg("Error processing from stores file")
					return err
				}
			}

			log.Debug().Str("report_file", reportFile).
				Str("outputFilePath", outputFilePath).Msg("Reconciliation report generated successfully")
			log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "reconcileRoots"))
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
			cmd.SilenceUsage = true

			// Specific Flags
			templateType, _ := cmd.Flags().GetString("type")
			format, _ := cmd.Flags().GetString("format")
			outputFilePath, _ := cmd.Flags().GetString("outputFilePath")
			storeType, _ := cmd.Flags().GetStringSlice("store-type")
			containerName, _ := cmd.Flags().GetStringSlice("container-name")
			collection, _ := cmd.Flags().GetStringSlice("collection")
			subjectName, _ := cmd.Flags().GetStringSlice("cn")

			// Debug + expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			informDebug(debugFlag)

			// Authenticate
			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
			kfClient, clErr := initClient(
				configFile, profile, providerType, providerProfile, noPrompt, authConfig,
				false,
			)
			if clErr != nil {
				log.Error().Err(clErr).Msg("Error initializing Keyfactor client")
				return clErr
			}

			stID := -1
			var storeData []api.GetCertificateStoreResponse
			var csvStoreData [][]string
			var csvCertData [][]string
			var rowLookup = make(map[string]bool)
			var errs []error

			if len(storeType) != 0 {
				log.Info().Strs("store_types", storeType).Msg("Processing store types")
				for _, s := range storeType {
					log.Debug().Str("store_type", s).Msg("Processing store type")
					var sType *api.CertificateStoreType
					var stErr error
					if s == "all" {
						log.Info().
							Str("store_type", s).
							Msg("Getting all store types")

						log.Trace().Msg("Creating empty store type for 'all' option")
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
						sInt, err := strconv.Atoi(s)

						if err == nil {
							log.Debug().Str("store_type", s).Msg("Getting store type by ID")
							sType, stErr = kfClient.GetCertificateStoreTypeById(sInt)
						} else {
							log.Debug().Str("store_type", s).Msg("Getting store type by name")
							sType, stErr = kfClient.GetCertificateStoreTypeByName(s)
						}
						if stErr != nil {
							//fmt.Printf("unable to get store type '%s' from Keyfactor Command: %s\n", s, stErr)
							errs = append(errs, stErr)
							continue
						}
						stID = sType.StoreType // This is the template type ID
					}

					if stID >= 0 || s == "all" {
						log.Debug().Str("store_type", s).
							Int("store_type_id", stID).
							Msg("Getting certificate stores")
						params := make(map[string]interface{})
						if stID >= 0 {
							params["StoreType"] = stID
						}

						log.Debug().Str("store_type", s).Msg("Getting certificate stores")
						stores, sErr := kfClient.ListCertificateStores(&params)
						if sErr != nil {
							log.Error().Err(sErr).
								Str("store_type", s).
								Int("store_type_id", stID).
								Interface("params", params).
								Msg("Error getting certificate stores")
							return sErr
						}
						if stores == nil {
							log.Warn().Str("store_type", s).Msg("No stores found")
							errs = append(errs, fmt.Errorf("no stores found for store type: %s", s))
							continue
						}
						for _, store := range *stores {
							log.Trace().Str("store_type", s).Msg("Processing stores of type")
							if store.CertStoreType == stID || s == "all" {
								storeData = append(storeData, store)
								if !rowLookup[store.Id] {
									log.Trace().Str("store_type", s).
										Str("store_id", store.Id).
										Msg("Constructing CSV row")
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
									log.Trace().Strs("line_data", lineData).Msg("Adding line data to CSV data")
									csvStoreData = append(csvStoreData, lineData)
									rowLookup[store.Id] = true
								}
							}
						}
					} else {
						errMsg := fmt.Errorf("Invalid input, must provide a store type of specify 'all'")
						log.Error().Err(errMsg).Msg("Invalid input")
						if len(errs) == 0 {
							errs = append(errs, errMsg)
						}
					}
				}
				log.Info().Strs("store_types", storeType).Msg("Store types processed")
			}

			if len(containerName) != 0 {
				log.Info().Strs("container_names", containerName).Msg("Processing container names")
				for _, c := range containerName {
					cStoresResp, scErr := kfClient.GetCertificateStoreByContainerID(c)
					if scErr != nil {
						fmt.Printf("[ERROR] getting store container: %s\n", scErr)
					}
					if cStoresResp != nil {
						for _, store := range *cStoresResp {
							sType, stErr := kfClient.GetCertificateStoreType(store.CertStoreType)
							if stErr != nil {
								fmt.Printf("[ERROR] getting store type: %s\n", stErr)
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
								csvStoreData = append(csvStoreData, lineData)
								rowLookup[store.Id] = true
							}
						}

					}
				}
				log.Info().Strs("container_names", containerName).Msg("Container names processed")
			}
			if len(collection) != 0 {
				log.Info().Strs("collections", collection).Msg("Processing collections")
				for _, c := range collection {
					q := make(map[string]string)
					q["collection"] = c
					certsResp, scErr := kfClient.ListCertificates(q)
					if scErr != nil {
						fmt.Printf("No certificates found in collection: %s\n", scErr)
					}
					if certsResp != nil {
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
								csvCertData = append(csvCertData, lineData)
								rowLookup[cert.Thumbprint] = true
							}
						}

					}
				}
				log.Info().Strs("collections", collection).Msg("Collections processed")
			}
			if len(subjectName) != 0 {
				log.Info().Strs("subject_names", subjectName).Msg("Processing subject names")
				for _, s := range subjectName {
					q := make(map[string]string)
					q["subject"] = s
					log.Debug().Str("subject_name", s).Msg("Getting certificates by subject name")
					certsResp, scErr := kfClient.ListCertificates(q)
					if scErr != nil {
						log.Error().Err(scErr).Str("subject_name", s).Msg("Error listing certificates by subject name")
						errs = append(errs, scErr)
					}

					if certsResp != nil {
						log.Debug().Str(
							"subject_name",
							s,
						).Msg("processing certificates returned from Keyfactor Command")
						for _, cert := range certsResp {
							log.Trace().Interface("cert", cert).Msg("Processing certificate")
							if !rowLookup[cert.Thumbprint] {
								log.Trace().
									Str("thumbprint", cert.Thumbprint).
									Str("subject_name", cert.IssuedCN).
									Str("not_before", cert.NotBefore).
									Str("not_after", cert.NotAfter).
									Msg("Adding certificate to CSV data")
								locationsFormatted := ""

								log.Debug().Str(
									"thumbprint",
									cert.Thumbprint,
								).Msg("Iterating over certificate locations")
								for _, loc := range cert.Locations {
									log.Trace().Str("thumbprint", cert.Thumbprint).Str(
										"location",
										loc.StoreMachine,
									).Msg("Processing location")
									locationsFormatted += fmt.Sprintf("%s:%s\n", loc.StoreMachine, loc.StorePath)
								}
								log.Trace().Str("thumbprint", cert.Thumbprint).Str(
									"locations",
									locationsFormatted,
								).Msg("Constructing CSV line data")
								lineData := []string{
									// "Thumbprint", "SubjectName", "Issuer", "CertID", "Locations", "LastQueriedDate"
									cert.Thumbprint,
									cert.IssuedCN,
									cert.IssuerDN,
									fmt.Sprintf("%d", cert.Id),
									locationsFormatted,
									getCurrentTime(""),
								}
								log.Trace().Strs("line_data", lineData).Msg("Adding line data to CSV data")
								csvCertData = append(csvCertData, lineData)
								rowLookup[cert.Thumbprint] = true
							}
						}

					}
				}
			}
			// Create CSV template file

			var filePath string
			if outputFilePath != "" {
				filePath = outputFilePath
			} else {
				filePath = fmt.Sprintf("%s_template.%s", templateType, format)
			}
			log.Info().Str("file_path", filePath).Msg("Creating template file")
			file, err := os.Create(filePath)
			if err != nil {
				log.Error().Err(err).Str("file_path", filePath).Msg("Error creating template file")
				return err
			}

			switch format {
			case "csv":
				log.Info().Str("file_path", filePath).Msg("Creating CSV writer")
				writer := csv.NewWriter(file)
				var data [][]string
				log.Debug().Str("template_type", templateType).Msg("Processing template type")
				switch templateType {
				case "stores":
					data = append(data, StoreHeader)
					if len(csvStoreData) != 0 {
						data = append(data, csvStoreData...)
					}
					log.Debug().Str("template_type", templateType).
						Interface("csv_data", csvStoreData).
						Msg("Writing CSV data to file")
				case "certs":
					data = append(data, CertHeader)
					if len(csvCertData) != 0 {
						data = append(data, csvCertData...)
					}
					log.Debug().Str("template_type", templateType).
						Interface("csv_data", csvCertData).
						Msg("Writing CSV data to file")
				case "actions":
					data = append(data, AuditHeader)
					log.Debug().Str("template_type", templateType).
						Interface("csv_data", csvCertData).
						Msg("Writing CSV data to file")
				}
				csvErr := writer.WriteAll(data)
				if csvErr != nil {
					log.Error().Err(csvErr).Str("file_path", filePath).Msg("Error writing CSV data to file")
					errs = append(errs, csvErr)
				}
				defer file.Close()

			case "json":
				log.Info().Str("file_path", filePath).Msg("Creating JSON file")
				log.Trace().Str("file_path", filePath).Msg("Creating JSON encoder")
				writer := bufio.NewWriter(file)
				_, err := writer.WriteString("StoreID,StoreType,StoreMachine,StorePath")
				if err != nil {
					log.Error().Err(err).Str("file_path", filePath).Msg("Error writing JSON data to file")
					errs = append(errs, err)
				}
			}
			if len(errs) != 0 {
				log.Error().Errs("errors", errs).Msg("Errors encountered while creating template file")
				errMsg := mergeErrsToString(&errs, false)
				return fmt.Errorf("errors encountered while creating template file: %s", errMsg)
			}
			fmt.Printf("Template file created at %s.\n", filePath)
			log.Info().Str("file_path", filePath).Msg("Template file created")
			log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "generateTemplate"))
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

func validateCertsInput(addRootsFile string, removeRootsFile string, client *api.Client) (
	string,
	string,
	error,
) {
	log.Debug().Str("add_certs_file", addRootsFile).
		Str("remove_certs_file", removeRootsFile).
		Bool("no_prompt", noPrompt).
		Msg(fmt.Sprintf(DebugFuncEnter, "validateCertsInput"))

	if addRootsFile == "" && removeRootsFile == "" && noPrompt {
		//cmd.SilenceUsage = false //todo: is this necessary?
		return addRootsFile, removeRootsFile, InvalidROTCertsInputErr
	}

	if addRootsFile == "" || removeRootsFile == "" {
		if addRootsFile == "" && !noPrompt {
			//prmpt := "Would you like to include a 'certs to add' CSV file?"
			prmpt := "Provide certificates to add and/or that should be present in selected stores?"
			provideAddFile := promptYesNo(prmpt)
			if provideAddFile {
				addSrcType := promptSelectFromAPIorFile("certificates")
				switch addSrcType {
				case "API":
					selectedCerts := promptSelectCerts(client)
					if len(selectedCerts) == 0 {
						return "", "", InvalidROTCertsInputErr
					}
					//create stores file
					addRootsFile = fmt.Sprintf("%s", DefaultROTAuditAddCertsOutfilePath)
					// create file
					f, ioErr := os.Create(addRootsFile)
					if ioErr != nil {
						log.Error().Err(ioErr).Str(
							"add_certs_file",
							addRootsFile,
						).Msg("Error creating certs to add file")
						return addRootsFile, removeRootsFile, ioErr
					}
					defer f.Close()
					// create CSV writer
					log.Debug().Str("add_certs_file", addRootsFile).Msg("Creating CSV writer")
					writer := csv.NewWriter(f)
					defer writer.Flush()
					// write header
					log.Debug().Str("add_certs_file", addRootsFile).Msg("Writing header to certs to add file")
					wErr := writer.Write(CertHeader)
					if wErr != nil {
						log.Error().Err(wErr).Str(
							"stores_file",
							addRootsFile,
						).Msg("Error writing header to stores file")
						return addRootsFile, removeRootsFile, wErr
					}
					// write selected stores
					for _, c := range selectedCerts {
						log.Debug().Str("cert_id", c).Msg("Adding cert to certs file")

						//parse certID, cn and thumbprint from selection `<id>: <cn> (<thumbprint>) - <issued_date>`

						//parse id from selection `<id>: <cn> (<thumbprint>) <issued_date>`
						certId := strings.Split(c, ":")[0]
						//remove () and white spaces from storeId
						certId = strings.Trim(certId, " ")
						certIdInt, cIdErr := strconv.Atoi(certId)
						if cIdErr != nil {
							log.Error().Err(cIdErr).Str("cert_id", certId).Msg("Error converting cert ID to int")
							certIdInt = -1
						}

						//parse the cn from the selection `<id>: <cn> (<thumbprint>) <issued_date>`
						cn := strings.Split(c, "(")[0]
						cn = strings.Split(cn, ":")[1]
						cn = strings.Trim(cn, " ")

						//parse thumbprint from selection `<id>: <cn> (<thumbprint>) <issued_date>`
						thumbprint := strings.Split(c, "(")[1]
						thumbprint = strings.Split(thumbprint, ")")[0]
						thumbprint = strings.Trim(strings.Trim(thumbprint, " "), ")")

						certInstance := ROTCert{
							ID:         certIdInt,
							ThumbPrint: thumbprint,
							CN:         cn,
							SANs:       []string{},
							Alias:      "",
							Locations:  []api.CertificateLocations{},
						}
						certLine := certInstance.toCSV()

						wErr = writer.Write(strings.Split(certLine, ","))
						if wErr != nil {
							log.Error().Err(wErr).Str(
								"add_certs_file",
								addRootsFile,
							).Msg("Error writing store to stores file")
							continue
						}
					}
					writer.Flush()
					f.Close()
				default:
					addRootsFile = promptForFilePath("Input a file path for the 'certs to add' CSV.")
				}
			}
		}
		if removeRootsFile == "" && !noPrompt {
			provideRemoveFile := promptYesNo("Would you like to include a 'certs to remove' CSV file?")
			if provideRemoveFile {
				removeRootsFile = promptForFilePath("Input a file path for the 'certs to remove' CSV. ")
			}
		}
		if addRootsFile == "" && removeRootsFile == "" {
			return "", "", InvalidROTCertsInputErr
		}
	}
	return addRootsFile, removeRootsFile, nil

}

func processFromStoresAndCertFiles(
	storesFile string,
	addRootsFile string,
	removeRootsFile string,
	reportFile string,
	outputFilePath string,
	minCerts int,
	maxLeaves int,
	maxKeys int,
	kfClient *api.Client,
	dryRun bool,
) error {
	// Read in the stores CSV
	log.Debug().Str("stores_file", storesFile).Msg("Reading in stores file")
	csvFile, _ := os.Open(storesFile)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	storeEntries, _ := reader.ReadAll()
	var stores = make(map[string]StoreCSVEntry)
	var lookupFailures []string
	var errs []error
	for i, row := range storeEntries {
		if len(row) == 0 {
			log.Warn().
				Str("stores_file", storesFile).
				Int("row", i).Msg("Skipping empty row")
			continue
		} else if row[0] == "StoreID" || row[0] == "StoreId" || i == 0 {
			log.Trace().Strs("row", row).Msg("Skipping header row")
			continue // Skip header
		}

		log.Debug().Strs("row", row).
			Str("store_id", row[0]).
			Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificateStoreByID"))
		apiResp, err := kfClient.GetCertificateStoreByID(row[0])
		if err != nil {
			errs = append(errs, err)
			log.Error().Err(err).Str("store_id", row[0]).Msg("failed to retrieve store from Keyfactor Command")
			lookupFailures = append(lookupFailures, row[0])
			continue
		}

		log.Debug().Str("store_id", row[0]).Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertStoreInventoryV1"))
		inventory, invErr := kfClient.GetCertStoreInventory(row[0])
		if invErr != nil {
			errs = append(errs, invErr)
			log.Error().Err(invErr).Str(
				"store_id",
				row[0],
			).Msg("failed to retrieve inventory for certificate store from Keyfactor Command")
			continue
		}

		if !isRootStore(apiResp, inventory, minCerts, maxLeaves, maxKeys) {
			log.Error().Str(
				"store_id",
				row[0],
			).Msg("Store is not considered a root of trust store and will be excluded.")
			errs = append(errs, fmt.Errorf("store '%s' is not considered a root of trust store", row[0]))
			continue
		}

		log.Info().Str("store_id", row[0]).Msg("Store is considered a root of trust store")
		log.Trace().Str("store_id", row[0]).Msg("Creating StoreCSVEntry object")
		stores[row[0]] = StoreCSVEntry{
			ID:          row[0],
			Type:        row[1],
			Machine:     row[2],
			Path:        row[3],
			Thumbprints: make(map[string]bool),
			Serials:     make(map[string]bool),
			Ids:         make(map[int]bool),
		}

		log.Debug().Str("store_id", row[0]).Msg(
			"Iterating over inventory for thumbprints, " +
				"serial numbers and cert IDs",
		)
		for _, cert := range *inventory {
			log.Trace().Str("store_id", row[0]).Interface("cert", cert).Msg("Processing inventory")
			thumb := cert.Thumbprints
			for t, v := range thumb {
				log.Trace().Str("store_id", row[0]).
					Str("value", v).
					Int("thumbprint", t).Msg("Adding cert thumbprint to store object")
				//stores[row[0]].Thumbprints[t] = v
			}
			for t, v := range cert.Serials {
				log.Trace().Str("store_id", row[0]).
					Str("value", v).
					Int("serial", t).Msg("Adding cert serial to store object")
				//stores[row[0]].Serials[t] = v
			}
			for t, v := range cert.Ids {
				log.Trace().Str("store_id", row[0]).
					Int("value", v).
					Int("cert_id", t).Msg("Adding cert ID to store object")
				//stores[row[0]].Ids[t] = v
			}
		}
	}
	if len(lookupFailures) > 0 {
		errMsg := fmt.Errorf("The following stores were not found:\r\n%s", strings.Join(lookupFailures, ",\r\n"))
		fmt.Printf(errMsg.Error())
		log.Error().Err(errMsg).
			Strs("lookup_failures", lookupFailures).
			Msg("The following stores could not be found")
		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			errMsg = fmt.Errorf("%s\r\n%s", errMsg, apiErrs)
		}
		return errMsg
	}
	if len(stores) == 0 {
		errMsg := fmt.Errorf("no root of trust stores found that meet the defined criteria")
		log.Error().
			Err(errMsg).
			Int("min_certs", minCerts).
			Int("max_leaves", maxLeaves).
			Int("max_keys", maxKeys).Send()

		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			errMsg = fmt.Errorf("%s\r\n%s", errMsg, apiErrs)
		}
		return errMsg
	}
	// Read in the add addCerts CSV
	var certsToAdd = make(map[string]string)
	var rErr error
	if addRootsFile == "" {
		log.Info().Msg("No add certs file specified, add operations will not be performed")
	} else {
		log.Info().Str("add_certs_file", addRootsFile).Msg("Reading certs to add file")
		log.Debug().Str("add_certs_file", addRootsFile).Msg(fmt.Sprintf(DebugFuncCall, "readCertsFile"))
		certsToAdd, rErr = readCertsFile(addRootsFile, kfClient)
		if rErr != nil {
			log.Error().Err(rErr).Str("add_certs_file", addRootsFile).Msg("Error reading certs to add file")
			if len(errs) > 0 {
				apiErrs := mergeErrsToString(&errs, false)
				rErr = fmt.Errorf("%s\r\n%s", rErr, apiErrs)
			}
			return rErr
		}
		log.Debug().Str("add_certs_file", addRootsFile).Msg("finished reading certs to add file")
	}

	// Read in the remove removeCerts CSV
	var certsToRemove = make(map[string]string)
	if removeRootsFile == "" {
		log.Info().Msg("No remove certs file specified, remove operations will not be performed")
	} else {
		log.Info().Str("remove_certs_file", removeRootsFile).Msg("Reading certs to remove file")
		log.Debug().Str("remove_certs_file", removeRootsFile).Msg(fmt.Sprintf(DebugFuncCall, "readCertsFile"))
		certsToRemove, rErr = readCertsFile(removeRootsFile, kfClient)
		if rErr != nil {
			log.Error().Err(rErr).Str("remove_certs_file", removeRootsFile).Msg("Error reading certs to remove file")
			if len(errs) > 0 {
				apiErrs := mergeErrsToString(&errs, false)
				rErr = fmt.Errorf("%s\r\n%s", rErr, apiErrs)
			}
			return rErr
		}
	}

	if len(certsToAdd) == 0 && len(certsToRemove) == 0 {
		log.Info().Msg("No add or remove operations specified, please verify your configuration")
		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			return fmt.Errorf(apiErrs)
		}
		fmt.Println("No add or remove operations specified, please verify your configuration")
		return nil
	}

	log.Trace().Interface("certs_to_add", certsToAdd).
		Interface("certs_to_remove", certsToRemove).
		Str("stores_file", storesFile).
		Msg("Generating audit report")

	log.Debug().
		Msg(fmt.Sprintf(DebugFuncCall, "generateAuditReport"))
	_, actions, err := generateAuditReport(certsToAdd, certsToRemove, stores, outputFilePath, kfClient)
	if err != nil {
		log.Error().
			Err(err).
			Str("outputFilePath", outputFilePath).
			Msg("Error generating audit report")
	}
	if len(actions) == 0 {
		msg := "No reconciliation actions to take, the specified root of trust stores are up-to-date"
		log.Info().
			Str("stores_file", storesFile).
			Str("add_certs_file", addRootsFile).
			Str("remove_certs_file", removeRootsFile).
			Msg(msg)
		fmt.Println("No reconciliation actions to take, root stores are up-to-date. Exiting.")
		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			return fmt.Errorf(apiErrs)
		}
		return nil
	}

	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "reconcileRoots"))
	rErr = reconcileRoots(actions, kfClient, reportFile, dryRun)
	if rErr != nil {
		log.Error().Err(rErr).Msg("Error reconciling root of trust stores")
		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			rErr = fmt.Errorf("%s\r\n%s", rErr, apiErrs)
		}
		return rErr
	}
	if lookupFailures != nil {
		errMsg := fmt.Errorf(
			"The following stores could not be found:\r\n%s", strings.Join(lookupFailures, ",\r\n"),
		)
		log.Error().Err(errMsg).Strs("lookup_failures", lookupFailures).Send()
		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			errMsg = fmt.Errorf("%s\r\n%s", errMsg, apiErrs)
			return errMsg
		}
		return errMsg
	}
	orchsURL := fmt.Sprintf(
		"https://%s/Keyfactor/Portal/AgentJobStatus/Index",
		kfClient.Hostname,
	) //todo: this path might not work for everyone

	log.Info().
		Str("orchs_url", orchsURL).
		Str("outputFilePath", outputFilePath).
		Msg("Reconciliation completed. Check orchestrator jobs for details.")
	fmt.Println(fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL))

	if len(lookupFailures) > 0 {
		lookupErrs := fmt.Errorf(
			"Reconciliation completed with failures, "+
				"the following stores could not be found:\r\n%s", strings.Join(
				lookupFailures,
				"\r\n",
			),
		)
		log.Error().Err(lookupErrs).Strs(
			"lookup_failures",
			lookupFailures,
		).Msg("The following stores could not be found")
		if len(errs) > 0 {
			apiErrs := mergeErrsToString(&errs, false)
			lookupErrs = fmt.Errorf("%s\r\n%s", lookupErrs, apiErrs)
		}
		return lookupErrs
	} else if len(errs) > 0 {
		apiErrs := mergeErrsToString(&errs, false)
		log.Error().Str("api_errors", apiErrs).Msg("Reconciliation completed with failures")
		return fmt.Errorf("Reconciliation completed with failures:\r\n%s", apiErrs)
	}
	return nil
}

func processCSVReportFile(reportFile string, kfClient *api.Client, dryRun bool) error {
	log.Debug().Str("report_file", reportFile).Bool("dry_run", dryRun).
		Msg("Parsing existing audit report")
	// Read in the CSV

	log.Debug().Str("report_file", reportFile).Msg("reading audit report file")
	csvFile, err := os.Open(reportFile)
	if err != nil {
		log.Error().Err(err).Str("report_file", reportFile).Msg("Error reading audit report file")
		return err
	}

	validHeader := false
	log.Trace().Str("report_file", reportFile).Msg("Creating CSV reader")
	aCSV := csv.NewReader(csvFile)
	aCSV.FieldsPerRecord = -1
	log.Debug().Str("report_file", reportFile).Msg("Reading CSV data")
	inFile, cErr := aCSV.ReadAll()
	if cErr != nil {
		log.Error().Err(cErr).Str("report_file", reportFile).Msg("Error reading CSV file")
		return cErr
	}

	actions := make(map[string][]ROTAction)
	fieldMap := make(map[int]string)

	log.Debug().Str("report_file", reportFile).
		Strs("csv_header", AuditHeader).
		Msg("Creating field map, index to header name")
	for i, field := range AuditHeader {
		log.Trace().Str("report_file", reportFile).Str("field", field).Int(
			"index",
			i,
		).Msg("Processing field")
		fieldMap[i] = field
	}

	log.Debug().Str("report_file", reportFile).Msg("Iterating over CSV rows")
	var errs []error
	for ri, row := range inFile {
		log.Trace().Str("report_file", reportFile).Strs("row", row).Msg("Processing row")
		if strings.EqualFold(strings.Join(row, ","), strings.Join(AuditHeader, ",")) {
			log.Trace().Str("report_file", reportFile).Strs("row", row).Msg("Skipping header row")
			validHeader = true
			continue // Skip header
		}
		if !validHeader {
			invalidHeaderErr := fmt.Errorf(
				"invalid header in audit report file please use '%s'", strings.Join(
					AuditHeader,
					",",
				),
			)
			log.Error().Err(invalidHeaderErr).Str(
				"report_file",
				reportFile,
			).Msg("Invalid header in audit report file")
			return invalidHeaderErr
		}

		log.Debug().Str("report_file", reportFile).Msg("Creating action map")
		action := make(map[string]interface{})
		for i, field := range row {
			log.Trace().Str("report_file", reportFile).Str("field", field).Int(
				"index",
				i,
			).Msg("Processing field")
			fieldInt, iErr := strconv.Atoi(field)
			if iErr != nil {
				log.Trace().Err(iErr).Str("report_file", reportFile).
					Str("field", field).
					Int("index", i).
					Msg("Field is not an integer, replacing with index value")
				action[fieldMap[i]] = field
			} else {
				log.Trace().Err(iErr).Str("report_file", reportFile).
					Str("field", field).
					Int("index", i).
					Msg("Field is an integer")
				action[fieldMap[i]] = fieldInt
			}
		}

		log.Debug().Str("report_file", reportFile).Msg("Processing add cert action")
		addCertStr, aOk := action["AddCert"].(string)
		if !aOk {
			log.Warn().Str("report_file", reportFile).Msg(
				"AddCert field not found in action, " +
					"using empty string",
			)
			addCertStr = ""
		}

		log.Trace().Str("report_file", reportFile).Str(
			"add_cert",
			addCertStr,
		).Msg("Converting addCertStr to bool")
		addCert, acErr := strconv.ParseBool(addCertStr)
		if acErr != nil {
			log.Warn().Str("report_file", reportFile).Err(acErr).Msg(
				"Unable to parse bool from addCertStr, defaulting to FALSE",
			)
			addCert = false
		}

		log.Debug().Str("report_file", reportFile).Msg("Processing remove cert action")
		removeCertStr, rOk := action["RemoveCert"].(string)
		if !rOk {
			log.Warn().Str("report_file", reportFile).Msg(
				"RemoveCert field not found in action, " +
					"using empty string",
			)
			removeCertStr = ""
		}
		log.Trace().Str("report_file", reportFile).Str(
			"remove_cert",
			removeCertStr,
		).Msg("Converting removeCertStr to bool")
		removeCert, rcErr := strconv.ParseBool(removeCertStr)
		if rcErr != nil {
			log.Warn().
				Str("report_file", reportFile).
				Err(rcErr).
				Msg("Unable to parse bool from removeCertStr, defaulting to FALSE")
			removeCert = false
		}

		log.Trace().Str("report_file", reportFile).Msg("Processing store type")
		sType, sOk := action["StoreType"].(string)
		if !sOk {
			log.Warn().Str("report_file", reportFile).Msg(
				"StoreType field not found in action, " +
					"using empty string",
			)
			sType = ""
		}

		log.Trace().Str("report_file", reportFile).Msg("Processing store path")
		sPath, pOk := action["Path"].(string)
		if !pOk {
			log.Warn().Str("report_file", reportFile).Msg(
				"Path field not found in action, " +
					"using empty string",
			)
			sPath = ""
		}

		log.Trace().Str("report_file", reportFile).Msg("Processing thumbprint")
		tp, tpOk := action["Thumbprint"].(string)
		if !tpOk {
			log.Warn().Str("report_file", reportFile).Msg(
				"Thumbprint field not found in action, " +
					"using empty string",
			)
			tp = ""
		}

		log.Trace().Str("report_file", reportFile).Msg("Processing cert id")
		cid, cidOk := action["CertID"].(int)
		if !cidOk {
			log.Warn().Str("report_file", reportFile).Msg(
				"CertID field not found in action, " +
					"using -1",
			)
			cid = -1
		}

		if !tpOk && !cidOk {
			errMsg := fmt.Errorf("row is missing Thumbprint or CertID")
			log.Error().Err(errMsg).
				Str("report_file", reportFile).
				Int("row", ri).
				Msg("Invalid row in audit report file")
			errs = append(errs, errMsg)
			continue
		}

		sId, sIdOk := action["StoreID"].(string)
		if !sIdOk {
			errMsg := fmt.Errorf("row is missing StoreID")
			log.Error().Err(errMsg).
				Str("report_file", reportFile).
				Int("row", ri).
				Msg("Invalid row in audit report file")
			errs = append(errs, errMsg)
			continue
		}
		if cid == -1 && tp != "" {
			log.Debug().Str("report_file", reportFile).
				Int("row", ri).
				Str("thumbprint", tp).
				Msg("Looking up certificate by thumbprint")
			certLookupReq := api.GetCertificateContextArgs{
				IncludeMetadata:  boolToPointer(true),
				IncludeLocations: boolToPointer(true),
				CollectionId:     nil, //todo: add support for collection ID
				Thumbprint:       tp,
				Id:               0, //force to 0 as -1 will error out the API request
			}
			log.Debug().Str("report_file", reportFile).
				Int("row", ri).
				Str("thumbprint", tp).
				Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificateContext"))

			certLookup, err := kfClient.GetCertificateContext(&certLookupReq)
			if err != nil {
				log.Error().Err(err).Str("report_file", reportFile).
					Int("row", ri).
					Str("thumbprint", tp).
					Msg("Error looking up certificate by thumbprint")
				continue
			}
			cid = certLookup.Id
			log.Debug().Str("report_file", reportFile).
				Int("row", ri).
				Str("thumbprint", tp).
				Int("cert_id", cid).
				Msg("Certificate found by thumbprint")
		}

		log.Trace().Str("report_file", reportFile).
			Int("row", ri).
			Str("store_id", sId).
			Str("store_type", sType).
			Str("store_path", sPath).
			Str("thumbprint", tp).
			Int("cert_id", cid).
			Bool("add_cert", addCert).
			Bool("remove_cert", removeCert).
			Msg("Creating reconciliation action")
		a := ROTAction{
			StoreID:    sId,
			StoreType:  sType,
			StorePath:  sPath,
			Thumbprint: tp,
			CertID:     cid,
			AddCert:    addCert,
			RemoveCert: removeCert,
		}

		log.Trace().Str("report_file", reportFile).
			Int("row", ri).Interface("action", a).Msg("Adding action to actions map")
		actions[a.Thumbprint] = append(actions[a.Thumbprint], a)
	}

	log.Info().Str("report_file", reportFile).Msg("Audit report parsed successfully")
	if len(actions) == 0 {
		rtMsg := "No reconciliation actions to take, root stores are up-to-date. Exiting."
		log.Info().Str("report_file", reportFile).
			Msg(rtMsg)
		fmt.Println(rtMsg)
		if len(errs) > 0 {
			errStr := mergeErrsToString(&errs, false)
			log.Error().Str("report_file", reportFile).
				Str("errors", errStr).
				Msg("Errors encountered while parsing audit report")
			return fmt.Errorf("errors encountered while parsing audit report: %s", errStr)
		}
		return nil
	}

	log.Debug().Str("report_file", reportFile).Msg(fmt.Sprintf(DebugFuncCall, "reconcileRoots"))
	rErr := reconcileRoots(actions, kfClient, reportFile, dryRun)
	if rErr != nil {
		log.Error().Err(rErr).Str("report_file", reportFile).Msg("Error reconciling roots")
		return rErr
	}
	defer csvFile.Close()

	orchsURL := fmt.Sprintf(
		"https://%s/Keyfactor/Portal/AgentJobStatus/Index",
		kfClient.Hostname,
	) //todo: this pathing might not work for everyone

	if len(errs) > 0 {
		errStr := mergeErrsToString(&errs, false)
		log.Error().Str("report_file", reportFile).
			Str("errors", errStr).
			Msg("Errors encountered while reconciling root of trust stores")
		return fmt.Errorf("errors encountered while reconciling roots:\r\n%s", errStr)

	}

	log.Info().Str("report_file", reportFile).
		Str("orchs_url", orchsURL).
		Msg("Reconciliation completed. Check orchestrator jobs for details")
	fmt.Println(fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL))

	return nil
}

func init() {
	var (
		stores          string
		addCerts        string
		removeCerts     string
		minCertsInStore int
		maxPrivateKeys  int
		maxLeaves       int
		tType           = tTypeCerts
		outputFilePath  string
		inputFile       string
		storeTypes      []string
		containerNames  []string
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
		&outputFilePath, "outputFilePath", "o", "",
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
		&outputFilePath, "outputFilePath", "o", "",
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
		&outputFilePath, "outputFilePath", "o", "",
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

	rotGenStoreTemplateCmd.RegisterFlagCompletionFunc("type", templateTypeCompletion)
	rotGenStoreTemplateCmd.MarkFlagRequired("type")
}

func promptYesNo(q string) bool {
	isYes := false
	promptMsg := fmt.Sprintf("%s", q)
	//check if prompt ends with ? and add it if not
	if !strings.HasSuffix(promptMsg, "?") {
		promptMsg = fmt.Sprintf("%s?", promptMsg)
	}
	prompt := &survey.Confirm{
		Message: promptMsg,
	}
	survey.AskOne(prompt, &isYes)
	return isYes
}

func promptForFilePath(msg string) string {
	file := ""
	if msg == "" {
		msg = "input a file path"
	}
	prompt := &survey.Input{
		Message: msg,
		Suggest: func(toComplete string) []string {
			files, _ := filepath.Glob(toComplete + "*")
			return files
		},
	}
	survey.AskOne(prompt, &file)
	return file
}

func promptSelectFromAPIorFile(resourceType string) string {
	var selected string

	selected = promptSingleSelect(
		fmt.Sprintf("Source %s from:", resourceType),
		DefaultSourceTypeOptions,
		DefaultMenuPageSizeSmall,
	)
	return selected

}

func promptSelectCerts(client *api.Client) []string {
	searchOpts := []string{
		"Certificate",
		"Collection",
	}
	var selectedCerts []string

	selectedSearch := promptMultiSelect("Select certs to include in audit by:", searchOpts)
	if len(selectedSearch) == 0 {
		fmt.Println("No search options selected defaulting to 'Certificate'")
		selectedSearch = []string{"Certificate"}
	}

	log.Debug().Strs("selected_search", selectedSearch).Msg("Processing selected search options")
	for _, s := range selectedSearch {
		log.Trace().Str("search_option", s).Msg("Processing search option")
		switch s {
		case "Certificate":
			log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "menuCertificates"))
			certOpts, certErr := menuCertificates(client, nil)
			if certErr != nil {
				log.Error().Err(certErr).Msg("Error fetching certificates from Keyfactor Command")
				continue
			} else if len(certOpts) == 0 {
				fmt.Println("No certificates returned from Keyfactor Command")
				continue
			}
			selectedCerts = append(
				selectedCerts,
				promptMultiSelect("Select certificates to audit:", certOpts)...,
			)

		case "Collection":
			log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "menuCollections"))
			collectionOpts, colErr := menuCollections(client)
			if colErr != nil {
				log.Error().Err(colErr).Msg("Error fetching collections from Keyfactor Command")
				// todo: prompt for collection name or ID
				continue
			}
			if len(collectionOpts) == 0 {
				fmt.Println("No collections returned from Keyfactor Command")
				continue
			}
			var selectedCollections []string
			selectedCollections = append(
				selectedCollections,
				promptMultiSelect(
					"Select certificates associated with collection(s) to audit:",
					collectionOpts,
				)...,
			)
			//fetch certs associated with selected collections
			log.Info().Msg("Fetching certificates associated with selected collections")
			for _, col := range selectedCollections {
				//parse collection ID from selected collection
				colVals := strings.Split(col, ":")
				colID, idErr := strconv.Atoi(colVals[0])
				if idErr != nil {
					log.Error().
						Err(idErr).
						Str("collection", col).
						Msg("Error parsing collection ID, unable to fetch certificates")
					continue
				}

				params := make(map[string]string)
				params["CollectionID"] = fmt.Sprintf("%d", colID)
				log.Debug().
					Str("collection", col).
					Int("collection_id", colID).
					Interface("params", params).
					Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificatesByCollection"))
				certOpts, certErr := menuCertificates(client, &params)
				if certErr != nil {
					log.Error().Err(certErr).Msg("Error fetching certificates from Keyfactor Command")
					continue
				}
				if len(certOpts) == 0 {
					log.Warn().Str("collection", col).Msg("No certificates found associated with selected collection")
					fmt.Println(fmt.Sprintf("No certificates found associated with collection %s", col))
					continue
				}
				selectedCerts = append(selectedCerts, certOpts...)
			}
		}
	}
	return selectedCerts
}

func promptSelectStores(client *api.Client) []string {
	searchOpts := []string{
		"Store",
		"StoreType",
		"Container",
		//"Collection",
	}
	var selectedStores []string

	selectedSearch := promptMultiSelect("Select cert stores to audit by:", searchOpts)
	if len(selectedSearch) == 0 {
		fmt.Println("No search options selected defaulting to 'Store'")
		selectedSearch = []string{"Store"}
	}

	for _, s := range selectedSearch {
		switch s {
		case "Container":
			contOpts, contErr := menuContainers(client)
			if contErr != nil {
				fmt.Println("Error fetching containers from Keyfactor Command: ", contErr)
				continue
			} else if contOpts == nil || len(contOpts) == 0 {
				fmt.Println("No containers found")
				continue
			}

			log.Debug().Msg("Prompting user to select containers")
			selectedStores = append(
				selectedStores,
				promptMultiSelect("Select stores associated with container(s) to audit:", contOpts)...,
			)
		// Collection based store collection not supported as stores are not associated with collections certificates
		// are associated with collections
		//case "Collection":
		//	collectionOpts, colErr := menuCollections(client)
		//	if colErr != nil {
		//		fmt.Println("Error fetching collections from Keyfactor Command: ", colErr)
		//		continue
		//	} else if collectionOpts == nil || len(collectionOpts) == 0 {
		//		fmt.Println("No collections found")
		//		continue
		//	}
		//	var selectedCollections []string
		//	selectedCollections = append(
		//		selectedCollections,
		//		promptMultiSelect(
		//			"Select stores associated with collection(s) to audit:",
		//			collectionOpts,
		//		)...,
		//	)
		//
		//	//fetch stores associated with selected collections
		//	log.Info().Msg("Fetching stores associated with selected collections")
		//	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetStoresByCollection"))
		//	stores, sErr := client.GetSt(selectedCollections)

		case "StoreType":
			storeTypeNames, stErr := menuStoreType(client)
			if stErr != nil {
				fmt.Println("Error fetching store types from Keyfactor Command: ", stErr)
				continue
			} else if len(storeTypeNames) == 0 {
				fmt.Println("No store types found")
				continue
			}

			log.Debug().Msg("Prompting user to select store types")
			var selectedStoreTypes []string
			selectedStoreTypes = append(
				selectedStoreTypes,
				promptMultiSelect(
					"Select stores associated with store type(s) to audit:",
					storeTypeNames,
				)...,
			)

			//lookup stores associated with selected store types
			log.Info().Msg("Fetching stores associated with selected store types")
			for _, st := range selectedStoreTypes {
				//parse storetype ID from selected store type
				stVals := strings.Split(st, ":")
				stID, idErr := strconv.Atoi(stVals[0])
				if idErr != nil {
					log.Error().
						Err(idErr).
						Str("store_type", st).
						Msg("Error parsing store type ID, unable to fetch stores of type")
					continue
				}

				log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetStoresByStoreType"))
				params := make(map[string]interface{})
				params["CertStoreType"] = stID
				stores, sErr := menuCertificateStores(client, &params)
				if sErr != nil {
					fmt.Println("Error fetching stores from Keyfactor Command: ", sErr)
					continue
				} else if len(stores) == 0 {
					log.Warn().
						Str("store_type", st).
						Msg("No stores found associated with selected store type")
					fmt.Println(fmt.Sprintf("No stores of type %s found", st)) //todo: propagate to top CLI
					continue
				}
				selectedStores = append(selectedStores, stores...)
			}

		default:
			stNames, stErr := menuCertificateStores(client, nil)
			if stErr != nil {
				fmt.Println("Error fetching stores from Keyfactor Command: ", stErr)
				continue
			} else if stNames == nil || len(stNames) == 0 {
				fmt.Println("No stores found")
				continue
			}

			log.Debug().Msg("Prompting user to select stores")
			selectedStores = append(
				selectedStores,
				promptMultiSelect("Select stores to audit:", stNames)...,
			)
		}
	}
	return selectedStores
}

func promptSingleSelect(msg string, opts []string, menuPageSize int) string {
	if menuPageSize <= 0 {
		menuPageSize = DefaultMenuPageSizeSmall
	}
	var choice string
	prompt := &survey.Select{
		Message:  msg,
		Options:  opts,
		PageSize: menuPageSize,
	}
	survey.AskOne(prompt, &choice, survey.WithPageSize(10))
	return choice
}

func promptMultiSelect(msg string, opts []string) []string {
	var choices []string
	prompt := &survey.MultiSelect{
		Message:  msg,
		Options:  opts,
		PageSize: 10,
	}
	survey.AskOne(prompt, &choices, survey.WithPageSize(10))
	return choices
}

func menuStoreType(client *api.Client) ([]string, error) {
	//fetch store type options from keyfactor command
	log.Info().Msg("Fetching store types from Keyfactor Command")
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.ListCertificateStoreTypes"))
	storeTypes, stErr := client.ListCertificateStoreTypes()
	if stErr != nil {
		log.Error().Err(stErr).Msg("Error fetching store types from Keyfactor Command")
		return nil, stErr
	} else if storeTypes == nil || len(*storeTypes) == 0 {
		log.Warn().Msg("No store types returned from Keyfactor Command")
		//fmt.Println("No store types found")
		return nil, nil
	}

	var storeTypeNames []string
	log.Trace().Interface("store_types", storeTypes).Msg("Formatting store type choices for prompt")
	for _, st := range *storeTypes {
		log.Trace().Interface("store_type", st).Msg("Adding store type to options")
		stName := fmt.Sprintf("%d: %s", st.StoreType, st.Name)
		log.Trace().Str("store_type_name", stName).Msg("Adding store type to options")
		storeTypeNames = append(storeTypeNames, stName)
		log.Trace().Strs("store_type_options", storeTypeNames).Msg("Store type options")
	}
	return storeTypeNames, nil
}

func menuContainers(client *api.Client) ([]string, error) {
	//fetch container options from keyfactor command
	log.Info().Msg("Fetching containers from Keyfactor Command")
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetStoreContainers"))
	containers, cErr := client.GetStoreContainers()
	if cErr != nil {
		log.Error().Err(cErr).Msg("Error fetching containers from Keyfactor Command")
		return nil, cErr
	} else if containers == nil || len(*containers) == 0 {
		log.Warn().Msg("No containers returned from Keyfactor Command")
		return nil, nil
	}
	var contOpts []string
	log.Trace().
		Interface("containers", containers).
		Msg("Formatting container choices for prompt")
	for _, c := range *containers {
		contName := fmt.Sprintf("%d: %s", c.Id, c.Name)
		log.Trace().Str("container_name", contName).Msg("Adding container to options")
		contOpts = append(contOpts, contName)
		log.Trace().Strs("container_options", contOpts).Msg("Container options")
	}
	return contOpts, nil
}

func menuCollections(client *api.Client) ([]string, error) {
	//fetch collection options from keyfactor command
	log.Info().Msg("Fetching collections from Keyfactor Command")
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCollections"))

	sdkClient, sdkErr := convertClient(client)
	if sdkErr != nil {
		log.Error().Err(sdkErr).Msg("Error converting client to v2")
		return nil, sdkErr
	}
	//createdPamProviderType, httpResponse, rErr := sdkClient.PAMProviderApi.PAMProviderCreatePamProviderType(context.Background()).
	//	XKeyfactorRequestedWith(XKeyfactorRequestedWith).XKeyfactorApiVersion(XKeyfactorApiVersion).
	//	Type_(*pamProviderType).
	//	Execute()
	collections, httpResponse, collErr := sdkClient.CertificateCollectionApi.
		CertificateCollectionGetCollections(context.Background()).
		XKeyfactorRequestedWith(XKeyfactorRequestedWith).
		XKeyfactorApiVersion(XKeyfactorApiVersion).
		Execute()

	defer httpResponse.Body.Close()

	switch {
	case collErr != nil:
		log.Error().Err(collErr).Msg("Error fetching collections from Keyfactor Command")
		return nil, collErr
	case collections == nil || len(collections) == 0:
		log.Warn().Msg("No collections returned from Keyfactor Command")
		return nil, nil
	case httpResponse.StatusCode != http.StatusOK:
		log.Warn().Int("status_code", httpResponse.StatusCode).Msg("No collections returned from Keyfactor Command")
		return nil, fmt.Errorf("%s - no collections returned from Keyfactor Command", httpResponse.Status)
	}

	var collectionOpts []string
	log.Trace().Interface("collections", collections).Msg("Formatting collection choices for prompt")
	for _, c := range collections {
		collName := fmt.Sprintf("%d: %s", *c.Id, *c.Name)
		log.Trace().Str("collection_name", collName).Msg("Adding collection to options")
		collectionOpts = append(collectionOpts, collName)
		log.Trace().Strs("collection_options", collectionOpts).Msg("Collection options")
	}
	return collectionOpts, nil
}

func convertClient(v1Client *api.Client) (*sdk.APIClient, error) {
	// todo add support to convert the v1 client to v2 but for now use inputs used to created the v1 client
	config := make(map[string]string)

	if v1Client != nil {
		config["host"] = v1Client.Hostname
		//todo: expose these values in the client
		//config["username"] = v1Client.Username
		//config["password"] = v1Client.Password
		//config["domain"] = v1Client.Domain
	} else {
		config["host"] = kfcHostName
		config["username"] = kfcUsername
		config["password"] = kfcPassword
		config["domain"] = kfcDomain
	}

	configuration := sdk.NewConfiguration(config)
	sdkClient := sdk.NewAPIClient(configuration)
	return sdkClient, nil
}

func menuCertificates(client *api.Client, params *map[string]string) ([]string, error) {
	//fetch certificate options from keyfactor command
	log.Info().Msg("Fetching certificates from Keyfactor Command")
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "menuCertificates"))
	if params == nil {
		params = &map[string]string{}
	}
	certs, cErr := client.ListCertificates(*params)
	if cErr != nil {
		log.Error().Err(cErr).Msg("Error fetching certificates from Keyfactor Command")
		return nil, cErr
	} else if len(certs) == 0 {
		log.Warn().Msg("No certificates returned from Keyfactor Command")
		return nil, nil
	}

	var certOpts []string
	log.Trace().Interface("certificates", certs).Msg("Formatting certificate choices for prompt")
	for _, c := range certs {
		certName := fmt.Sprintf("%d: %s (%s) - %s", c.Id, c.IssuedCN, c.Thumbprint, c.NotBefore)
		log.Trace().Str("certificate_name", certName).Msg("Adding certificate to options")
		certOpts = append(certOpts, certName)
		log.Trace().Strs("certificate_options", certOpts).Msg("Certificate options")
	}
	log.Debug().Int("certificates", len(certOpts)).Msg(fmt.Sprintf(DebugFuncExit, "menuCertificates"))
	//sort certOps
	sort.Strings(certOpts)
	return certOpts, nil

}

func menuCertificateStores(client *api.Client, params *map[string]interface{}) ([]string, error) {
	// fetch all stores from keyfactor command
	log.Info().Msg("Fetching stores from Keyfactor Command")
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.ListCertificateStores"))
	stores, sErr := client.ListCertificateStores(params)
	if sErr != nil {
		log.Error().Err(sErr).Msg("Error fetching stores from Keyfactor Command")
		fmt.Println("Error fetching stores from Keyfactor Command: ", sErr)
		return nil, sErr
	} else if stores == nil || len(*stores) == 0 {
		log.Info().Msg("No stores returned from Keyfactor Command")
		fmt.Println("No stores found")
		return nil, nil
	}

	log.Trace().Interface("stores", stores).Msg("Formatting store choices for prompt")
	var stNames []string
	var storeTypesLookup = make(map[int]string)
	for _, st := range *stores {
		//lookup store type name
		var stName = fmt.Sprintf("%d", st.CertStoreType)
		if _, ok := storeTypesLookup[st.CertStoreType]; !ok {
			log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "kfClient.GetCertificateStoreType"))
			storeType, stErr := client.GetCertificateStoreType(st.CertStoreType)
			if stErr != nil {
				log.Error().Err(stErr).Msg("Error fetching store type name from Keyfactor Command")
			} else {
				storeTypesLookup[st.CertStoreType] = storeType.Name
				stName = storeType.Name
			}
		} else {
			stName = storeTypesLookup[st.CertStoreType]
		}

		log.Trace().Interface("store", st).Msg("Adding store to options")
		stMenuName := fmt.Sprintf(
			"%s/%s [%s]: (%s)", st.ClientMachine,
			st.StorePath, stName, st.Id,
		)
		log.Trace().Str("store_name", stMenuName).Msg("Adding store to options")
		stNames = append(stNames, stMenuName)
	}
	sort.Strings(stNames)
	return stNames, nil
}
