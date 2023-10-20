/*
Copyright 2023 The Keyfactor Command Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package extensions

import (
	"os"
	"testing"
)

func GetGithubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func TestNewGithubReleaseFetcher(t *testing.T) {
	fetcher := NewGithubReleaseFetcher("", GetGithubToken())

	list, err := fetcher.GetExtensionList()
	if err != nil {
		t.Error(err)
	}

	if len(list) == 0 {
		t.Error("No extensions returned")
	}
}

func TestGithubReleaseFetcher_ExtensionExists(t *testing.T) {
	fetcher := NewGithubReleaseFetcher("", GetGithubToken())

	extension, err := fetcher.GetFirstExtension()
	if err != nil {
		t.Error(err)
	}

	versions, err := fetcher.GetExtensionVersions(extension)
	if err != nil {
		t.Error(err)
	}

	exists, err := fetcher.ExtensionExists(extension, versions[0])
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Error("Expected extension to exist")
	}
}

func TestGithubReleaseFetcher_Get(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		errorExpected bool
		object        any
	}{
		{
			name:          "ValidURL",
			url:           "https://api.github.com/orgs/keyfactor/repos?type=public&page=1&per_page=5",
			errorExpected: false,
			object:        []GithubRepo{},
		},
		{
			name:          "InvalidURL",
			url:           "https://api.git^hub.com/orgs/keyfactor/repos?type=public&page=1&per_page=5",
			errorExpected: true,
			object:        []GithubRepo{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := NewGithubReleaseFetcher("", GetGithubToken())

			err := f.Get(test.url, &test.object)
			if err != nil && !test.errorExpected {
				t.Errorf("Unexpected error: %s", err)
			}

			if err == nil && test.errorExpected {
				t.Error("Expected error, got none")
			}
		})
	}
}
