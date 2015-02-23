package main

import (
	"fmt"
	"github.com/vaughan0/go-ini"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Role        string
	ClusterPort int
	Peers       []string
}

var	actions chan string
var conf Config

func init() {
	actions = make(chan string)
	conf = Config{
		Role:        "Monitor",
		ClusterPort: 1234,
		Peers: []string{},
	}

	if len(os.Args) < 2 {
		panic("where is my config file bro?")
	}

	file, err := ini.LoadFile(os.Args[1])
	if err != nil {
		panic("failed to load config file: " + err.Error())
	}

	// no conversion required for strings.
	if role, ok := file.Get("config", "role"); ok {
		fmt.Print(role)
		if role != "monitor" && role != "primary" && role != "secondary" {
			panic("That Role is NOT OK!")
		}
		conf.Role = role
	}

	parseInt(&conf.ClusterPort, file, "config", "cluster_port")
	parseArr(&conf.Peers, file, "config", "peers")
}

func parseInt(val *int, file ini.File, section, name string) {
	if port, ok := file.Get(section, name); ok {
		i, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			panic(name + " is not an int")
		}
		*val = int(i)
	}
}

func parseArr(val *[]string, file ini.File, section, name string) {
	if peers, ok := file.Get(section, name); ok {
		*val = strings.Split(peers, ",")
	}
}
