package main

import (
  "database/sql"
  _ "github.com/lib/pq"
	"fmt"
	"time"
	"os"
	"os/exec"
	"syscall"
	"strings"
	"github.com/hoisie/mustache"
)

type Piper struct {
	Prefix string
	// just need a couple methods
}

func (p Piper) Write(d []byte) (int, error) {
	log.Info(p.Prefix+" %s", d)
	return len(d), nil
}

var	cmd	*exec.Cmd

// Listen on the action channel and perform the action
func ActionStart() error {
	go func() {
		for {
			select {
			case act := <-actions:
				log.Info("[action] new action: " + act)
				doAction(act)
			}
		}
	}()

	return nil
}

// switch through the actions and perform the requested action
func doAction(act string) {
	switch act {
	case "kill":
		killDB()
	case "master":
		startMaster()
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
	status.SetState("configuring")
	configurePGConf(false)
	// set pg_hba.conf
	configureHBAConf()
	// delete recovery.conf
	destroyRecovery()
	// start the database
	status.SetState("starting")
	startDB()
	// connect to DB and tell it to start backup
  db, err := sql.Open("postgres", fmt.Sprintf("user=postgres sslmode=disable host=localhost port=%d", conf.PGPort))
  if err != nil {
  	log.Fatal("[action.startMaster] Couldnt establish Database connection (%s)", err.Error())
  	log.Close()
  	os.Exit(1)
  }
  defer db.Close()

  _, err = db.Exec("select pg_start_backup('replication')")
  if err != nil {
  	log.Fatal("[action.startMaster] Couldnt start backup (%s)", err.Error())
  	log.Close()
  	os.Exit(1)
  }
	// rsync my files over to the slave server
	status.SetState("syncing")
	self := myself()
	other, _ := Whois(otherRole(self))
	// rsync -a {{local_dir}} {{slave_ip}}:{{slave_dir}}
  sync := mustache.Render(conf.SyncCommand, map[string]string{"local_dir":conf.DataDir,"slave_ip":other.Ip,"slave_dir":other.DataDir})
  cmd := strings.Split(sync, " ")
  sc := exec.Command(cmd[0], cmd[1:]...)
	sc.Stdout = Piper{"[sync.stdout]"}
	sc.Stderr = Piper{"[sync.stderr]"}
  
	if err = sc.Run(); err != nil {
		log.Error("[action] sync failed.")
	}

	// connect to DB and tell it to stop backup
  _, err = db.Exec("select pg_stop_backup()")
  if err != nil {
  	log.Fatal("[action.startMaster] Couldnt start backup (%s)", err.Error())
  	log.Close()
  	os.Exit(1)
  }

	// set postgresql.conf as master
	configurePGConf(true)

	// start refresh/restart server
	restartDB()

	// make sure slave is connected and in sync
	status.SetState("waiting")
  defer status.SetState("running")
	for {
	  rows, err := db.Query("SELECT application_name, client_addr, state, sync_state, pg_xlog_location_diff(pg_current_xlog_location(), sent_location) FROM pg_stat_replication")
	  if err != nil {
	  	
	  }
	  for rows.Next() {
      var name string
      var addr string
      var state string
      var sync string
      var behind int64
      err = rows.Scan(&name, &addr, &state, &sync, &behind)
      if err != nil { panic(err) }
      if behind > 0 {
      	log.Info("Sync is catching up (name:%s,address:%s,state:%s,sync:%s,behind:%d)", name, addr, state, sync, behind)
      } else {
      	return
      }
	  }
	  time.Sleep(time.Second)
	}


}

// Starts the database as a slave node after waiting for master to come online
func startSlave() {
	// wait for master server to be running
	self := myself()
	for {
		other, _ := Whois(otherRole(self))
		if other.State == "running" || other.State == "waiting" {
			break
		}
		time.Sleep(time.Second)
	}
	// make sure we have a database in the data folder
	initDB()
	// set postgresql.conf as not master
  status.SetState("configuring")
	configurePGConf(false)
	// set pg_hba.conf
	configureHBAConf()
	// set recovery.conf
	createRecovery()
	// start the database
	status.SetState("starting")
	startDB()
  status.SetState("running")
}

// Starts the database as a single node 
func startSingle() {
	// make sure we have a database in the data folder
	initDB()
	// set postgresql.conf as not master
	configurePGConf(false)
	// set pg_hba.conf
	configureHBAConf()
	// delete recovery.conf
	destroyRecovery()
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
	cmd.Process.Signal(syscall.SIGQUIT)

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
	cmd = exec.Command("postgres", "-D", conf.DataDir)
	cmd.Stdout = Piper{"[postgres.stdout]"}
	cmd.Stderr = Piper{"[postgres.stderr]"}
	cmd.Start()
	time.Sleep(10 * time.Second)
	if cmd.ProcessState.Exited() {
		log.Fatal("I just started the database and it is not running")
		log.Close()
		os.Exit(1)
	}
}

func restartDB() {
	killDB()
	startDB()
}

func initDB() {
	if _, err := os.Stat(conf.DataDir+"/postgresql.conf"); os.IsNotExist(err) {
		init := exec.Command("initdb", conf.DataDir)
		init.Stdout = Piper{"[INITDB.stdout]"}
		init.Stderr = Piper{"[INITDB.stderr]"}

		if err = init.Run(); err != nil {
			log.Fatal("[action] initdb failed. Are you missing your postgresql.conf")
			log.Close()
			os.Exit(1)
		}
	}
}
