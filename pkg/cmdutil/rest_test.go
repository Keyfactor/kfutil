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
