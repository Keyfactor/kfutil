// Copyright 2025 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

type IntegrationManifestV2 struct {
	Schema          string  `json:"$schema"`
	IntegrationType string  `json:"integration_type"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	LinkGithub      bool    `json:"link_github"`
	UpdateCatalog   bool    `json:"update_catalog"`
	SupportLevel    string  `json:"support_level"`
	ReleaseDir      string  `json:"release_dir"`
	ReleaseProject  string  `json:"release_project"`
	Description     string  `json:"description"`
	About           AboutV2 `json:"about"`
}

type IntegrationManifestV3 struct {
	Schema          string  `json:"$schema"`
	IntegrationType string  `json:"integration_type"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	LinkGithub      bool    `json:"link_github"`
	UpdateCatalog   bool    `json:"update_catalog"`
	SupportLevel    string  `json:"support_level"`
	ReleaseDir      string  `json:"release_dir"`
	ReleaseProject  string  `json:"release_project"`
	Description     string  `json:"description"`
	About           AboutV3 `json:"about"`
}

type AboutV2 struct {
	Orchestrator Orchestrator `json:"orchestrator,omitempty"`
	PAM          PAMV2        `json:"pam,omitempty"`
}

type AboutV3 struct {
	Orchestrator Orchestrator `json:"orchestrator,omitempty"`
	PAM          PAMV3        `json:"pam,omitempty"`
}

type Orchestrator struct {
	UOFramework              string                     `json:"UOFramework"`
	PAMSupport               bool                       `json:"pam_support"`
	KeyfactorPlatformVersion string                     `json:"keyfactor_platform_version"`
	StoreTypes               []api.CertificateStoreType `json:"store_types"`
}

type PAMV2 struct {
	Name                    string                                   `json:"providerName"`
	AssemblyName            string                                   `json:"assemblyName"`
	DBName                  string                                   `json:"dbName"`
	FullyQualifiedClassName string                                   `json:"fullyQualifiedClassName"`
	PAMTypes                map[string]api.ProviderTypeCreateRequest `json:"pam_types"`
}

type PAMV3 struct {
	Name                    string                          `json:"providerName"`
	AssemblyName            string                          `json:"assemblyName"`
	DBName                  string                          `json:"dbName"`
	FullyQualifiedClassName string                          `json:"fullyQualifiedClassName"`
	PAMTypes                []api.ProviderTypeCreateRequest `json:"pam_types"`
}
