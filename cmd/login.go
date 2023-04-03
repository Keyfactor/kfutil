// Package cmd Copyright 2022 Keyfactor
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
// and limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Keyfactor/keyfactor-go-client/api"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"io"

	"log"
	"os"
	"os/signal"
	"strings"
)

const DefaultConfigFileName = "command_config.json"

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "User interactive login to Keyfactor. Stores the credentials in the config file '$HOME/.keyfactor/command_config.json'.",
	Long: `Will prompt the user for a username and password and then attempt to login to Keyfactor.
You can provide the --config flag to specify a config file to use. If not provided, the default
config file will be used. The default config file is located at $HOME/.keyfactor/command_config.json.
To prevent the prompt for username and password, use the --no-prompt flag. If this flag is provided then
the CLI will default to using the environment variables: KEYFACTOR_HOSTNAME, KEYFACTOR_USERNAME, 
KEYFACTOR_PASSWORD and KEYFACTOR_DOMAIN.

WARNING: The username and password will be stored in the config file in plain text at: 
'$HOME/.keyfactor/command_config.json.'
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(io.Discard)
		//log.SetOutput(os.Stdout)
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")

		authenticated := authConfigFile(configFile, noPrompt)
		if !authenticated {
			fmt.Println("Login failed.")
			log.Fatal("Unable to authenticate")
		}
		fmt.Println("Login successful!")
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}

func authConfigFile(configFile string, noPrompt bool) bool {
	config := make(map[string]string)
	userHomeDir, hErr := os.UserHomeDir()
	if configFile == "" {
		// Set up home directory config
		if hErr != nil {
			fmt.Println("Error getting user home directory: ", hErr)
		} else {
			userHomeDir = fmt.Sprintf("%s/.keyfactor", userHomeDir)
		}
		_, err := os.Stat(userHomeDir)

		if os.IsNotExist(err) {
			errDir := os.MkdirAll(userHomeDir, 0700)
			if errDir != nil {
				fmt.Println("Unable to create login config file. ", errDir)
				log.Printf("[ERROR] creating directory: %s", errDir)
			}
		}
		config = loadConfigFile(fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName), nil)
	} else {
		// Load config from specified file
		config = loadConfigFile(configFile, nil)
		return true
	}

	// Get the Keyfactor Command URL
	envHostName, hostSet := os.LookupEnv("KEYFACTOR_HOSTNAME")
	if !hostSet {
		log.Println("[INFO] Hostname not set. Please set the KEYFACTOR_HOSTNAME environment variable.")
	}
	var host string
	if noPrompt {
		//fmt.Println("Connecting to Keyfactor Command host: ", envHostName)
		if len(envHostName) == 0 {
			envHostName = config["host"]
		}
		host = envHostName
	} else {
		fmt.Printf("Enter Keyfactor Command host URL [%s]: \n", envHostName)
		_, phErr := fmt.Scanln(&host)
		if phErr != nil {
			if phErr.Error() != "unexpected newline" {
				fmt.Println("Error getting hostname: ", phErr)
				log.Println("[ERROR] getting hostname: ", phErr)
			}
		}
		if len(host) == 0 {
			host = envHostName
		}
	}
	ehErr := os.Setenv("KEYFACTOR_HOSTNAME", host)
	if ehErr != nil {
		fmt.Println("Error setting hostname: ", ehErr)
		log.Fatal("[ERROR] setting hostname: ", ehErr)
	}

	// Get the username
	envUserName, userSet := os.LookupEnv("KEYFACTOR_USERNAME")
	if !userSet {
		log.Println("[INFO] Username not set. Please set the KEYFACTOR_USERNAME environment variable.")
	}
	var username string
	if noPrompt {
		if len(envUserName) == 0 {
			envUserName = config["username"]
		}
		username = envUserName
	} else {
		fmt.Printf("Enter your Keyfactor Command username [%s]: \n", envUserName)
		_, puErr := fmt.Scanln(&username)
		if puErr != nil {
			if puErr.Error() != "unexpected newline" {
				fmt.Println("Error getting username: ", puErr)
				log.Println("[ERROR] getting username: ", puErr)
			}
		}
	}
	if len(username) == 0 {
		if len(envUserName) == 0 {
			envUserName = config["username"]
		}
		username = envUserName
	}
	euErr := os.Setenv("KEYFACTOR_USERNAME", username)
	if euErr != nil {
		fmt.Println("Error setting username: ", euErr)
		log.Fatal("[ERROR] setting username: ", euErr)
	}

	// Get the password.
	envPassword, passSet := os.LookupEnv("KEYFACTOR_PASSWORD")
	if !passSet {
		log.Println("[INFO] Password not set. Please set the KEYFACTOR_PASSWORD environment variable.")
	}
	var p string
	if noPrompt {
		if len(envPassword) == 0 {
			envPassword = config["password"]
		}
		p = envPassword
	} else {
		if len(envPassword) > 0 {
			p = getPassword("password: [<from env KEYFACTOR_PASSWORD>]")
		} else {
			p = getPassword("password: ")
		}

		if len(p) == 0 {
			p = envPassword
		}
	}
	epErr := os.Setenv("KEYFACTOR_PASSWORD", p)
	if epErr != nil {
		fmt.Println("Error setting password: ", epErr)
		log.Fatal("[ERROR] setting password: ", epErr)
	}

	// Get AD domain if not provided in the username or config file
	envDomain, domainSet := os.LookupEnv("KEYFACTOR_DOMAIN")
	var domain string
	if !domainSet {
		if strings.Contains(username, "@") {
			envDomain = strings.Split(username, "@")[1]
		} else if strings.Contains(username, "\\") {
			envDomain = strings.Split(username, "\\")[0]
		} else {
			log.Println("[INFO] Domain not set. Please set the KEYFACTOR_DOMAIN environment variable.")
		}
	}
	if noPrompt {
		//fmt.Println("Using domain: ", envDomain)
		if len(envDomain) == 0 {
			envDomain = config["domain"]
		}
		domain = envDomain
	} else {
		fmt.Printf("Enter your Keyfactor Command AD domain [%s]: \n", envDomain)
		_, sdErr := fmt.Scanln(&domain)
		if sdErr != nil {
			if sdErr.Error() != "unexpected newline" {
				fmt.Println("Error getting domain: ", sdErr)
				log.Println("[ERROR] getting domain: ", sdErr)
			}

		}
		if len(domain) == 0 {
			domain = envDomain
		}
	}
	edErr := os.Setenv("KEYFACTOR_DOMAIN", domain)
	if edErr != nil {
		fmt.Println("Error setting domain: ", edErr)
		log.Fatal("[ERROR] setting domain: ", edErr)
	}

	authConfig := api.AuthConfig{
		Hostname: host,
		Username: username,
		Password: p,
		Domain:   domain,
	}
	// Since there's no login command in the API, we'll just try to get the list of CAs
	_, kfcErr := api.NewKeyfactorClient(&authConfig)
	if kfcErr != nil {
		log.Println("[ERROR] initializing Keyfactor client: ", kfcErr)
		fmt.Println(fmt.Sprintf("Unable to initialize Keyfactor client. %s. Please check your configuration and try again.", kfcErr))
		return false
	}

	config["host"] = host
	config["username"] = username
	config["domain"] = domain
	config["password"] = p
	f, fErr := os.OpenFile(fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName), os.O_CREATE|os.O_RDWR, 0700)
	defer f.Close()
	if fErr != nil {
		fmt.Println("[ERROR] creating config file: ", fErr)
	}
	encoder := json.NewEncoder(f)
	enErr := encoder.Encode(&config)
	if enErr != nil {
		fmt.Println("Unable to read config file due to invalid format. ", enErr)
		log.Println("[ERROR] encoding config file: ", enErr)
	}
	return true
}

func getPassword(prompt string) string {
	// Get the initial state of the terminal.
	initialTermState, e1 := terminal.GetState(int(os.Stdin.Fd()))
	if e1 != nil {
		panic(e1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = terminal.Restore(int(os.Stdin.Fd()), initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Print(prompt)
	p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println("")
	if err != nil {
		panic(err)
	}

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	// Return the password as a string.
	return string(p)
}

func loadConfigFile(path string, filter func(map[string]interface{}) bool) map[string]string {
	data := make(map[string]string)

	f, _ := os.ReadFile(path)

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		log.Println("[ERROR] decoding config file: ", jErr)
	}

	//filteredData := []map[string]interface{}{}

	//for _, data := range data {
	//	// Do some filtering
	//	if filter(data) {
	//		filteredData = append(filteredData, data)
	//	}
	//}

	return data
}
