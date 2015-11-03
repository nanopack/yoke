// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package monitor

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"github.com/hoisie/mustache"
	_ "github.com/lib/pq"
	"github.com/nanobox-io/yoke/config"
	"github.com/nanobox-io/yoke/state"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	Done = errors.New("done")
)

type (
	Performer interface {
		TransitionToActive()
		TransitionToBackup()
		TransitionToSingle()
		Stop()
		Loop() error
	}

	performer struct {
		sync.Mutex
		step   map[string]bool
		me     state.State
		other  state.State
		err    chan error
		done   chan interface{}
		cmd    *exec.Cmd
		config config.Config
	}
)

func NewPrefix(prefix string) io.Writer {
	read, write := io.Pipe()
	scan := bufio.NewScanner(read)
	go func() {
		for scan.Scan() {
			fmt.Printf("%v %v\n", prefix, scan.Text())
		}
	}()
	return write
}

func NewPerformer(me state.State, other state.State, config config.Config) *performer {
	perform := performer{
		config: config,
		step: map[string]bool{
			"trigger": true, // this should only be there if the trigger file exists
		},
		me:    me,
		other: other,
		err:   make(chan error),
		done:  make(chan interface{}),
	}

	return &perform
}

func (performer *performer) Loop() error {
	config.Log.Info("Going to Init")
	if err := performer.Initialize(); err != nil {
		return err
	}
	config.Log.Info("Waiting for error")
	return <-performer.err
}

func (performer *performer) Stop() {
	config.Log.Info("going to stop")
	performer.Lock()
	defer performer.Unlock()
	config.Log.Info("stopping")
	performer.stop()
	config.Log.Info("stopped")
}

func (performer *performer) TransitionToSingle() {
	performer.Lock()
	defer performer.Unlock()

	// backups or actives can transition to single
	// it just means that the other node went down
	role, err := performer.me.GetDBRole()
	if err != nil {
		performer.err <- err
		return
	}
	if role == "single" {
		return
	}

	err = performer.Single()
	if err != nil {
		performer.err <- err
	}
}

func (performer *performer) TransitionToActive() {
	performer.Lock()
	defer performer.Unlock()

	role, err := performer.me.GetDBRole()
	if err != nil {
		performer.err <- err
		return
	}
	switch role {
	case "active":
		return
	case "backup":
		// backups must transition to single before they can become active.
		panic("something went seriously wrong, backups cannot transition to active.")
	}

	err = performer.Active()
	if err != nil {
		performer.err <- err
	}
}

func (performer *performer) TransitionToBackup() {
	performer.Lock()
	defer performer.Unlock()

	role, err := performer.me.GetDBRole()
	if err != nil {
		performer.err <- err
		return
	}
	if role == "backup" {
		return
	}

	err = performer.Backup()
	if err != nil {
		performer.err <- err
	}
}

func (performer *performer) stop() error {
	if performer.step["started"] {
		fmt.Println("sending signal")
		err := performer.cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			return err
		}
		fmt.Println("waiting for stop")
		<-performer.done
		fmt.Println("really stopped")
		performer.step["started"] = false
	}
	return nil
}

func (performer *performer) Initialize() error {
	_, err := os.Stat(performer.config.DataDir)
	switch {
	case os.IsNotExist(err):
		config.Log.Info("creating database")
		init := exec.Command("initdb", performer.config.DataDir)
		init.Stdout = NewPrefix("[initdb.stdout]")
		init.Stderr = NewPrefix("[initdb.stderr]")
		if err = init.Run(); err != nil {
			return err
		}
	default:
		config.Log.Info("database has already been created... skipping.")
	}
	return err
}

func (performer *performer) Start() error {
	config.Log.Info("going to start")
	performer.Lock()
	defer performer.Unlock()
	config.Log.Info("starting")
	return performer.startDB()
}

// The Single state.
func (performer *performer) Single() error {
	config.Log.Info("transitioning to Single")

	// disable syncronus transaction commits.
	if err := performer.setSync(false, nil); err != nil {
		return err
	}

	if err := performer.replicate(false); err != nil {
		return err
	}

	config.Log.Info("[action] running DB as single")

	performer.addVip()
	performer.roleChangeCommand("single")
	performer.me.SetDBRole("single")

	return nil
}

