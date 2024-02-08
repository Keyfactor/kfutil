// Copyright 2024 Keyfactor
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
	"path/filepath"
	"strings"
	"testing"
)

func Test_Stores_HelpCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"stores", "--help"})
	err := testCmd.Execute()

	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"stores", "-h"})
	err = testCmd.Execute()
	assert.NoError(t, err)

	// test root halp
	testCmd.SetArgs([]string{"stores", "--halp"})
	err = testCmd.Execute()

	assert.Error(t, err)
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}

func Test_Stores_ListCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"stores", "list", "--exp"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var stores []interface{}
	if err := json.Unmarshal([]byte(output), &stores); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// assert slice is len >= 0
	assert.GreaterOrEqual(t, len(stores), 0)

	if len(stores) > 0 {
		for _, store := range stores {
			// assert that each store has a name
			assert.NotEmpty(t, store.(map[string]interface{})["DisplayName"])
			// assert that each store has an ID
			assert.NotEmpty(t, store.(map[string]interface{})["Id"])
			// assert that each store has a type
			assert.NotEmpty(t, store.(map[string]interface{})["CertStoreType"])
		}
	}
}

func Test_Stores_GetCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"stores", "list", "--exp"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})
	var stores []interface{}
	if err := json.Unmarshal([]byte(output), &stores); err != nil {
		t.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// assert slice is len >= 0
	assert.GreaterOrEqual(t, len(stores), 0)

	if len(stores) > 0 {
		for _, store := range stores {
			// assert that each store has a name
			assert.NotEmpty(t, store.(map[string]interface{})["DisplayName"])
			// assert that each store has an ID
			assert.NotEmpty(t, store.(map[string]interface{})["Id"])
			// assert that each store has a type
			assert.NotEmpty(t, store.(map[string]interface{})["CertStoreType"])

			testCmd.SetArgs([]string{"stores", "get", "--id", store.(map[string]interface{})["Id"].(string)})
			output = captureOutput(func() {
				err := testCmd.Execute()
				assert.NoError(t, err)
			})

			var store map[string]interface{}
			if err := json.Unmarshal([]byte(output), &store); err != nil {
				t.Fatalf("Error unmarshalling JSON: %v", err)
			}

			// assert that each store has a name
			assert.NotEmpty(t, store["DisplayName"])
			// assert that each store has an ID
			assert.NotEmpty(t, store["Id"])
			// assert that each store has a type
			assert.NotEmpty(t, store["CertStoreType"])

			break // only test one store
		}
	}
}

func Test_Stores_CreateCmd(t *testing.T) {
	// TODO: test create command
}

func Test_Stores_UpdateCmd(t *testing.T) {
	// TODO: test update command
}

func Test_Stores_DeleteCmd(t *testing.T) {
	// This is covered by import/export tests
}

func Test_Stores_ImportCmd(t *testing.T) {

	// first export a store
	_, files := testExportStore(t, "k8ssecret")

	// delete all stores defined in files
	for _, f := range files {
		// open file as csv
		header, csvErr := readCSVHeader(f)
		assert.Nil(t, csvErr)
		assert.NotEmpty(t, header)

		// assert that header contains "Id" column
		assert.Contains(t, header, "Id")

		csvData, csvErr := csvToMap(f)
		assert.Nil(t, csvErr)
		assert.NotEmpty(t, csvData)
		var modifiedCSVData []map[string]string
		for _, row := range csvData {
			// assert that each row has an ID
			assert.NotEmpty(t, row["Id"])
			// delete store
			//deleteStoreTest(t, row["Id"], true)

			// modify row
			row["Id"] = ""
			row["StorePath"] = fmt.Sprintf("%s-test-%s", row["StorePath"], generateRandomUUID())
			row["Properties.ServerUsername"] = "kubeconfig"
			row["Properties.ServerPassword"] = "test"

			// append modified row to modifiedCSVData
			modifiedCSVData = append(modifiedCSVData, row)
		}

		// write modifiedCSVData to file
		outFileName := strings.Replace(f, "export", "import", 1)
		convErr := mapToCSV(modifiedCSVData, outFileName)
		assert.NoError(t, convErr)

		testCmd := RootCmd
		// test
		testCmd.SetArgs([]string{"stores", "import", "csv", "--file", outFileName, "--store-type-name", "k8ssecret", "--exp"})
		output := captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "records processed")
		assert.Contains(t, output, "results written to")
		assert.NotContains(t, output, "rows had errors")

		// append _results to outFileName between filename and extension
		resultsFileName := strings.Replace(outFileName, ".csv", "_results.csv", 1)
		testDeleteImportResults(t, resultsFileName)

		// remove files
		err := os.Remove(outFileName)
		assert.NoError(t, err)
		err = os.Remove(f)
		assert.NoError(t, err)
		err = os.Remove(resultsFileName)
		assert.NoError(t, err)
	}
}

func Test_Stores_ExportCmd(t *testing.T) {
	// test
	_, files := testExportStore(t, "k8ssecret")

	// remove all files
	for _, f := range files {
		// fetch header from csv
		header, csvErr := readCSVHeader(f)
		assert.Nil(t, csvErr)
		assert.NotEmpty(t, header)

		// parse filename from path
		filename := filepath.Base(f)

		// validate header
		testValidateCSVHeader(t, filename, header, bulkStoreImportCSVHeader)

		// remove file
		err := os.Remove(f)
		assert.NoError(t, err)
	}
}

