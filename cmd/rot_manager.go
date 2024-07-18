package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
)

type RootOfTrustManager struct {
	Client         *api.Client
	OutputFilePath string
	addCerts       map[string]string      // map[alias]certId
	removeCerts    map[string]string      // map[alias]certId
	actions        map[string][]ROTAction // map[alias]ROTAction
	//stores              map[string]StoreCSVEntry
	Stores              map[string]*TrustStore
	data                [][]string
	ReportFilePath      string
	StoresFilePath      string
	AddCertsFilePath    string
	RemoveCertsFilePath string
	IsDryRun            bool
	TrustStoreCriteria  RootOfTrustCriteria
}

type TrustStore struct {
	StoreID       string
	StoreType     string
	StoreMachine  string
	StorePath     string
	Inventory     []api.CertStoreInventory
	ContainerName string
	ContainerID   int
	LeafCount     int
	KeyCount      int
	CertCount     int
	ThumbPrints   map[string]bool
	Serials       map[string]bool
	Aliases       map[string]bool
	CertIDs       map[int]bool
}

func (t *TrustStore) generateMaps() error {
	// check if inventory is empty
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "generateMaps"))
	if len(t.Inventory) == 0 {
		log.Warn().Msg("Inventory is empty, unable to generate maps")
		log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "generateMaps"))
		return nil
	}
	log.Debug().Msg("Generating thumbprint, serial number, and certificate ID maps")
	for _, cert := range t.Inventory {
		log.Trace().Str("alias", cert.Name).Msg("Adding alias to map")
		t.Aliases[cert.Name] = true
		// Thumbprints
		for _, thumbprint := range cert.Thumbprints {
			log.Trace().Str("thumbprint", thumbprint).Msg("Adding thumbprint to map")
			t.ThumbPrints[thumbprint] = true
		}
		// Serials
		for _, serial := range cert.Serials {
			log.Trace().Str("serial", serial).Msg("Adding serial number to map")
			t.Serials[serial] = true
		}
		// Cert IDs
		for _, id := range cert.Ids {
			log.Trace().Int("cert_id", id).Msg("Adding certificate ID to map")
			t.CertIDs[id] = true
		}
	}
	return nil
}

func (t *TrustStore) InventoryCSV(includeHeader bool) string {
	// order CSV col based InventoryHeader
	InventoryHeader := []string{
		"Thumbprint",
		"SerialNumber",
		"CertificateID",
	}
	var output [][]string

	// add header
	if includeHeader {
		output = append(output, InventoryHeader)
	}

	for _, h := range t.Inventory {
		row := make([]string, len(InventoryHeader))
		for i, header := range InventoryHeader {
			switch header {
			case "Alias":
				row[i] = h.Name
			case "Thumbprint":
				row[i] = strings.Join(h.Thumbprints, ",")
				// escape any commas in the fields
				row[i] = strings.ReplaceAll(row[i], ",", "\\,")
			case "SerialNumbers":
				row[i] = strings.Join(h.Serials, ",")
				// escape any commas in the fields
				row[i] = strings.ReplaceAll(row[i], ",", "\\,")
			case "CertificateID":
				// join int slice to string
				var certIDs []string
				for _, id := range h.Ids {
					certIDs = append(certIDs, strconv.Itoa(id))
				}
				row[i] = strings.Join(certIDs, ",")
				// escape any commas in the fields
				row[i] = strings.ReplaceAll(row[i], ",", "\\,")
			}
		}
		output = append(output, row)
	}
	// flatten into single string with newlines
	var csvOutput []string
	for _, o := range output {
		//escape any commas in the fields
		csvOutput = append(csvOutput, strings.Join(o, ","))
	}
	return strings.Join(csvOutput, "\n")
}

func (t *TrustStore) ToCSV() string {
	// order CSV col based StoreHeader
	output := make([]string, len(StoreHeader))
	for i, h := range StoreHeader {
		switch h {
		case "StoreID":
			output[i] = t.StoreID
		case "StoreType":
			output[i] = t.StoreType
		case "StoreMachine":
			output[i] = t.StoreMachine
		case "StorePath":
			output[i] = t.StorePath
		case "Inventory":
			output[i] = fmt.Sprintf("%v", len(t.Inventory))
		case "ContainerName":
			output[i] = t.ContainerName
		case "ContainerID":
			output[i] = fmt.Sprintf("%d", t.ContainerID)
		}
	}
	//escape any commas in the fields
	for i, o := range output {
		output[i] = strings.ReplaceAll(o, ",", "\\,")
	}
	return strings.Join(output, ",")
}