func (performer *performer) sync(command string) error {
	sc := exec.Command("bash", "-c", command)
	sc.Stdout = NewPrefix("[pre-sync.stdout]")
	sc.Stderr = NewPrefix("[pre-sync.stderr]")
	config.Log.Info("[action] running pre-sync")
	config.Log.Debug("[action] pre-sync command(%s)", command)

	return sc.Run()
}

func (performer *performer) replicate(enabled bool) error {
	if performer.step["trigger"] == enabled {
		return nil
	}
	performer.step["trigger"] = enabled

	trigger := performer.config.StatusDir + "/i-am-primary"
	switch enabled {
	case true:
		return os.Remove(trigger)
	default:
		// this is a trigger file, it should stop this node from replicating from a
		// remote node,
		f, err := os.Create(trigger)
		if err != nil {
			return err
		}
		f.Close()
		return nil
	}
}

func (performer *performer) pgConnect() (*sql.DB, error) {
	fmt.Println("opening new connection to db")
	return sql.Open("postgres", fmt.Sprintf("user=%s database=postgres sslmode=disable host=localhost port=%d", performer.config.SystemUser, performer.config.PGPort))
}

func (performer *performer) setSync(enabled bool, db *sql.DB) error {
	if db == nil {
		var err error
		db, err = performer.pgConnect()
		if err != nil {
			return err
		}
		defer db.Close()
	}
	var sync string
	switch enabled {
	case true:
		sync = "on"
	default:
		sync = "off"
	}
	_, err := db.Exec(fmt.Sprintf(
		`BEGIN;
SET LOCAL synchronous_commit=off;
ALTER USER %v SET synchronous_commit=%v;
COMMIT;`, performer.config.SystemUser, sync))
	return err

}

// The Active state.
func (performer *performer) Active() error {
	config.Log.Info("transitioning to Active")
	if err := performer.replicate(false); err != nil {
		return err
	}

	// do an initial copy of files which might be corrupt because they are not consistant
	// this will be fixed later. we do this now so that a majority of the data will make it across without
	// having to pause the Durablility (ACID compliance) of postgres
	config.Log.Debug("[action] pre-backup started")
	dataDir, err := performer.other.GetDataDir()
	if err != nil {
		return err
	}
	location := performer.other.Location()
	ip, _, err := net.SplitHostPort(location)
	if err != nil {
		return err
	}
	sync := mustache.Render(performer.config.SyncCommand, map[string]string{"local_dir": performer.config.DataDir, "slave_ip": ip, "slave_dir": dataDir})

	if err := performer.sync(sync); err != nil {
		return err
	}

	db, err := performer.pgConnect()
	if err != nil {
		return err
	}
	defer db.Close()

	// this informs postgres to make the files on disk consistant for copying,
	// all changes are kept in memory from this point on
	_, err = db.Exec("select pg_start_backup('replication')")
	if err != nil {
		return err
	}

	config.Log.Debug("[action] backup started")

	if err := performer.sync(sync); err != nil {
		// stop the backup, if it fails, there is nothing we can do so we return the original error
		db.Exec("select pg_stop_backup()")

		// something went wrong, we are the master still, so lets wait for the slave to reconnect
		return nil
	}

	// connect to DB and tell it to stop backup
	if _, err = db.Exec("select pg_stop_backup()"); err != nil {
		return err
	}

	config.Log.Debug("[action] backup complete")

	// if we were unsucessfull at setting the sync flag on the other node
	// then we need to start all over
	if performer.other.SetSynced(true) != nil {
		return nil
	}

	// enable syncronus transaction commits.
	if err := performer.setSync(true, db); err != nil {
		return err
	}

	performer.addVip()
	performer.roleChangeCommand("master")

	performer.me.SetDBRole("active")
	return nil
}

