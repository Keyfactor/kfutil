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

type TrustStoreCriteria struct {
	MinCerts int
	MaxKeys  int
	MaxLeaf  int
}

func (t *TrustStoreCriteria) String() string {
	return fmt.Sprintf("MinCerts: %d, MaxKeys: %d, MaxLeaf: %d", t.MinCerts, t.MaxKeys, t.MaxLeaf)
}

var trustCriteria = TrustStoreCriteria{
	MinCerts: 1,
	MaxKeys:  0,
	MaxLeaf:  1,
}

type KFCStore struct {
	ApiResponse api.GetCertificateStoreResponse
	Inventory   []api.CertStoreInventory
}

type KFCStores struct {
	Stores map[string]KFCStore
}

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

func readCertsFile(certsFilePath string) (map[string]string, error) {
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
		log.Trace().Int("row", i).Msg("Processing row")
		if i == 0 {
			for j, h := range entry {
				headerMap[h] = j
			}
			continue
		}

		log.Trace().Strs("entry", entry).Msg("Processing row")
		switch entry[0] {
		case "CertID", "thumbprint", "id", "CertId", "Thumbprint",
			"Alias", "alias": //todo: is there a way to do this with a var?
			log.Trace().Strs("entry", entry).Msg("Skipping header row")
			continue // Skip header
		}
		alias := entry[headerMap["Alias"]]
		if alias == "" {
			tp := entry[headerMap["Thumbprint"]]
			if tp == "" {
				log.Warn().Strs("entry", entry).Msg("'Alias' and 'Thumbprint' are empty, skipping")
				continue
			}
			alias = tp
		}

		cId := entry[headerMap["CertID"]]
		if cId == "" {
			log.Warn().Strs("entry", entry).Msg("CertID is empty, skipping")
			continue
		}

		log.Trace().Strs("entry", entry).Msg("Adding thumbprint to map")
		certs[alias] = cId
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
	log.Debug().
		Int("min_certs", minCerts).
		Int("max_keys", maxKeys).
		Int("max_leaf", maxLeaf).
		Msg(fmt.Sprintf(DebugFuncEnter, "isRootStore"))
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
		Int("min_certs", minCerts).
		Int("leaf_count", leafCount).
		Int("max_leaves", maxLeaf).
		Int("key_count", keyCount).
		Int("max_keys", maxKeys).
		Msg("Store meets criteria to be considered a root of trust")
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "isRootStore"))
	return true
}

