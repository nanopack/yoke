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

var (
	status *Status
	store  *scribble.Driver
)

//
func StatusStart() error {

	//
	s := Status{CRole: conf.Role, State: "booting"}

	//
	store = scribble.New("./status", log)
	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: "node", Container: &s}
	if err := store.Transact(t); err != nil {
		return err
	}

	//
	rpc.Register(s)

	// RPC SERVER
	l, err := net.Listen("tcp", ":" + strconv.FormatInt(int64(conf.ClusterPort+1), 10))
	if err != nil {
		return err
	}

	//
	go func(l net.Listener) {
		for {
			if conn, err := l.Accept(); err != nil {
				fmt.Println("accept error: " + err.Error())
			} else {
				fmt.Printf("new connection established\n")
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
		status, err := Whois(m.Name)
		if err != nil {
			return nil, err
		}

		//
		members = append(members, status)
	}

	return members, nil
}

//
func Demote() error {
	return nil
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

// 'private' functions

//
func get(role string, v interface{}) error {
	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &v}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

//
func save(v interface{}) error {
	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: "node", Container: &v}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
