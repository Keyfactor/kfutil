// Package cmd Copyright 2023 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func Test_PAMHelpCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"pam", "--help"})
	err := testCmd.Execute()

	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"pam", "-h"})
	err = testCmd.Execute()
	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"pam", "--halp"})
	err = testCmd.Execute()

	assert.Error(t, err)
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}

func Test_PAMListCmd(t *testing.T) {
	// list providers
	pamProviders, err := testListPamProviders(t)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to list PAM providers: %v", err)
	}

	if len(pamProviders) <= 0 {
		t.Fatalf("0 PAM providers found, cannot test list")
	}
}

func Test_PAMTypesListCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"pam", "types-list"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var pTypes []interface{}
	if err := json.Unmarshal([]byte(output), &pTypes); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// assert slice is len >= 0
	assert.GreaterOrEqual(t, len(pTypes), 0)

	if len(pTypes) > 0 {
		for _, p := range pTypes {
			providerConfig := p.(map[string]interface{})
			// assert that each p has a name
			assert.NotEmpty(t, providerConfig["Name"])
			// assert that each p has an ID
			assert.NotEmpty(t, providerConfig["Id"])
			// assert that each p has a type
			//emptyParams := assert.NotEmpty(t, providerConfig["ProviderTypeParams"])
			//if !emptyParams {
			//	t.Logf("ProviderTypeParams is empty for %s", providerConfig["Name"])
			//}

			// Check params is a list of maps
			pTypeParams := providerConfig["ProviderTypeParams"].([]interface{})
			//assert.NotEmpty(t, pTypeParams)
			//assert.GreaterOrEqual(t, len(pTypeParams), 0)
			if len(pTypeParams) > 0 {
				for _, param := range pTypeParams {
					assert.NotEmpty(t, param.(map[string]interface{})["Id"])
					assert.NotEmpty(t, param.(map[string]interface{})["Name"])
					assert.NotEmpty(t, param.(map[string]interface{})["DataType"])
				}
			} else {
				t.Logf("ProviderTypeParams is empty for %s (%s)", providerConfig["Name"], providerConfig["Id"])
			}
		}
	}
}

func Test_PAMGetCmd(t *testing.T) {
	testCmd := RootCmd
	// list providers
	pamProviders, err := testListPamProviders(t)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to list PAM providers: %v", err)
	}

	if len(pamProviders) > 0 {
		for _, p := range pamProviders {
			providerConfig := p.(map[string]interface{})
			// assert that each p has a name
			assert.NotEmpty(t, providerConfig["Name"])
			// assert that each p has an ID
			assert.NotEmpty(t, providerConfig["Id"])
			// assert that each p has a type
			assert.NotEmpty(t, providerConfig["ProviderType"])

			// Check params is a list of maps
			pTypeParams := providerConfig["ProviderType"].(map[string]interface{})["ProviderTypeParams"].([]interface{})
			assert.NotEmpty(t, pTypeParams)
			assert.GreaterOrEqual(t, len(pTypeParams), 0)
			if len(pTypeParams) > 0 {
				for _, param := range pTypeParams {
					assert.NotEmpty(t, param.(map[string]interface{})["Id"])
					assert.NotEmpty(t, param.(map[string]interface{})["Name"])
					assert.NotEmpty(t, param.(map[string]interface{})["DataType"])
				}
			}

			// test
			idInt := int(providerConfig["Id"].(float64))
			idStr := strconv.Itoa(idInt)
			testCmd.SetArgs([]string{"pam", "get", "--id", idStr})
			output := captureOutput(func() {
				err := testCmd.Execute()
				assert.NoError(t, err)
			})
			var pamProvider interface{}
			if err := json.Unmarshal([]byte(output), &pamProvider); err != nil {
				t.Fatalf("Error unmarshalling JSON: %v", err)
			}
			// assert that each p has a name
			assert.NotEmpty(t, pamProvider.(map[string]interface{})["Name"])
			// assert that each p has an ID
			assert.NotEmpty(t, pamProvider.(map[string]interface{})["Id"])
			// assert that each p has a type
			assert.NotEmpty(t, pamProvider.(map[string]interface{})["ProviderType"])
		}
	} else {
		t.Fatalf("0 PAM providers found, cannot test get")
	}
}

