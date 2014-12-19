package speedtestdotnet

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	pingTimeout      = time.Second * 5
	latencyTestCount = 9
)

var (
	errInvalidServerResponse = errors.New("Invalid server response")
	errPingFailure           = errors.New("Failed to complete ping test")
)

type durations []time.Duration

//Ping runs a latency test against the server and stores the median latency
func (ts *Testserver) Ping() error {
	//establish connection to the host
	conn, err := net.DialTimeout("tcp", ts.Host, pingTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	durs := []time.Duration{}
	buff := make([]byte, 256)
	for i := 0; i < latencyTestCount; i++ {
		t := time.Now()
		fmt.Fprintf(conn, "PING %d\n", uint(t.UnixNano()/1000000))
		conn.SetReadDeadline(time.Now().Add(pingTimeout))
		n, err := conn.Read(buff)
		if err != nil {
			return err
		}
		conn.SetReadDeadline(time.Time{})
		d := time.Since(t)
		flds := strings.Fields(strings.TrimRight(string(buff[0:n]), "\n"))
		if len(flds) != 2 {
			return errInvalidServerResponse
		}
		if flds[0] != "PONG" {
			return errInvalidServerResponse
		}
		if _, err = strconv.ParseInt(flds[1], 10, 64); err != nil {
			return errInvalidServerResponse
		}
		durs = append(durs, d)
	}
	if len(durs) != latencyTestCount {
		return errPingFailure
	}
	sort.Sort(durations(durs))
	ts.Latency = durs[1]
	return nil
}

func (d durations) Len() int           { return len(d) }
func (d durations) Less(i, j int) bool { return d[i].Nanoseconds() < d[j].Nanoseconds() }
func (d durations) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
