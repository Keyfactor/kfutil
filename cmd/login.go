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
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/signal"
	"strings"
)

const DEFAULT_CONFIG_FILE_NAME = "command_config.json"

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "User interactive login to Keyfactor.",
	Long: `Will prompt the user for a username and password and then attempt to login to Keyfactor.
You can provide the --config flag to specify a config file to use. If not provided, the default
config file will be used. The default config file is located at $HOME/.keyfactor/command_config.json.
To prevent the prompt for username and password, use the --no-prompt flag. If this flag is provided then
the CLI will default to using the environment variables: KEYFACTOR_HOSTNAME, KEYFACTOR_USERNAME, 
KEYFACTOR_PASSWORD and KEYFACTOR_DOMAIN.
`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		userHomeDir, hErr := os.UserHomeDir()
		config := make(map[string]string)

		if configFile == "" {
			// Set up home directory config
			if hErr != nil {
				fmt.Println("[ERROR] getting user home directory: ", hErr)
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
			config = loadConfigFile(fmt.Sprintf("%s/%s", userHomeDir, DEFAULT_CONFIG_FILE_NAME), nil)
		} else {
			// Load config from specified file
			config = loadConfigFile(configFile, nil)
			return
		}

		// Get the Keyfactor Command URL
		hostName, hostSet := os.LookupEnv("KEYFACTOR_HOSTNAME")
		if !hostSet {
			log.Println("[INFO] Hostname not set. Please set the KEYFACTOR_HOSTNAME environment variable.")
		}
		var host string
		if noPrompt {
			fmt.Println("Connecting to Keyfactor Command host: ", hostName)
			host = hostName
		} else {
			fmt.Printf("Enter Keyfactor Command host URL [%s]: \n", hostName)
			fmt.Scanln(&host)
			if len(host) == 0 {
				host = hostName
			}
		}

		// Get the username
		envUserName, userSet := os.LookupEnv("KEYFACTOR_USERNAME")
		if !userSet {
			fmt.Println("[INFO] Username not set. Please set the KEYFACTOR_USERNAME environment variable.")
		}
		var username string
		if noPrompt {
			fmt.Println("Logging in with username: ", envUserName)
			username = envUserName
		} else {
			fmt.Printf("Enter your Keyfactor Command username [%s]: \n", envUserName)
			fmt.Scanln(&username)
		}
		if len(username) == 0 {
			username = envUserName
		}
		os.Setenv("KEYFACTOR_USERNAME", username)

		// Get the password.
		envPassword, passSet := os.LookupEnv("KEYFACTOR_PASSWORD")
		if !passSet {
			log.Println("[INFO] Password not set. Please set the KEYFACTOR_PASSWORD environment variable.")
		}
		var p string
		if noPrompt {
			p = envPassword
		} else {
			p = getPassword("password: [<from env KEYFACTOR_PASSWORD>]")
			if len(p) == 0 {
				p = envPassword
			}
		}
		os.Setenv("KEYFACTOR_PASSWORD", p)

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
			fmt.Println("Using domain: ", envDomain)
			domain = envDomain
		} else {
			fmt.Printf("Enter your Keyfactor Command AD domain [%s]: \n", envDomain)
			fmt.Scanln(&domain)
			if len(domain) == 0 {
				domain = envDomain
			}
		}
		os.Setenv("KEYFACTOR_DOMAIN", domain)

		// Since there's no login command in the API, we'll just try to get the list of CAs
		kfClient, kfcErr := initClient()
		if kfcErr != nil {
			fmt.Println("[ERROR] initializing Keyfactor client: ", kfcErr)
		}
		kfClient.GetCAList()

		config["host"] = host
		config["username"] = username
		config["domain"] = domain
		config["password"] = p
		file, fErr := os.OpenFile(fmt.Sprintf("%s/%s", userHomeDir, DEFAULT_CONFIG_FILE_NAME), os.O_CREATE|os.O_RDWR, 0700)
		defer file.Close()
		if fErr != nil {
			fmt.Println("[ERROR] creating config file: ", fErr)
		}
		encoder := json.NewEncoder(file)
		enErr := encoder.Encode(&config)
		if enErr != nil {
			fmt.Println("Unable to read config file due to invalid format. ", enErr)
			log.Println("[ERROR] encoding config file: ", enErr)
		}
		fmt.Println("Login successful!")
	},
}

func init() {
	var (
		configFile string
		noPrompt   bool
	)

	RootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.keyfactor/%s)")
	loginCmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "Do not prompt for username and password")
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

	json.Unmarshal(f, &data)

	//filteredData := []map[string]interface{}{}

	//for _, data := range data {
	//	// Do some filtering
	//	if filter(data) {
	//		filteredData = append(filteredData, data)
	//	}
	//}

	return data
}