func Test_PAMTypesCreateCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	randomName := generateRandomUUID()
	t.Logf("randomName: %s", randomName)
	testCmd.SetArgs([]string{"pam", "types-create", "--repo", "hashicorp-vault-pam", "--name", randomName})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var createResponse interface{}
	if err := json.Unmarshal([]byte(output), &createResponse); err != nil {
		t.Log(output)
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}
	assert.NotEmpty(t, createResponse.(map[string]interface{})["Id"])
	assert.NotEmpty(t, createResponse.(map[string]interface{})["Name"])
	assert.Equal(t, createResponse.(map[string]interface{})["Name"], randomName)
	assert.NotEmpty(t, createResponse.(map[string]interface{})["Parameters"])
}

func Test_PAMCreateCmd(t *testing.T) {
	// test

	// get current working dir
	cwd, _ := os.Getwd()
	t.Logf("cwd: %s", cwd)

	providerName := "Delinea-SecretServer-test"
	t.Logf("providerName: %s", providerName)
	inputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-template.json")
	t.Logf("inputFileName: %s", inputFileName)
	invalidInputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-invalid.json")
	t.Logf("invalidInputFileName: %s", invalidInputFileName)
	//cProviderTypeName := "Delinea-SecretServer"

	// read input file into a map[string]interface{}
	updatedFileName, fErr := testFormatPamCreateConfig(t, inputFileName, "", false)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}

	// Test invalid config file
	createResponse, err := testCreatePamProvider(t, updatedFileName, providerName, false)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to create a PAM provider: %v", err)
	}

	createdObject := createResponse.(map[string]interface{})
	createdId := int(createdObject["Id"].(float64))

	// Test creating same provider again
	_, err = testCreatePamProvider(t, inputFileName, providerName, true)
	assert.Error(t, err)
	if err == nil {
		t.Fatalf("this test should have failed to create a duplicate PAM provider: %v", err)
	}

	// Delete the provider we just created
	err = testDeletePamProvider(t, createdId, false)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to delete a PAM provider %d: %v", createdId, err)
	}

	// Test invalid config file
	_, err = testCreatePamProvider(t, invalidInputFileName, providerName, true)
	assert.Error(t, err)
	if err == nil {
		t.Fatalf("this test should have failed to create a PAM provider: %v", err)
	}

	// delete the updated file
	os.Remove(updatedFileName)
}

func Test_PAMUpdateCmd(t *testing.T) {
	// test
	// get current working dir
	cwd, _ := os.Getwd()
	t.Logf("cwd: %s", cwd)

	providerName := "Delinea-SecretServer-test"
	t.Logf("providerName: %s", providerName)
	inputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-template.json")
	t.Logf("inputFileName: %s", inputFileName)
	invalidInputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-invalid.json")
	t.Logf("invalidInputFileName: %s", invalidInputFileName)

	// read input file into a map[string]interface{}
	updatedFileName, fErr := testFormatPamCreateConfig(t, inputFileName, "", false)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}
	// Create a provider to delete, doesn't matter if it fails, assume it exists then delete it
	testCreatePamProvider(t, updatedFileName, providerName, true)

	updatedFileName, fErr = testFormatPamCreateConfig(t, inputFileName, providerName, true)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}

	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"pam", "update", "--from-file", updatedFileName})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})

	var updateResponse interface{}
	if err := json.Unmarshal([]byte(output), &updateResponse); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}
	assert.NotEmpty(t, updateResponse.(map[string]interface{})["Id"])
	assert.NotEmpty(t, updateResponse.(map[string]interface{})["Name"])
	assert.Equal(t, updateResponse.(map[string]interface{})["Name"], providerName)
	assert.NotEmpty(t, updateResponse.(map[string]interface{})["ProviderType"])
	assert.NotEmpty(t, updateResponse.(map[string]interface{})["ProviderTypeParamValues"])

	// delete the pam provider we just created
	testDeletePamProvider(t, int(updateResponse.(map[string]interface{})["Id"].(float64)), false)

	// delete the updated file
	os.Remove(updatedFileName)
}

