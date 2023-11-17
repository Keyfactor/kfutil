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
	"os"
	"reflect"
	"testing"
)

func TestIntegrationManifest_CopyInto(t *testing.T) {
	// newStoreJsonString intentionally has values that don't exist in StoreType.
	newStoreJsonString := `
{
  "Name": "CloudDBCert",
  "ShortName": "CloudDBCert",
  "Capability": "CertificateManagement",
  "LocalStore": true,
  "SupportedOperations": {
    "Add": true,
    "Create": true,
    "Discovery": false,
    "Enrollment": true,
    "Remove": true
  },
  "Properties": [
    {
      "Name": "CloudRegion",
      "DisplayName": "Cloud Region",
      "Type": "String",
      "DependsOn": null,
      "DefaultValue": "us-west-1",
      "Required": true
    },
    {
      "Name": "ResourceGroupName",
      "DisplayName": "Resource Group Name",
      "Type": "String",
      "DependsOn": null,
      "DefaultValue": null,
      "Required": true
    },
    {
      "Name": "ResourceType",
      "DisplayName": "Resource Type",
      "Type": "String",
      "DependsOn": null,
      "DefaultValue": "SQLDatabase",
      "Required": true
    },
    {
      "Name": "AdminUsername",
      "DisplayName": "Admin Username",
      "Type": "String",
      "DependsOn": null,
      "DefaultValue": null,
      "Required": true
    },
    {
      "Name": "AdminPassword",
      "DisplayName": "Admin Password",
      "Type": "Secret",
      "DependsOn": null,
      "DefaultValue": null,
      "Required": true
    },
    {
      "Name": "UseEncryption",
      "DisplayName": "Use Encryption",
      "Type": "Bool",
      "DependsOn": null,
      "DefaultValue": "false",
      "Required": false
    }
  ],
  "EntryParameters": [
    {
      "Name": "Environment",
      "DisplayName": "Environment",
      "Type": "Dropdown",
      "Values": ["Development", "Staging", "Production"],
      "DefaultValue": "Development",
      "Required": true
    }
  ],
  "PasswordOptions": {
    "EntrySupported": true,
    "StoreRequired": true,
    "Style": "PasswordComplex"
  },
  "StorePathType": "CloudService",
  "StorePathValue": "/subscriptions/subid/resourceGroups/rgname/providers/Microsoft.DBforMySQL/servers/servername",
  "PrivateKeyAllowed": "Optional",
  "JobProperties": ["Backup", "Rotation"],
  "ServerRequired": false,
  "PowerShell": true,
  "BlueprintAllowed": true,
  "CustomAliasAllowed": "Allowed",
  "ServerRegistration": 2,
  "InventoryEndpoint": "/CloudInventory/Update",
  "InventoryJobType": "ac4b1d4a-46e0-4f13-95d6-1e57b314d2b1",
  "ManagementJobType": "b1a5f93b-e480-4f8a-9850-e3b125fba5f9",
  "DiscoveryJobType": "e3d1c3b9-6e92-4893-a5d5-67d5c7053a6b",
  "EnrollmentJobType": "1d2c5f59-5be4-4f5d-9ed6-d6b458aadf58"
}
`

	// Marshal newStoreJsonString into a StoreType struct for later
	var newStore StoreType
	err := json.Unmarshal([]byte(newStoreJsonString), &newStore)
	if err != nil {
		t.Fatalf("MarshalJSON returned an error: %v", err)
	}

	t.Run("FromBlankIM", func(t *testing.T) {
		// Initialize an IntegrationManifest struct
		im := IntegrationManifest{}

		// Use CopyInto to copy data from the newStoreJsonString into the IntegrationManifest struct
		err = im.CopyIntoStoreType(newStoreJsonString)
		if err != nil {
			t.Fatalf("CopyInto returned an error: %v", err)
		}

		// Compare the values in the IntegrationManifest struct to the values in the StoreType struct
		if !reflect.DeepEqual(im.About.Orchestrator.StoreTypes[0], newStore) {
			t.Errorf("CopyInto did not copy the values from the StoreType struct into the IntegrationManifest struct")
		}
	})

	t.Run("FromExistingIM", func(t *testing.T) {
		// Initialize an IntegrationManifest struct with test data
		im := getTestIntegrationManifest()

		// Use CopyInto to copy data from the newStoreJsonString into the IntegrationManifest struct
		err = im.CopyIntoStoreType(newStoreJsonString)
		if err != nil {
			t.Fatalf("CopyInto returned an error: %v", err)
		}

		// Compare the values in the IntegrationManifest struct to the values in the StoreType struct
		if !reflect.DeepEqual(im.About.Orchestrator.StoreTypes[1], newStore) {
			t.Errorf("CopyInto did not copy the values from the StoreType struct into the IntegrationManifest struct")
		}
	})

	// Testing to ensure that CopyInto will overwrite existing stores with the same name
	t.Run("FromExistingIMWithExistingStore", func(t *testing.T) {
		// Initialize an IntegrationManifest struct with test data
		im := getTestIntegrationManifest()

		// Change the name of the existing store to match the new store
		im.About.Orchestrator.StoreTypes[0].Name = newStore.Name

		// Use CopyInto to copy data from the newStoreJsonString into the IntegrationManifest struct
		err = im.CopyIntoStoreType(newStoreJsonString)
		if err != nil {
			t.Fatalf("CopyInto returned an error: %v", err)
		}

		// Compare the values in the IntegrationManifest struct to the values in the StoreType struct
		if !reflect.DeepEqual(im.About.Orchestrator.StoreTypes[0], newStore) {
			t.Errorf("CopyInto did not copy the values from the StoreType struct into the IntegrationManifest struct")
		}
	})
}

