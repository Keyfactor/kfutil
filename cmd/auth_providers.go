package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"io"
	"log"
	"net/http"
)

func (apaz AuthProviderAzureIDParams) authAzureIdentity() (azcore.AccessToken, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
		return azcore.AccessToken{}, err
	}

	return cred.GetToken(
		nil,
		policy.TokenRequestOptions{
			Scopes: []string{"https://vault.azure.net/.default"},
		},
	)
}

func (apaz AuthProviderAzureIDParams) authenticate() (ConfigurationFile, error) {
	metadataURL := "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://vault.azure.net"

	client := &http.Client{}
	config := ConfigurationFile{}
	log.Println("Creating request to:", metadataURL)
	req, respBodyErr := http.NewRequest("GET", metadataURL, nil)
	if respBodyErr != nil {
		log.Println("Error creating request:", respBodyErr)
		return config, respBodyErr
	} else if req == nil {
		log.Println("Error, request is nil")
		return config, fmt.Errorf("error, request is nil when requesting %s", metadataURL)
	}

	req.Header.Set("Metadata", "true")
	log.Println("Sending request to: ", metadataURL)
	resp, metaRespErr := client.Do(req)
	if metaRespErr != nil {
		fmt.Println("Error making request:", metaRespErr)
		return config, metaRespErr
	}

	defer resp.Body.Close()

	body, respBodyErr := io.ReadAll(resp.Body)
	if respBodyErr != nil {
		log.Println("Error reading response:", respBodyErr)
		return config, respBodyErr
	}

	log.Println("Response code:", resp.Status)
	if resp.StatusCode != 200 {
		log.Println("Response Body:", string(body))
		return config, fmt.Errorf("error, response code is %d", resp.StatusCode)
	}
	//log.Println("Response Body:", string(body))

	// Parse the access token from the response JSON
	accessToken, tokenErr := apaz.parseAccessToken(body)

	if tokenErr != nil {
		log.Println("Error parsing access token:", tokenErr)
		return config, tokenErr
	}

	if accessToken != "" {
		log.Println("Access Token:", accessToken)
	}

	secretURL := fmt.Sprintf("https://%s.vault.azure.net/secrets/%s?api-version=7.0", apaz.AzureVaultName, apaz.SecretName)

	return apaz.getCommandCredsFromAzureKeyVault(secretURL, accessToken)
}

func (apaz AuthProviderAzureIDParams) getCommandCredsFromAzureKeyVault(secretURL string, accessToken string) (ConfigurationFile, error) {
	// Create a new secret in Azure Key Vault
	client := &http.Client{}
	config := ConfigurationFile{}
	log.Println("Creating request to:", secretURL)
	req, jsonErr := http.NewRequest("GET", secretURL, nil)
	if jsonErr != nil {
		log.Println("Error creating request:", jsonErr)
		return config, jsonErr
	} else if req == nil {
		log.Println("Error, request is nil")
		return config, fmt.Errorf("error, request is nil when requesting %s", secretURL)
	} else if accessToken == "" {
		log.Println("Error, access token is empty")
		return config, fmt.Errorf("error, access token is empty when requesting %s", secretURL)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := client.Do(req)
	if respErr != nil {
		log.Println("Error making request:", jsonErr)
		return config, respErr
	}
	defer resp.Body.Close()

	body, respBodyErr := io.ReadAll(resp.Body)
	if respBodyErr != nil {
		fmt.Println("Error reading response:", jsonErr)
		return config, respBodyErr
	}
	// Check the response code
	log.Println("Response code:", resp.Status)
	if resp.StatusCode != 200 {
		log.Println("Response Body:", string(body))
		return config, fmt.Errorf("error, response code is %d", resp.StatusCode)
	}
	// Convert response body to json
	var f interface{}
	log.Println("Unmarshalling response body to JSON")
	jsonErr = json.Unmarshal(body, &f)
	if jsonErr != nil {
		fmt.Println("Error unmarshalling JSON:", jsonErr)
		return config, jsonErr
	}

	//Parse KeyVault "value" from json
	m := f.(map[string]interface{})

	// Pretty print the json
	j, _ := json.MarshalIndent(m["value"], "", "  ")
	log.Println("Secret value:", string(j))

	// try to convert value to ConfigurationFile
	//read escaped json string into ConfigurationFile
	jsonErr = json.Unmarshal(j, &config)

	if jsonErr != nil {
		var configInt interface{}
		jsonErr = json.Unmarshal(j, &configInt)
		switch configInt.(type) {
		case string:
			// convert configInt to ConfigurationFile
			jsonErr = json.Unmarshal([]byte(configInt.(string)), &config)
			if jsonErr == nil {
				return config, jsonErr
			}
		}
		// Check if it's an instance of ConfigurationFileEntry
		var configEntry ConfigurationFileEntry
		jsonErr2 := json.Unmarshal(j, &configEntry)
		if jsonErr2 != nil {
			log.Println("Error unmarshalling JSON:", jsonErr2)
			return config, jsonErr2
		}
		log.Println("Secret value: ", configEntry)

		// Convert to ConfigurationFile
		config = ConfigurationFile{
			Servers: map[string]ConfigurationFileEntry{
				"default": configEntry,
			},
		}
	}
	return config, nil
}

func (apaz AuthProviderAzureIDParams) parseAccessToken(body []byte) (string, error) {
	var f interface{}
	err := json.Unmarshal(body, &f)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return "", err
	}

	m := f.(map[string]interface{})
	return m["access_token"].(string), nil
}
