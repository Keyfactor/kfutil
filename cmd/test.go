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
