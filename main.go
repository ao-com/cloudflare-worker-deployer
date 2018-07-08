package main

import (
	"flag"
	"fmt"
	"net/http"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func main() {
	var authEmail string
	flag.StringVar(&authEmail, "email", "authemail@example.com", "cloudflare auth email")
	var authKey string
	flag.StringVar(&authKey, "key", "someauthkey", "cloudflare auth key")
	var zoneName string
	flag.StringVar(&zoneName, "zone", "somewebsite.com", "zone to deploy the worker code to")
	var scriptLocation string
	flag.StringVar(&scriptLocation, "script-location", "./worker.js", "cloudflare worker script location")
	flag.Parse()
	httpClient = &http.Client{}
	api, err := cloudflare.New(authKey, authEmail)
	if err != nil {
		panic(err)
	}
	zones, err := api.ListZones(zoneName)
	if err != nil {
		panic(err)
	}
	zone := getZoneByName(zoneName, zones)
	if zone == nil {
		panic(fmt.Errorf("couldn't find zone %s", zoneName))
	}
	err = Deploy(authEmail, authKey, *zone, scriptLocation)
	if err != nil {
		panic(err)
	}
}
