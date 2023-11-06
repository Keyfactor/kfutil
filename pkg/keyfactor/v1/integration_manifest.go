/*
Copyright 2023 The Keyfactor Command Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifestv1

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"os"
)

const (
	jsonIndent = "  "
	imSchema   = "https://keyfactor.github.io/integration-manifest-schema.json"
)

// IntegrationManifest is the structure of the integration manifest file generated from
// https://keyfactor.github.io/integration-manifest-schema.json
type IntegrationManifest struct {
	Schema          string           `json:"$schema"`
	Name            string           `json:"name"`
	IntegrationType string           `json:"integration_type"`
	Status          string           `json:"status"`
	Description     string           `json:"description"`
	SupportLevel    string           `json:"support_level"`
	ReleaseDir      string           `json:"release_dir"`
	LinkGithub      bool             `json:"link_github"`
	UpdateCatalog   bool             `json:"update_catalog"`
	WinSupport      string           `json:"winSupport,omitempty"`
	LinSupport      string           `json:"linSupport,omitempty"`
	PamRegDLL       string           `json:"pamRegDLL,omitempty"`
	About           IntegrationAbout `json:"about,omitempty"`
}

type IntegrationAbout struct {
	Orchestrator OrchestratorDetails `json:"orchestrator,omitempty"`
	Pam          PamDetails          `json:"pam,omitempty"`
}

type OrchestratorDetails struct {
	UOFramework string       `json:"UOFramework,omitempty"`
	PamSupport  bool         `json:"pam_support"`
	Win         WinSupport   `json:"win,omitempty"`
	Linux       LinuxSupport `json:"linux,omitempty"`
	StoreTypes  []StoreType  `json:"store_types,omitempty"`
}

type WinSupport struct {
	SupportsManagementAdd    bool `json:"supportsManagementAdd"`
	SupportsManagementRemove bool `json:"supportsManagementRemove"`
	SupportsCreateStore      bool `json:"supportsCreateStore"`
	SupportsDiscovery        bool `json:"supportsDiscovery"`
	SupportsReenrollment     bool `json:"supportsReenrollment"`
	SupportsInventory        bool `json:"supportsInventory"`
}

type LinuxSupport struct {
	SupportsManagementAdd    bool `json:"supportsManagementAdd"`
	SupportsManagementRemove bool `json:"supportsManagementRemove"`
	SupportsCreateStore      bool `json:"supportsCreateStore"`
	SupportsDiscovery        bool `json:"supportsDiscovery"`
	SupportsReenrollment     bool `json:"supportsReenrollment"`
	SupportsInventory        bool `json:"supportsInventory"`
}

type StoreType struct {
	Name                string           `json:"Name,omitempty"`
	ShortName           string           `json:"ShortName,omitempty"`
	Capability          string           `json:"Capability,omitempty"`
	LocalStore          bool             `json:"LocalStore"`
	SupportedOperations OperationSupport `json:"SupportedOperations,omitempty"`
	Properties          []PropertyDetail `json:"Properties,omitempty"`
	EntryParameters     []EntryParameter `json:"EntryParameters,omitempty"`
	PasswordOptions     interface{}      `json:"PasswordOptions,omitempty"`
	StorePathType       string           `json:"StorePathType,omitempty"`
	StorePathValue      string           `json:"StorePathValue,omitempty"`
	PrivateKeyAllowed   string           `json:"PrivateKeyAllowed,omitempty"`
	JobProperties       []string         `json:"JobProperties,omitempty"`
	ServerRequired      bool             `json:"ServerRequired"`
	PowerShell          bool             `json:"PowerShell"`
	BlueprintAllowed    bool             `json:"BlueprintAllowed"`
	CustomAliasAllowed  string           `json:"CustomAliasAllowed,omitempty"`
}

type OperationSupport struct {
	Add        bool `json:"Add"`
	Remove     bool `json:"Remove"`
	Enrollment bool `json:"Enrollment"`
	Discovery  bool `json:"Discovery"`
	Inventory  bool `json:"Inventory"`
}

type PropertyDetail struct {
	Name         string `json:"Name,omitempty"`
	DisplayName  string `json:"DisplayName,omitempty"`
	Type         string `json:"Type,omitempty"`
	DependsOn    string `json:"DependsOn,omitempty"`
	DefaultValue string `json:"DefaultValue,omitempty"`
	Required     bool   `json:"Required"`
}

type EntryParameter struct {
	Name         string      `json:"Name,omitempty"`
	DisplayName  string      `json:"DisplayName,omitempty"`
	Type         string      `json:"Type,omitempty"`
	RequiredWhen Requirement `json:"RequiredWhen,omitempty"`
	DefaultValue string      `json:"DefaultValue,omitempty"`
	DependsOn    string      `json:"DependsOn,omitempty"`
	Options      string      `json:"Options,omitempty"`
}

type Requirement struct {
	HasPrivateKey  bool `json:"HasPrivateKey"`
	OnAdd          bool `json:"OnAdd"`
	OnRemove       bool `json:"OnRemove"`
	OnReenrollment bool `json:"OnReenrollment"`
}

type PamDetails struct {
	AssemblyName            string `json:"assemblyName,omitempty"`
	ProviderName            string `json:"providerName,omitempty"`
	FullyQualifiedClassName string `json:"fullyQualifiedClassName,omitempty"`
	DbName                  string `json:"dbName,omitempty"`
}

func NewIntegrationManifest() *IntegrationManifest {
	return &IntegrationManifest{
		Schema: imSchema,
	}
}

// LoadFromFilesystem loads an integration manifest from the current working directory.
// If the file does not exist, it returns nil
func (im *IntegrationManifest) LoadFromFilesystem() error {
	// Load the integration manifest from the current working directory
	manifestFileBytes, err := os.Open("integration-manifest.json")
	if err != nil {
		log.Debug().Msg("Could not open integration-manifest.json - it may not exist")
		return nil
	}

	// Marshal the integration manifest into the IntegrationManifest struct
	err = json.NewDecoder(manifestFileBytes).Decode(&im)
	if err != nil {
		return err
	}

	return nil
}

// SaveToFilesystem saves an integration manifest to the current working directory
func (im *IntegrationManifest) SaveToFilesystem() error {
	imJsonString, err := im.Marshal()
	if err != nil {
		return err
	}

	// Overwrite or create the integration manifest file
	manifestFile, err := os.OpenFile("integration-manifest.json", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	// Write the integration manifest to the file
	_, err = manifestFile.WriteString(imJsonString)
	if err != nil {
		return err
	}

	return nil
}

func (im *IntegrationManifest) CopyIntoStoreType(source string) error {
	// Marshal the source string into a StoreType
	var storeType StoreType
	err := json.Unmarshal([]byte(source), &storeType)
	if err != nil {
		return err
	}

	// Either overwrite an existing StoreType or append a new one
	for i, st := range im.About.Orchestrator.StoreTypes {
		if st.Name == storeType.Name {
			im.About.Orchestrator.StoreTypes[i] = storeType
			return nil
		}
	}
	im.About.Orchestrator.StoreTypes = append(im.About.Orchestrator.StoreTypes, storeType)

	return nil
}

func (im *IntegrationManifest) Marshal() (string, error) {
	// Marshal the IntegrationManifest struct into a string
	manifestBytes, err := json.MarshalIndent(im, "", jsonIndent)
	if err != nil {
		log.Debug().Msg("Could not marshal integration manifest")
		return "", err
	}

	return string(manifestBytes), nil
}
