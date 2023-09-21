package cmd

import (
	"fmt"
	"strings"
)

type AuthProvider struct {
	Type       string      `json:"type"`
	Profile    string      `json:"profile"`
	Parameters interface{} `json:"parameters"`
}

func (a AuthProvider) String() string {
	return fmt.Sprintf("Type: %s, Profile: %s, Parameters: %s", a.Type, a.Profile, a.Parameters)
}

type AuthProviderAzureIDParams struct {
	SecretName     string `json:"secret_name"`
	AzureVaultName string `json:"vault_name"`
	TenantID       string `json:"tenant_id;omitempty"`
	SubscriptionID string `json:"subscription_id;omitempty"`
	ResourceGroup  string `json:"resource_group;omitempty"`
}

func (apaz AuthProviderAzureIDParams) String() string {
	return fmt.Sprintf("SecretName: %s, AzureVaultName: %s, TenantId: %s, SubscriptionId: %s, ResourceGroup: %s", apaz.SecretName, apaz.AzureVaultName, apaz.TenantID, apaz.SubscriptionID, apaz.ResourceGroup)
}

type ConfigurationFile struct {
	Servers map[string]ConfigurationFileEntry `json:"servers"`
}

func (c ConfigurationFile) String() string {
	var servers string
	for name, server := range c.Servers {
		servers += fmt.Sprintf("%s: %s\n", name, server)
	}
	// remove trailing newline
	if len(servers) > 0 {
		servers = strings.TrimRight(servers, "\n")
	}
	return fmt.Sprintf("Servers: %s", servers)
}

type ConfigurationFileEntry struct {
	Hostname     string       `json:"host"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	Domain       string       `json:"domain"`
	APIPath      string       `json:"api_path"`
	AuthProvider AuthProvider `json:"auth_provider"`
}

func (c ConfigurationFileEntry) String() string {
	if !logInsecure {
		return fmt.Sprintf("\n\tHostname: %s,\n\tUsername: %s,\n\tPassword: %s,\n\tDomain: %s,\n\tAPIPath: %s,\n\tAuthProvider: %s", c.Hostname, c.Username, hashSecretValue(c.Password), c.Domain, c.APIPath, c.AuthProvider)
	}
	return fmt.Sprintf("\n\tHostname: %s,\n\tUsername: %s,\n\tPassword: %s,\n\tDomain: %s,\n\tAPIPath: %s,\n\tAuthProvider: %s", c.Hostname, c.Username, c.Password, c.Domain, c.APIPath, c.AuthProvider)
}

type NewStoreCSVEntry struct {
	Id                string `json:"Id"`
	CertStoreType     string `json:"CertStoreType"`
	ClientMachine     string `json:"ClientMachine"`
	Storepath         string `json:"StorePath"`
	Properties        string `json:"Properties"`
	Approved          bool   `json:"Approved"`
	CreateIfMissing   bool   `json:"CreateIfMissing"`
	AgentID           string `json:"AgentId"`
	InventorySchedule string `json:"InventorySchedule"`
}

func (n NewStoreCSVEntry) String() string {
	return fmt.Sprintf("Id: %s, CertStoreType: %s, ClientMachine: %s, Storepath: %s, Properties: %s, Approved: %t, CreateIfMissing: %t, AgentId: %s, InventorySchedule: %s", n.Id, n.CertStoreType, n.ClientMachine, n.Storepath, n.Properties, n.Approved, n.CreateIfMissing, n.AgentID, n.InventorySchedule)
}
