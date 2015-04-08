// status.go defines a Status struct for each node in the cluster, providing
// four attributes (CRole, DBRole, State, UpdateAt) as a way of determining each
// nodes role in the cluster, current state, and the role of the pgqsl running
// inside each (non-monitor) node.
//
// status provides methods for updating the DBRole and State as the clusters
// environment changes due to outages, and also provides methods of retrieving
// information about each node in the cluster or about the cluster as a whole.

package main

import (
	"errors"
	"fmt"
	"github.com/nanobox-core/scribble"
	"net"
	"net/rpc"
	"os"
	"time"
)

// Status represents the Status of a node in the cluser
type Status struct {
	CRole     string    // the nodes 'role' in the cluster (primary, secondary, monitor)
	DataDir   string    // directory of the postgres database
	DBRole    string    // the 'role' of the running pgsql instance inside the node (master, slave)
	Ip        string    // advertise_ip
	PGPort    int       //
	State     string    // the current state of the node
	UpdatedAt time.Time // the last time the node state was updated
}

var (
	status *Status
	store  *scribble.Driver
)

// StatusStart
func StatusStart() error {
	log.Info("[(%s) status.StatusStart]", conf.Role)

	var err error

	// create a new scribble store
	store = scribble.New(conf.StatusDir, log)
	status = &Status{}

	// determine if the current node already has a record in scribble
	fi, err := os.Stat(conf.StatusDir + "cluster/" + conf.Role)
	if err != nil {
		log.Warn("[(%s) status.StatusStart] Failed to %s", conf.Role, err)
	}

	// if no record found that matches the current node; create a new record in
	// scribble
	if fi == nil {
		log.Warn("[(%s) status.StatusStart] 404 Not found: Creating record for '%s'", conf.Role, conf.Role)

		//
		status = &Status{
			CRole:     conf.Role,
			DataDir:   conf.DataDir,
			DBRole:    "initialized",
			Ip:        conf.AdvertiseIp,
			PGPort:    conf.PGPort,
			State:     "booting",
			UpdatedAt: time.Now(),
		}

		save(status)

		// record found; set nodes status information
	} else {
		status = Whoami()
	}

	log.Debug("[(%s) status] Node Status: %#v\n", status.CRole, status)

	// register our Status struct with RPC
	rpc.Register(status)

	fmt.Printf("[(%s) status] Starting RPC server... ", status.CRole)

	// fire up an RPC (tcp) server
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.AdvertisePort))
	if err != nil {
		log.Error("[(%s) status] Unable to start server!\n%s\n", status.CRole, err)
		return err
	}

	fmt.Printf("success (listening on port %d)\n", conf.AdvertisePort)

	// daemonize the server
	go func(l net.Listener) {

		// accept connections on the RPC server
		for {
			if conn, err := l.Accept(); err != nil {
				log.Error("[(%s) status] RPC server - failed to accept connection!\n%s", status.CRole, err.Error())
			} else {
				log.Debug("[(%s) status] RPC server - new connection established!\n", status.CRole)

				//
				go rpc.ServeConn(conn)
			}
		}
	}(l)

	return nil
}

// Whoami attempts to pull a matching record from scribble for the local node
// returned from memberlist
func Whoami() *Status {
	log.Debug("[status.Whoami] %s", status.CRole)

	v := &Status{}

	// attempt to pull records from scribble... if unsuccessful fail.
	for i := 0; i < 10; i++ {

		// attempt to pull a record from scribble for the current node
		if err := get(conf.Role, v); err == nil {
			return v
		} else {
			log.Warn("[(%s) status.Whoami] Unable to retrieve record! retrying... (%s)", status.CRole, err)
		}
	}

	log.Fatal("[(%s) status.Whoami] Failed to retrieve record!", status.CRole)
	os.Exit(1)

	return nil
}

