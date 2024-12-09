package cmd

import (
	"github.com/Keyfactor/keyfactor-go-client/v3/api"
)

type IntegrationManifest struct {
	Schema          string `json:"$schema"`
	IntegrationType string `json:"integration_type"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	LinkGithub      bool   `json:"link_github"`
	UpdateCatalog   bool   `json:"update_catalog"`
	SupportLevel    string `json:"support_level"`
	ReleaseDir      string `json:"release_dir"`
	ReleaseProject  string `json:"release_project"`
	Description     string `json:"description"`
	About           About  `json:"about"`
}

type About struct {
	Orchestrator Orchestrator `json:"orchestrator"`
}

type Orchestrator struct {
	UOFramework              string                     `json:"UOFramework"`
	PAMSupport               bool                       `json:"pam_support"`
	KeyfactorPlatformVersion string                     `json:"keyfactor_platform_version"`
	StoreTypes               []api.CertificateStoreType `json:"store_types"`
}