func Test_PAMDeleteCmd(t *testing.T) {
	// test
	// get current working dir
	cwd, _ := os.Getwd()
	t.Logf("cwd: %s", cwd)

	providerName := "Delinea-SecretServer-test"
	t.Logf("providerName: %s", providerName)
	inputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-template.json")
	t.Logf("inputFileName: %s", inputFileName)
	invalidInputFileName := path.Join(filepath.Dir(cwd), "artifacts/pam/pam-create-invalid.json")
	t.Logf("invalidInputFileName: %s", invalidInputFileName)

	//cProviderTypeName := "Delinea-SecretServer"

	// read input file into a map[string]interface{}
	updatedFileName, fErr := testFormatPamCreateConfig(t, inputFileName, "", false)
	assert.NoError(t, fErr)
	if fErr != nil {
		t.Fatalf("failed to format PAM provider config file '%s': %v", inputFileName, fErr)
		return
	}
	// Create a provider to delete, doesn't matter if it fails, assume it exists then delete it
	testCreatePamProvider(t, updatedFileName, providerName, true)

	// list providers
	providersList, err := testListPamProviders(t)
	assert.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to list PAM providers: %v", err)
	}
	if len(providersList) > 0 {
		//find the one named providerName
		isDeleted := false
		for _, p := range providersList {
			providerConfig := p.(map[string]interface{})
			if providerConfig["Name"] == providerName {
				// test
				idInt := int(providerConfig["Id"].(float64))
				//idStr := strconv.Itoa(idInt)
				dErr := testDeletePamProvider(t, idInt, false)
				assert.NoError(t, dErr)
				isDeleted = true
				break
			}
		}
		if !isDeleted {
			t.Fatalf("failed to find PAM provider %s to delete", providerName)
		}
	} else {
		t.Fatalf("0 PAM providers found, cannot test delete")
	}
	// delete the updated file
	os.Remove(updatedFileName)
}

func testListPamProviders(t *testing.T) ([]interface{}, error) {
	var output string
	var pamProviders []interface{}
	var err error

	t.Run("Listing PAM provider instances", func(t *testing.T) {
		testCmd := RootCmd
		// test
		testCmd.SetArgs([]string{"pam", "list"})
		output = captureOutput(func() {
			err = testCmd.Execute()
			assert.NoError(t, err)
		})

		if err = json.Unmarshal([]byte(output), &pamProviders); err != nil {
			t.Fatalf("Error unmarshalling JSON: %v", err)
		}

		// assert slice is len >= 0
		assert.GreaterOrEqual(t, len(pamProviders), 0)

		if len(pamProviders) > 0 {
			for _, p := range pamProviders {
				providerConfig := p.(map[string]interface{})
				// assert that each p has a name
				assert.NotEmpty(t, providerConfig["Name"])
				// assert that each p has an ID
				assert.NotEmpty(t, providerConfig["Id"])
				// assert that each p has a type
				assert.NotEmpty(t, providerConfig["ProviderType"])

				// Check params is a list of maps
				pTypeParams := providerConfig["ProviderType"].(map[string]interface{})["ProviderTypeParams"].([]interface{})
				assert.NotEmpty(t, pTypeParams)
				assert.GreaterOrEqual(t, len(pTypeParams), 0)
				if len(pTypeParams) > 0 {
					for _, param := range pTypeParams {
						assert.NotEmpty(t, param.(map[string]interface{})["Id"])
						assert.NotEmpty(t, param.(map[string]interface{})["Name"])
						assert.NotEmpty(t, param.(map[string]interface{})["DataType"])
					}
				}
			}
		}
	})
	if err != nil {
		t.Log(output)
		return nil, err
	}
	return pamProviders, nil
}

