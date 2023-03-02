package kustoutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
)

// WebHookClient holds the http client for the webhook
type WebHookClient struct {
	HttpClient *http.Client
}

// Query holds the query parameters for the webhook
type Query struct {
	Cluster string
	Label   string
}

// Response is the expected Webhook response
type Response struct {
	DBS []string `json:"dbs"`
}

// NewWebHookClient creates a new `WebHookClient`
func NewWebHookClient(httpClient *http.Client) *WebHookClient {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &WebHookClient{
		HttpClient: httpClient,
	}
}

// PerformQuery calls the webhook with the provided parameters
func (c *WebHookClient) PerformQuery(url, server, label string) ([]string, error) {
	a := Query{Cluster: server, Label: label}
	buf := &bytes.Buffer{}
	fmt.Printf("teamplte to use: %s", url)
	t := template.Must(template.New("t2").Parse(url))
	err := t.Execute(buf, a)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute the url query template - please review the template")
		return nil, err
	}
	r, err := http.NewRequest(http.MethodGet, buf.String(), nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate http request")
		return nil, err
	}
	resp, err := c.HttpClient.Do(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get db list from web-hool")
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("Unauthorized")
	}
	body, _ := io.ReadAll(resp.Body)
	res := Response{}
	err = json.Unmarshal(body, &res)
	return res.DBS, err
}
