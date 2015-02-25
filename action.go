package main

import (
	"fmt"
	"time"
	"os/exec"
	"os/signal"
)
// syscall.Kill(cpid, syscall.SIGHUP)
var	cmd	*exec.Command

//
func ActionStart() error {
	for {
		select {
		case act := <-actions:
			log.Info("[action] new action: " + act)
			doAction(act)
		}
	}

	return nil
}

//
func doAction(act string) {
	switch act {
	case "kill":
		killDB()
	case "master":
		startMaster()
		// status.SetState("configuring")
		// time.Sleep(time.Second * 20)
		// status.SetState("starting/restarting")
		// time.Sleep(time.Second * 20)
		// status.SetState("running")
	case "slave":
		startSlave()
	case "single":
		startSingle()
	default:
		fmt.Println("i dont know what to do with this action: " + act)
	}
}

// Starts the database as a master node and sends its data file to the slave
func startMaster() {
	// make sure we have a database in the data folder
	initDB()
	// set postgresql.conf as not master
	BuildPGConf(false)
	// set pg_hba.conf
	BuildHBAConf()
	// delete recovery.conf
	RemoveRecovery()
	// start the database
	startDB()
	// connect to DB and tell it to start backup

	// rsync my files over to the slave server

	// connect to DB and tell it to stop backup

	// set postgresql.conf as master
	BuildPGConf(true)
	// start refresh/restart server

	// make sure slave is connected and in sync
}

// Starts the database as a slave node after waiting for master to come online
func startSlave() {
	// wait for master server to be running
	// make sure we have a database in the data folder
	initDB()
	// set postgresql.conf as not master
	BuildPGConf(false)
	// set pg_hba.conf
	BuildHBAConf()
	// set recovery.conf
	BuildRecovery()
	// start the database
	startDB()
}

// Starts the database as a single node 
func startSingle() {
	// make sure we have a database in the data folder
	initDB()
	// set postgresql.conf as not master
	BuildPGConf(false)
	// set pg_hba.conf
	BuildHBAConf()
	// delete recovery.conf
	RemoveRecovery()
	// start the database
	startDB()
}

// this will kill the database that is running. reguardless of its current state
func killDB() {
	defer func() {
		status.SetState("down")
	}()
	// return if it was never created or up
	if cmd == nil {
		return
	}
	// if it was stopped set the cmd to nil and return
	if cmd.ProcessState.Exited() {
		return
	}

	// if it is running kill it and wait for it to go down
	status.SetState("signaling")
	cmd.Process.Signal(signal.SIGQUIT)

	// waiting for shutdown
	status.SetState("waiting")
	cmd.Wait()
	cmd = nil
}

func startDB() {
	// if we still happen to have a cmd running kill it
	if cmd != nil {
		killDB()
	}
	cmd = os.Command("postgres", "-D", conf.DataDir)
	pipeOutput(cmd)
	cmd.Start()
}

func restartDB() {
	killDB()
	startDB()
}

func initDB() {
	
}

func pipeOutput(cmd) {

}
