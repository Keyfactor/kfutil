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

package cmd

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	manifestv1 "kfutil/pkg/keyfactor/v1"
	"os"
	"testing"
)

func Test_StoreTypesGet(t *testing.T) {
	t.Run("WithName", func(t *testing.T) {
		testCmd := RootCmd
		// Attempt to get the AWS store type because it comes with the product
		testCmd.SetArgs([]string{"store-types", "get", "--name", "PEM"})
		output := captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})
		var storeType map[string]interface{}
		if err := json.Unmarshal([]byte(output), &storeType); err != nil {
			t.Fatalf("Error unmarshalling JSON: %v", err)
		}

		assert.NotNil(t, storeType["Name"], "Expected store type to have a name")
		assert.NotNil(t, storeType["ShortName"], "Expected store type to have short name")
		assert.NotNil(t, storeType["StoreType"], "Expected store type to have a store type")

		// verify that the store type is an integer
		_, ok := storeType["StoreType"].(float64)
		assert.True(t, ok, "Expected store type to be an integer")
		// verify short name is a string
		_, ok = storeType["ShortName"].(string)
		assert.True(t, ok, "Expected short name to be a string")
		// verify name is a string
		_, ok = storeType["Name"].(string)
		assert.True(t, ok, "Expected name to be a string")
		// check that shortname == AWS
		assert.Equal(t, storeType["ShortName"], "PEM", "Expected short name to be PEM")
	})

	t.Run("GenericOutput", func(t *testing.T) {
		testCmd := RootCmd
		// Attempt to get the AWS store type because it comes with the product
		testCmd.SetArgs([]string{"store-types", "get", "--name", "PEM", "-g"})
		output := captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})
		var storeType map[string]interface{}
		if err := json.Unmarshal([]byte(output), &storeType); err != nil {
			t.Fatalf("Error unmarshalling JSON: %v", err)
		}

		assert.NotNil(t, storeType["Name"], "Expected store type to have a Name")
		assert.NotNil(t, storeType["ShortName"], "Expected store type to have ShortName")

		assert.Nil(t, storeType["StoreType"], "Expected StoreType to to be nil")
		assert.Nil(t, storeType["InventoryJobType"], "Expected InventoryJobType to be nil")
		assert.Nil(t, storeType["InventoryEndpoint"], "Expected InventoryEndpoint to be nil")
		assert.Nil(t, storeType["ManagementJobType"], "Expected ManagementJobType to be nil")
		assert.Nil(t, storeType["DiscoveryJobType"], "Expected DiscoveryJobType to be nil")
		assert.Nil(t, storeType["EnrollmentJobType"], "Expected EnrollmentJobType to be nil")
		assert.Nil(t, storeType["ImportType"], "Expected ImportType to be nil")

		// verify short name is a string
		_, ok := storeType["ShortName"].(string)
		assert.True(t, ok, "Expected short name to be a string")
		// verify name is a string
		_, ok = storeType["Name"].(string)
		assert.True(t, ok, "Expected name to be a string")
		// check that shortname == PEM
		assert.Equal(t, storeType["ShortName"], "PEM", "Expected short name to be PEM")
	})

	t.Run("OutputToManifest", func(t *testing.T) {
		testCmd := RootCmd
		// Attempt to get the AWS store type because it comes with the product
		testCmd.SetArgs([]string{"store-types", "get", "--name", "PEM", "--output-to-integration-manifest"})
		captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})

		// Verify that integration-manifest.json was created
		manifest := manifestv1.IntegrationManifest{}
		err := manifest.LoadFromFilesystem()
		if err != nil {
			t.Fatalf("Error loading integration manifest: %v", err)
		}

		if len(manifest.About.Orchestrator.StoreTypes) != 1 {
			t.Fatalf("Expected 1 store type, got %d", len(manifest.About.Orchestrator.StoreTypes))
		}

		// Clean up
		err = os.Remove("integration-manifest.json")
		if err != nil {
			t.Errorf("Error removing integration-manifest.json: %v", err)
		}
	})
}
