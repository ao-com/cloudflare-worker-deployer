package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
)

var (
	httpClient       httpExecutor
	cloudflareAPIURL = "https://api.cloudflare.com/client/v4"
)

type httpExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

func getUploadRequest(authEmail string, authKey string, zoneID string, script string) (*http.Request, error) {
	url := fmt.Sprintf("%s/zones/%s/workers/script", cloudflareAPIURL, zoneID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(script)))
	if err != nil {
		return &http.Request{}, err
	}
	req.Header.Add("Content-Type", "application/javascript")
	req.Header.Add("X-Auth-Email", authEmail)
	req.Header.Add("X-Auth-Key", authKey)
	return req, nil
}

func getZoneByName(name string, zones []cloudflare.Zone) *cloudflare.Zone {
	for _, zone := range zones {
		if name == zone.Name {
			return &zone
		}
	}
	return nil
}

// Deploy deploys the Cloudflare Worker script for a given zone
func Deploy(authEmail string, authKey string, zone cloudflare.Zone, scriptLocation string) error {
	file, err := ioutil.ReadFile(scriptLocation)
	if err != nil {
		panic(err)
	}
	req, err := getUploadRequest(authEmail, authKey, zone.ID, string(file))
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response cloudflare.Response
	json.Unmarshal(respBytes, &response)
	if !response.Success {
		return errors.New(response.Errors[0].Message)
	}
	return nil
}
