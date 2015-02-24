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

	log.Info("[STATUS - StatusStart] ...\n")

	//
	status = &Status{CRole: conf.Role, DBRole: "initialized", State: "booting", UpdatedAt: time.Now()}

	log.Debug("[STATUS] Created status: %+v\n", status)

	//
	store = scribble.New("./status", log)
	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: status}
	if err := store.Transact(t); err != nil {
		return err
	}

	//
	rpc.Register(status)

	log.Info("[STATUS] Starting RPC server...\n")

	// RPC SERVER
	l, err := net.Listen("tcp", ":" + strconv.FormatInt(int64(conf.ClusterPort+1), 10))
	if err != nil {
		log.Error("[STATUS] Unable to start server! %v\n", err)
		return err
	}

	//
	go func(l net.Listener) {
		for {
			if conn, err := l.Accept(); err != nil {
				log.Error("[STATUS] RPC server - failed to accept connection\n", err.Error())
			} else {
				log.Trace("[STATUS] RPC server - new connection established\n")
				go rpc.ServeConn(conn)
			}
		}
	}(l)

	return nil
}

// 'public' methods

//
func (s *Status) SetDBRole(role string) {
	log.Info("[STATUS - SetDBRole] setting role '%s' on node '%s'\n", role, s.CRole)

	s.DBRole = role

	if err := save(s); err != nil {
		log.Fatal("[STATUS - SetDBRole] Failed to save status! %s", err)
		panic(err)
	}

	s.UpdatedAt = time.Now()
}

//
func (s *Status) SetState(state string) {
	log.Info("[STATUS - SetState] setting '%s' on '%s'\n", state, s.CRole)

	s.State = state

	if err := save(s); err != nil {
		log.Fatal("[STATUS - SetDBRole] Failed to save status! %s", err)
		panic(err)
	}
}

// 'public' function

//
func Whoami() (*Status, error) {
	log.Info("[STATUS - Whoami] I am '%s'", list.LocalNode().Name)
	log.Debug("[STATUS - Whoami] list.LocalNode() - %+v", list.LocalNode())

	s := &Status{}

	//
	if err := get(list.LocalNode().Name, s); err != nil {
		return nil, err
	}

	return s, nil
}

//
func Whois(role string) (*Status, error) {
	log.Info("[STATUS - Whois] Who is '%s'?", role)

	var conn string

	for _, m := range list.Members() {
		if m.Name == role {
			conn = fmt.Sprintf("%s:%s", m.Addr, strconv.FormatInt(int64(m.Port+1), 10))
		}
	}

	log.Debug("[STATUS - Whois] connection - %s", conn)

	//
	client, err := rpc.Dial("tcp", conn)
	if err != nil {
		log.Error("[STATUS - Whois] RPC Client unable to dial! %s", err.Error())
		return nil, err
	}

	s := &Status{}

	if err := client.Call("Status.Ping", role, s); err != nil {
		log.Error("[STATUS - Whois] RPC Client unable to call! %s", err.Error())
		return nil, err
	}

	//
	client.Close()

	return s, nil
}

//
func Cluster() ([]*Status, error) {
	// log.Info("[STATUS - Cluster]")

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

	log.Info("[STATUS - Cluster] members - %+v", members)

	return members, nil
}

// RPC methods

//
func (s *Status) Ping(role string, status *Status) error {

	log.Info("[STATUS - Ping] pinging '%s'...", role)

	//
	for _, m := range list.Members() {
		if m.Name == role {
			if err := get(role, status); err != nil {
				log.Error("[STATUS - Ping] Unable to read '%s'! %s", role, err.Error())
				return err
			}
		}
	}

	return nil
}

//
func (s *Status) Demote(source string, status *Status) error {
	log.Info("[STATUS - Demote] Advising demote...")

	advice <- "demote"

	return nil
}

// 'private' functions

//
func get(role string, status *Status) error {
	log.Debug("[STATUS - get] Attempting to get node '%s'", role)

	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

//
func save(status *Status) error {
	log.Debug("[STATUS - get] Attempting to save node '%s'", status.CRole)

	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
