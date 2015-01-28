speedtest
=========

Command line speedtest which interacts with the speedtest.net infrastructure

Example below:


	$ /opt/mygo/bin/speedtest -s="Phoe"
	1 Matching servers:
	-------  ----------------  -----------------  ------------------
	    ID              Name            Sponsor       Distance (km) 
	-------  ----------------  -----------------  ------------------
	     0       Phoenix, AZ       Pavlov Media             1111.75 
	-------  ----------------  -----------------  ------------------
	Enter server ID for bandwidth test, or "quit" to exit
	ID> 0
	Latency: ██▁▁▃▁▁▁▃▁▃▃▅▃▁▃▅▃█▃	80ms avg	80ms median	83ms max	79ms min
	Download: 23.01 Mb/s
	Upload:   2.95 Mb/s
	$ /opt/mygo/bin/speedtest
	Gathering server list and testing...
	5 Closest responding servers:
	-------  --------------------  -------------------------------------  ------------------  -----------------
	    ID                  Name                                Sponsor       Distance (km)       Latency (ms) 
	-------  --------------------  -------------------------------------  ------------------  -----------------
	     0       Idaho Falls, ID                              Microserv                7.65       125.885984ms 

	     1       Idaho Falls, ID                       Syringa Networks                7.65       138.980347ms 

	     2           Rexburg, ID       Brigham Young University - Idaho               48.50       130.055828ms 

	     3           Bozeman, MT                    Montana Opticom LLC              262.15       114.208524ms 

	     4           Bozeman, MT                             Global Net              262.15       116.363797ms 
	-------  --------------------  -------------------------------------  ------------------  -----------------
	Enter server ID for bandwidth test, or "quit" to exit
	ID> 0
	Latency: █▇▃▂▂▃▂▃▂▃▂▃▅▁▃▃▂▃▃▂	123ms avg	123ms median	128ms max	122ms min
	Download: 9.73 Mb/s
	Upload:   2.89 Mb/s
	$