func Test_Stores_GenerateImportTemplateCmd(t *testing.T) {
	testCmd := RootCmd
	// test
	testCmd.SetArgs([]string{"stores", "import", "generate-template", "--store-type-name", "k8ssecret", "--exp"})
	output := captureOutput(func() {
		err := testCmd.Execute()
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Template file for store type with id")
	assert.Contains(t, output, "csv")

	// split output by spaces
	outputSplit := strings.Split(output, " ")
	assert.NotEmpty(t, outputSplit)

	// get last element in outputSplit
	outfileName := outputSplit[len(outputSplit)-1]
	// remove newline from outfileName
	outfileName = strings.Replace(outfileName, "\n", "", 1)

	assert.NotEmpty(t, outfileName)
	assert.Contains(t, outfileName, "csv")

	// Verify f exists
	_, err := os.Stat(outfileName)
	assert.NoError(t, err)

	// Verify f is not empty
	f, err := os.Open(outfileName)
	assert.NoError(t, err)
	assert.NotNil(t, f)

	// Verify f has content
	fileInfo, err := f.Stat()
	assert.NoError(t, err)
	assert.NotNil(t, fileInfo)
	assert.NotZero(t, fileInfo.Size())

	// Verify f is a csv
	assert.Contains(t, fileInfo.Name(), "csv")
	header, csvErr := readCSVHeader(outfileName)
	assert.NoError(t, csvErr)
	assert.NotEmpty(t, header)

	// Verify header contains all fields in bulkStoreImportCSVHeader
	testValidateCSVHeader(t, outfileName, header, bulkStoreImportCSVHeader)

	// remove file
	err = os.Remove(outfileName)
	assert.NoError(t, err)

}

func testExportStore(t *testing.T, storeTypeName string) (string, []string) {
	var (
		output string
		files  []string
		err    error
	)
	t.Run(fmt.Sprintf("Export Stores of type %s", storeTypeName), func(t *testing.T) {
		testCmd := RootCmd
		testCmd.SetArgs([]string{"stores", "export", "--store-type-name", storeTypeName})
		output = captureOutput(func() {
			err := testCmd.Execute()
			assert.NoError(t, err)
		})

		// assert that output is not empty
		assert.NotEmpty(t, output)

		// assert that output is a string
		assert.IsType(t, "", output)

		// assert that output does not contain 'error'
		assert.NotContains(t, output, "error")

		// assert that output does not contain 'Error'
		assert.NotContains(t, output, "Error")

		// assert that output does not contain 'ERROR'
		assert.NotContains(t, output, "ERROR")

		// assert that contains "exported for store type with id"
		assert.Contains(t, output, "exported for store type with id")

		// assert that contains .csv
		assert.Contains(t, output, ".csv")

		// assert that a csv file was created in current working directory with a filename that contains 'export_store_*.csv'
		files, err = findMatchingFiles("export_stores_*.csv")
		assert.Nil(t, err)
		assert.NotEmpty(t, files)
	})
	return output, files
}

func deleteStoreTest(t *testing.T, storeID string, allowFail bool) {
	t.Run(fmt.Sprintf("Delete Store %s", storeID), func(t *testing.T) {
		testCmd := RootCmd
		testCmd.SetArgs([]string{"stores", "delete", "--id", storeID})
		deleteStoreOutput := captureOutput(func() {
			err := testCmd.Execute()
			if !allowFail {
				assert.NoError(t, err)
			}
		})
		if !allowFail {
			if strings.Contains(deleteStoreOutput, "does not exist") {
				t.Errorf("Store %s does not exist", storeID)
			}
			if strings.Contains(deleteStoreOutput, "cannot be deleted") {
				assert.Fail(t, fmt.Sprintf("Store %s already exists", storeID))
			}
			if strings.Contains(deleteStoreOutput, "error processing the request") {
				assert.Fail(t, fmt.Sprintf("Store %s was not deleted: %s", storeID, deleteStoreOutput))
			}
			assert.Contains(t, deleteStoreOutput, "deleted")
			assert.Contains(t, deleteStoreOutput, storeID)
		}
	})
}

func testValidateCSVHeader(t *testing.T, filename string, header []string, expected []string) {
	// iterate bulkStoreImportCSVHeader and verify that each header is in the csv header
	t.Run(fmt.Sprintf("Validate CSV header %s", filename), func(t *testing.T) {
		for _, h := range expected {
			if h != "Properties" {
				assert.Contains(t, header, h)
			}
		}

		var props []string
		for _, h := range header {
			if strings.Contains(h, "Properties") {
				props = append(props, h)
			}
		}
		assert.NotEmpty(t, props)
	})
}

func testDeleteImportResults(t *testing.T, fileName string) {
	// get filename from path
	fName := filepath.Base(fileName)
	t.Run(fmt.Sprintf("Deleting imported stores %s", fName), func(t *testing.T) {
		csvData, err := csvToMap(fileName)
		assert.NoError(t, err)

		for _, row := range csvData {
			// assert that each row has an ID
			//assert.NotEmpty(t, row[""])
			// delete store
			if row["Id"] != "" && row["Id"] != "error" && row["Id"] != "Id" {
				deleteStoreTest(t, row["Id"], false)
			}
		}
	})
}
