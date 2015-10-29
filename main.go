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
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//
func main() {
	store, err := scribble.New("dir", config.Log)
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
	switch config.Conf.Role {
	case "primary":
		location := config.Conf.Secondary
		other = state.NewRemoteState("tcp", location, time.Second)
	case "secondary":
		location := config.Conf.Secondary
		other = state.NewRemoteState("tcp", location, time.Second)
	default:
		// nothing as the monitor does not need to monitor anything
		// the monitor just acts as a secondary mode of communication in network
		// splits
	}

	mon := state.NewRemoteState("tcp", config.Conf.Monitor, time.Second)

	var perform monitor.Performer
	finished := make(chan error)
	if other != nil {
		meCan := me.(monitor.Candidate)
		otherCan := other.(monitor.Candidate)
		monMon := mon.(monitor.Monitor)

		perform = monitor.NewPerformer(meCan, otherCan)
		decide := monitor.NewDecider(meCan, otherCan, monMon, perform)

		go decide.Loop(time.Second * 10)
		go func() {
			err := perform.Loop()
			if err != nil {
				finished <- err
			}
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
				if perform != nil {
					// stop the database, then wait for it to be stopped
					config.Log.Info("shutting down the database")
					perform.Stop()
					perform = nil
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