func TestIntegrationManifest_LoadFromFilesystem(t *testing.T) {
	// Marshal test IM to a string and write it to an IM
	originalIm := getTestIntegrationManifest()
	imString, err := json.Marshal(originalIm)
	if err != nil {
		t.Fatalf("MarshalJSON returned an error: %v", err)
	}

	err = os.WriteFile("integration-manifest.json", imString, 0644)
	if err != nil {
		t.Fatalf("WriteFile returned an error: %v", err)
	}

	// Initialize a blank IntegrationManifest struct
	im := IntegrationManifest{}

	// Load the integration manifest from the current working directory
	err = im.LoadFromFilesystem()
	if err != nil {
		t.Fatalf("LoadFromFilesystem returned an error: %v", err)
	}

	// Test if the old and new IntegrationManifest structs are deeply equal
	if !reflect.DeepEqual(im, originalIm) {
		t.Errorf("LoadFromFilesystem did not load the integration manifest correctly")
	}

	// Clean up the integration manifest file
	err = os.Remove("integration-manifest.json")
	if err != nil {
		t.Fatalf("Remove returned an error: %v", err)
	}
}

func TestIntegrationManifest_Marshal(t *testing.T) {
	// Marshal test IM to a string
	originalIm := getTestIntegrationManifest()
	imString, err := json.MarshalIndent(originalIm, "", jsonIndent)
	if err != nil {
		t.Fatalf("MarshalJSON returned an error: %v", err)
	}

	// Marshal the IntegrationManifest struct into a string
	im, err := originalIm.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned an error: %v", err)
	}

	// Test if the old and new IntegrationManifest structs are deeply equal
	if !reflect.DeepEqual(im, string(imString)) {
		t.Errorf("Marshal did not marshal the integration manifest correctly")
	}
}

func TestIntegrationManifest_SaveToFilesystem(t *testing.T) {
	im := getTestIntegrationManifest()
	err := im.SaveToFilesystem()
	if err != nil {
		t.Fatalf("SaveToFilesystem returned an error: %v", err)
	}

	// Load the integration manifest from the current working directory
	err = im.LoadFromFilesystem()
	if err != nil {
		t.Fatalf("LoadFromFilesystem returned an error: %v", err)
	}

	// Test if the old and new IntegrationManifest structs are deeply equal
	if !reflect.DeepEqual(im, getTestIntegrationManifest()) {
		t.Errorf("SaveToFilesystem did not save the integration manifest correctly")
	}

	// Clean up the integration manifest file
	err = os.Remove("integration-manifest.json")
	if err != nil {
		t.Errorf("Remove returned an error: %v", err)
	}
}

func getTestIntegrationManifest() IntegrationManifest {
	return IntegrationManifest{
		Schema:          imSchema,
		Name:            "ExampleIntegration",
		IntegrationType: "orchestrator",
		Status:          "production",
		Description:     "This is an example integration for demonstration purposes.",
		SupportLevel:    "community",
		ReleaseDir:      "/path/to/release/directory",
		LinkGithub:      true,
		UpdateCatalog:   true,
		WinSupport:      "Full",
		LinSupport:      "Partial",
		PamRegDLL:       "example.dll",
		About: IntegrationAbout{
			Orchestrator: OrchestratorDetails{
				UOFramework: "ExampleFramework",
				PamSupport:  true,
				Win: WinSupport{
					SupportsManagementAdd:    true,
					SupportsManagementRemove: true,
					SupportsCreateStore:      true,
					SupportsDiscovery:        true,
					SupportsReenrollment:     true,
					SupportsInventory:        true,
				},
				Linux: LinuxSupport{
					SupportsManagementAdd:    true,
					SupportsManagementRemove: true,
					SupportsCreateStore:      true,
					SupportsDiscovery:        true,
					SupportsReenrollment:     true,
					SupportsInventory:        true,
				},
				StoreTypes: []StoreType{
					{
						Name:       "ExampleStore",
						ShortName:  "ExStore",
						Capability: "ExStore",
						LocalStore: false,
						SupportedOperations: OperationSupport{
							Add:        true,
							Remove:     true,
							Enrollment: true,
							Discovery:  true,
							Inventory:  true,
						},
						Properties: []PropertyDetail{
							{
								Name:         "exampleProperty",
								DisplayName:  "Example Property",
								Type:         "String",
								DependsOn:    "previousProperty",
								DefaultValue: "defaultValue",
								Required:     true,
							},
						},
						EntryParameters: []EntryParameter{
							{
								Name:        "exampleParameter",
								DisplayName: "Example Parameter",
								Type:        "int",
								RequiredWhen: Requirement{
									HasPrivateKey:  true,
									OnAdd:          true,
									OnRemove:       false,
									OnReenrollment: true,
								},
								DefaultValue: "10",
								DependsOn:    "otherParameter",
								Options:      "10,20,30",
							},
						},
						PasswordOptions:    nil,
						StorePathType:      "FileSystem",
						StorePathValue:     "/path/to/store",
						PrivateKeyAllowed:  "Optional",
						JobProperties:      []string{"prop1", "prop2"},
						ServerRequired:     true,
						PowerShell:         true,
						BlueprintAllowed:   true,
						CustomAliasAllowed: "ByPolicy",
					},
				},
			},
			Pam: PamDetails{
				AssemblyName:            "Example.Assembly",
				ProviderName:            "ExampleProvider",
				FullyQualifiedClassName: "Example.Namespace.ProviderClass",
				DbName:                  "ExampleDB",
			},
		},
	}
}
