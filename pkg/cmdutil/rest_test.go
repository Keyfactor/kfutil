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

package cmdutil

import "testing"

func TestNewSimpleRestClient(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		bearer        string
		errorExpected bool
	}{
		{
			name:          "TestGet",
			url:           "https://www.google.com",
			bearer:        "",
			errorExpected: false,
		},
		{
			name:          "TestGetWithBearer",
			url:           "https://www.google.com",
			bearer:        "1234567890",
			errorExpected: false,
		},
		{
			name:          "TestInvalidUrl",
			url:           "http://exa^mple.com",
			errorExpected: true,
		},
		{
			name:          "TestNonExistentUrl",
			url:           "https://www.thiswillprobablybe404.com",
			errorExpected: true,
		},
		{
			name:          "TestNon200Response",
			url:           "https://www.google.com/404",
			bearer:        "",
			errorExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewSimpleRestClient()
			c.SetBearerToken(test.bearer)
			_, err := c.Get(test.url)
			if err != nil && !test.errorExpected {
				t.Errorf("Error not expected: %v", err)
			}
			if err == nil && test.errorExpected {
				t.Errorf("Error expected but not found")
			}
		})
	}
}
