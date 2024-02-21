// Copyright 2024 Keyfactor
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
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func (apaz AuthProviderAzureIDParams) authAzureIdentity() (azcore.AccessToken, error) {
	log.Debug().Msg("enter: AuthProviderAzureIDParams.authAzureIdentity()")
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error().Err(err).Msg("unable to authenticate with Azure Identity")
		return azcore.AccessToken{}, err
	}
	log.Debug().Msg("return: AuthProviderAzureIDParams.authAzureIdentity()")
	return cred.GetToken(
		nil,
		policy.TokenRequestOptions{
			Scopes: []string{"https://vault.azure.net/.default"},
		},
	)
}

func (apaz AuthProviderAzureIDParams) authenticate() (ConfigurationFile, error) {
	log.Info().Msg("Authenticating with Azure Identity")
	log.Debug().Msg("enter: AuthProviderAzureIDParams.authenticate()")
	metadataURL := "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://vault.azure.net"
	log.Debug().Str("metadataURL", metadataURL).Msg("fetching metadata")

	client := &http.Client{}
	config := ConfigurationFile{}
	//log.Println("Creating request to:", metadataURL)
	log.Debug().Msg("Creating HTTP request to Azure Metadata Service")
	req, respBodyErr := http.NewRequest("GET", metadataURL, nil)

	if respBodyErr != nil {
		log.Error().Err(respBodyErr).Msg("Error creating request")
		return config, respBodyErr
	} else if req == nil {
		rErr := fmt.Errorf("error, request is nil when requesting %s", metadataURL)
		log.Error().Err(rErr)
		return config, rErr
	}

	req.Header.Set("Metadata", "true")

	log.Debug().Msg("Sending HTTP request to Azure Metadata Service")
	resp, metaRespErr := client.Do(req)
	log.Trace().Interface("resp", resp).Msg("response from Azure Metadata Service")
	if metaRespErr != nil {
		log.Error().Err(metaRespErr).Msg("failed to auth with Azure Identity")
		return config, returnHttpErr(resp, metaRespErr)
	}
	log.Debug().Str("responseStatusCode", resp.Status).Send()

	defer resp.Body.Close()

	body, respBodyErr := io.ReadAll(resp.Body)
	if respBodyErr != nil {
		log.Error().Err(respBodyErr).Msg("invalid response from Azure")
		return config, respBodyErr
	}

	if resp.StatusCode != 200 {
		log.Trace().Str("responseBody", string(body)).Msg("response body from Azure")
		return config, fmt.Errorf("invalid response code '%d'", resp.StatusCode)
	}

	// Parse the access token from the response JSON
	log.Debug().Msg("call: parseAccessToken()")
	accessToken, tokenErr := apaz.parseAccessToken(body)
	log.Debug().Msg("returned: parseAccessToken()")

	if tokenErr != nil {
		log.Error().Err(tokenErr).Msg("unable to parse access token from Azure response")
		return config, tokenErr
	}

	if accessToken != "" {
		log.Debug().Str("accessToken", hashSecretValue(accessToken)).Msg("access token from Azure response")
	}

	secretURL := fmt.Sprintf("https://%s.vault.azure.net/secrets/%s?api-version=7.0", apaz.AzureVaultName, apaz.SecretName)
	log.Debug().Str("secretURL", secretURL).Msg("returning secret URL for Azure Key Vault secret")
	log.Debug().Msg("return: AuthProviderAzureIDParams.authenticate()")
	return apaz.getCommandCredsFromAzureKeyVault(secretURL, accessToken)
}

