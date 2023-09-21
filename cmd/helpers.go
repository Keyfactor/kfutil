package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"time"
)

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

func loadCSVAsMap(filename string) ([]map[string]string, error) {
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

func logGlobals() {

	if !logInsecure {
		log.Debug().Str("configFile", configFile).
			Str("profile", profile).
			Str("providerType", providerType).
			Str("providerProfile", providerProfile).
			Str("providerConfig", providerConfig).
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
			Str("providerConfig", providerConfig).
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

func removeLastColumn(inputFile, outputFile string) error {
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

func storeTypeIdentifierFlagCheck(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("store-type-name") && !cmd.Flags().Changed("store-type-id") {
		inputErr := fmt.Errorf("'store-type-id' or 'store-type-name' must be provided")
		cmd.Usage()
		return inputErr
	}
	return nil
}

func warnExperimentalFeature(expEnabled bool, isExperimental bool) error {
	_, expErr := IsExperimentalFeatureEnabled(expEnabled, isExperimental)
	if expErr != nil {
		//fmt.Println(fmt.Sprintf("WARNING this is an expEnabled feature, %s", expErr))
		log.Error().Err(expErr)
		return expErr
	}
	return nil
}
