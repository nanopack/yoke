// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package main

import (
	"fmt"
	"github.com/nanobox-io/golang-scribble"
	"github.com/nanobox-io/yoke/config"
	"github.com/nanobox-io/yoke/monitor"
	"github.com/nanobox-io/yoke/state"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//
func main() {
	if len(os.Args) != 2 {
		fmt.Println("missing required config file!")
		os.Exit(1)
	}
	config.Init(os.Args[1])

	config.ConfigurePGConf("0.0.0.0", config.Conf.PGPort)

	store, err := scribble.New(config.Conf.StatusDir, config.Log)
	if err != nil {
		config.Log.Fatal("scribble did not setup correctly %v", err)
		os.Exit(1)
	}

	location := fmt.Sprintf("%v:%d", config.Conf.AdvertiseIp, config.Conf.AdvertisePort)
	me, err := state.NewLocalState(config.Conf.Role, location, config.Conf.DataDir, store)
	if err != nil {
		panic(err)
	}

	me.ExposeRPCEndpoint("tcp", location)

	var other state.State
	var host string
	switch config.Conf.Role {
	case "primary":
		location := config.Conf.Secondary
		other = state.NewRemoteState("tcp", location, time.Second)
		host, _, err = net.SplitHostPort(location)
		if err != nil {
			panic(err)
		}
	case "secondary":
		location := config.Conf.Primary
		other = state.NewRemoteState("tcp", location, time.Second)
		host, _, err = net.SplitHostPort(location)
		if err != nil {
			panic(err)
		}
	default:
		// nothing as the monitor does not need to monitor anything
		// the monitor just acts as a secondary mode of communication in network
		// splits
	}

	mon := state.NewRemoteState("tcp", config.Conf.Monitor, time.Second)

	var perform monitor.Performer
	finished := make(chan error)
	if other != nil {

		perform = monitor.NewPerformer(me, other, config.Conf)

		if err := perform.Initialize(); err != nil {
			panic(err)
		}

		if err := config.ConfigureHBAConf(host); err != nil {
			panic(err)
		}

		if err := config.ConfigurePGConf("0.0.0.0", config.Conf.PGPort); err != nil {
			panic(err)
		}

		if err := perform.Start(); err != nil {
			panic(err)
		}

		go func() {
			decide := monitor.NewDecider(me, other, mon, perform)
			decide.Loop(time.Second * 2)
		}()

		go func() {
			err := perform.Loop()
			if err != nil {
				finished <- err
			}
			// how do I stop the decide loop?
			close(finished)
		}()
	}

	// signal Handle
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Kill, syscall.SIGQUIT, syscall.SIGALRM)

	// Block until a signal is received.
	for {
		select {
		case err := <-finished:
			// the performer is finished, something triggered a stop.
			if err != nil {
				panic(err)
			}
			config.Log.Info("the database was shut down")
			return
		case signal := <-signals:
			switch signal {
			case syscall.SIGINT, os.Kill, syscall.SIGQUIT, syscall.SIGTERM:
				config.Log.Info("shutting down")
				if perform != nil {
					// stop the database, then wait for it to be stopped
					config.Log.Info("shutting down the database")
					perform.Stop()
					perform = nil
					config.Log.Info("waiting for the database")
				} else {
					return
				}
			case syscall.SIGALRM:
				config.Log.Info("Printing Stack Trace")
				stacktrace := make([]byte, 8192)
				length := runtime.Stack(stacktrace, true)
				fmt.Println(string(stacktrace[:length]))
			}
		}
	}
}