func (apaz AuthProviderAzureIDParams) getCommandCredsFromAzureKeyVault(secretURL string, accessToken string) (ConfigurationFile, error) {
	log.Debug().Str("secretURL", secretURL).
		Str("accessToken", hashSecretValue(accessToken)).
		Msg("enter: AuthProviderAzureIDParams.getCommandCredsFromAzureKeyVault()")
	client := &http.Client{}
	config := ConfigurationFile{}
	log.Info().Str("secretURL", secretURL).Msg("fetching secret from Azure Key Vault")
	log.Debug().Msg("Creating HTTP request to Azure Key Vault")
	req, jsonErr := http.NewRequest("GET", secretURL, nil)
	if jsonErr != nil {
		log.Error().Err(jsonErr).Msg("unable to create request")
		return config, jsonErr
	} else if req == nil {
		rErr := fmt.Errorf("error, request is nil when requesting %s", secretURL)
		log.Error().Err(rErr).Msg("unable to create request")
		return config, rErr
	} else if accessToken == "" {
		aErr := fmt.Errorf("error, access token is empty when requesting %s", secretURL)
		log.Error().Err(aErr).Msg("access token is empty unable to fetch from Azure Key Vault")
		return config, aErr
	}

	log.Debug().Msg("Setting request headers")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	log.Debug().Msg("Sending HTTP request to Azure Key Vault")
	resp, respErr := client.Do(req)
	log.Debug().Msg("returned from HTTP request to Azure Key Vault")
	log.Trace().Interface("resp", resp).Msg("response from Azure Key Vault")
	if respErr != nil {
		log.Error().Err(respErr).Msg("unable to fetch secret from Azure Key Vault")
		return config, returnHttpErr(resp, respErr)
	}
	defer resp.Body.Close()

	log.Debug().Msg("Reading response body")
	body, respBodyErr := io.ReadAll(resp.Body)
	if respBodyErr != nil {
		log.Error().Err(respBodyErr).Msg("unable to read response body")
		return config, respBodyErr
	}
	// Check the response code
	log.Debug().Str("responseStatusCode", resp.Status).Msg("response status code from Azure Key Vault")
	if resp.StatusCode != 200 {
		log.Error().Int("responseStatusCode", resp.StatusCode).Msg("invalid response code")
		return config, fmt.Errorf("invalid status code '%d'", resp.StatusCode)
	}
	// Convert response body to json
	var f interface{}
	log.Debug().Msg("Converting response body to JSON")
	jsonErr = json.Unmarshal(body, &f)
	if jsonErr != nil {
		log.Error().Err(jsonErr).Msg("invalid response from Azure Key Vault")
		return config, jsonErr
	}

	//Parse KeyVault "value" from json
	m := f.(map[string]interface{})

	// Pretty print the json
	log.Debug().Msg("Formatting JSON response")
	j, _ := json.MarshalIndent(m["value"], "", "  ")
	log.Trace().Str("json", string(j)).Msg("json response from Azure Key Vault")

	// try to convert value to ConfigurationFile
	//read escaped json string into ConfigurationFile
	log.Debug().Msg("Converting JSON response to ConfigurationFile")
	jsonErr = json.Unmarshal(j, &config)

	if jsonErr != nil {
		log.Error().Err(jsonErr).Msg("unable to unmarshal JSON response")
		var configInt interface{}
		log.Debug().Msg("Converting JSON response to interface{}")
		jsonErr = json.Unmarshal(j, &configInt)
		switch configInt.(type) {
		case string:
			// convert configInt to ConfigurationFile
			log.Debug().Msg("Converting interface{} to ConfigurationFile")
			jsonErr = json.Unmarshal([]byte(configInt.(string)), &config)
			if jsonErr == nil {
				log.Error().Err(jsonErr).Msg("unable to convert Azure Key Vault secret to kfutil config")
				return config, jsonErr
			}
		}
		// Check if it's an instance of ConfigurationFileEntry
		log.Debug().Msg("Converting JSON response to ConfigurationFileEntry")
		var configEntry ConfigurationFileEntry
		jsonErr2 := json.Unmarshal(j, &configEntry)
		if jsonErr2 != nil {
			log.Error().Err(jsonErr2).Msg("unable to convert Azure Key Vault secret to kfutil config")
			return config, jsonErr2
		}
		log.Trace().Interface("configEntry", configEntry).Msg("configEntry")

		// Convert to ConfigurationFile
		config = ConfigurationFile{
			Servers: map[string]ConfigurationFileEntry{
				"default": configEntry,
			},
		}
	}
	log.Debug().Msg("return: AuthProviderAzureIDParams.getCommandCredsFromAzureKeyVault()")
	log.Info().Msg("successfully fetched secret from Azure Key Vault")
	return config, nil
}

func (apaz AuthProviderAzureIDParams) parseAccessToken(body []byte) (string, error) {
	log.Debug().Msg("enter: AuthProviderAzureIDParams.parseAccessToken()")
	var f interface{}
	log.Debug().Msg("Converting response body to JSON")
	err := json.Unmarshal(body, &f)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse access token from Azure response")
		return "", err
	}

	m := f.(map[string]interface{})
	aToken := m["access_token"].(string)
	log.Debug().Str("accessToken", hashSecretValue(aToken)).Msg("access token from Azure response")
	log.Debug().Msg("return: AuthProviderAzureIDParams.parseAccessToken()")
	return aToken, nil
}
