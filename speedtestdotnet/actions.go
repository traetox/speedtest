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

package speedtestdotnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxDownstreamTestCount = 4
	maxTransferSize        = 8 * 1024 * 1024
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
	startBlockSize           = uint64(4096) //4KB
	dataBlock                []byte

	ErrTimeout = errors.New("Timeout")
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
		return errRet, ErrTimeout
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

//Ping will run count number of latency tests and return the results of each
func (ts *Testserver) Ping(count int) ([]time.Duration, error) {
	return ts.ping(count)
}

//throwBytes chucks bytes at the remote server then listens for a response
func throwBytes(conn io.ReadWriter, count uint64) error {
	var writeBytes uint64
	var b []byte
	buff := make([]byte, 128)
	for writeBytes < count {
		if (count - writeBytes) >= uint64(len(dataBlock)) {
			b = dataBlock
		} else {
			b = dataBlock[0:(count - writeBytes)]
		}
		n, err := conn.Write(b)
		if err != nil {
			return err
		}
		writeBytes += uint64(n)
	}
	//read the response
	n, err := conn.Read(buff)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("Failed to get OK on upload")
	}
	if !strings.HasPrefix(string(buff[0:n]), "OK ") {
		return fmt.Errorf("Failed to get OK on upload")
	}
	return nil
}

//readBytes reads until we get a newline or an error
func readBytes(rdr io.Reader, count uint64) error {
	var rBytes uint64
	buff := make([]byte, 4096)
	for rBytes < count {
		n, err := rdr.Read(buff)
		if err != nil {
			return err
		}
		rBytes += uint64(n)
		if n == 0 {
			break
		}
		if buff[n-1] == '\n' {
			break
		}
	}
	if rBytes != count {
		return fmt.Errorf("Failed entire read: %d != %d", rBytes, count)
	}
	return nil
}

//Upstream measures upstream bandwidth in bps
func (ts *Testserver) Upstream(duration int) (uint64, error) {
	var currBps uint64
	sz := startBlockSize
	conn, err := net.DialTimeout("tcp", ts.Host, speedTestTimeout)
	if err != nil {
		return 0, ErrTimeout
	}
	targetTestDuration := time.Second * time.Duration(duration)
	defer conn.Close()

	//we repeat the tests until we have a test that lasts at least N seconds
	for i := 0; i < maxDownstreamTestCount; i++ {
		//request a download of size sz and set a deadline
		if err = conn.SetWriteDeadline(time.Now().Add(cmdTimeout)); err != nil {
			return 0, err
		}
		cmdStr := fmt.Sprintf("UPLOAD %d 0\n", sz)
		if _, err := conn.Write([]byte(cmdStr)); err != nil {
			return 0, err
		}
		if err = conn.SetWriteDeadline(time.Time{}); err != nil {
			return 0, err
		}

		ts := time.Now() //set start time mark
		if err = conn.SetWriteDeadline(time.Now().Add(speedTestTimeout)); err != nil {
			return 0, err
		}
		if err := throwBytes(conn, sz-uint64(len(cmdStr))); err != nil {
			return 0, err
		}
		if err = conn.SetReadDeadline(time.Time{}); err != nil {
			return 0, err
		}
		//check if our test was a reasonable timespan
		dur := time.Since(ts)
		currBps = bps(sz, dur)
		if dur.Nanoseconds() > targetTestDuration.Nanoseconds() || sz == maxTransferSize {
			_, err = fmt.Fprintf(conn, "QUIT\n")
			return bps(sz, dur), err
		}
		//test was too short, try again
		sz = calcNextSize(sz, dur)
		if sz > maxTransferSize {
			sz = maxTransferSize
		}
	}

	_, err = fmt.Fprintf(conn, "QUIT\n")
	return currBps, err
}

//Downstream measures upstream bandwidth in bps
func (ts *Testserver) Downstream(duration int) (uint64, error) {
	var currBps uint64
	sz := startBlockSize
	conn, err := net.DialTimeout("tcp", ts.Host, speedTestTimeout)
	if err != nil {
		return 0, ErrTimeout
	}
	defer conn.Close()

	targetTestDuration := time.Second * time.Duration(duration)
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
		if err = readBytes(conn, sz); err != nil {
			return 0, err
		}
		if err = conn.SetReadDeadline(time.Time{}); err != nil {
			return 0, err
		}
		//check if our test was a reasonable timespan
		dur := time.Since(ts)
		currBps = bps(sz, dur)
		if dur.Nanoseconds() > targetTestDuration.Nanoseconds() || sz == maxTransferSize {
			_, err = fmt.Fprintf(conn, "QUIT\n")
			return bps(sz, dur), err
		}
		//test was too short, try again
		sz = calcNextSize(sz, dur)
		if sz > maxTransferSize {
			sz = maxTransferSize
		}
	}

	_, err = fmt.Fprintf(conn, "QUIT\n")
	return currBps, err
}

//calcNextSize takes the current preformance metrics and
//attempts to calculate what the next size should be
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
