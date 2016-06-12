package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	endpointWorkspaces     = "https://www.toggl.com/api/v8/workspaces"
	endpointClients        = "https://www.toggl.com/api/v8/clients"
	endpointReportWeekly   = "https://toggl.com/reports/api/v2/weekly"
	endpointReportDetailed = "https://toggl.com/reports/api/v2/details"
	endpointReportSummary  = "https://toggl.com/reports/api/v2/summary"
	endpointStartTime      = "https://www.toggl.com/api/v8/time_entries/start"

	// APISecret is specified from toggl
	apiSecret       = "api_token"
	contentTypeJSON = "application/json"
	userAgent       = "toggl-go/0.1"
)

// APIKey store API token
type APIKey struct {
	Token  string
	Secret string
}

// Resources is slice of endpoint
type Resources map[string]Endpoint

// AddEndpoint is the method for add an API endpoint.
func (r *Resources) AddEndpoint(name string, endpoint Endpoint) error {
	_, ok := (*r)[name]
	if ok {
		return fmt.Errorf("%s is already used.\n", name)
	}
	(*r)[name] = endpoint
	return nil
}

// GetURL return API endpoint url.URL of given name
func (r *Resources) GetURL(name string) (*url.URL, error) {
	endpoint, ok := (*r)[name]
	if !ok {
		return nil, fmt.Errorf("%s is not registered as a resource.\n", name)
	}
	return endpoint.URL(), nil
}

// Endpoint represent each REST endpoint
type Endpoint interface {
	URLString() string
	URL() *url.URL
}

// Client store basic information for use toggl API
type Client struct {
	resources   *Resources
	apiKey      *APIKey
	contentType string
	userAgent   string
}

// NewClient return a Client instance if not return error
func NewClient(apiKey *APIKey, resources *Resources) (*Client, error) {
	return &Client{
		resources:   resources,
		apiKey:      apiKey,
		contentType: contentTypeJSON,
		userAgent:   userAgent,
	}, nil
}

func (c *Client) buildURL(resource string) (*url.URL, error) {
	return c.resources.GetURL(resource)
}

func (c *Client) buildRequest(method, path string, body io.Reader) (req *http.Request, err error) {
	endpoint, err := c.buildURL(path)
	if err != nil {
		return
	}
	req, err = http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return
	}

	req.SetBasicAuth(c.apiKey.Token, c.apiKey.Secret)
	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("Content-Type", c.contentType)
	return
}

func (c *Client) request(req *http.Request, body interface{}) (err error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		body := struct {
			Error errorResponse `json:"error"`
		}{}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&body)
		if err != nil {
			return errorResponse{
				Code:    resp.StatusCode,
				Message: resp.Status,
			}
		}

		return body.Error
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&body)
	return
}

func (c *Client) encodeJSON(object interface{}) (reader io.Reader, err error) {
	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	err = encoder.Encode(object)
	if err != nil {
		return
	}

	reader = buffer
	return
}

// GetRequest sends GET request
func (c *Client) GetRequest(name string) (err error) {
	url, err := c.buildURL(name)
	if err != nil {
		return
	}
	req, err := c.buildRequest("GET", url.Path, nil)
	if err != nil {
		return
	}
	err = c.request(req, nil)
	if err != nil {
		return
	}
	return
}
