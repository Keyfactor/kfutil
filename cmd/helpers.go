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
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func mergeErrsToString(errs *[]error, indent bool) string {
	var errStr string
	if errs == nil || len(*errs) == 0 {
		return ""
	}
	for _, err := range *errs {
		if indent {
			errStr += fmt.Sprintf(" \t%s\r\n", err)
			continue
		}
		errStr += fmt.Sprintf("%s\r\n", err)
	}
	return errStr
}

func boolToPointer(b bool) *bool {
	return &b
}

func checkDebug(v bool) bool {
	envDebug := os.Getenv("KFUTIL_DEBUG")
	envValue, _ := strconv.ParseBool(envDebug)
	switch {
	case (envValue && !v) || (envValue && v):
		//log.SetOutput(os.Stdout)
		return envValue
	case v:
		//log.SetOutput(os.Stdout)
		return v
	default:
		//log.SetOutput(io.Discard)
		return v
	}
}

func csvRemoveLastColumn(inputFile, outputFile string) error {
	// Open the input CSV file
	input, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	// Create a CSV reader for the input file
	reader := csv.NewReader(input)

	// Open the output CSV file
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Create a CSV writer for the output file
	writer := csv.NewWriter(output)

	// Read and process each row
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		// Remove the last column (i.e., the last element in the record slice)
		if len(record) > 0 {
			record = record[:len(record)-1]
		}

		// Write the modified row to the output CSV file
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	// Flush any buffered data to the output file
	writer.Flush()

	// Check for errors during the writing process
	return writer.Error()
}

func csvToMap(filename string) ([]map[string]string, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader
	csvReader := csv.NewReader(file)

	// Read the CSV header to get column names
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	// Initialize a slice to store the maps
	var data []map[string]string

	// Read and process each row
	for {
		// Read a row
		row, err := csvReader.Read()
		if err == nil {
			// Create a map for the row data
			rowMap := make(map[string]string)

			// Populate the map with data from the row
			for i, column := range header {
				rowMap[column] = row[i]
			}

			// Append the map to the data slice
			data = append(data, rowMap)
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}

	return data, nil
}

func findMatchingFiles(pattern string) ([]string, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Use the filepath package to create a glob pattern
	globPattern := filepath.Join(currentDir, pattern)

	// Use the filepath package to perform the glob operation
	matchingFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, err
	}

	return matchingFiles, nil
}

func getCurrentTime(f string) string {
	switch f {
	case "unix":
		return strconv.FormatInt(time.Now().Unix(), 10)
	case "unixNano":
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	case "date":
		return time.Now().Format("2006-01-02")
	case "time":
		return time.Now().Format("15:04:05")
	default:
		return time.Now().Format(time.RFC3339)
	}
}

