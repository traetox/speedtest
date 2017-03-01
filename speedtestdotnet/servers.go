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
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kellydunn/golang-geo"
)

const (
	serversConfigUrl string        = `http://www.speedtest.net/speedtest-servers-static.php?x=whysosad`
	clientConfigUrl  string        = `http://www.speedtest.net/speedtest-config.php`
	getTimeout       time.Duration = 2 * time.Second
)

type Testserver struct {
	Name     string
	Sponsor  string
	Country  string
	Lat      float64
	Long     float64
	Distance float64 //distance from server in KM
	URLs     []string
	Host     string
	Latency  time.Duration //latency in ms
}
type testServerlist []Testserver

type Config struct {
	LicenseKey string
	IP         net.IP
	Lat        float64
	Long       float64
	ISP        string
	Servers    []Testserver
}

type sconfig struct {
	XMLName   xml.Name `xml:"server-config"`
	Threads   int      `xml:"threadcount,attr"`
	IgnoreIDs string   `xml:"ignoreids,attr"`
}

type cconfig struct {
	XMLName  xml.Name `xml:"client"`
	Ip       string   `xml:"ip,attr"`
	Lat      float64  `xml:"lat,attr"`
	Long     float64  `xml:"lon,attr"`
	ISP      string   `xml:"isp,attr"`
	ISPUpAvg uint     `xml:"ispulavg,attr"`
	ISPDlAvg uint     `xml:"ispdlavg,attr"`
}

type speedtestConfig struct {
	XMLName      xml.Name `xml:"settings"`
	License      string   `xml:"licensekey"`
	ClientConfig cconfig  `xml:"client"`
	ServerConfig sconfig  `xml:"server-config"`
}

type server struct {
	XMLName xml.Name `xml:"server"`
	Url     string   `xml:"url,attr"`
	Url2    string   `xml:"url2,attr"`
	Lat     float64  `xml:"lat,attr"`
	Long    float64  `xml:"lon,attr"`
	Name    string   `xml:"name,attr"`
	Country string   `xml:"country,attr"`
	CC      string   `xml:"cc,attr"`
	Sponsor string   `xml:"sponsor,attr"`
	ID      uint     `xml:"id,attr"`
	Host    string   `xml:"host,attr"`
}

type settings struct {
	XMLName xml.Name `xml:"settings"`
	Servers []server `xml:"servers>server"`
}

//GetServerList returns a list of servers in the native speedtest.net structure
func GetServerList() ([]server, error) {
	//get a list of servers
	clnt := http.Client{
		Timeout: getTimeout,
	}
	//get the server configs
	req, err := http.NewRequest("GET", serversConfigUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1")
	resp, err := clnt.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		x, _ := ioutil.ReadAll(resp.Body)
		println(string(x))
		return nil, fmt.Errorf("Invalid status %d", resp.StatusCode)
	}
	xmlDec := xml.NewDecoder(resp.Body)
	sts := settings{}
	if err := xmlDec.Decode(&sts); err != nil {
		return nil, err
	}
	return sts.Servers, nil
}

//GetConfig returns a configuration containing information about our client and a list of acceptable servers sorted by distance
func GetConfig() (*Config, error) {
	//get a client configuration
	clnt := http.Client{
		Timeout: getTimeout,
	}
	//get the server configs
	req, err := http.NewRequest("GET", clientConfigUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1")

	resp, err := clnt.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		x, _ := ioutil.ReadAll(resp.Body)
		println(string(x))
		return nil, fmt.Errorf("Invalid status %d", resp.StatusCode)
	}
	xmlDec := xml.NewDecoder(resp.Body)
	cc := speedtestConfig{}
	if err := xmlDec.Decode(&cc); err != nil {
		return nil, err
	}
	cfg := Config{
		LicenseKey: cc.License,
		IP:         net.ParseIP(cc.ClientConfig.Ip),
		Lat:        cc.ClientConfig.Lat,
		Long:       cc.ClientConfig.Long,
		ISP:        cc.ClientConfig.ISP,
	}
	ignoreIDs := make(map[uint]bool, 1)
	strIDs := strings.Split(cc.ServerConfig.IgnoreIDs, ",")
	for i := range strIDs {
		x, err := strconv.ParseUint(strIDs[i], 10, 32)
		if err != nil {
			continue
		}
		ignoreIDs[uint(x)] = false
	}
	srvs, err := GetServerList()
	if err != nil {
		return nil, err
	}
	if err := populateServers(&cfg, srvs, ignoreIDs); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func populateServers(cfg *Config, srvs []server, ignore map[uint]bool) error {
	for i := range srvs {
		//checking if we are ignoring this server
		_, ok := ignore[srvs[i].ID]
		if ok {
			continue
		}
		srv := Testserver{
			Name:    srvs[i].Name,
			Sponsor: srvs[i].Sponsor,
			Country: srvs[i].Country,
			Lat:     srvs[i].Lat,
			Long:    srvs[i].Long,
			Host:    srvs[i].Host,
		}
		if srvs[i].Url != "" {
			srv.URLs = append(srv.URLs, srvs[i].Url)
		}
		if srvs[i].Url2 != "" {
			srv.URLs = append(srv.URLs, srvs[i].Url2)
		}
		p := geo.NewPoint(cfg.Lat, cfg.Long)
		if p == nil {
			return errors.New("Invalid client lat/long")
		}
		sp := geo.NewPoint(srvs[i].Lat, srvs[i].Long)
		if sp == nil {
			return errors.New("Invalid server lat/long")
		}
		srv.Distance = p.GreatCircleDistance(sp)
		cfg.Servers = append(cfg.Servers, srv)
	}
	sort.Sort(testServerlist(cfg.Servers))
	return nil
}

func (tsl testServerlist) Len() int           { return len(tsl) }
func (tsl testServerlist) Swap(i, j int)      { tsl[i], tsl[j] = tsl[j], tsl[i] }
func (tsl testServerlist) Less(i, j int) bool { return tsl[i].Distance < tsl[j].Distance }
