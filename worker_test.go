package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/assert"
)

type successfulUploadClient struct {
}

func (client *successfulUploadClient) Do(req *http.Request) (*http.Response, error) {
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	responseBody := fmt.Sprintf(`{
		"success": true,
		"errors": [],
		"messages": [],
		"result": {
			"script": "%s",
			"etag": "ea95132c15732412d22c1476fa83f27a",
			"size": 1024,
			"modified_on": "2017-01-01T00:00:00Z"
		}
	}`, string(bodyBytes))
	return &http.Response{
		Body: ioutil.NopCloser(bytes.NewReader([]byte(responseBody))),
	}, nil
}

type failedUploadClient struct {
}

func (client *failedUploadClient) Do(req *http.Request) (*http.Response, error) {
	responseBody := `{
		"success": false,
		"errors": [{
			"code": "10021",
			"message": "Script content failed validation checks, but was otherwise parseable"
		}],
		"messages": [],
		"result": {}
	}`
	return &http.Response{
		Body: ioutil.NopCloser(bytes.NewReader([]byte(responseBody))),
	}, nil
}

func TestDeploy(t *testing.T) {
	tests := []struct {
		client   httpExecutor
		expected error
	}{
		{
			client:   &successfulUploadClient{},
			expected: nil,
		},
		{
			client:   &failedUploadClient{},
			expected: errors.New("Script content failed validation checks, but was otherwise parseable"),
		},
	}

	for _, test := range tests {
		httpClient = test.client
		result := Deploy("", "", cloudflare.Zone{}, "./fixtures/worker.js")
		assert.Equal(t, test.expected, result)
	}
}

func TestGetUploadRequest(t *testing.T) {
	tests := []struct {
		authEmail string
		authKey   string
		zoneID    string
		url       string
		script    string
	}{
		{
			authEmail: "test@example.com",
			authKey:   "somekey",
			zoneID:    "nnEJBXpi3DnQze66Pej7wN0mA4gChXS5",
			url:       fmt.Sprintf("%s/zones/nnEJBXpi3DnQze66Pej7wN0mA4gChXS5/workers/script", cloudflareAPIURL),
			script:    "somescript",
		},
	}

	for _, test := range tests {
		result, err := getUploadRequest(test.authEmail, test.authKey, test.zoneID, test.script)
		if err != nil {
			t.Error()
		}
		body, _ := ioutil.ReadAll(result.Body)
		assert.Contains(t, test.url, result.URL.Path)
		assert.Equal(t, test.authEmail, result.Header.Get("X-Auth-Email"))
		assert.Equal(t, test.authKey, result.Header.Get("X-Auth-Key"))
		assert.Equal(t, "application/javascript", result.Header.Get("Content-Type"))
		assert.Equal(t, test.script, string(body))
	}
}

func TestGetZoneByName(t *testing.T) {
	file, _ := ioutil.ReadFile("./fixtures/cloudflare-zones-response.json")
	var zonesResponse cloudflare.ZonesResponse
	json.Unmarshal(file, &zonesResponse)
	tests := []struct {
		zones    []cloudflare.Zone
		name     string
		expected *cloudflare.Zone
	}{
		{
			zones:    zonesResponse.Result,
			name:     "website-one.com",
			expected: &zonesResponse.Result[0],
		},
		{
			zones:    zonesResponse.Result,
			name:     "website-two.com",
			expected: &zonesResponse.Result[1],
		},
		{
			zones:    zonesResponse.Result,
			name:     "none.com",
			expected: nil,
		},
	}

	for _, test := range tests {
		result := getZoneByName(test.name, test.zones)
		assert.Equal(t, result, test.expected)
	}
}