func (r *RootOfTrustManager) findTrustStores(
	containerName string,
) (*KFCStores, error) {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "findTrustStores"))
	trustStores := KFCStores{
		Stores: map[string]KFCStore{},
	}

	log.Info().Msg("Finding root of trust stList")
	log.Debug().Msg("Iterating over stList")

	// fetch list of stList from Keyfactor
	params := make(map[string]interface{})
	if containerName != "" {
		//check if name is an int
		_, err := strconv.Atoi(containerName)
		if err == nil {
			params["ContainerId"] = containerName
		} else {
			params["ContainerName"] = containerName
		}
	}
	log.Debug().
		Str("container", containerName).
		Interface("params", params).
		Msg(fmt.Sprintf(DebugFuncCall, "c.ListCertificateStores"))
	stList, stErr := r.Client.ListCertificateStores(&params)
	if stErr != nil {
		log.Error().Err(stErr).Msg("Error fetching stList from Keyfactor Command")
		return nil, stErr
	} else if stList == nil {
		log.Error().
			Interface("params", params).
			Msg("No stList returned from Keyfactor Command")
		return nil, fmt.Errorf("no stList returned from Keyfactor Command")
	}

	log.Debug().Str("stList", fmt.Sprintf("%v", stList)).Msg("Stores fetched successfully")

	var stLkErrs []string
	for _, st := range *stList {
		log.Debug().Str("store_id", st.Id).
			Str("store_id", st.Id).
			Str("store_path", st.StorePath).
			Str("client_machine", st.ClientMachine).
			Msg(fmt.Sprintf(DebugFuncCall, "GetCertStoreInventory"))
		inventory, invErr := r.Client.GetCertStoreInventory(st.Id)
		if invErr != nil {
			log.Error().Err(invErr).Str("store_id", st.Id).Msg("Error getting cert store inventory")
			errLine := fmt.Sprintf("%s,%s,%s,%s\n", st.Id, st.StorePath, st.ClientMachine, st.CertStoreType)
			stLkErrs = append(stLkErrs, errLine)
			continue
		} else if inventory == nil {
			log.Error().Str(
				"store_id",
				st.Id,
			).Msg("No inventory response returned for store from Keyfactor Command")
			errLine := fmt.Sprintf("%s,%s,%s,%s\n", st.Id, st.StorePath, st.ClientMachine, st.CertStoreType)
			stLkErrs = append(stLkErrs, errLine)
			continue
		}

		log.Debug().Str("store_id", st.Id).
			Int("min_certs", r.TrustStoreCriteria.MinCerts).
			Int("max_keys", r.TrustStoreCriteria.MaxKeys).
			Int("max_leaf", r.TrustStoreCriteria.MaxLeaves).
			Str("store_id", st.Id).
			Str("store_path", st.StorePath).
			Str("client_machine", st.ClientMachine).
			Msg(fmt.Sprintf(DebugFuncCall, "isRootStore"))
		if isRootStore(
			&st, inventory, r.TrustStoreCriteria.MinCerts, r.TrustStoreCriteria.MaxKeys,
			r.TrustStoreCriteria.MaxLeaves,
		) {
			log.Info().
				Str("store_id", st.Id).
				Str("store_path", st.StorePath).
				Str("client_machine", st.ClientMachine).
				Msg("certificate store is considered a 'trust' store")
			trstSt := KFCStore{
				ApiResponse: st,
				Inventory:   *inventory,
			}
			trustStores.Stores[st.Id] = trstSt
			continue
		}
		log.Info().
			Int("min_certs", r.TrustStoreCriteria.MinCerts).
			Int("max_keys", r.TrustStoreCriteria.MaxKeys).
			Int("max_leaf", r.TrustStoreCriteria.MaxLeaves).
			Str("store_id", st.Id).
			Str("store_path", st.StorePath).
			Str("client_machine", st.ClientMachine).
			Msg("certificate store is NOT considered a 'trust' store")
	}

	if len(stLkErrs) > 0 {
		log.Error().
			Strs("stList", stLkErrs).
			Msg("Error looking up inventory for stList")
		log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "findTrustStores"))
		return &trustStores, fmt.Errorf("error looking up inventory for stList: %s", strings.Join(stLkErrs, ","))
	}

	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "findTrustStores"))
	return &trustStores, nil
}