func (t *TrustStore) String() string {
	return fmt.Sprintf(
		"StoreID: %s, StoreType: %s, StoreMachine: %s, StorePath: %s, Inventory: %v",
		t.StoreID,
		t.StoreType,
		t.StoreMachine,
		t.StorePath,
		len(t.Inventory),
	)
}

type RootOfTrustCriteria struct {
	MinCerts  int
	MaxLeaves int
	MaxKeys   int
}

func (r *RootOfTrustManager) String() string {
	output := "StoresFilePath: " + r.StoresFilePath + "\n"
	output += "AddCertsFilePath: " + r.AddCertsFilePath + "\n"
	output += "RemoveCertsFilePath: " + r.RemoveCertsFilePath + "\n"
	output += "ReportFilePath: " + r.ReportFilePath + "\n"
	output += "MinCerts: " + strconv.Itoa(r.TrustStoreCriteria.MinCerts) + "\n"
	output += "MaxLeaves: " + strconv.Itoa(r.TrustStoreCriteria.MaxLeaves) + "\n"
	output += "MaxKeys: " + strconv.Itoa(r.TrustStoreCriteria.MaxKeys) + "\n"
	output += "AddCerts: \n"
	for k, v := range r.addCerts {
		output += "\t" + k + ": " + v + "\n"
	}
	output += "RemoveCerts: \n"
	for k, v := range r.removeCerts {
		output += "\t" + k + ": " + v + "\n"
	}
	output += "Actions: \n"
	for k, v := range r.actions {
		output += "\t" + k + ": \n"
		for _, a := range v {
			output += "\t\t" + a.Thumbprint + "\n"
			output += "\t\t\tAddCert: " + fmt.Sprintf("%v", a.AddCert) + "\n"
			output += "\t\t\tRemoveCert: " + fmt.Sprintf("%v", a.RemoveCert) + "\n"
		}
	}

	return output
}

func (r *RootOfTrustManager) generateAuditReport() error {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "generateAuditReport"))
	log.Info().Str("output_file", r.OutputFilePath).Msg("Generating audit report")
	var (
		data [][]string
	)

	data = append(data, AuditHeader)
	var csvFile *os.File
	var fErr error
	log.Debug().Str("output_file", r.OutputFilePath).Msg("Checking for output file")
	if r.OutputFilePath == "" {
		log.Debug().Str("output_file", reconcileDefaultFileName).Msg("No output file specified, using default")
		csvFile, fErr = os.Create(reconcileDefaultFileName)
		r.OutputFilePath = reconcileDefaultFileName
	} else {
		log.Debug().Str("output_file", r.OutputFilePath).Msg("Creating output file")
		csvFile, fErr = os.Create(r.OutputFilePath)
	}

	if fErr != nil {
		fmt.Printf("%s", fErr)
		log.Error().Err(fErr).Str("output_file", r.OutputFilePath).Msg("Error creating output file")
	}

	log.Trace().Str("output_file", r.OutputFilePath).Msg("Creating CSV writer")
	csvWriter := csv.NewWriter(csvFile)
	log.Debug().Str("output_file", r.OutputFilePath).Strs("csv_header", AuditHeader).Msg("Writing header to CSV")
	cErr := csvWriter.Write(AuditHeader)
	if cErr != nil {
		log.Error().Err(cErr).Str("output_file", r.OutputFilePath).Msg("Error writing header to CSV")
		return cErr
	}

	log.Trace().Str("output_file", r.OutputFilePath).Msg("Creating actions map")
	actions := make(map[string][]ROTAction)

	var errs []error

	// process certs to add
	addData, addActions, addErrs := r.processAddCerts(csvWriter)
	errs = append(
		errs,
		addErrs...,
	)
	data = append(data, addData...)
	for k, v := range addActions {
		actions[k] = append(actions[k], v...)
	}

	// process certs to remove
	removeData, removeActions, removeErrs := r.processRemoveCerts(csvWriter)
	errs = append(
		errs,
		removeErrs...,
	)
	data = append(data, removeData...)
	for k, v := range removeActions {
		actions[k] = append(actions[k], v...)
	}
	log.Trace().
		Str("output_file", r.OutputFilePath).
		Msg("Flushing CSV writer")
	csvWriter.Flush()
	log.Trace().
		Str("output_file", r.OutputFilePath).
		Msg("Closing CSV file")
	ioErr := csvFile.Close()
	if ioErr != nil {
		log.Error().
			Err(ioErr).
			Str("output_file", r.OutputFilePath).
			Msg("Error closing CSV file")
	}
	log.Info().
		Str("output_file", r.OutputFilePath).
		Msg("Audit report written to disk successfully")
	fmt.Printf("Audit report written to %s\n", r.OutputFilePath) //todo: send to output formatter
	fmt.Printf(
		"Please review the report and run `kfutil stores rot reconcile --import-csv --input"+
			"-file %s` apply the changes\n", r.OutputFilePath,
	) //todo: send to output formatter

	r.actions = actions
	r.data = data

	if len(errs) > 0 {
		errStr := mergeErrsToString(&errs, false)
		log.Trace().Str("output_file", r.OutputFilePath).Str(
			"errors",
			errStr,
		).Msg("The following errors occurred while generating audit report")
		return fmt.Errorf("the following errors occurred while generating audit report:\r\n%s", errStr)
	}
	log.Debug().Msg(fmt.Sprintf(DebugFuncExit, "generateAuditReport"))
	return nil
}

