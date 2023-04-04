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
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/signal"
	"strings"
)

const DefaultConfigFileName = "command_config.json"
const DefaultConfigurationFileName = "kfcmd_config.json"

type ConfigurationFile struct {
	Servers map[string]ConfigurationFileEntry `json:"servers"`
}

type ConfigurationFileEntry struct {
	Hostname string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
}

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
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Global flags
		debugFlag, _ := cmd.Flags().GetBool("debug")
		configFile, _ := cmd.Flags().GetString("config")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")
		profile, _ := cmd.Flags().GetString("profile")

		debugModeEnabled := checkDebug(debugFlag)
		log.Println("Debug mode enabled: ", debugModeEnabled)

		authenticated := authConfigFile(configFile, noPrompt, profile)
		if !authenticated {
			fmt.Println("Login failed.")
			log.Fatal("[FATAL] Unable to authenticate")
		}
		fmt.Println(fmt.Sprintf("Login successful!"))
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}

func authConfigFile(configFile string, noPrompt bool, profile string) bool {
	var config map[string]string
	var configurationFile ConfigurationFile

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
		//config, _ = loadConfigFile(fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName), true)
		configurationFile, _ = loadConfigurationFile(fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName), true)
		configFile = fmt.Sprintf("%s/%s", userHomeDir, DefaultConfigFileName)
	} else {
		// Load config from specified file
		//config, _ = loadConfigFile(configFile, false)
		configurationFile, _ = loadConfigurationFile(configFile, false)
		return true
	}

	if config == nil {
		config = make(map[string]string)
	}

	if configurationFile.Servers == nil {
		configurationFile.Servers = make(map[string]ConfigurationFileEntry)
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
	var userDomain string
	var configDomain string
	if !domainSet {
		if strings.Contains(username, "@") {
			userDomain = strings.Split(username, "@")[1]
		} else if strings.Contains(username, "\\") {
			userDomain = strings.Split(username, "\\")[0]
		} else {
			configDomain = config["domain"]
		}
	}
	if noPrompt {
		//fmt.Println("Using domain: ", envDomain)
		if len(envDomain) == 0 && len(userDomain) == 0 && len(configDomain) == 0 {

			fmt.Println("Domain not set and unable to be inferred. Please set the KEYFACTOR_DOMAIN environment variable.")
			log.Fatal("[FATAL] Domain not set. Please set the KEYFACTOR_DOMAIN environment variable.")
			return false
		}
		if len(configDomain) == 0 {
			if len(userDomain) > 0 {
				log.Println("[INFO] Domain not set. Using domain from username ", userDomain)
				domain = userDomain
			} else if len(envDomain) > 0 {
				log.Println("[INFO] Domain not set. Using domain from environment variable KEYFACTOR_DOMAIN", envDomain)
				domain = envDomain
			}
		}
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

	configuration := ConfigurationFileEntry{
		Hostname: host,
		Username: username,
		Password: p,
		Domain:   domain,
	}

	//confErr := createConfigFile(config, configFile)
	configurationErr := createConfigurationFile(configuration, profile, configFile)
	if configurationErr != nil {
		log.Fatal("[FATAL] Login failed due to an issue with the configuration file: ", configurationErr)
	}
	//if confErr != nil {
	//	log.Fatal("[FATAL] Login failed due ot an issue with the config file: ", confErr)
	//}

	return true
}

func createConfigFile(kfcfg map[string]string, configPath string) error {
	log.Println("[INFO] creating kfcfg file")
	log.Println("[DEBUG] kfcfg path: ", configPath)

	entry := ConfigurationFileEntry{
		Hostname: kfcfg["host"],
		Username: kfcfg["username"],
		Password: kfcfg["password"],
		Domain:   kfcfg["domain"],
	}

	log.Println("[DEBUG] kfcfg entry: ", entry)

	f, fErr := os.OpenFile(fmt.Sprintf("%s", configPath), os.O_CREATE|os.O_RDWR, 0600)
	defer f.Close()
	if fErr != nil {
		fmt.Println("Error creating kfcfg file: ", fErr)
		log.Println("[ERROR] creating kfcfg file: ", fErr)
		return fErr
	}
	encoder := json.NewEncoder(f)
	enErr := encoder.Encode(&kfcfg)
	if enErr != nil {
		fmt.Println("Unable to read kfcfg file due to invalid format. ", enErr)
		log.Println("[ERROR] encoding kfcfg file: ", enErr)
		return enErr
	}
	return nil
}

func createConfigurationFile(cfgFile ConfigurationFileEntry, configName string, configPath string) error {
	log.Println("[INFO] creating kfcfg file")
	log.Println("[DEBUG] kfcfg path: ", configPath)

	if len(configName) == 0 {
		configName = "default"
	}

	existingConfig, exsErr := loadConfigurationFile(configPath, true)
	if exsErr != nil {
		log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", configName))
		existingConfig.Servers = make(map[string]ConfigurationFileEntry)
		existingConfig.Servers[configName] = cfgFile
	} else if len(existingConfig.Servers) > 0 {
		// check if the config name already exists
		if _, ok := existingConfig.Servers[configName]; ok {
			log.Println(fmt.Sprintf("[WARN] config name '%s' already exists. Overwriting existing config.", configName))
			log.Println(fmt.Sprintf("[DEBUG] existing config: %v", existingConfig.Servers[configName]))
			log.Println(fmt.Sprintf("[DEBUG] new config: %v", cfgFile))
			// print out the diff between the two configs
			diff := cmp.Diff(existingConfig.Servers[configName], cfgFile)
			log.Println(fmt.Sprintf("[DEBUG] diff: %s", diff))
		} else {
			log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", configName))
		}
		existingConfig.Servers[configName] = cfgFile
	} else {
		log.Println(fmt.Sprintf("[INFO] adding new config name '%s'", configName))
		existingConfig.Servers = make(map[string]ConfigurationFileEntry)
		existingConfig.Servers[configName] = cfgFile
	}

	log.Println("[DEBUG] kfcfg entry: ", cfgFile)

	f, fErr := os.OpenFile(fmt.Sprintf("%s", configPath), os.O_CREATE|os.O_RDWR, 0600)
	defer f.Close()
	if fErr != nil {
		fmt.Println("Error creating kfcfg file: ", fErr)
		log.Println("[ERROR] creating kfcfg file: ", fErr)
		return fErr
	}
	encoder := json.NewEncoder(f)
	enErr := encoder.Encode(&existingConfig)
	if enErr != nil {
		fmt.Println("Unable to read kfcfg file due to invalid format. ", enErr)
		log.Println("[ERROR] encoding kfcfg file: ", enErr)
		return enErr
	}
	return nil
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

func loadConfigFile(path string, silent bool) (map[string]string, error) {
	data := make(map[string]string) // todo: make this a struct and support multiple configs in a single file

	f, rFErr := os.ReadFile(path)
	if rFErr != nil {
		if !silent {
			fmt.Println(fmt.Sprintf("Unable to read config file '%s'.", rFErr))
			log.Fatal("[FATAL] Error reading config file: ", rFErr)
		}
		return nil, rFErr
	}

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		log.Println("[ERROR] decoding config file: ", jErr)
	}

	return data, nil
}

func loadConfigurationFile(path string, silent bool) (ConfigurationFile, error) {

	//data := ConfigurationFile{Servers: make(map[string]ConfigurationFileEntry)}
	data := ConfigurationFile{}
	f, rFErr := os.ReadFile(path)
	if rFErr != nil {
		if !silent {
			fmt.Println(fmt.Sprintf("Unable to read config file '%s'.", rFErr))
			log.Fatal("[FATAL] Error reading config file: ", rFErr)
		}
		return data, rFErr
	}

	jErr := json.Unmarshal(f, &data)
	if jErr != nil {
		//fmt.Println("Unable to read config file due to invalid format. ", jErr)
		log.Println("[ERROR] decoding config file: ", jErr)
	}

	return data, nil
}
