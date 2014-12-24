package speedtestdotnet

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxDownstreamTestCount = 4
	pingTimeout            = time.Second * 5
	speedTestTimeout       = time.Second * 10
	cmdTimeout             = time.Second
	latencyMaxTestCount    = 60
	dataBlockSize          = 16 * 1024 //16KB
)

var (
	errInvalidServerResponse = errors.New("Invalid server response")
	errPingFailure           = errors.New("Failed to complete ping test")
	errDontBeADick           = errors.New("requested ping count too high")
	targetTestDuration       = time.Second * 3 //5 seconds
	startBlockSize           = uint64(4096)    //4KB
	dataBlock                []byte
)

func init() {
	base := []byte("ABCDEFGHIJ")
	dataBlock = make([]byte, dataBlockSize)
	for i := range dataBlock {
		dataBlock[i] = base[i%len(base)]
	}
}

type durations []time.Duration

func (ts *Testserver) ping(count int) ([]time.Duration, error) {
	var errRet []time.Duration
	if count > latencyMaxTestCount {
		return errRet, errDontBeADick
	}
	//establish connection to the host
	conn, err := net.DialTimeout("tcp", ts.Host, pingTimeout)
	if err != nil {
		return errRet, err
	}
	defer conn.Close()

	durs := []time.Duration{}
	buff := make([]byte, 256)
	for i := 0; i < count; i++ {
		t := time.Now()
		fmt.Fprintf(conn, "PING %d\n", uint(t.UnixNano()/1000000))
		conn.SetReadDeadline(time.Now().Add(pingTimeout))
		n, err := conn.Read(buff)
		if err != nil {
			return errRet, err
		}
		conn.SetReadDeadline(time.Time{})
		d := time.Since(t)
		flds := strings.Fields(strings.TrimRight(string(buff[0:n]), "\n"))
		if len(flds) != 2 {
			return errRet, errInvalidServerResponse
		}
		if flds[0] != "PONG" {
			return errRet, errInvalidServerResponse
		}
		if _, err = strconv.ParseInt(flds[1], 10, 64); err != nil {
			return errRet, errInvalidServerResponse
		}
		durs = append(durs, d)
	}
	if len(durs) != count {
		return errRet, errPingFailure
	}
	return durs, nil
}

//MedianPing runs a latency test against the server and stores the median latency
func (ts *Testserver) MedianPing(count int) (time.Duration, error) {
	var errRet time.Duration
	durs, err := ts.ping(count)
	if err != nil {
		return errRet, err
	}
	sort.Sort(durations(durs))
	ts.Latency = durs[count/2]
	return durs[count/2], nil
}

//ping will run count number of latency tests and return the results of each
func (ts *Testserver) Ping(count int) ([]time.Duration, error) {
	return ts.ping(count)
}

//Upstream measures upstream bandwidth in bps
func (ts *Testserver) Upstream() (uint64, error) {
	return 0, errors.New("not ready")
}

//Downstream measures upstream bandwidth in bps
func (ts *Testserver) Downstream() (uint64, error) {
	var currBps uint64
	sz := startBlockSize
	conn, err := net.DialTimeout("tcp", ts.Host, speedTestTimeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	rdr := bufio.NewReader(conn)

	//we repeat the tests until we have a test that lasts at least N seconds
	for i := 0; i < maxDownstreamTestCount; i++ {
		//request a download of size sz and set a deadline
		if err = conn.SetWriteDeadline(time.Now().Add(cmdTimeout)); err != nil {
			return 0, err
		}
		fmt.Fprintf(conn, "DOWNLOAD %d\n", sz)
		if err = conn.SetWriteDeadline(time.Time{}); err != nil {
			return 0, err
		}

		ts := time.Now() //set start time mark
		if err = conn.SetReadDeadline(time.Now().Add(speedTestTimeout)); err != nil {
			return 0, err
		}
		//read until we get a newline
		if _, err = rdr.ReadBytes('\n'); err != nil {
			return 0, err
		}
		if err = conn.SetReadDeadline(time.Time{}); err != nil {
			return 0, err
		}
		//check if our test was a reasonable timespan
		dur := time.Since(ts)
		currBps = bps(sz, dur)
		if dur.Nanoseconds() > targetTestDuration.Nanoseconds() {
			return bps(sz, dur), nil
		}
		//test was too short, try again
		sz = calcNextSize(sz, dur)
	}
	return currBps, nil
}

func calcNextSize(b uint64, dur time.Duration) uint64 {
	if b == 0 {
		return startBlockSize
	}
	target := time.Second * 5
	return (b * uint64(target.Nanoseconds())) / uint64(dur.Nanoseconds())
}

//take the byte count and duration and calcuate a bits per second
func bps(byteCount uint64, dur time.Duration) uint64 {
	bits := byteCount * 8
	return uint64((bits * 1000000000) / uint64(dur.Nanoseconds()))
}

func (d durations) Len() int           { return len(d) }
func (d durations) Less(i, j int) bool { return d[i].Nanoseconds() < d[j].Nanoseconds() }
func (d durations) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