func (r *RootOfTrustManager) processRemoveCerts(
	csvWriter *csv.Writer,
) (
	[][]string,
	map[string][]ROTAction,
	[]error,
) {
	var (
		data    [][]string
		errs    []error
		actions = make(map[string][]ROTAction)
	)
	for tp, cId := range r.removeCerts {
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
			Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertificateContext"))
		certLookup, err := r.Client.GetCertificateContext(&certLookupReq)
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
		log.Debug().Int("certID", certID).Msg("Processing cert to remove")
		//certIDStr := certLookup.Id
		log.Debug().Str("thumbprint", tp).Msg("Iterating over stores")
		for _, store := range r.Stores {
			log.Debug().Str("thumbprint", tp).Str(
				"store_id",
				store.StoreID,
			).Msg("Checking if cert is deployed to store")
			//TODO: This logic should be replaced by receiver method on TrustStore
			//if _, ok := store.Inventory; !ok {
			//	// Cert is already in the store do nothing
			//	log.Info().Str("thumbprint", tp).Str("store_id", store.ID).Msg("Cert is not deployed to store")
			//	row := []string{
			//		//todo: this should be a toCSV field on whatever object this is
			//		tp,
			//		certIDStr,
			//		certLookup.IssuedDN,
			//		certLookup.IssuerDN,
			//		store.ID,
			//		store.Type,
			//		store.Machine,
			//		store.Path,
			//		"false", // Add to store
			//		"false", // Remove from store
			//		"false", // Is Deployed
			//		getCurrentTime(""),
			//	}
			//	log.Trace().Str("thumbprint", tp).Strs("row", row).Msg("Appending data row")
			//	data = append(data, row)
			//	log.Trace().Str("thumbprint", tp).Strs("row", row).Msg("Writing data row to CSV")
			//	wErr := csvWriter.Write(row)
			//	if wErr != nil {
			//		log.Error().
			//			Err(wErr).
			//			Str("thumbprint", tp).
			//			Str("output_file", r.OutputFilePath).
			//			Strs("row", row).
			//			Msg("Error writing row to CSV")
			//	}
			//} else {
			//	// Cert is deployed to this store and will need to be removed
			//	log.Info().
			//		Str("thumbprint", tp).
			//		Str("store_id", store.ID).
			//		Msg("Cert is deployed to store")
			//	row := []string{
			//		//todo: this should be a toCSV
			//		tp,
			//		certIDStr,
			//		certLookup.IssuedDN,
			//		certLookup.IssuerDN,
			//		store.ID,
			//		store.Type,
			//		store.Machine,
			//		store.Path,
			//		"false", // Add to store
			//		"true",  // Remove from store
			//		"true",  // Is Deployed
			//		getCurrentTime(""),
			//	}
			//	log.Trace().
			//		Str("thumbprint", tp).
			//		Strs("row", row).
			//		Msg("Appending data row")
			//	data = append(data, row)
			//	log.Debug().
			//		Str("thumbprint", tp).
			//		Strs("row", row).
			//		Msg("Writing data row to CSV")
			//	wErr := csvWriter.Write(row)
			//	if wErr != nil {
			//		log.Error().
			//			Err(wErr).
			//			Str("thumbprint", tp).
			//			Str("output_file", r.OutputFilePath).
			//			Strs("row", row).
			//			Msg("Error writing row to CSV")
			//	}
			//	log.Debug().
			//		Str("thumbprint", tp).
			//		Msg("Adding 'remove' action to actions map")
			//	actions[tp] = append(
			//		actions[tp], ROTAction{
			//			Thumbprint: tp,
			//			StoreAlias: "", //TODO get this value
			//			CertID:     certID,
			//			StoreID:    store.ID,
			//			StoreType:  store.Type,
			//			StorePath:  store.Path,
			//			AddCert:    false,
			//			RemoveCert: true,
			//			Deployed:   true,
			//		},
			//	)
			//}
		}
	}
	return data, actions, errs
}