// Whois takes a 'role' string and iterates over all the nodes in memberlist looking
// for a matching node. If a matching node is found it then creates an RPC client
// which is used to make an RPC call to the matching node, which returns a Status
// object for that node
func Whois(role string) (*Status, error) {
	log.Debug("[(%s) status.Whois] Who is '%s'?", status.CRole, role)

	conn := getConn(role)
	if conn == "" {
		return nil, errors.New("I could not find a connection string for role(" + role + ")")
	}

	log.Debug("[(%s) status.Whois] connection - %s", status.CRole, conn)

	// create an RPC client that will connect to the matching node
	client, err := rpcClient(conn)
	if err != nil {
		log.Error("[(%s) status.Whois] RPC Client unable to dial!\n%s", status.CRole, err)
		return nil, err
	}

	//
	defer client.Close()

	reply := &Status{}

	//
	c := make(chan error, 1)
	go func() { c <- client.Call("Status.RPCEnsureWhois", status.CRole, reply) }()
	select {
	case err := <-c:
		if err != nil {
			return nil, err
		}
		return reply, err
	case <-time.After(15 * time.Second):
		return nil, errors.New("Timeout waiting for method call")
	}

	return reply, nil
}

// Whoisnot takes a 'role' string and attempts to find the 'other' node that does
// not match the role provided
func Whoisnot(not string) (*Status, error) {
	log.Debug("[(%s) status.Whoisnot] Who is not '%s'?", status.CRole, not)

	var role string

	// set role equal to the 'opposite' of the given role
	switch not {
	case "primary":
		role = "secondary"
	case "secondary":
		role = "primary"
	}

	// return the node that does not match the give role
	return Whois(role)
}

// Cluster iterates over all the nodes in member list, running a Whois(), and
// storing each corresponding Status into a slice and returning the collection
func Cluster(who ...string) []Status {
	var members = &[]Status{}

	if len(who) == 0 {
		log.Debug("[(%s) status.Cluster] Retrieving cluster stats (local)...", status.CRole)

		if err := status.RPCCluster(conf.Role, members); err != nil {
			log.Error("[(%s) status.Cluster] Retrieving cluster stats (%s)", status.CRole, err.Error())
		}
		return *members
	}

	for _, role := range who {

		conn := getConn(role)
		if conn == "" {
			continue
		}
		log.Debug("[(%s) status.Cluster] connection - %s", status.CRole, conn)

		// create an RPC client that will connect to the matching node
		client, err := rpcClient(conn)
		if err != nil {
			log.Error("[(%s) status.Cluster] RPC Client unable to dial!\n%s", status.CRole, err)
			continue
		}
		//
		defer client.Close()

		//
		membs := &[]Status{}
		c := make(chan error, 1)
		go func() { c <- client.Call("Status.RPCCluster", status.CRole, membs) }()
		select {
		case err := <-c:
			if err != nil {
				continue
			}
			*members = append(*members, *membs...)
		case <-time.After(30 * time.Second):
			continue
		}
	}

	return *members
}

// SetDBRole takes a 'role' string and attempts to set the Status.DBRole, and then
// update the record via scribble
func (s *Status) SetDBRole(role string) {
	log.Debug("[(%s) status.SetDBRole] setting role '%s' on node '%s'", s.CRole, role, s.CRole)

	s.DBRole = role

	if err := save(s); err != nil {
		log.Fatal("[(%s) status.SetDBRole] Failed to save status! %s", s.CRole, err)
		panic(err)
	}

	s.UpdatedAt = time.Now()
}

// SetState takes a 'state' string and attempts to set the Status.State, and then
// update the record via scribble
func (s *Status) SetState(state string) {
	log.Debug("[(%s) status.SetState] setting '%s' on '%s'", s.CRole, state, s.CRole)

	s.State = state

	if err := save(s); err != nil {
		log.Fatal("[(%s) status.SetDBRole] Failed to save status! %s", s.CRole, err)
		panic(err)
	}
}