func informDebug(debugFlag bool) {
	debugModeEnabled := checkDebug(debugFlag)
	zerolog.SetGlobalLevel(zerolog.Disabled)
	if debugModeEnabled {
		//zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

func intToPointer(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func isExperimentalFeatureEnabled(expFlag bool, isExperimental bool) (bool, error) {
	envExp := os.Getenv("KFUTIL_EXP")
	envValue, _ := strconv.ParseBool(envExp)
	if envValue {
		return envValue, nil
	}
	if isExperimental && !expFlag {
		return false, fmt.Errorf("experimental features are not enabled. To enable experimental features, use the `--exp` flag or set the `KFUTIL_EXP` environment variable to true")
	}
	return envValue, nil
}

func generateRandomUUID() string {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		// Handle the error if UUID generation fails.
		panic(err)
	}
	return uuidObj.String()
}

func loadJSONFile(filename string) (map[string]interface{}, error) {
	// Read the JSON file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a map to unmarshal the JSON data into
	var result map[string]interface{}

	// Create a decoder to decode the JSON data from the file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func logGlobals() {

	if !logInsecure {
		log.Debug().Str("configFile", configFile).
			Str("profile", profile).
			Str("providerType", providerType).
			Str("providerProfile", providerProfile).
			//Str("providerConfig", providerConfig).
			Bool("noPrompt", noPrompt).
			Bool("expEnabled", expEnabled).
			Bool("debugFlag", debugFlag).
			Str("kfcUsername", kfcUsername).
			Str("kfcHostName", kfcHostName).
			Str("kfcPassword", hashSecretValue(kfcPassword)).
			Str("kfcDomain", kfcDomain).
			Str("kfcAPIPath", kfcAPIPath).
			Msg("Global Flags")
	} else {
		log.Debug().Str("configFile", configFile).
			Str("profile", profile).
			Str("providerType", providerType).
			Str("providerProfile", providerProfile).
			//Str("providerConfig", providerConfig).
			Bool("noPrompt", noPrompt).
			Bool("expEnabled", expEnabled).
			Bool("debugFlag", debugFlag).
			Str("kfcUsername", kfcUsername).
			Str("kfcHostName", kfcHostName).
			Str("kfcPassword", kfcPassword).
			Str("kfcDomain", kfcDomain).
			Str("kfcAPIPath", kfcAPIPath).
			Msg("Global Flags")
	}

}

func mapToCSV(data []map[string]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header using keys from the first map
	var header []string
	if len(data) > 0 {
		for key := range data[0] {
			header = append(header, key)
		}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	// Write map data to CSV
	for _, row := range data {
		var record []string
		for _, key := range header {
			record = append(record, row[key])
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func outputError(err error, isFatal bool, format string) {
	if isFatal {
		if format == "json" {
			fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err))
		} else {
			fmt.Errorf(fmt.Sprintf("Fatal error: %s", err))
		}
	}
	if format == "json" {
		fmt.Println(fmt.Sprintf("{\"error\": \"%s\"}", err))
	} else {
		fmt.Println(fmt.Sprintf("Error: %s", err))
	}
}

func outputResult(result interface{}, format string) {
	if format == "json" {
		fmt.Println(result)
	} else {
		fmt.Println(fmt.Sprintf("%s", result))
	}
}

func readCSVHeader(filename string) ([]string, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader
	csvReader := csv.NewReader(file)

	// Read the header row
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	return header, nil
}

func stringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func storeTypeIdentifierFlagCheck(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("store-type-name") && !cmd.Flags().Changed("store-type-id") {
		inputErr := fmt.Errorf("'store-type-id' or 'store-type-name' must be provided")
		cmd.Usage()
		return inputErr
	}
	return nil
}

func warnExperimentalFeature(expEnabled bool, isExperimental bool) error {
	_, expErr := isExperimentalFeatureEnabled(expEnabled, isExperimental)
	if expErr != nil {
		//fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
		log.Error().Err(expErr)
		return expErr
	}
	return nil
}

func writeJSONFile(filename string, data interface{}) error {
	// Open the JSON file for writing
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create an encoder to encode the data to JSON and write it to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func returnHttpErr(resp *http.Response, err error) error {
	if resp == nil {
		log.Error().Err(err).Msg("unable to create PAM provider - no response")
		return err
	}
	if resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Err(err).Str("httpResponseCode", resp.Status).
			Str("httpResponseBody", string(body)).
			Msg("unable to create PAM provider")

		// parse response body into map[string]interface{}
		var responseMap map[string]interface{}
		if jErr := json.Unmarshal(body, &responseMap); jErr != nil {
			// check of the response body is plaintext
			if string(body) != "" {
				errMsg := fmt.Errorf("%s- %s", resp.Status, string(body))
				return errMsg
			}
			log.Error().Err(jErr).Msg("unable to parse response body")
			return jErr
		}

		errMsg := fmt.Errorf("%s- %s", resp.Status, responseMap["Message"].(string))

		return errMsg
	}
	log.Error().Err(err).Str("httpResponseCode", resp.Status).
		Msg("unable to create PAM provider")
	return err
}