func (r *RootOfTrustManager) processAddCerts(
	csvWriter *csv.Writer,
) (
	[][]string,
	map[string][]ROTAction,
	[]error,
) {
	var (
		data    [][]string
		errs    []error
		actions = make(map[string][]ROTAction)
	)
	for tp, cId := range r.addCerts {
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
			Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertificateContext"))
		certLookup, err := r.Client.GetCertificateContext(&certLookupReq)
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
		log.Debug().Int("certID", certID).Msg("Processing cert to add")
		log.Debug().Str("thumbprint", tp).Msg("Iterating over stores")
		for _, store := range r.Stores {
			log.Debug().Str("thumbprint", tp).Str(
				"store_id",
				store.StoreID,
			).Msg("Checking if cert is deployed to store")
			//TODO: this should all be handled by a receiver function on TrustStore
			//if _, ok := store.Thumbprints[tp]; ok {
			//	// Cert is already in the store do nothing
			//	log.Info().Str("thumbprint", tp).Str("store_id", store.ID).Msg("Cert is already deployed to store")
			//	row := []string{
			//		//todo: this should be a toCSV field on whatever object this is
			//		tp,
			//		certIDStr,
			//		certLookup.IssuedDN,
			//		certLookup.IssuerDN,
			//		store.ID,
			//		store.Type,
			//		store.Machine,
			//		store.Path,
			//		"false",
			//		"false",
			//		"true",
			//		getCurrentTime(""),
			//	}
			//	log.Trace().Str("thumbprint", tp).Strs("row", row).Msg("Appending data row")
			//	data = append(data, row)
			//	log.Trace().Str("thumbprint", tp).Strs("row", row).Msg("Writing data row to CSV")
			//	wErr := csvWriter.Write(row)
			//	if wErr != nil {
			//		log.Error().
			//			Err(wErr).
			//			Str("thumbprint", tp).
			//			Str("output_file", r.OutputFilePath).
			//			Strs("row", row).
			//			Msg("Error writing row to CSV")
			//	}
			//} else {
			//	// Cert is not deployed to this store and will need to be added
			//	log.Info().
			//		Str("thumbprint", tp).
			//		Str("store_id", store.ID).
			//		Msg("Cert is not deployed to store")
			//	row := []string{
			//		//todo: this should be a toCSV
			//		tp,
			//		certIDStr,
			//		certLookup.IssuedDN,
			//		certLookup.IssuerDN,
			//		store.ID,
			//		store.Type,
			//		store.Machine,
			//		store.Path,
			//		"true",
			//		"false",
			//		"false",
			//		getCurrentTime(""),
			//	}
			//	log.Trace().
			//		Str("thumbprint", tp).
			//		Strs("row", row).
			//		Msg("Appending data row")
			//	data = append(data, row)
			//	log.Debug().
			//		Str("thumbprint", tp).
			//		Strs("row", row).
			//		Msg("Writing data row to CSV")
			//	wErr := csvWriter.Write(row)
			//	if wErr != nil {
			//		log.Error().
			//			Err(wErr).
			//			Str("thumbprint", tp).
			//			Str("output_file", r.OutputFilePath).
			//			Strs("row", row).
			//			Msg("Error writing row to CSV")
			//	}
			//	log.Debug().
			//		Str("thumbprint", tp).
			//		Msg("Adding 'add' action to actions map")
			//	actions[tp] = append(
			//		actions[tp], ROTAction{
			//			Thumbprint: tp,
			//			CertID:     certID,
			//			StoreID:    store.ID,
			//			StoreType:  store.Type,
			//			StorePath:  store.Path,
			//			StoreAlias: "", //TODO get this value
			//			AddCert:    true,
			//			RemoveCert: false,
			//			Deployed:   false,
			//		},
			//	)
			//}
		}
	}
	return data, actions, errs
}