func testCreatePamProvider(t *testing.T, fileName string, providerName string, allowFail bool) (interface{}, error) {
	var err error
	var createResponse interface{}
	var testName string
	if allowFail {
		testName = fmt.Sprintf("Create PAM provider '%s' allow fail", providerName)
	} else {
		testName = fmt.Sprintf("Create PAM provider '%s'", providerName)
	}
	t.Run(testName, func(t *testing.T) {
		testCmd := RootCmd

		testCmd.SetArgs([]string{"pam", "create", "--from-file", fileName})
		output := captureOutput(func() {
			err = testCmd.Execute()
			if !allowFail {
				assert.NoError(t, err)
			}
		})
		if err = json.Unmarshal([]byte(output), &createResponse); err != nil {
			if allowFail {
				t.Logf("Error unmarshalling JSON: %v", err)
			} else {
				t.Fatalf("failed to create a PAM provider: %v", err)
			}
			return
		}

		if !allowFail {
			assert.NotEmpty(t, createResponse.(map[string]interface{})["Id"])
			assert.NotEmpty(t, createResponse.(map[string]interface{})["Name"])
			assert.Equal(t, createResponse.(map[string]interface{})["Name"], providerName)
			assert.NotEmpty(t, createResponse.(map[string]interface{})["ProviderType"])
		}
	})

	return createResponse, err
}

func testDeletePamProvider(t *testing.T, pID int, allowFail bool) error {
	var err error
	var output string
	t.Run(fmt.Sprintf("Deleting PAM provider %d", pID), func(t *testing.T) {
		testCmd := RootCmd

		testCmd.SetArgs([]string{"pam", "delete", "--id", strconv.Itoa(pID)})
		output = captureOutput(func() {
			err = testCmd.Execute()
			if !allowFail {
				assert.NoError(t, err)
			}
		})
		if !allowFail {
			assert.Contains(t, output, fmt.Sprintf("Deleted PAM provider with ID %d", pID))
		}
	})
	if err != nil && !allowFail {
		t.Log(output)
		return err
	}
	return nil
}

func testListPamProviderTypes(t *testing.T, name string, allowFail bool, allowEmpty bool) (interface{}, error) {
	var err error
	var output string
	var pvTypes interface{}

	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"pam", "types-list"})
	output = captureOutput(func() {
		err = testCmd.Execute()
		if !allowFail {
			assert.NoError(t, err)
		}
	})
	var pTypes []interface{}
	if err = json.Unmarshal([]byte(output), &pTypes); err != nil && !allowFail {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// assert slice is len >= 0
	if !allowEmpty {
		assert.GreaterOrEqual(t, len(pTypes), 0)
	}

	if len(pTypes) > 0 {
		for _, p := range pTypes {
			providerConfig := p.(map[string]interface{})

			if !allowFail {
				// assert that each p has a name
				assert.NotEmpty(t, providerConfig["Name"])
				// assert that each p has an ID
				assert.NotEmpty(t, providerConfig["Id"])

				if providerConfig["Name"] == name {
					pvTypes = p
				}
			}

			// Check params is a list of maps
			pTypeParams := providerConfig["ProviderTypeParams"].([]interface{})
			//assert.NotEmpty(t, pTypeParams)
			//assert.GreaterOrEqual(t, len(pTypeParams), 0)
			if len(pTypeParams) > 0 {
				for _, param := range pTypeParams {
					if !allowFail {
						assert.NotEmpty(t, param.(map[string]interface{})["Id"])
						assert.NotEmpty(t, param.(map[string]interface{})["Name"])
						assert.NotEmpty(t, param.(map[string]interface{})["DataType"])
					}
				}
			} else {
				t.Logf("ProviderTypeParams is empty for %s (%s)", providerConfig["Name"], providerConfig["Id"])
			}
		}
	}
	return pvTypes, err
}

