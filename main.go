// The MIT License (MIT)

// Copyright (c) 2014, 2016 traetox

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Bowery/prompt"
	"github.com/bndr/gotabulate"
	"github.com/joliv/spark"

	stdn "github.com/traetox/speedtest/speedtestdotnet"
	"github.com/traetox/speedtest/version"
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
	search            = flag.String("s", "", "Server name substring to search candidate servers")
	auto              = flag.Bool("a", false, "Auto-select nearest candidate server")
	vrs               bool
)

func init() {
	flag.BoolVar(&vrs, "version", false, "print version and exit")
	flag.BoolVar(&vrs, "v", false, "print version and exit (shorthand)")
	flag.Parse()
	if vrs {
		fmt.Printf("Speedtest v%s\n", version.Version)
		os.Exit(0)
	}
	if *speedtestDuration <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid test duration")
		os.Exit(-1)
	}
}

func main() {
	cfg, err := stdn.GetConfig()
	if err != nil {
		fmt.Printf("Failed to get server list configuration: %v\n", err)
		os.Exit(-1)
	}
	if len(cfg.Servers) <= 0 {
		fmt.Printf("No acceptable servers found\n")
		os.Exit(-1)
	}
	var headers []string
	var data [][]string
	var testServers []stdn.Testserver
	if *search == "" {
		fmt.Printf("Gathering server list and testing...\n")
		if testServers, err = autoGetTestServers(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(-1)
		}
		fmt.Printf("%d Closest responding servers:\n", len(testServers))
		for i := range testServers {
			data = append(data, []string{fmt.Sprintf("%d", i),
				testServers[i].Name, testServers[i].Sponsor,
				fmt.Sprintf("%.02f", testServers[i].Distance),
				fmt.Sprintf("%s", testServers[i].Latency)})
		}
		headers = []string{"ID", "Name", "Sponsor", "Distance (km)", "Latency (ms)"}
	} else {
		if testServers, err = getSearchServers(cfg, *search); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(-1)
		}
		headers = []string{"ID", "Name", "Sponsor", "Distance (km)"}
		fmt.Printf("%d Matching servers:\n", len(testServers))
		for i := range testServers {
			data = append(data, []string{fmt.Sprintf("%d", i),
				testServers[i].Name, testServers[i].Sponsor,
				fmt.Sprintf("%.02f", testServers[i].Distance)})
		}

	}
	t := gotabulate.Create(data)
	t.SetHeaders(headers)
	t.SetWrapStrings(false)
	fmt.Printf("%s", t.Render(tableFormat))

	// Define server variable to be used for either auto-selection or manual selection
	var selServer stdn.Testserver

	if *auto {
		// Double check the existence of a server again to avoid out-of bound panic
		if len(testServers) > 0 {
			selServer = testServers[0]
		} else {
			fmt.Println("No servers found")
			os.Exit(-1)
		}
		fmt.Printf("\nAuto-selecting closest server for bandwidth test: %s / %s\n", selServer.Name, selServer.Sponsor)
	} else {
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
			if strings.HasPrefix(strings.ToLower(s), "quit") {
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

			// We are done, set ID accordingly
			selServer = testServers[id]
			break
		}
	}

	// Perform the actual test
	if err = fullTest(selServer); err != nil {
		switch err {
		case io.EOF:
			fmt.Fprintf(os.Stderr, "Error, the remote server kicked us.\n")
			fmt.Fprintf(os.Stderr, "Maximum request size may have changed\n")
		case stdn.ErrTimeout:
			fmt.Fprintf(os.Stderr, "Test failed due to connection timeout.  The server may be down, or rejecting us")
		default:
			fmt.Fprintf(os.Stderr, "Test failed with unknown error: %v\n", err)
		}
		os.Exit(-1)
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

func autoGetTestServers(cfg *stdn.Config) ([]stdn.Testserver, error) {
	//get the first 5 closest servers
	testServers := []stdn.Testserver{}
	failures := 0
	for i := range cfg.Servers {
		if failures >= maxFailureCount {
			if len(testServers) > 0 {
				return testServers, nil
			}
			return nil, fmt.Errorf("Failed to perform latency test on closest servers\n")
		}
		if len(testServers) >= initialTestCount {
			return testServers, nil
		}
		//get a latency from the server, the last latency will also be store in the
		//server structure
		if _, err := cfg.Servers[i].MedianPing(basePingCount); err != nil {
			failures++
			continue
		}
		testServers = append(testServers, cfg.Servers[i])
	}
	return testServers, nil
}

func getSearchServers(cfg *stdn.Config, query string) ([]stdn.Testserver, error) {
	//get the first 5 closest servers
	testServers := []stdn.Testserver{}
	for i := range cfg.Servers {
		if strings.Contains(strings.ToLower(cfg.Servers[i].Name), strings.ToLower(query)) {
			testServers = append(testServers, cfg.Servers[i])
		}
	}
	if len(testServers) == 0 {
		return nil, errors.New("no servers found")
	}
	return testServers, nil
}
