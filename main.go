package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Bowery/prompt"
	"github.com/bndr/gotabulate"
	"github.com/joliv/spark"
	stdn "github.com/traetox/speedtest/speedtestdotnet"
	//stdn "./speedtestdotnet" //for testing
)

const (
	tableFormat      = "simple"
	maxFailureCount  = 3
	initialTestCount = 5
	basePingCount    = 5
	fullTestCount    = 20
)

var (
	speedtestDuration = flag.Int("t", 3, "Target duration for speedtests (in seconds)")
)

func init() {
	flag.Parse()
	if *speedtestDuration <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid test duration")
		os.Exit(-1)
	}
}

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
		//get a latency from the server, the last latency will also be store in the
		//server structure
		if _, err := cfg.Servers[i].MedianPing(basePingCount); err != nil {
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
	fmt.Printf("Enter server ID for bandwidth test, or \"quit\" to exit\n")
	for {
		s, err := prompt.Basic("ID> ", true)
		if err != nil {
			fmt.Printf("input failure \"%v\"\n", err)
			os.Exit(-1)
		}
		//be REALLY forgiving on exit logic
		if strings.HasPrefix(strings.ToLower(s), "exit") {
			os.Exit(0)
		}

		//try to convert the string to a number
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\"%s\" is not a valid id\n", s)
			continue
		}
		if id > uint64(len(testServers)) {
			fmt.Fprintf(os.Stderr, "No server with ID \"%d\" available\n", id)
			continue
		}
		if err = fullTest(testServers[id]); err != nil {
			fmt.Fprintf(os.Stderr, "bandwidth test failed: %v\n", err)
			os.Exit(-1)
		} else {
			break //we are done
		}
	}
}

func testLatency(server stdn.Testserver) error {
	//perform a full latency test
	durs, err := server.Ping(fullTestCount)
	if err != nil {
		return err
	}
	var avg, max, min uint64
	var latencies []float64
	for i := range durs {
		ms := uint64(durs[i].Nanoseconds() / 1000000)
		latencies = append(latencies, float64(ms))
		avg += ms
		if ms > max {
			max = ms
		}
		if ms < min || min == 0 {
			min = ms
		}
	}
	avg = avg / uint64(len(durs))
	median := durs[len(durs)/2].Nanoseconds() / 1000000
	sparkline := spark.Line(latencies)
	fmt.Printf("Latency: %s\t%dms avg\t%dms median\t%dms max\t%dms min\n", sparkline, avg, median, max, min)
	return nil
}

func testDownstream(server stdn.Testserver) error {
	bps, err := server.Downstream(*speedtestDuration)
	if err != nil {
		return err
	}
	fmt.Printf("Download: %s\n", stdn.HumanSpeed(bps))
	return nil
}

func testUpstream(server stdn.Testserver) error {
	bps, err := server.Upstream(*speedtestDuration)
	if err != nil {
		return err
	}
	fmt.Printf("Upload:   %s\n", stdn.HumanSpeed(bps))
	return nil
}

func fullTest(server stdn.Testserver) error {
	if err := testLatency(server); err != nil {
		return err
	}
	if err := testDownstream(server); err != nil {
		return err
	}
	if err := testUpstream(server); err != nil {
		return err
	}
	return nil
}