// RPCEnsureWhois ensures full duplex communication between nodes before sending
// back status information for a requested node
func (s *Status) RPCEnsureWhois(asking string, v *Status) error {
	log.Debug("[(%s) status.RPCEnsureWhois] '%s' requesting '%s's' status...", s.CRole, asking, s.CRole)

	conn := getConn(asking)
	if conn == "" {
		return errors.New("I could not find a connection string for role(" + asking + ")")
	}

	log.Debug("[(%s) status.RPCEnsureWhois] connection - %s", status.CRole, conn)

	// create an RPC client that will connect to the matching node
	client, err := rpcClient(conn)
	if err != nil {
		log.Error("[(%s) status.RPCEnsureWhois] RPC Client unable to dial!\n%s", s.CRole, err)
		return err
	}

	//
	defer client.Close()

	tmp := &Status{}
	// 'pingback' to the requesting node to ensure duplex communication
	c := make(chan error, 1)
	go func() { c <- client.Call("Status.RPCWhois", status.CRole, tmp) }()
	select {
	case err := <-c:
		if err != nil {
			return err
		}
		return s.RPCWhois(asking, v)
	case <-time.After(10 * time.Second):
		return errors.New("Timeout waiting for method RPCWhois")
	}

	// if err := client.Call("Status.RPCWhois", asking, tmp); err != nil {
	// 	log.Error("[(%s) status.RPCEnsureWhois] RPC Client unable to call!\n%s", s.CRole, err)
	// 	return err
	// }

	// return status
	return errors.New("Execution failed!")
}

// RPCWhois is the response to an RPC call made from Whois requesting the status
// information for the provided 'role'
func (s *Status) RPCWhois(asking string, v *Status) error {
	log.Debug("[(%s) status.RPCWhois] '%s' providing status to '%s'...", s.CRole, s.CRole, asking)

	// attempt to retrieve that Status of the node from scribble
	if err := get(status.CRole, v); err != nil {
		log.Error("[(%s) status.RPCWhois] Failed to retrieve status!\n%s\n", status.CRole, err.Error())
		return err
	}

	return nil
}

// RPCCluster
func (s *Status) RPCCluster(source string, v *[]Status) error {
	log.Debug("[(%s) status.RPCCluster] Requesting cluster stats...", s.CRole)

	errHandle := func(s *Status, err error) *Status {
		if err != nil {
			log.Warn("[(%s) status.Cluster] Failed to retrieve status!\n%s", status.CRole, err)
		}
		return s
	}

	self := Whoami()
	*v = append(*v, *s)
	switch self.CRole {
	case "monitor":
		if s := errHandle(Whois("primary")); s != nil {
			*v = append(*v, *s)
		}
		if s := errHandle(Whois("secondary")); s != nil {
			*v = append(*v, *s)
		}
	case "primary":
		if s := errHandle(Whois("monitor")); s != nil {
			*v = append(*v, *s)
		}
		if s := errHandle(Whois("secondary")); s != nil {
			*v = append(*v, *s)
		}
	case "secondary":
		if s := errHandle(Whois("monitor")); s != nil {
			*v = append(*v, *s)
		}
		if s := errHandle(Whois("primary")); s != nil {
			*v = append(*v, *s)
		}
	}

	return nil
}

// Demote is used as a way to 'advise' the current node that it needs to demote
func (s *Status) Demote(source string, v *Status) error {
	log.Debug("[(%s) status.Demote] Advising demote...", s.CRole)

	go func() { advice <- "demote" }()

	return nil
}

func rpcClient(conn string) (*rpc.Client, error) {
	c := make(chan error, 1)
	var client *rpc.Client
	var err error

	go func() {
		client, err = rpc.Dial("tcp", conn)
		c <- err
	}()
	select {
	case err := <-c:
		return client, err
	case <-time.After(2 * time.Second):
		return nil, errors.New("Timeout Waiting for rpc.Dial")
	}
}

func getConn(role string) string {
	switch role {
	case "monitor":
		return conf.Monitor
	case "primary":
		return conf.Primary
	case "secondary":
		return conf.Secondary
	}
	return ""
}

// get retrieves a Status from scribble by 'role'
func get(role string, v *Status) error {
	log.Debug("[(%s) status.get] Attempting to get node '%s'", status.CRole, role)

	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &v}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

// save saves a Status to scribble by 'role'
func save(v *Status) error {
	log.Debug("[(%s) status.save] Attempting to save node '%s'", status.CRole, status.CRole)

	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: &v}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