func testFormatPamCreateConfig(t *testing.T, inputFileName string, providerName string, isUpdate bool) (string, error) {
	pConfig, pErr := loadJSONFile(inputFileName)

	assert.NoError(t, pErr)
	if pErr != nil {
		t.Fatalf("failed to load PAM provider config file '%s': %v", inputFileName, pErr)
		return "", pErr
	}

	// parse provider type name
	cProviderType := pConfig["ProviderType"].(map[string]interface{})
	cProviderTypeName := cProviderType["Name"].(string)

	//cProviderTypeParams := cProviderType["ProviderTypeParams"].([]interface{})
	cProviderTypeParamValues := pConfig["ProviderTypeParamValues"].([]interface{})
	//var providerTypeId int

	// find the provider type ID

	// todo: for some reason calling this function mutates pConfig
	apiProviderType, pvtErr := testListPamProviderTypes(t, cProviderTypeName, false, false)
	switch apiProviderType.(type) {
	case nil:
		t.Fatalf("failed to find PAM provider type '%s' unable to create PAM provider: %v", cProviderTypeName, pvtErr)
		break
	case map[string]interface{}:
		aProviderType := apiProviderType.(map[string]interface{})
		cProviderType["Id"] = aProviderType["Id"]
		// override the config file params with the API params so you have the IDs
		cProviderType["ProviderTypeParams"] = aProviderType["ProviderTypeParams"]
		// iterate over each param and set the ID value on cProviderTypeParamValues
		nameToIdMap := make(map[string]int)
		for _, cParam := range cProviderType["ProviderTypeParams"].([]interface{}) {
			paramId := cParam.(map[string]interface{})["Id"]
			paramName := cParam.(map[string]interface{})["Name"]
			nameToIdMap[paramName.(string)] = int(paramId.(float64))
		}
		for idx, pValue := range cProviderTypeParamValues {
			pValueMap := pValue.(map[string]interface{})
			paramInfo := pValueMap["ProviderTypeParam"].(map[string]interface{})
			paramInfo["Id"] = nameToIdMap[paramInfo["Name"].(string)]
			pValueMap["ProviderTypeParam"] = paramInfo
			cProviderTypeParamValues[idx] = pValueMap
		}
		break
	default:
		t.Fatalf("failed to find PAM provider type '%s' unable to create PAM provider: %v", cProviderTypeName, pvtErr)
	}

	// reload the config file because it was mutated
	pConfig, pErr = loadJSONFile(inputFileName)
	assert.NoError(t, pErr)
	if pErr != nil {
		t.Fatalf("failed to load PAM provider config file '%s': %v", inputFileName, pErr)
		return "", pErr
	}

	// update pConfig with updated provider type info
	pConfig["ProviderType"] = cProviderType
	pConfig["ProviderTypeParamValues"] = cProviderTypeParamValues

	if providerName != "" {
		pConfig["Name"] = providerName
	}

	if isUpdate {
		// list providers
		providersList, err := testListPamProviders(t)
		assert.NoError(t, err)
		if err != nil {
			t.Fatalf("failed to list PAM providers: %v", err)
		}
		if len(providersList) > 0 {
			//find the one named providerName
			for _, p := range providersList {
				providerConfig := p.(map[string]interface{})
				if providerConfig["Name"] == providerName {
					// test
					idInt := int(providerConfig["Id"].(float64))
					//idStr := strconv.Itoa(idInt)
					pConfig["Id"] = idInt
					break
				}
			}
		} else {
			t.Fatalf("0 PAM providers found, cannot test delete")
		}
	}

	// write the updated config file
	//replace -template.json with .json
	updatedFileName := strings.Replace(inputFileName, "-template.json", ".json", 1)
	wErr := writeJSONFile(updatedFileName, pConfig)
	if wErr != nil {
		t.Fatalf("failed to write updated PAM provider config file '%s': %v", inputFileName, wErr)
		return "", wErr
	}
	return updatedFileName, nil
}
