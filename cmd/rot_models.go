package cmd

import (
	"fmt"
	"strings"

	"github.com/Keyfactor/keyfactor-go-client/v2/api"
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
	CertHeader = []string{"CertID", "Thumbprint", "SubjectName", "Issuer", "Alias", "Locations", "LastQueriedDate"}
)

type ROTStore struct {
	StoreID       string `json:"StoreID,omitempty"`
	StoreType     string `json:"StoreType,omitempty"`
	StoreMachine  string `json:"StoreMachine,omitempty"`
	StorePath     string `json:"StorePath,omitempty"`
	ContainerId   string `json:"ContainerId,omitempty"`
	ContainerName string `json:"ContainerName,omitempty"`
	LastQueried   string `json:"LastQueried,omitempty"`
}

func (r *ROTStore) toCSV() string {
	return fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s,%s",
		r.StoreID,
		r.StoreType,
		r.StoreMachine,
		r.StorePath,
		r.ContainerId,
		r.ContainerName,
		r.LastQueried,
	)
}

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
	ThumbPrint string                     `json:"thumbprint"`
	ID         int                        `json:"id"`
	CN         string                     `json:"cn"`
	SANs       []string                   `json:"sans"`
	Alias      string                     `json:"alias"`
	Locations  []api.CertificateLocations `json:"locations"`
	Issuer     string                     `json:"issuer"`
}

func (r *ROTCert) toCSV() string {
	subjectName := strings.Join(r.SANs, ";")
	// check if CN is not in subject_name
	if !strings.Contains(subjectName, r.CN) {
		if subjectName != "" {
			subjectName = fmt.Sprintf("%s;%s", r.CN, subjectName)
		} else {
			subjectName = r.CN
		}
	}

	return fmt.Sprintf(
		"%d,%s,%s,%s,%s,%v,%s",
		r.ID,
		r.ThumbPrint,
		subjectName,
		r.Issuer,
		r.Alias,
		//strings.Join(r.SANs, ";"),
		r.Locations,
		getCurrentTime(""), // LastQueriedDate
	)
}

type ROTAction struct {
	Thumbprint string `json:"thumbprint" mapstructure:"Thumbprint"`
	StoreAlias string `json:"alias" mapstructure:"Alias;omitempty"`
	CertID     int    `json:"cert_id" mapstructure:"CertID"`
	CertDN     string `json:"cert_dn" mapstructure:"SubjectName"`
	CertSANs   string `json:"cert_sans,omitempty" mapstructure:"CertSANs,omitempty"`
	Issuer     string `json:"issuer" mapstructure:"Issuer"`
	StoreID    string `json:"store_id" mapstructure:"StoreID"`
	StoreType  string `json:"store_type" mapstructure:"StoreType"`
	Machine    string `json:"client_machine" mapstructure:"Machine"`
	StorePath  string `json:"store_path" mapstructure:"Path"`
	Alias      string `json:"alias" mapstructure:"Alias,omitempty"`
	AddCert    bool   `json:"add" mapstructure:"AddCert"`
	RemoveCert bool   `json:"remove"  mapstructure:"RemoveCert"`
	Deployed   bool   `json:"deployed" mapstructure:"Deployed"`
	AuditDate  string `json:"audit_date" mapstructure:"AuditDate"`
}

func (r *ROTAction) getAuditHeaderMap() map[string]int {
	headerMap := make(map[string]int)
	for i, h := range AuditHeader {
		headerMap[h] = i
	}
	return headerMap
}

func (r *ROTAction) getAuditCSVHeader() []string {
	//return []string{
	//	"Thumbprint",
	//	"CertID",
	//	"SubjectName",
	//	"Issuer",
	//	"StoreID",
	//	"StoreType",
	//	"Machine",
	//	"Path",
	//	"AddCert",
	//	"RemoveCert",
	//	"Deployed",
	//	"AuditDate",
	//}
	return AuditHeader
}

func (r *ROTAction) getReconciledCSVHeader() []string {
	//return []string{
	//	"Thumbprint",
	//	"CertID",
	//	"SubjectName",
	//	"Issuer",
	//	"StoreID",
	//	"StoreType",
	//	"Machine",
	//	"Path",
	//	"AddCert",
	//	"RemoveCert",
	//	"Deployed",
	//	"AuditDate",
	//}
	return ReconciledAuditHeader
}

func (r *ROTAction) toCSV(rowType string) string {

	switch rowType {
	case "audit":
		headerMap := r.getAuditHeaderMap()

		//create csv row with fields arranged in order of the header map
		row := make([]string, len(AuditHeader))
		row[headerMap["Thumbprint"]] = r.Thumbprint
		row[headerMap["CertID"]] = fmt.Sprintf("%d", r.CertID)
		row[headerMap["SubjectName"]] = r.CertDN
		row[headerMap["Issuer"]] = r.Issuer
		row[headerMap["StoreID"]] = r.StoreID
		row[headerMap["StoreType"]] = r.StoreType
		row[headerMap["Machine"]] = ""
		row[headerMap["Path"]] = r.StorePath
		row[headerMap["AddCert"]] = fmt.Sprintf("%t", r.AddCert)
		row[headerMap["RemoveCert"]] = fmt.Sprintf("%t", r.RemoveCert)
		row[headerMap["Deployed"]] = ""
		row[headerMap["AuditDate"]] = getCurrentTime("")

		return strings.Join(row, ",")
	}
	return "invalid format"
}