func (r *RootOfTrustManager) validateStoresInput(storesFile *string, noPrompt *bool) error {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "validateStoresInput"))

	if noPrompt == nil {
		noPrompt = boolToPointer(false)
	}

	if (storesFile == nil || *storesFile == "") && r.StoresFilePath != "" {
		log.Debug().
			Str("stores_file", r.StoresFilePath).
			Bool("no_prompt", *noPrompt).
			Msg("Setting stores file path from struct")
		storesFile = &r.StoresFilePath
	}

	log.Debug().Str("stores_file", *storesFile).Bool("no_prompt", *noPrompt).Msg("Validating stores input")

	if storesFile == nil || *storesFile == "" {
		if *noPrompt {
			return fmt.Errorf("stores file is required, use flag `--stores` to specify 1 or more file paths")
		}
		apiOrFile := promptSelectRotStores("certificate stores")
		switch apiOrFile {
		case "All":
			selectedStores, sErr := r.Client.ListCertificateStores(nil)
			if sErr != nil {
				return sErr
			}
			if len(*selectedStores) == 0 {
				return errors.New("no certificate stores selected, unable to continue")
			}
			//create stores file
			storesFile = stringToPointer(DefaultROTAuditStoresOutfilePath)
			// create file
			f, ioErr := os.Create(*storesFile)
			if ioErr != nil {
				log.Error().Err(ioErr).Str("stores_file", *storesFile).Msg("Error creating stores file")
				return ioErr
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
				return wErr
			}
			// write selected stores
			r.Stores = make(map[string]*TrustStore)
			for _, store := range *selectedStores {
				log.Debug().Str("store_id", store.Id).Msg("Adding store to stores file")
				//parse ID from selection `<id>: <name>`
				storeId := store.Id
				//remove () and white spaces from storeId
				storeId = strings.Trim(strings.Trim(strings.Trim(storeId, " "), "("), ")")

				tStore := TrustStore{
					StoreID:       storeId,
					StoreType:     fmt.Sprintf("%d", store.CertStoreType), //todo: look up name
					StoreMachine:  store.ClientMachine,
					StorePath:     store.StorePath,
					ContainerName: store.ContainerName,
					ContainerID:   store.ContainerId,
					Inventory:     []api.CertStoreInventory{},
				}

				r.Stores[storeId] = &tStore

				storeInstance := ROTStore{
					StoreID:       storeId,
					StoreType:     fmt.Sprintf("%d", store.CertStoreType), //todo: look up name
					StoreMachine:  store.ClientMachine,
					StorePath:     store.StorePath,
					ContainerId:   fmt.Sprintf("%d", store.ContainerId),
					ContainerName: store.ContainerName,
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
			r.StoresFilePath = *storesFile
			return nil
		case "Manual Select":
			selectedStores := promptSelectStores(r.Client)
			if len(selectedStores) == 0 {
				return errors.New("no certificate stores selected, unable to continue")
			}
			//create stores file
			storesFile = stringToPointer(fmt.Sprintf("%s", DefaultROTAuditStoresOutfilePath))
			// create file
			f, ioErr := os.Create(*storesFile)
			if ioErr != nil {
				log.Error().Err(ioErr).Str("stores_file", *storesFile).Msg("Error creating stores file")
				return ioErr
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
				return wErr
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
			r.StoresFilePath = *storesFile
			return nil
		case "File":

			r.StoresFilePath = promptForFilePath("Input a file path for the CSV file containing stores to audit.")
			return nil
		case "Search":
			promptForCriteria()
			trusts, sErr := r.findTrustStores("")
			if sErr != nil {
				return sErr
			} else if trusts == nil || trusts.Stores == nil || len(trusts.Stores) == 0 {
				return fmt.Errorf("no trust stores found using the following criteria:\n%s", trustCriteria.String())
			}
			storesFile = stringToPointer(DefaultROTAuditStoresOutfilePath)
			// create file
			f, ioErr := os.Create(*storesFile)
			if ioErr != nil {
				log.Error().Err(ioErr).Str("stores_file", *storesFile).Msg("Error creating stores file")
				return ioErr
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
				return wErr
			}
			for _, store := range trusts.Stores {
				storeInstance := ROTStore{
					StoreID:       store.ApiResponse.Id,
					StoreType:     fmt.Sprintf("%d", store.ApiResponse.CertStoreType),
					StoreMachine:  store.ApiResponse.ClientMachine,
					StorePath:     store.ApiResponse.StorePath,
					ContainerId:   fmt.Sprintf("%d", store.ApiResponse.ContainerId),
					ContainerName: store.ApiResponse.ContainerName,
					LastQueried:   getCurrentTime(""),
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
			r.StoresFilePath = f.Name()
			return nil
		default:
			errors.New("invalid selection")
		}
	}
	r.StoresFilePath = *storesFile
	return nil
}

func (r *RootOfTrustManager) validateCertsInput(addRootsFile string, removeRootsFile string, noPrompt bool) error {
	log.Debug().Str("add_certs_file", addRootsFile).
		Str("remove_certs_file", removeRootsFile).
		Bool("no_prompt", noPrompt).
		Msg(fmt.Sprintf(DebugFuncEnter, "validateCertsInput"))

	if addRootsFile == "" && removeRootsFile == "" && noPrompt {
		return InvalidROTCertsInputErr
	}

	if addRootsFile == "" || removeRootsFile == "" {
		if addRootsFile == "" && !noPrompt {
			//prmpt := "Would you like to include a 'certs to add' CSV file?"
			prmpt := "Provide certificates to add to and/or that should be present in selected stores?"
			provideAddFile := promptYesNo(prmpt)
			if provideAddFile {
				addSrcType := promptSelectFromAPIorFile("certificates")
				switch addSrcType {
				case "API":
					selectedCerts := promptSelectCerts(r.Client)
					if len(selectedCerts) == 0 {
						return InvalidROTCertsInputErr
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
						return ioErr
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
						return wErr
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
			prmpt := "Provide certificates to remove from and/or that should NOT be present in selected stores?"
			provideRemoveFile := promptYesNo(prmpt)
			if provideRemoveFile {
				//removeRootsFile = promptForFilePath("Input a file path for the 'certs to remove' CSV. ")
				remSrcType := promptSelectFromAPIorFile("certificates")
				switch remSrcType {
				case "API":
					selectedCerts := promptSelectCerts(r.Client)
					if len(selectedCerts) == 0 {
						return InvalidROTCertsInputErr
					}
					//create stores file
					removeRootsFile = fmt.Sprintf("%s", DefaultROTAuditRemoveCertsOutfilePath)
					// create file
					f, ioErr := os.Create(removeRootsFile)
					if ioErr != nil {
						log.Error().Err(ioErr).Str(
							"remove_certs_file",
							removeRootsFile,
						).Msg("Error creating certs to remove file")
						return ioErr
					}
					defer f.Close()
					// create CSV writer
					log.Debug().Str("remove_certs_file", removeRootsFile).Msg("Creating CSV writer")
					writer := csv.NewWriter(f)
					defer writer.Flush()
					// write header
					log.Debug().Str("remove_certs_file", removeRootsFile).Msg("Writing header to certs to remove file")
					wErr := writer.Write(CertHeader)
					if wErr != nil {
						log.Error().Err(wErr).Str(
							"stores_file",
							removeRootsFile,
						).Msg("Error writing header to stores file")
						return wErr
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
								"remove_certs_file",
								removeRootsFile,
							).Msg("Error writing store to stores file")
							continue
						}
					}
					writer.Flush()
					f.Close()
				default:
					removeRootsFile = promptForFilePath("Input a file path for the 'certs to remove' CSV.")
				}
			}
		}
		if addRootsFile == "" && removeRootsFile == "" {
			return InvalidROTCertsInputErr
		}
	}
	r.AddCertsFilePath = addRootsFile
	r.RemoveCertsFilePath = removeRootsFile

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
		collections     []string
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
		&outputFilePath, "OutputFilePath", "o", "",
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
		&outputFilePath, "OutputFilePath", "o", "",
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
		&outputFilePath, "OutputFilePath", "o", "",
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
		&collections,
		"collection",
		[]string{},
		"Certificate collection name(s) to pre-populate the stores template with. If not specified, the template will be empty.",
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

func promptSelectRotStores(resourceType string) string {
	var selected string

	opts := []string{
		"Manual Select",
		"Search",
		"File",
		"All",
	}
	//sort ops
	sort.Strings(opts)

	selected = promptSingleSelect(
		fmt.Sprintf("Source %s from:", resourceType),
		opts,
		DefaultMenuPageSizeSmall,
	)
	return selected

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
					Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertificatesByCollection"))
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
		//	collectionOpts, colErr := menuCollections(Client)
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
		//	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.GetStoresByCollection"))
		//	stores, sErr := Client.GetSt(selectedCollections)

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

				log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.GetStoresByStoreType"))
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
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.ListCertificateStoreTypes"))
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
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.GetStoreContainers"))
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
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCollections"))

	sdkClient, sdkErr := convertClient(client)
	if sdkErr != nil {
		log.Error().Err(sdkErr).Msg("Error converting Client to v2")
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
	// todo add support to convert the v1 Client to v2 but for now use inputs used to created the v1 Client
	config := make(map[string]string)

	if v1Client != nil {
		config["host"] = v1Client.Hostname
		//todo: expose these values in the Client
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
	log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.ListCertificateStores"))
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
			log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertificateStoreType"))
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

func promptForCriteria() error {
	var maxKeys int
	prompt := &survey.Input{
		Message: "Enter max private keys:",
		Help: "Enter the maximum number of private keys allowed in a certificate store for it to be considered" +
			" a trusted root store",
		Default: fmt.Sprintf("%d", trustCriteria.MaxKeys),
	}
	survey.AskOne(prompt, &maxKeys)

	var minCerts int
	prompt = &survey.Input{
		Message: "Enter min certs in store:",
		Help: "Enter the minimum number of certificates allowed in a certificate store for it to be considered" +
			" a trusted root store",
		Default: fmt.Sprintf("%d", trustCriteria.MinCerts),
	}
	survey.AskOne(prompt, &minCerts)

	var maxLeaves int
	prompt = &survey.Input{
		Message: "Enter max leaf certs in store:",
		Help: "Enter the maximum number of non-root certificates allowed in a certificate store for it to be considered" +
			" a trusted root store",
		Default: fmt.Sprintf("%d", trustCriteria.MaxLeaf),
	}
	survey.AskOne(prompt, &maxLeaves)

	trustCriteria.MaxKeys = maxKeys
	trustCriteria.MinCerts = minCerts
	trustCriteria.MaxLeaf = maxLeaves
	return nil
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
		Example:                "kfutil stores rot audit",
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
			outputFilePath, _ := cmd.Flags().GetString("OutputFilePath")

			// Debug + expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			informDebug(debugFlag)

			log.Debug().Str("trust_criteria", fmt.Sprintf("%s", trustCriteria.String())).
				Str("add_file", addRootsFile).
				Str("remove_file", removeRootsFile).
				Int("min_certs", minCerts).
				Int("max_keys", maxKeys).
				Int("max_leaves", maxLeaves).
				Bool("dry_run", dryRun).
				Msg("Root of trust audit command")

			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
			kfClient, cErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
			if cErr != nil {
				log.Error().Err(cErr).Msg("Error initializing Keyfactor Client")
				return cErr
			}

			rotCriteria := RootOfTrustCriteria{
				MinCerts:  minCerts,
				MaxLeaves: maxLeaves,
				MaxKeys:   maxKeys,
			}

			rotClient := RootOfTrustManager{
				AddCertsFilePath:    addRootsFile,
				RemoveCertsFilePath: removeRootsFile,
				StoresFilePath:      storesFile,
				ReportFilePath:      "",
				TrustStoreCriteria:  rotCriteria,
				OutputFilePath:      outputFilePath,
				IsDryRun:            dryRun,
				Client:              kfClient,
			}

			// validate flags
			var storesErr error
			log.Debug().Str("stores_file", storesFile).Bool("no_prompt", noPrompt).
				Msg(fmt.Sprintf(DebugFuncCall, "validateStoresInput"))
			storesErr = rotClient.validateStoresInput(&storesFile, &noPrompt)
			if storesErr != nil {
				return storesErr
			}

			log.Debug().Str("add_file", addRootsFile).Str("remove_file", removeRootsFile).Bool("no_prompt", noPrompt).
				Msg(fmt.Sprintf(DebugFuncCall, "validateCertsInput"))
			var certsErr error
			certsErr = rotClient.validateCertsInput(
				addRootsFile, removeRootsFile, noPrompt,
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

			//Process stores file
			sErr := rotClient.processStoresFile()
			if sErr != nil {
				log.Error().Err(sErr).Msg("Error processing stores file")
				return sErr
			}

			ctErr := rotClient.processCertsFiles()
			if ctErr != nil {
				log.Error().Err(ctErr).Msg("Error processing certs files")
				return ctErr
			}

			log.Debug().Msg(fmt.Sprintf(DebugFuncCall, "generateAuditReport"))
			gErr := rotClient.generateAuditReport()
			if gErr != nil {
				log.Error().Err(gErr).Msg("Error generating audit report")
				return gErr
			}

			log.Info().
				Str("OutputFilePath", outputFilePath).
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
			outputFilePath, _ := cmd.Flags().GetString("OutputFilePath")

			// Debug + expEnabled checks
			isExperimental := false
			debugErr := warnExperimentalFeature(expEnabled, isExperimental)
			if debugErr != nil {
				return debugErr
			}
			informDebug(debugFlag)

			// Check KFC connection
			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
			kfClient, clErr := initClient(configFile, profile, "", "", noPrompt, authConfig, false)
			if clErr != nil {
				log.Error().Err(clErr).Msg("Error initializing Keyfactor Client")
				return clErr
			}

			log.Info().Str("stores_file", storesFile).
				Str("add_file", addRootsFile).
				Str("remove_file", removeRootsFile).
				Bool("dry_run", dryRun).
				Msg("Performing root of trust reconciliation")

			rotCriteria := RootOfTrustCriteria{
				MinCerts:  minCerts,
				MaxLeaves: maxLeaves,
				MaxKeys:   maxKeys,
			}

			rotClient := RootOfTrustManager{
				AddCertsFilePath:    addRootsFile,
				RemoveCertsFilePath: removeRootsFile,
				StoresFilePath:      storesFile,
				ReportFilePath:      reportFile,
				TrustStoreCriteria:  rotCriteria,
				OutputFilePath:      outputFilePath,
				IsDryRun:            dryRun,
				Client:              kfClient,
			}

			// Parse existing audit report
			if isCSV && reportFile != "" {
				log.Debug().Str("report_file", reportFile).Msg("Processing audit report")
				err := rotClient.processCSVReportFile()
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

				log.Trace().Str("rotClient", fmt.Sprintf("%s", rotClient.String())).Msg("Root of trust Client")

				//Process stores file
				sErr := rotClient.processStoresFile()
				if sErr != nil {
					log.Error().Err(sErr).Msg("Error processing stores file")
					return sErr
				}

				cErr := rotClient.processCertsFiles()
				if cErr != nil {
					log.Error().Err(cErr).Msg("Error processing certs files")
					return cErr
				}

				gErr := rotClient.generateAuditReport()
				if gErr != nil {
					log.Error().Err(gErr).Msg("Error generating audit report")
					return gErr
				}

				rErr := rotClient.reconcileRoots()
				if rErr != nil {
					log.Error().Err(rErr).Msg("Error reconciling roots")
					return rErr
				}

				orchsURL := fmt.Sprintf(
					"https://%s/Keyfactor/Portal/AgentJobStatus/Index",
					kfClient.Hostname,
				) //todo: this path might not work for everyone

				log.Info().
					Str("orchs_url", orchsURL).
					Str("OutputFilePath", outputFilePath).
					Msg("Reconciliation completed. Check orchestrator jobs for details.")
				fmt.Println(fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL))
			}

			log.Debug().Str("report_file", reportFile).
				Str("OutputFilePath", outputFilePath).Msg("Reconciliation report generated successfully")
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
			outputFilePath, _ := cmd.Flags().GetString("OutputFilePath")
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
				log.Error().Err(clErr).Msg("Error initializing Keyfactor Client")
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
