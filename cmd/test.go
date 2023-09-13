package cmd

import (
	"bytes"
	"io"
	"os"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

type testEnv struct {
	CommandHostname            string
	CommandUsername            string
	CommandDomain              string
	CommandPassword            string
	CommandAPIPath             string
	CommandConfig              string
	CommandProfile             string
	CommandAuthProvider        string
	CommandAuthProviderProfile string
	CommandExpEnabled          string
}

func getTestEnv() (testEnv, error) {
	commandHostname := os.Getenv("KEYFACTOR_HOSTNAME")
	commandUsername := os.Getenv("KEYFACTOR_USERNAME")
	commandDomain := os.Getenv("KEYFACTOR_DOMAIN")
	commandPassword := os.Getenv("KEYFACTOR_PASSWORD")
	//command_api_path := os.Getenv("KEYFACTOR_API_PATH")
	commandConfig := os.Getenv("KEYFACTOR_CONFIG")
	commandProfile := os.Getenv("KEYFACTOR_PROFILE")

	testEnv := testEnv{
		CommandHostname: commandHostname,
		CommandUsername: commandUsername,
		CommandDomain:   commandDomain,
		CommandPassword: commandPassword,
		//CommandApiPath:             commandApiPath,
		CommandConfig:  commandConfig,
		CommandProfile: commandProfile,
		//CommandAuthProvider:        commandAuthProvider,
		//CommandAuthProviderProfile: commandAuthProviderProfile,
		//CommandExpEnabled:          commandExpEnabled,
	}

	return testEnv, nil

}
