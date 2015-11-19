// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package config

import (
	"os"
	"net"
	"os/exec"
	"os/user"
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
	Monitor           string
	Primary           string
	Secondary         string
	DataDir           string
	StatusDir         string
	SyncCommand       string
	DecisionTimeout   int
	Vip               string
	VipAddCommand     string
	VipRemoveCommand  string
	RoleChangeCommand string
	SystemUser        string
}

// establish constants
// these are singleton values that are used throughout
// the package.
var (
	Conf = Config{
		AdvertisePort:   4400,
		PGPort:          5432,
		DataDir:         "/data/",
		StatusDir:       "./status/",
		SyncCommand:     "rsync -a --delete {{local_dir}} {{slave_ip}}:{{slave_dir}}",
		DecisionTimeout: 10,
		SystemUser:      SystemUser(),
	}
	Log = lumber.NewConsoleLogger(lumber.INFO)
)

// init Initializeds the config file and the other constants
func Init(path string) {

	//
	file, err := ini.LoadFile(path)
	if err != nil {
		Log.Error("[config.init]: Failed to load config file!\n%s\n", err)
		os.Exit(1)
	}

	// no conversion required for strings.
	if role, ok := file.Get("config", "role"); ok {
		Conf.Role = role
	}

	if dDir, ok := file.Get("config", "data_dir"); ok {
		Conf.DataDir = dDir
	}

	// make sure the datadir ends with a slash this should make it easier to handle
	if !strings.HasSuffix(Conf.DataDir, "/") {
		Conf.DataDir = Conf.DataDir + "/"
	}

	if sDir, ok := file.Get("config", "status_dir"); ok {
		Conf.StatusDir = sDir
	}

	if sMonitor, ok := file.Get("config", "monitor"); ok {
		Conf.Monitor = sMonitor
	}
	if sPrimary, ok := file.Get("config", "primary"); ok {
		Conf.Primary = sPrimary
	}
	if sSecondary, ok := file.Get("config", "secondary"); ok {
		Conf.Secondary = sSecondary
	}

	if !strings.HasSuffix(Conf.StatusDir, "/") {
		Conf.StatusDir = Conf.StatusDir + "/"
	}

	if sync, ok := file.Get("config", "sync_command"); ok {
		Conf.SyncCommand = sync
	}

	if ip, ok := file.Get("config", "advertise_ip"); ok {
		Conf.AdvertiseIp = ip
	}

	if vip, ok := file.Get("vip", "ip"); ok {
		Conf.Vip = vip
	}
	if vipAddCommand, ok := file.Get("vip", "add_command"); ok {
		Conf.VipAddCommand = vipAddCommand
	}
	if vipRemoveCommand, ok := file.Get("vip", "remove_command"); ok {
		Conf.VipRemoveCommand = vipRemoveCommand
	}

	if rcCommand, ok := file.Get("role_change", "command"); ok {
		Conf.RoleChangeCommand = rcCommand
	}

	parseInt(&Conf.AdvertisePort, file, "config", "advertise_port")
	parseInt(&Conf.PGPort, file, "config", "pg_port")
	parseInt(&Conf.DecisionTimeout, file, "config", "decision_timeout")

	if logLevel, ok := file.Get("config", "Log_level"); ok {
		switch logLevel {
		case "TRACE", "trace":
			Log.Level(lumber.TRACE)
		case "DEBUG", "debug":
			Log.Level(lumber.DEBUG)
		case "INFO", "info":
			Log.Level(lumber.INFO)
		case "WARN", "warn":
			Log.Level(lumber.WARN)
		case "ERROR", "error":
			Log.Level(lumber.ERROR)
		case "FATAL", "fatal":
			Log.Level(lumber.FATAL)
		}
	}
	confirmPeers()
	confirmRole()
	confirmAdvertiseIp()
	confirmAdvertisePort()

}

func confirmPeers() {
	if Conf.Monitor == "" || Conf.Primary == "" || Conf.Secondary == "" {
		Log.Fatal("I need connection Credentials for monitor, primary and secondary")
		Log.Close()
		os.Exit(1)
	}
}

func confirmRole() {
	if Conf.Role == "" {
		Conf.Role = getRole()
	}
	if Conf.Role != "monitor" && Conf.Role != "primary" && Conf.Role != "secondary" {
		Log.Fatal("I could not find the appropriate role (role:'%s').", Conf.Role)
		Log.Close()
		os.Exit(1)
	}
}

func confirmAdvertiseIp() {
	if Conf.AdvertiseIp == "" || Conf.AdvertiseIp == "0.0.0.0" {
		getAdvertiseData()
	}
	if Conf.AdvertiseIp == "" || Conf.AdvertiseIp == "0.0.0.0" {
		Log.Fatal("I could not find the appropriate AdvertiseIP (ip:'%s').", Conf.AdvertiseIp)
		Log.Close()
		os.Exit(1)
	}
}

func confirmAdvertisePort() {
	if Conf.AdvertisePort == 0 {
		Log.Fatal("I could not find the appropriate Port to listen on (port:'0').")
		Log.Close()
		os.Exit(1)
	}
}

func getRole() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			str := strings.Split(addr.String(), "/")[0]
			switch {
			case strings.HasPrefix(Conf.Monitor, str):
				return "monitor"
			case strings.HasPrefix(Conf.Primary, str):
				return "primary"
			case strings.HasPrefix(Conf.Monitor, str):
				return "secondary"
			}
		}
	}
	return ""
}

func getAdvertiseData() {
	if Conf.AdvertiseIp == "" || Conf.AdvertiseIp == "0.0.0.0" || Conf.AdvertisePort == 0 {
		Log.Info(Conf.AdvertiseIp)
		var self string
		switch Conf.Role {
		case "monitor":
			self = Conf.Monitor
		case "primary":
			self = Conf.Primary
		case "secondary":
			self = Conf.Secondary
		}
		Log.Info(self)
		connArr := strings.Split(self, ":")
		if len(connArr) == 2 {
			Conf.AdvertiseIp = connArr[0]

			i, err := strconv.ParseInt(connArr[1], 10, 64)
			if err == nil {
				Conf.AdvertisePort = int(i)
			}
		}

	}
}

//
func parseInt(val *int, file ini.File, section, name string) {
	if port, ok := file.Get(section, name); ok {
		i, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			Log.Fatal(name + " is not an int")
			Log.Close()
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

func SystemUser() string {
	username := "postgres"
	usr, err := user.Current()
	if err != nil {
		cmd := exec.Command("bash", "-c", "whoami")
		bytes, err := cmd.Output()
		if err == nil {
			str := string(bytes)
			return strings.TrimSpace(str)
		}
	}

	username = usr.Username
	return username
}