// The Backup state.
func (performer *performer) Backup() error {
	config.Log.Info("transitioning to Backup")
	performer.removeVip()

	// TODO figure out if the recover.conf file needs to be regenerated.

	// wait for master server to be running
	for {
		ready, err := performer.me.HasSynced()
		if err != nil {
			return err
		}
		if ready {
			break
		}
		time.Sleep(time.Second)
	}

	config.Log.Debug("[action] starting database")
	performer.startDB()
	performer.roleChangeCommand("backup")
	return performer.me.SetDBRole("backup")
}

// this will kill the database that is running. reguardless of its current state
func (performer *performer) killDB() {
	config.Log.Debug("[action] KillingDB")

	if performer.cmd == nil {
		config.Log.Debug("[action] nothing to kill")
		return
	}

	err := performer.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		config.Log.Error("[action] Kill Signal error: %s", err.Error())
	}
}

func (performer *performer) startDB() error {
	if !performer.step["started"] {
		config.Log.Info("[action] starting db")
		cmd := exec.Command("postgres", "-D", performer.config.DataDir)
		cmd.Stdout = NewPrefix("[postgres.stdout]")
		cmd.Stderr = NewPrefix("[postgres.stderr]")
		err := cmd.Start()
		if err != nil {
			return err
		}
		performer.cmd = cmd
		go performer.reportExit()

		// wait for postgres to exit, or for it to start correctly
		for {
			if performer.cmd == nil {
				return <-performer.err
			}
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", performer.config.PGPort))
			fmt.Println("checking if postgres is up", conn, err, performer.config.PGPort)
			if err == nil {
				conn.Close()
				break
			}
			<-time.After(time.Second)
		}
		config.Log.Info("[action] db started")
		performer.step["started"] = true
	}
	return nil
}

func (performer *performer) reportExit() {
	err := performer.cmd.Wait()
	performer.cmd = nil
	fmt.Println("it exited", err)

	if err != nil {
		performer.err <- err
	}
	close(performer.done)
}

func (performer *performer) roleChangeCommand(role string) {
	if performer.config.RoleChangeCommand != "" {
		rcc := exec.Command("bash", "-c", fmt.Sprintf("%s %s", performer.config.RoleChangeCommand, role))
		rcc.Stdout = NewPrefix("[RoleChangeCommand.stdout]")
		rcc.Stderr = NewPrefix("[RoleChangeCommand.stderr]")
		if err := rcc.Run(); err != nil {
			config.Log.Error("[action] RoleChangeCommand failed.")
			config.Log.Debug("[RoleChangeCommand.error] message: %s", err.Error())
		}
	}
}

func (performer *performer) addVip() {
	if performer.vipable() {
		config.Log.Info("[action] Adding VIP")
		vAddCmd := exec.Command("bash", "-c", fmt.Sprintf("%s %s", performer.config.VipAddCommand, performer.config.Vip))
		vAddCmd.Stdout = NewPrefix("[VIPAddCommand.stdout]")
		vAddCmd.Stderr = NewPrefix("[VIPAddCommand.stderr]")
		if err := vAddCmd.Run(); err != nil {
			config.Log.Error("[action] VIPAddCommand failed.")
			config.Log.Debug("[VIPAddCommand.error] message: %s", err.Error())
		}
	}
}

func (performer *performer) removeVip() {
	if performer.vipable() {
		config.Log.Info("[action] Removing VIP")
		vRemoveCmd := exec.Command("bash", "-c", fmt.Sprintf("%s %s", performer.config.VipRemoveCommand, performer.config.Vip))
		vRemoveCmd.Stdout = NewPrefix("[VIPRemoveCommand.stdout]")
		vRemoveCmd.Stderr = NewPrefix("[VIPRemoveCommand.stderr]")
		if err := vRemoveCmd.Run(); err != nil {
			config.Log.Error("[action] VIPRemoveCommand failed.")
			config.Log.Debug("[VIPRemoveCommand.error] message: %s", err.Error())
		}
	}
}

func (performer *performer) vipable() bool {
	return performer.config.Vip != "" && performer.config.VipAddCommand != "" && performer.config.VipRemoveCommand != ""
}
