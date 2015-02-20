package main

import "github.com/vaughan0/go-ini"
import "flag"
import "strconv"
import "strings"


type Config struct {
	Role string
	ClusterPort int
	Peers []string
}

var conf Config

func init() {
	conf = Config{
		Role: "Monitor",
		ClusterPort: 1234,
	}

	var c = flag.String("config", "", "Config file")
	flag.Parse()
	file, err := ini.LoadFile(*c)	
	if err != nil {
		panic("failed to load config file: " + err.Error())
	}
	if role, ok := file.Get("config", "role"); ok {
		if role != "monitor" || role != "primary" || role != "secondary" {
			panic("That Role is NOT OK!")
		}
		conf.Role = role
	}
	if port, ok := file.Get("config", "cluster_port"); ok {
		i, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			panic("cluster_port is not an ip")
		}
		conf.ClusterPort = int(i)
	}
	if peers, ok := file.Get("config", "peers"); ok {
		conf.Peers = strings.Split(peers, ",")
	}

}
