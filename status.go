package main

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"time"

	"github.com/nanobox-core/scribble"
)

//
type (

	//
	Status struct {
		CRole    	string
		DBRole   	string
		State    	string
		UpdatedAt time.Time
	}
)

var(
	status *Status
	store *scribble.Driver
)


//
func StatusStart() error {

	//
	status = &Status{CRole: conf.Role, State: "booting", UpdatedAt: time.Now()}

	//
	store = scribble.New("./status", log)
	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: status}
	if err := store.Transact(t); err != nil {
		return err
	}

	//
	rpc.Register(status)

	// RPC SERVER
	l, err := net.Listen("tcp", ":" + strconv.FormatInt(int64(conf.ClusterPort+1), 10))
	if err != nil {
		return err
	}

	//
	go func(l net.Listener) {
		for {
			if conn, err := l.Accept(); err != nil {
				log.Error("accept error: " + err.Error())
			} else {
				log.Trace("new connection established\n")
				go rpc.ServeConn(conn)
			}
		}
	}(l)

	return nil
}

// 'public' methods

//
func (s *Status) SetDBRole(role string) {
	s.DBRole = role

	if err := save(s); err != nil {
		log.Fatal("BONK!", err)
		panic("Unable to set db role! " + err.Error())
	}

	s.UpdatedAt = time.Now()
}

//
func (s *Status) SetState(state string) {
	s.State = state
	if err := save(s); err != nil {
		log.Fatal("BONK!", err)
		panic("Unable to set state! " + err.Error())
	}
}

// 'public' function

//
func Whoami() (*Status, error) {
	s := &Status{}

	//
	if err := get(list.LocalNode().Name, s); err != nil {
		return nil, err
	}

	return s, nil
}

//
func Whois(role string) (*Status, error) {

	var conn string

	for _, m := range list.Members() {
		if m.Name == role {
			conn = fmt.Sprintf("%s:%s", m.Addr, strconv.FormatInt(int64(m.Port+1), 10))
		}
	}

	//
	client, err := rpc.Dial("tcp", conn)
	if err != nil {
		return nil, err
	}

	s := &Status{}

	if err := client.Call("Status.Ping", role, s); err != nil {
		return nil, err
	}

	//
	client.Close()

	return s, nil
}

//
func Cluster() ([]*Status, error) {
	var members = []*Status{}

	//
	for _, m := range list.Members() {

		//
		s, err := Whois(m.Name)
		if err != nil {
			return nil, err
		}

		//
		members = append(members, s)
	}

	return members, nil
}

// RPC methods

//
func (s *Status) Ping(role string, status *Status) error {

	//
	for _, m := range list.Members() {
		if m.Name == role {
			if err := get(role, status); err != nil {
				return err
			}
		}
	}

	return nil
}

//
func (s *Status) Demote(source string, status *Status) error {
	advice <- "demote"

	return nil
}

// 'private' functions

//
func get(role string, status *Status) error {
	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

//
func save(status *Status) error {
	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
