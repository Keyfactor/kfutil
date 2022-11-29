/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const DEFAULT_CONFIG_FILE_NAME = "command_config.json"

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	initialTermState, e1 := terminal.GetState(syscall.Stdin)
	if e1 != nil {
		panic(e1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = terminal.Restore(syscall.Stdin, initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Print(prompt)
	p, err := terminal.ReadPassword(syscall.Stdin)
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
