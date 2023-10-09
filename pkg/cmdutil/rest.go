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

import (
	"fmt"
	"io"
	"net/http"
)

type SimpleRestClient struct {
	bearerToken string
	client      *http.Client
}

func NewSimpleRestClient() *SimpleRestClient {
	return &SimpleRestClient{
		client: http.DefaultClient,
	}
}

func (c *SimpleRestClient) SetBearerToken(token string) {
	c.bearerToken = token
}

func (c *SimpleRestClient) Get(url string) ([]byte, error) {
	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add Authorization header
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	// Send the request
	get, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if get.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", get.StatusCode)
	}

	// Read the body of the response
	body, err := io.ReadAll(get.Body)
	if err != nil {
		return nil, err
	}

	// Return the body
	return body, nil
}
