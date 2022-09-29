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
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
)

type StoreCSVEntry struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Machine string `json:"address"`
	Path    string `json:"path"`
}

type RotCert struct {
	//Id         string `json:"id"`
	ThumbPrint string `json:"thumbprint"`
}

// rotCmd represents the rot command
var rotCmd = &cobra.Command{
	Use:   "rot",
	Short: "Root Of Trust",
	Long:  `Root Of Trust: Will parse a CSV and attempt to enroll a cert or set of certs into a list of cert stores.`,
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
			apiResp, err := kfClient.GetCertificateStoreByID(entry[0])
			if err != nil {
				//log.Fatalf("Error getting cert store: %s", err)
				log.Printf("[ERROR] Error getting cert store: %s", err)
				lookupFailures = append(lookupFailures, strings.Join(entry, ","))
				continue
			}
			stores[entry[0]] = StoreCSVEntry{
				Id:      entry[0],
				Type:    entry[1],
				Machine: entry[2],
				Path:    entry[3],
			}
			log.Printf("[DEBUG] Store: %s", apiResp)
		}
		storesJson, _ := json.Marshal(stores)
		fmt.Println(string(storesJson))

		// Read in the add addCerts CSV
		var addCerts = make(map[string]RotCert)
		if addRootsFile != "" {
			addCerts, err := readCertsFile(addRootsFile)
			if err != nil {
				log.Fatalf("Error reading addCerts file: %s", err)
			}
			addCertsJson, _ := json.Marshal(addCerts)
			fmt.Printf("[DEBUG] add certs JSON: %s", string(addCertsJson))
			fmt.Println("add rot called")
		} else {
			log.Printf("[DEBUG] No addCerts file specified")
			log.Printf("[DEBUG] No addCerts = %s", addCerts)
		}

		// Read in the remove removeCerts CSV
		var removeCerts = make(map[string]RotCert)
		if removeRootsFile != "" {
			removeCerts, err := readCertsFile(removeRootsFile)
			if err != nil {
				log.Fatalf("Error reading removeCerts file: %s", err)
			}
			removeCertsJson, _ := json.Marshal(removeCerts)
			fmt.Println(string(removeCertsJson))
			fmt.Println("remove rot called")
		} else {
			log.Printf("[DEBUG] No removeCerts file specified")
			log.Printf("[DEBUG] No removeCerts = %s", removeCerts)
		}
	},
}

func readCertsFile(certsFilePath string) (map[string]RotCert, error) {
	// Read in the cert CSV
	csvFile, _ := os.Open(certsFilePath)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	certEntries, _ := reader.ReadAll()
	var certs = make(map[string]RotCert)
	for _, entry := range certEntries {
		switch entry[0] {
		case "CertId", "thumbprint", "id", "certId", "Thumbprint":
			continue // Skip header
		}

		certs[entry[0]] = RotCert{
			ThumbPrint: entry[0],
		}
		// Get certificate context
		//args := &api.GetCertificateContextArgs{
		//	IncludeMetadata:  boolToPointer(true),
		//	IncludeLocations: boolToPointer(true),
		//	CollectionId:     nil,
		//	Id:               certificateIdInt,
		//}
		//cResp, err := r.p.client.GetCertificateContext(args)
	}
	return certs, nil
}

var rotGenStoreTemplateCmd = &cobra.Command{
	Use:   "generate-template-rot",
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

func isRootStore(client *api.Client) bool {
	//client.GetCertInventory()
	return true
}

func initClient() (*api.Client, error) {
	var clientAuth api.AuthConfig
	clientAuth.Username = os.Getenv("KEYFACTOR_USERNAME")
	log.Printf("[DEBUG] Username: %s", clientAuth.Username)
	clientAuth.Password = os.Getenv("KEYFACTOR_PASSWORD")
	log.Printf("[DEBUG] Password: %s", clientAuth.Password)
	clientAuth.Domain = os.Getenv("KEYFACTOR_DOMAIN")
	log.Printf("[DEBUG] Domain: %s", clientAuth.Domain)
	clientAuth.Hostname = os.Getenv("KEYFACTOR_HOSTNAME")
	log.Printf("[DEBUG] Hostname: %s", clientAuth.Hostname)

	c, err := api.NewKeyfactorClient(&clientAuth)

	if err != nil {
		log.Fatalf("Error creating Keyfactor client: %s", err)
	}
	return c, err
}

func init() {
	storesCmd.AddCommand(rotCmd)
	var stores string
	var certs string
	rotCmd.Flags().StringVarP(&stores, "stores", "s", "", "CSV file containing cert stores to enroll into")
	rotCmd.MarkFlagRequired("stores")
	rotCmd.Flags().StringVarP(&certs, "add-certs", "a", "", "CSV file containing cert(s) to enroll into the defined cert stores")
	rotCmd.Flags().StringVarP(&certs, "remove-certs", "r", "", "CSV file containing cert(s) to remove from the defined cert stores")

	rotCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rotCmd.MarkFlagRequired("certs")
	storesCmd.AddCommand(rotGenStoreTemplateCmd)
	rotGenStoreTemplateCmd.Flags().String("outpath", "template.csv", "Output file to write the template to")
	rotGenStoreTemplateCmd.Flags().String("format", "csv", "The type of template to generate. Only `csv` is supported at this time.")
	rotGenStoreTemplateCmd.Flags().String("type", "stores", "The type of template to generate. Only `certs|stores` are supported at this time.")
	//rotGenStoreTemplateCmd.MarkFlagRequired("type")
	//rotGenStoreTemplateCmd.MarkFlagRequired("format")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
