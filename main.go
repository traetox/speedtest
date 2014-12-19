package main

import (
	//"flag"
	"fmt"

	stdn "github.com/traetox/speedtest/speedtestdotnet"
)

func main() {
	cfg, err := stdn.GetConfig()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	if len(cfg.Servers) <= 0 {
		fmt.Printf("No acceptable servers found\n")
		return
	}
	//get the first 5 closest servers
	testServers := []stdn.Server{}
	for i := range cfg.Servers {
		if len(testServers) >= 5 {
			break
		}
		testServers = append(testServers, cfg.Servers[i])
	}

	fmt.Printf("5 Closest servers:\n")
	for i := range testServers {
		fmt.Printf("%s (%s) - %f Km\n", testServers[i].Name, testServers[i].Sponsor, testServers[i].Distance)
	}
}

func test() {
	servers, err := stdn.GetServerList()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for i := range servers {
		fmt.Printf("%s %f %f\n", servers[i].Name, servers[i].Lat, servers[i].Long)
	}
}
