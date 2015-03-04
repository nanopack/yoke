package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/jcelliott/lumber"
	"github.com/vaughan0/go-ini"
)

// Config is the struct of all global configuration data
// it is set by a config file that is the first arguement
// given to the exec
type Config struct {
	Role              string
	AdvertiseIp       string
	AdvertisePort     int
	PGPort            int
	Peers             []string
	DataDir           string
	StatusDir         string
	SyncCommand       string
	DecisionTimeout   int
	Vip               string
	VipAddCommand     string
	VipRemoveCommand  string
	RoleChangeCommand string
}

// establish constants
// these are singleton values that are used throughout
// the package.
var (
	advice  chan string
	actions chan string
	conf    Config
	log     *lumber.ConsoleLogger
)

// init Initializeds the config file and the other constants
func init() {
	advice = make(chan string)
	actions = make(chan string)

	//
	log = lumber.NewConsoleLogger(lumber.INFO)

	//
	conf = Config{
		Role:            "monitor",
		AdvertisePort:   4400,
		PGPort:          5432,
		Peers:           []string{},
		DataDir:         "/data/",
		StatusDir:       "./status/",
		SyncCommand:     "rsync -a --delete {{local_dir}} {{slave_ip}}:{{slave_dir}}",
		DecisionTimeout: 10,
	}

	//
	if len(os.Args) < 2 {
		log.Error("[config.init]: Config file required, run 'yoke path/to/config.ini' to start! Exiting...")
		os.Exit(1)
	}

	//
	file, err := ini.LoadFile(os.Args[1])
	if err != nil {
		log.Error("[config.init]: Failed to load config file!\n%s\n", err)
		os.Exit(1)
	}

	// no conversion required for strings.
	if role, ok := file.Get("config", "role"); ok {
		if role != "monitor" && role != "primary" && role != "secondary" {
			panic("That Role is NOT OK!")
		}
		conf.Role = role
	}

	if dDir, ok := file.Get("config", "data_dir"); ok {
		conf.DataDir = dDir
	}

	// make sure the datadir ends with a slash
	// this should make it easier to handle
	if !strings.HasSuffix(conf.DataDir, "/") {
		conf.DataDir = conf.DataDir + "/"
	}

	if sDir, ok := file.Get("config", "status_dir"); ok {
		conf.StatusDir = sDir
	}

	if !strings.HasSuffix(conf.StatusDir, "/") {
		conf.StatusDir = conf.StatusDir + "/"
	}

	if sync, ok := file.Get("config", "sync_command"); ok {
		conf.SyncCommand = sync
	}

	if ip, ok := file.Get("config", "advertise_ip"); ok {
		conf.AdvertiseIp = ip
	}

	if vip, ok := file.Get("vip", "ip"); ok {
		conf.Vip = vip
	}
	if vipAddCommand, ok := file.Get("vip", "add_command"); ok {
		conf.VipAddCommand = vipAddCommand
	}
	if vipRemoveCommand, ok := file.Get("vip", "remove_command"); ok {
		conf.VipRemoveCommand = vipRemoveCommand
	}

	if rcCommand, ok := file.Get("role_change", "command"); ok {
		conf.RoleChangeCommand = rcCommand
	}

	if conf.AdvertiseIp == "" || conf.AdvertiseIp == "0.0.0.0" {
		log.Fatal("advertise_ip (" + conf.AdvertiseIp + ") is not a valid ip")
		log.Close()
		os.Exit(1)
	}

	parseInt(&conf.AdvertisePort, file, "config", "advertise_port")
	parseInt(&conf.PGPort, file, "config", "pg_port")
	parseInt(&conf.DecisionTimeout, file, "config", "decision_timeout")
	parseArr(&conf.Peers, file, "config", "peers")

	if logLevel, ok := file.Get("config", "log_level"); ok {
		switch logLevel {
		case "TRACE", "trace":
			log.Level(lumber.TRACE)
		case "DEBUG", "debug":
			log.Level(lumber.DEBUG)
		case "INFO", "info":
			log.Level(lumber.INFO)
		case "WARN", "warn":
			log.Level(lumber.WARN)
		case "ERROR", "error":
			log.Level(lumber.ERROR)
		case "FATAL", "fatal":
			log.Level(lumber.FATAL)
		}
	}

}

//
func parseInt(val *int, file ini.File, section, name string) {
	if port, ok := file.Get(section, name); ok {
		i, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			log.Fatal(name + " is not an int")
			log.Close()
			os.Exit(1)
		}
		*val = int(i)
	}
}

//
func parseArr(val *[]string, file ini.File, section, name string) {
	if peers, ok := file.Get(section, name); ok {
		*val = strings.Split(peers, ",")
	}
}
