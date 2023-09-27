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