func (r *RootOfTrustManager) reconcileRoots() error {
	log.Debug().Msg(fmt.Sprintf(DebugFuncEnter, "reconcileRoots"))
	if len(r.actions) == 0 {
		log.Info().Msg("No actions to reconcile detected, root of trust stores are up-to-date.")
		return nil
	}
	log.Info().Msg("Reconciling root of trust stores")

	rFileName := fmt.Sprintf("%s_reconciled.csv", strings.Split(r.ReportFilePath, ".csv")[0])
	log.Debug().
		Str("report_file", r.ReportFilePath).
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
	log.Info().Str("report_file", r.ReportFilePath).Msg("Processing reconciliation actions")
	var errs []error
	for thumbprint, action := range r.actions {
		for _, a := range action {
			if a.AddCert {
				if !r.IsDryRun {
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
					).Msg(fmt.Sprintf(DebugFuncCall, "Client.AddCertificateToStores"))
					_, err := r.Client.AddCertificateToStores(&addReq)
					if err != nil {
						fmt.Printf(
							"ERROR adding cert %s(%d) to store %s: %s\n",
							a.Thumbprint,
							a.CertID,
							a.StoreID,
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
				if !r.IsDryRun {
					log.Info().Str("thumbprint", thumbprint).Str(
						"store_id",
						a.StoreID,
					).Msg("Attempting to remove cert from store")
					cStore := api.CertificateStore{
						CertificateStoreId: a.StoreID,
						Alias:              a.Thumbprint, //todo: support non-thumbprint aliases
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
					).Msg(fmt.Sprintf(DebugFuncCall, "Client.RemoveCertificateFromStores"))
					_, err := r.Client.RemoveCertificateFromStores(&removeReq)
					if err != nil {
						log.Error().Err(err).Str("thumbprint", thumbprint).Str(
							"store_id",
							a.StoreID,
						).Str("store_path", a.StorePath).Msg("unable to remove cert from store")
						fmt.Printf(
							"ERROR removing cert %s(%d) from store %s: %s\n",
							a.Thumbprint,
							a.CertID,
							a.StoreID,
							err,
						)
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

func (r *RootOfTrustManager) processStoresFile() error {
	log.Debug().Str("stores_file", r.StoresFilePath).Msg("Reading in stores file")
	csvFile, _ := os.Open(r.StoresFilePath)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	storeEntries, _ := reader.ReadAll()
	var stores = make(map[string]StoreCSVEntry)
	var lookupFailures []string
	var errs []error
	for i, row := range storeEntries {
		if len(row) == 0 {
			log.Warn().
				Str("stores_file", r.StoresFilePath).
				Int("row", i).Msg("Skipping empty row")
			continue
		} else if row[0] == "StoreID" || row[0] == "StoreId" || i == 0 {
			log.Trace().Strs("row", row).Msg("Skipping header row")
			continue // Skip header
		}

		log.Debug().Strs("row", row).
			Str("store_id", row[0]).
			Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertificateStoreByID"))
		apiResp, err := r.Client.GetCertificateStoreByID(row[0])
		if err != nil {
			errs = append(errs, err)
			log.Error().Err(err).Str("store_id", row[0]).Msg("failed to retrieve store from Keyfactor Command")
			lookupFailures = append(lookupFailures, row[0])
			continue
		}

		log.Debug().Str("store_id", row[0]).Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertStoreInventoryV1"))
		inventory, invErr := r.Client.GetCertStoreInventory(row[0])
		if invErr != nil {
			errs = append(errs, invErr)
			log.Error().Err(invErr).Str(
				"store_id",
				row[0],
			).Msg("failed to retrieve inventory for certificate store from Keyfactor Command")
			continue
		}

		if !isRootStore(
			apiResp,
			inventory,
			r.TrustStoreCriteria.MinCerts,
			r.TrustStoreCriteria.MaxLeaves,
			r.TrustStoreCriteria.MaxKeys,
		) {
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
			Aliases:     make(map[string]bool),
		}

		log.Debug().Str("store_id", row[0]).Msg(
			"Iterating over inventory for thumbprints, " +
				"serial numbers and cert IDs",
		)

		if _, exists := r.Stores[row[0]]; !exists {
			log.Trace().Str("store_id", row[0]).Msg("Store not found in TrustStore map")
			return fmt.Errorf("unable to process stores file, store '%s' not found in TrustStore map", row[0])
		}

		// Directly modify the Inventory field without needing to reassign the struct
		r.Stores[row[0]].Inventory = *inventory
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
			Int("min_certs", r.TrustStoreCriteria.MinCerts).
			Int("max_leaves", r.TrustStoreCriteria.MaxLeaves).
			Int("max_keys", r.TrustStoreCriteria.MaxKeys).Send()
		errs = append(errs, errMsg)
	}

	if len(errs) > 0 {
		apiErrs := mergeErrsToString(&errs, false)
		return fmt.Errorf(apiErrs)
	}

	//r.stores = stores
	return nil
}

func (r *RootOfTrustManager) processAddCertsFile() (map[string]string, error) {
	var errs []error
	log.Info().Str("add_certs_file", r.AddCertsFilePath).Msg("Reading certs to add file")
	log.Debug().Str("add_certs_file", r.AddCertsFilePath).Msg(fmt.Sprintf(DebugFuncCall, "readCertsFile"))
	certsToAdd, rErr := readCertsFile(r.AddCertsFilePath)
	if rErr != nil {
		log.Error().Err(rErr).Str("add_certs_file", r.AddCertsFilePath).Msg("Error reading certs to add file")
		errs = append(errs, rErr)
	}

	if len(errs) > 0 {
		apiErrs := mergeErrsToString(&errs, false)
		return certsToAdd, fmt.Errorf(apiErrs)
	}

	log.Debug().Str("add_certs_file", r.AddCertsFilePath).Msg("finished reading certs to add file")
	return certsToAdd, nil
}

func (r *RootOfTrustManager) processRemoveCertsFile() (map[string]string, error) {
	var errs []error
	log.Info().Str("remove_certs_file", r.RemoveCertsFilePath).Msg("Reading certs to remove file")
	log.Debug().Str("remove_certs_file", r.RemoveCertsFilePath).Msg(fmt.Sprintf(DebugFuncCall, "readCertsFile"))
	certsToRemove, rErr := readCertsFile(r.RemoveCertsFilePath)
	if rErr != nil {
		log.Error().Err(rErr).Str("remove_certs_file", r.RemoveCertsFilePath).Msg("Error reading certs to remove file")
		errs = append(errs, rErr)
	}

	if len(errs) > 0 {
		apiErrs := mergeErrsToString(&errs, false)
		return certsToRemove, fmt.Errorf(apiErrs)
	}

	log.Debug().Str("remove_certs_file", r.RemoveCertsFilePath).Msg("finished reading certs to remove file")
	return certsToRemove, nil
}

func (r *RootOfTrustManager) processCertsFiles() error {
	var errs []error
	var certsToAdd = make(map[string]string)
	var rErr error
	if r.AddCertsFilePath == "" {
		log.Info().Msg("No add certs file specified, add operations will not be performed")
	} else {
		certsToAdd, rErr = r.processAddCertsFile()
		if rErr != nil {
			errs = append(errs, rErr)
		}
	}

	// Read in the remove removeCerts CSV
	var certsToRemove = make(map[string]string)
	if r.RemoveCertsFilePath == "" {
		log.Info().Msg("No remove certs file specified, remove operations will not be performed")
	} else {
		certsToRemove, rErr = r.processRemoveCertsFile()
		if rErr != nil {
			errs = append(errs, rErr)
		}
	}

	if len(errs) > 0 {
		apiErrs := mergeErrsToString(&errs, false)
		return fmt.Errorf(apiErrs)
	}

	if len(certsToAdd) == 0 && len(certsToRemove) == 0 {
		errMsg := "no ADD or REMOVE operations specified, please verify your configuration"
		e := fmt.Errorf(errMsg)
		log.Error().Err(e).Send()
		fmt.Println(errMsg)
		return e
	}

	r.addCerts = certsToAdd
	r.removeCerts = certsToRemove

	return nil
}

func (r *RootOfTrustManager) processAuditReportFile() error {
	log.Trace().
		Interface("certs_to_add", r.addCerts).
		Interface("certs_to_remove", r.removeCerts).
		Str("stores_file", r.StoresFilePath).
		Msg("Generating audit report")

	log.Debug().
		Msg(fmt.Sprintf(DebugFuncCall, "generateAuditReport"))
	err := r.generateAuditReport()
	log.Trace().Interface("data", r.data).Msg("Audit report data")
	if err != nil {
		log.Error().
			Err(err).
			Str("OutputFilePath", r.OutputFilePath).
			Msg("Error generating audit report")
		return err
	}
	if len(r.actions) == 0 {
		msg := "no reconciliation actions to take, the specified root of trust stores are up-to-date"
		log.Warn().
			Str("stores_file", r.StoresFilePath).
			Str("add_certs_file", r.AddCertsFilePath).
			Str("remove_certs_file", r.RemoveCertsFilePath).
			Msg(msg)
		fmt.Println(msg) //todo send to output formatter
	}
	return nil
}

func (r *RootOfTrustManager) processCSVReportFile() error {
	log.Debug().
		Str("report_file", r.ReportFilePath).
		Bool("dry_run", r.IsDryRun).
		Msg("Parsing existing audit report")
	// Read in the CSV

	log.Debug().
		Str("report_file", r.ReportFilePath).
		Msg("reading audit report file")

	csvFile, err := os.Open(r.ReportFilePath)
	if err != nil {
		log.Error().Err(err).Str("report_file", r.ReportFilePath).Msg("Error reading audit report file")
		return err
	}

	validHeader := false
	log.Trace().Str("report_file", r.ReportFilePath).Msg("Creating CSV reader")
	aCSV := csv.NewReader(csvFile)
	aCSV.FieldsPerRecord = -1
	log.Debug().Str("report_file", r.ReportFilePath).Msg("Reading CSV data")
	inFile, cErr := aCSV.ReadAll()
	if cErr != nil {
		log.Error().Err(cErr).Str("report_file", r.ReportFilePath).Msg("Error reading CSV file")
		return cErr
	}

	actions := make(map[string][]ROTAction)
	fieldMap := make(map[int]string)

	log.Debug().Str("report_file", r.ReportFilePath).
		Strs("csv_header", AuditHeader).
		Msg("Creating field map, index to header name")
	for i, field := range AuditHeader {
		log.Trace().Str("report_file", r.ReportFilePath).Str("field", field).Int(
			"index",
			i,
		).Msg("Processing field")
		fieldMap[i] = field
	}

	log.Debug().Str("report_file", r.ReportFilePath).Msg("Iterating over CSV rows")
	var errs []error
	for ri, row := range inFile {
		log.Trace().Str("report_file", r.ReportFilePath).Strs("row", row).Msg("Processing row")
		if strings.EqualFold(strings.Join(row, ","), strings.Join(AuditHeader, ",")) {
			log.Trace().Str("report_file", r.ReportFilePath).Strs("row", row).Msg("Skipping header row")
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
				r.ReportFilePath,
			).Msg("Invalid header in audit report file")
			return invalidHeaderErr
		}

		log.Debug().Str("report_file", r.ReportFilePath).Msg("Creating action map")
		action := make(map[string]interface{})
		for i, field := range row {
			log.Trace().Str("report_file", r.ReportFilePath).Str("field", field).Int(
				"index",
				i,
			).Msg("Processing field")
			fieldInt, iErr := strconv.Atoi(field)
			if iErr != nil {
				log.Trace().Err(iErr).Str("report_file", r.ReportFilePath).
					Str("field", field).
					Int("index", i).
					Msg("Field is not an integer, replacing with index value")
				action[fieldMap[i]] = field
			} else {
				log.Trace().Err(iErr).Str("report_file", r.ReportFilePath).
					Str("field", field).
					Int("index", i).
					Msg("Field is an integer")
				action[fieldMap[i]] = fieldInt
			}
		}

		log.Debug().Str("report_file", r.ReportFilePath).Msg("Processing add cert action")
		addCertStr, aOk := action["AddCert"].(string)
		if !aOk {
			log.Warn().Str("report_file", r.ReportFilePath).Msg(
				"AddCert field not found in action, " +
					"using empty string",
			)
			addCertStr = ""
		}

		log.Trace().Str("report_file", r.ReportFilePath).Str(
			"add_cert",
			addCertStr,
		).Msg("Converting addCertStr to bool")
		addCert, acErr := strconv.ParseBool(addCertStr)
		if acErr != nil {
			log.Warn().Str("report_file", r.ReportFilePath).Err(acErr).Msg(
				"Unable to parse bool from addCertStr, defaulting to FALSE",
			)
			addCert = false
		}

		log.Debug().Str("report_file", r.ReportFilePath).Msg("Processing remove cert action")
		removeCertStr, rOk := action["RemoveCert"].(string)
		if !rOk {
			log.Warn().Str("report_file", r.ReportFilePath).Msg(
				"RemoveCert field not found in action, " +
					"using empty string",
			)
			removeCertStr = ""
		}
		log.Trace().Str("report_file", r.ReportFilePath).Str(
			"remove_cert",
			removeCertStr,
		).Msg("Converting removeCertStr to bool")
		removeCert, rcErr := strconv.ParseBool(removeCertStr)
		if rcErr != nil {
			log.Warn().
				Str("report_file", r.ReportFilePath).
				Err(rcErr).
				Msg("Unable to parse bool from removeCertStr, defaulting to FALSE")
			removeCert = false
		}

		log.Trace().Str("report_file", r.ReportFilePath).Msg("Processing store type")
		sType, sOk := action["StoreType"].(string)
		if !sOk {
			log.Warn().Str("report_file", r.ReportFilePath).Msg(
				"StoreType field not found in action, " +
					"using empty string",
			)
			sType = ""
		}

		log.Trace().Str("report_file", r.ReportFilePath).Msg("Processing store path")
		sPath, pOk := action["Path"].(string)
		if !pOk {
			log.Warn().Str("report_file", r.ReportFilePath).Msg(
				"Path field not found in action, " +
					"using empty string",
			)
			sPath = ""
		}

		log.Trace().Str("report_file", r.ReportFilePath).Msg("Processing thumbprint")
		tp, tpOk := action["Thumbprint"].(string)
		if !tpOk {
			log.Warn().Str("report_file", r.ReportFilePath).Msg(
				"Thumbprint field not found in action, " +
					"using empty string",
			)
			tp = ""
		}

		log.Trace().Str("report_file", r.ReportFilePath).Msg("Processing cert id")
		cid, cidOk := action["CertID"].(int)
		if !cidOk {
			log.Warn().Str("report_file", r.ReportFilePath).Msg(
				"CertID field not found in action, " +
					"using -1",
			)
			cid = -1
		}

		if !tpOk && !cidOk {
			errMsg := fmt.Errorf("row is missing Thumbprint or CertID")
			log.Error().Err(errMsg).
				Str("report_file", r.ReportFilePath).
				Int("row", ri).
				Msg("Invalid row in audit report file")
			errs = append(errs, errMsg)
			continue
		}

		sId, sIdOk := action["StoreID"].(string)
		if !sIdOk {
			errMsg := fmt.Errorf("row is missing StoreID")
			log.Error().Err(errMsg).
				Str("report_file", r.ReportFilePath).
				Int("row", ri).
				Msg("Invalid row in audit report file")
			errs = append(errs, errMsg)
			continue
		}
		if cid == -1 && tp != "" {
			log.Debug().Str("report_file", r.ReportFilePath).
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
			log.Debug().Str("report_file", r.ReportFilePath).
				Int("row", ri).
				Str("thumbprint", tp).
				Msg(fmt.Sprintf(DebugFuncCall, "Client.GetCertificateContext"))

			certLookup, err := r.Client.GetCertificateContext(&certLookupReq)
			if err != nil {
				log.Error().Err(err).Str("report_file", r.ReportFilePath).
					Int("row", ri).
					Str("thumbprint", tp).
					Msg("Error looking up certificate by thumbprint")
				continue
			}
			cid = certLookup.Id
			log.Debug().Str("report_file", r.ReportFilePath).
				Int("row", ri).
				Str("thumbprint", tp).
				Int("cert_id", cid).
				Msg("Certificate found by thumbprint")
		}

		log.Trace().Str("report_file", r.ReportFilePath).
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

		log.Trace().Str("report_file", r.ReportFilePath).
			Int("row", ri).Interface("action", a).Msg("Adding action to actions map")
		actions[a.Thumbprint] = append(actions[a.Thumbprint], a)
	}

	log.Info().Str("report_file", r.ReportFilePath).Msg("Audit report parsed successfully")
	if len(actions) == 0 {
		rtMsg := "No reconciliation actions to take, root stores are up-to-date. Exiting."
		log.Info().Str("report_file", r.ReportFilePath).
			Msg(rtMsg)
		fmt.Println(rtMsg)
		if len(errs) > 0 {
			errStr := mergeErrsToString(&errs, false)
			log.Error().Str("report_file", r.ReportFilePath).
				Str("errors", errStr).
				Msg("Errors encountered while parsing audit report")
			return fmt.Errorf("errors encountered while parsing audit report: %s", errStr)
		}
		return nil
	}

	r.actions = actions
	log.Debug().Str("report_file", r.ReportFilePath).Msg(fmt.Sprintf(DebugFuncCall, "reconcileRoots"))
	rErr := r.reconcileRoots()
	if rErr != nil {
		log.Error().Err(rErr).Str("report_file", r.ReportFilePath).Msg("Error reconciling roots")
		return rErr
	}
	defer csvFile.Close()

	orchsURL := fmt.Sprintf(
		"https://%s/Keyfactor/Portal/AgentJobStatus/Index",
		r.Client.Hostname,
	) //todo: this pathing might not work for everyone

	if len(errs) > 0 {
		errStr := mergeErrsToString(&errs, false)
		log.Error().Str("report_file", r.ReportFilePath).
			Str("errors", errStr).
			Msg("Errors encountered while reconciling root of trust stores")
		return fmt.Errorf("errors encountered while reconciling roots:\r\n%s", errStr)

	}

	log.Info().Str("report_file", r.ReportFilePath).
		Str("orchs_url", orchsURL).
		Msg("Reconciliation completed. Check orchestrator jobs for details")
	fmt.Println(fmt.Sprintf("Reconciliation completed. Check orchestrator jobs for details. %s", orchsURL))

	return nil
}
