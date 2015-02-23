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

	log.Info("[YOKE - StatusStart] ...\n")

	//
	status = &Status{CRole: conf.Role, DBRole: "initialized", State: "booting", UpdatedAt: time.Now()}

	log.Info("[YOKE] Creating Status %+v\n", status)

	//
	store = scribble.New("./status", log)
	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: status}
	if err := store.Transact(t); err != nil {
		return err
	}

	//
	rpc.Register(status)

	log.Info("[YOKE] Starting RPC server...\n")

	// RPC SERVER
	l, err := net.Listen("tcp", ":" + strconv.FormatInt(int64(conf.ClusterPort+1), 10))
	if err != nil {
		log.Error("[YOKE] Unable to start server! %v\n", err)
		return err
	}

	//
	go func(l net.Listener) {
		for {
			if conn, err := l.Accept(); err != nil {
				log.Error("[YOKE] RPC server - failed to accept connection\n", err.Error())
			} else {
				log.Trace("[YOKE] RPC server - new connection established\n")
				go rpc.ServeConn(conn)
			}
		}
	}(l)

	return nil
}

// 'public' methods

//
func (s *Status) SetDBRole(role string) {
	log.Info("[YOKE - SetDBRole] setting '%s' on '%s'\n", role, s.CRole)

	s.DBRole = role

	if err := save(s); err != nil {
		log.Fatal("BONK!", err)
		panic("Unable to set db role! " + err.Error())
	}

	s.UpdatedAt = time.Now()
}

//
func (s *Status) SetState(state string) {
	log.Info("[YOKE - SetState] setting '%s' on '%s'\n", state, s.CRole)

	s.State = state

	if err := save(s); err != nil {
		log.Fatal("BONK!", err)
		panic("Unable to set state! " + err.Error())
	}
}

// 'public' function

//
func Whoami() (*Status, error) {
	log.Info("[YOKE - Whoami] I am '%s'", list.LocalNode().Name)
	log.Debug("[YOKE - Whoami] list.LocalNode() - %+v", list.LocalNode())

	s := &Status{}

	//
	if err := get(list.LocalNode().Name, s); err != nil {
		return nil, err
	}

	return s, nil
}

//
func Whois(role string) (*Status, error) {
	log.Info("[YOKE - Whois] Who is '%s'", role)

	var conn string

	for _, m := range list.Members() {
		if m.Name == role {
			conn = fmt.Sprintf("%s:%s", m.Addr, strconv.FormatInt(int64(m.Port+1), 10))
		}
	}

	log.Debug("[YOKE - Whois] connection - %s", conn)

	//
	client, err := rpc.Dial("tcp", conn)
	if err != nil {
		log.Error("[YOKE - Whois] RPC Client unable to dial! %s", err.Error())
		return nil, err
	}

	s := &Status{}

	if err := client.Call("Status.Ping", role, s); err != nil {
		log.Error("[YOKE - Whois] RPC Client unable to call! %s", err.Error())
		return nil, err
	}

	//
	client.Close()

	return s, nil
}

//
func Cluster() ([]*Status, error) {
	log.Info("[YOKE - Cluster]")

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

	log.Info("[YOKE - Cluster] members - %+v", members)

	return members, nil
}

// RPC methods

//
func (s *Status) Ping(role string, status *Status) error {

	log.Info("[YOKE - Ping] pinging '%s'", role)

	//
	for _, m := range list.Members() {
		if m.Name == role {
			if err := get(role, status); err != nil {
				log.Error("[YOKE - Ping] Unable to read '%s'! %s", role, err.Error())
				return err
			}
		}
	}

	return nil
}

//
func (s *Status) Demote(source string, status *Status) error {
	log.Info("[YOKE - Demote] Advising demote...")

	advice <- "demote"

	return nil
}

// 'private' functions

//
func get(role string, status *Status) error {
	log.Info("[YOKE - get] Attempting to get '%s'", role)

	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

//
func save(status *Status) error {
	log.Info("[YOKE - get] Attempting to save '%s'", status.CRole)

	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
