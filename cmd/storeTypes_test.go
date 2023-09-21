package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_StoreTypesHelpCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"store-types", "--help"})
	err := testCmd.Execute()

	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"store-types", "-h"})
	err = testCmd.Execute()
	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"store-types", "--halp"})
	err = testCmd.Execute()

	assert.Error(t, err)
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}

func Test_StoreTypesListCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"store-types", "list"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var storeTypes []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &storeTypes); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// iterate over the store types and verify that each has a name shortname and storetype
	for _, storeType := range storeTypes {
		assert.NotNil(t, storeType["Name"], "Expected store type to have a Name")
		assert.NotNil(t, storeType["ShortName"], "Expected store type to have ShortName")
		assert.NotNil(t, storeType["StoreType"], "Expected store type to have a StoreType")

		// verify that the store type is an integer
		_, ok := storeType["StoreType"].(float64)
		assert.True(t, ok, "Expected store type to be an integer")
		// verify short name is a string
		_, ok = storeType["ShortName"].(string)
		assert.True(t, ok, "Expected short name to be a string")
		// verify name is a string
		_, ok = storeType["Name"].(string)
		assert.True(t, ok, "Expected name to be a string")
	}
}

func Test_StoreTypesFetchTemplatesCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"store-types", "templates-fetch"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var storeTypes map[string]interface{}
	if err := json.Unmarshal([]byte(output), &storeTypes); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// iterate over the store types and verify that each has a name shortname and storetype
	for sType := range storeTypes {
		storeType := storeTypes[sType].(map[string]interface{})
		assert.NotNil(t, storeType["Name"], "Expected store type to have a name")
		assert.NotNil(t, storeType["ShortName"], "Expected store type to have short name")

		// verify short name is a string
		_, ok := storeType["ShortName"].(string)
		assert.True(t, ok, "Expected short name to be a string")
		// verify name is a string
		_, ok = storeType["Name"].(string)
		assert.True(t, ok, "Expected name to be a string")
	}
}

func Test_StoreTypesGetCmd(t *testing.T) {
	testCmd := RootCmd
	// Attempt to get the AWS store type because it comes with the product
	testCmd.SetArgs([]string{"store-types", "get", "--name", "IIS"})
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
	assert.Equal(t, storeType["ShortName"], "IIS", "Expected short name to be IIS")
}

func Test_StoreTypesGetGenericCmd(t *testing.T) {
	testCmd := RootCmd
	// Attempt to get the AWS store type because it comes with the product
	testCmd.SetArgs([]string{"store-types", "get", "--name", "IIS", "--generic"})
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
	// check that shortname == IIS
	assert.Equal(t, storeType["ShortName"], "IIS", "Expected short name to be IIS")
}

func Test_StoreTypesCreateFromTemplatesCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"store-types", "templates-fetch"})
	templatesOutput := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var storeTypes map[string]interface{}
	if err := json.Unmarshal([]byte(templatesOutput), &storeTypes); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Verify that the length of the response is greater than 0
	assert.True(t, len(storeTypes) >= 0, "Expected non-empty list of store types")

	// iterate over the store types and verify that each has a name shortname and storetype
	for sType := range storeTypes {
		storeType := storeTypes[sType].(map[string]interface{})
		assert.NotNil(t, storeType["Name"], "Expected store type to have a name")
		assert.NotNil(t, storeType["ShortName"], "Expected store type to have short name")

		// verify short name is a string
		_, ok := storeType["ShortName"].(string)
		assert.True(t, ok, "Expected short name to be a string")
		// verify name is a string
		_, ok = storeType["Name"].(string)
		assert.True(t, ok, "Expected name to be a string")

		// Attempt to create the store type
		shortName := storeType["ShortName"].(string)
		createStoreTypeTest(t, shortName)
	}
	createAllStoreTypes(t, storeTypes)
}

func createAllStoreTypes(t *testing.T, storeTypes map[string]interface{}) {
	t.Run(fmt.Sprintf("Create ALL StoreTypes"), func(t *testing.T) {
		testCmd := RootCmd
		// Attempt to get the AWS store type because it comes with the product
		testCmd.SetArgs([]string{"store-types", "create", "--all"})
		output := captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})
		assert.NotNil(t, output, "No output returned from create all command")

		// iterate over the store types and verify that each has a name shortname and storetype
		for sType := range storeTypes {
			storeType := storeTypes[sType].(map[string]interface{})
			assert.NotNil(t, storeType["Name"], "Expected store type to have a name")
			assert.NotNil(t, storeType["ShortName"], "Expected store type to have short name")

			// verify short name is a string
			_, ok := storeType["ShortName"].(string)
			assert.True(t, ok, "Expected short name to be a string")
			// verify name is a string
			_, ok = storeType["Name"].(string)
			assert.True(t, ok, "Expected name to be a string")

			// Attempt to create the store type
			shortName := storeType["ShortName"].(string)

			assert.Contains(t, output, fmt.Sprintf("Certificate store type %s created with ID", shortName), "Expected output to contain store type created message")

			// Delete again after create
			deleteStoreTypeTest(t, shortName, true)
		}
	})
}

func deleteStoreTypeTest(t *testing.T, shortName string, allowFail bool) {
	t.Run(fmt.Sprintf("Delete StoreType %s", shortName), func(t *testing.T) {
		testCmd := RootCmd
		testCmd.SetArgs([]string{"store-types", "delete", "--name", shortName})
		deleteStoreOutput := captureOutput(func() {
			err := testCmd.Execute()
			if !allowFail {
				assert.NoError(t, err)
			}
		})
		if !allowFail {
			if strings.Contains(deleteStoreOutput, "does not exist") {
				t.Errorf("Store type %s does not exist", shortName)
			}
			if strings.Contains(deleteStoreOutput, "cannot be deleted") {
				assert.Fail(t, fmt.Sprintf("Store type %s already exists", shortName))
			}
			if !strings.Contains(deleteStoreOutput, "deleted") {
				assert.Fail(t, fmt.Sprintf("Store type %s was not deleted: %s", shortName, deleteStoreOutput))
			}
			if strings.Contains(deleteStoreOutput, "error processing the request") {
				assert.Fail(t, fmt.Sprintf("Store type %s was not deleted: %s", shortName, deleteStoreOutput))
			}
		}
	})
}

func createStoreTypeTest(t *testing.T, shortName string) {
	t.Run(fmt.Sprintf("CreateStore %s", shortName), func(t *testing.T) {
		testCmd := RootCmd
		deleteStoreTypeTest(t, shortName, true)
		testCmd.SetArgs([]string{"store-types", "create", "--name", shortName})
		createStoreOutput := captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})

		if strings.Contains(createStoreOutput, "already exists") {
			assert.Fail(t, fmt.Sprintf("Store type %s already exists", shortName))
		} else if !strings.Contains(createStoreOutput, "created with ID") {
			assert.Fail(t, fmt.Sprintf("Store type %s was not created: %s", shortName, createStoreOutput))
		}
		// Delete again after create
		deleteStoreTypeTest(t, shortName, false)
	})
}
