// status.go - status defines a Status struct for each node in the cluster,
// providing four attributes (CRole, DBRole, State, UpdateAt) as a way of
// determining each nodes role in the cluster, current state, and the role of
// the pgqsl running inside each (non-monitor) node.
//
// status provides methods for updating the DBRole and State as the clusters
// environment changes due to outages, and also provides methods of retrieving
// information about each node in the cluster or about the cluster as a whole.

package main

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"time"

	"github.com/nanobox-core/scribble"
)

// Status represents the Status of node in the cluser
type Status struct {
	CRole     string    // the nodes 'role' in the cluster (primary, secondary, monitor)
	DBRole    string    // the 'role' of the running pgsql instance inside the node (master, slave)
	State     string    // the current state of the node
	UpdatedAt time.Time // the last time the node state was updated
}

var (
	status *Status
	store  *scribble.Driver
)

// StatusStart
func StatusStart() error {

	// create a new scribble store
	store = scribble.New("./status", log)

	// the current node will attempt to discover who they are in the event of an
	// outage, or be created for the first time
	s, err := Whoami()
	if err != nil {
		log.Warn("[STATUS - StatusStart] 404 Not found: No record found for '%s'\n", conf.Role)

		status = &Status{CRole: conf.Role, DBRole: "initialized", State: "booting", UpdatedAt: time.Now()}

		log.Debug("[STATUS - StatusStart] Creating record for '%s'\n", conf.Role)
		save(status)
	}

	status = s

	log.Debug("[STATUS] Node Status: %+v\n", status)

	// register our Status struct with RPC
	rpc.Register(status)

	log.Info("[STATUS] Starting RPC server...\n")

	// fire up an RPC (tcp) server
	l, err := net.Listen("tcp", ":"+strconv.FormatInt(int64(conf.ClusterPort+1), 10))
	if err != nil {
		log.Error("[STATUS] Unable to start server! %v\n", err)
		return err
	}

	// daemonize the server
	go func(l net.Listener) {

		// accept connections on the RPC server
		for {
			if conn, err := l.Accept(); err != nil {
				log.Error("[STATUS] RPC server - failed to accept connection\n", err.Error())
			} else {
				log.Debug("[STATUS] RPC server - new connection established\n")

				//
				go rpc.ServeConn(conn)
			}
		}
	}(l)

	return nil
}

// Whoami attempts to pull a matching record from scribble for the local node
// returned from memberlist
func Whoami() (*Status, error) {
	log.Debug("[STATUS - Whoami] list.LocalNode() - %+v", list.LocalNode())

	s := &Status{}

	// attempt to pull a record from scribble for the current node
	if err := get(list.LocalNode().Name, s); err != nil {
		return nil, err
	}

	return s, nil
}

// Whois takes a 'role' string and iterates over all the nodes in memberlist looking
// for a matching node. If a matching node is found it then creates an RPC client
// which is used to make an RPC call to the matching node, which returns a Status
// object for that node
func Whois(role string) (*Status, error) {
	log.Debug("[STATUS - Whois] Who is '%s'?", role)

	var conn string

	// find a matching node for the desired 'role'
	for _, m := range list.Members() {
		if m.Name == role {
			conn = fmt.Sprintf("%s:%s", m.Addr, strconv.FormatInt(int64(m.Port+1), 10))
		}
	}

	log.Debug("[STATUS - Whois] connection - %s", conn)

	// create an RPC client that will connect to the matching node
	client, err := rpc.Dial("tcp", conn)
	if err != nil {
		log.Error("[STATUS - Whois] RPC Client unable to dial! %s", err.Error())
		return nil, err
	}

	s := &Status{}

	// 'ping' the matching node to retrieve its Status
	if err := client.Call("Status.Ping", role, s); err != nil {
		log.Error("[STATUS - Whois] RPC Client unable to call! %s", err.Error())
		return nil, err
	}

	//
	client.Close()

	return s, nil
}

// Cluster iterates over all the nodes in member list, running a Whois(), and
// storing each corresponding Status into a slice and returning the collection
func Cluster() ([]*Status, error) {
	// log.Info("[STATUS - Cluster]")

	var members = []*Status{}

	// iterate over all nodes in member list
	for _, m := range list.Members() {

		// retrieve each nodes Status
		s, err := Whois(m.Name)
		if err != nil {
			return nil, err
		}

		// append each status into our slice of statuses
		members = append(members, s)
	}

	log.Debug("[STATUS - Cluster] members - %+v", members)

	return members, nil
}

// SetDBRole takes a 'role' string and attempts to set the Status.DBRole, and then
// update the record via scribble
func (s *Status) SetDBRole(role string) {
	log.Debug("[STATUS - SetDBRole] setting role '%s' on node '%s'\n", role, s.CRole)

	s.DBRole = role

	if err := save(s); err != nil {
		log.Fatal("[STATUS - SetDBRole] Failed to save status! %s", err)
		panic(err)
	}

	s.UpdatedAt = time.Now()
}

// SetState takes a 'state' string and attempts to set the Status.State, and then
// update the record via scribble
func (s *Status) SetState(state string) {
	log.Debug("[STATUS - SetState] setting '%s' on '%s'\n", state, s.CRole)

	s.State = state

	if err := save(s); err != nil {
		log.Fatal("[STATUS - SetDBRole] Failed to save status! %s", err)
		panic(err)
	}
}

// Ping makes an RPC call to a desired node (by 'role') attempting to return the
// Status of that node. It will iterate through each node in memberlist until it
// finds a matching node
func (s *Status) Ping(role string, status *Status) error {

	log.Info("[STATUS - Ping] pinging '%s'...", role)

	// iterate through each node in memberlist looking for a node whos name matches
	// the desired 'role'
	for _, m := range list.Members() {
		if m.Name == role {

			// attempt to retrieve that Status of the node from scribble
			if err := get(role, status); err != nil {
				log.Error("[STATUS - Ping] Unable to read '%s'! %s", role, err.Error())
				return err
			}
		}
	}

	return nil
}

// Demote is used as a way to 'advise' the current node that it needs to demote
func (s *Status) Demote(source string, status *Status) error {
	log.Info("[STATUS - Demote] Advising demote...")

	advice <- "demote"

	return nil
}

// get retrieves a Status from scribble by 'role'
func get(role string, status *Status) error {
	log.Debug("[STATUS - get] Attempting to get node '%s'", role)

	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

// save saves a Status to scribble by 'role'
func save(status *Status) error {
	log.Debug("[STATUS - get] Attempting to save node '%s'", status.CRole)

	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
