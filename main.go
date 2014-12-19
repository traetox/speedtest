package main

import (
	"fmt"
	"os"

	"github.com/bndr/gotabulate"
	stdn "github.com/traetox/speedtest/speedtestdotnet"
)

const (
	tableFormat      = "simple"
	maxFailureCount  = 3
	initialTestCount = 5
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
	testServers := []stdn.Testserver{}
	failures := 0
	fmt.Printf("Gathering server list and testing...\n")
	for i := range cfg.Servers {
		if failures >= maxFailureCount {
			if len(testServers) > 0 {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to perform latency test\n")
			os.Exit(-1)
		}
		if len(testServers) >= initialTestCount {
			break
		}
		//get a latency from the server
		if err := cfg.Servers[i].Ping(); err != nil {
			failures++
			continue
		}
		testServers = append(testServers, cfg.Servers[i])
	}

	fmt.Printf("%d Closest responding servers:\n", len(testServers))
	data := [][]string{}
	for i := range testServers {
		data = append(data, []string{fmt.Sprintf("%d", i),
			testServers[i].Name, testServers[i].Sponsor,
			fmt.Sprintf("%.02f", testServers[i].Distance),
			fmt.Sprintf("%s", testServers[i].Latency)})
	}
	t := gotabulate.Create(data)
	t.SetHeaders([]string{"ID", "Name", "Sponsor", "Distance (km)", "Latency (ms)"})
	t.SetWrapStrings(false)
	fmt.Printf("%s", t.Render(tableFormat))

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
