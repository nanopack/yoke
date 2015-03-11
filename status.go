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
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"github.com/nanobox-core/scribble"
)

// Status represents the Status of node in the cluser
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
	fi, err := os.Stat(conf.StatusDir + "/cluster/" + conf.Role)
	if err != nil {
		log.Warn("[(%s) status.StatusStart] Failed to read '%s'\n%s\n", conf.Role, conf.StatusDir+"/"+conf.Role, err)
	}

	// if no record found that matches the current node; create a new record in
	// scribble
	if fi == nil {
		log.Warn("[(%s) status.StatusStart] 404 Not found: No record found for '%s'\n", conf.Role, conf.Role)

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

		log.Debug("[(%s) status.StatusStart] Creating record for '%s'\n", status.CRole, conf.Role)
		save(status)

		// record found; set nodes status information
	} else {
		status = Whoami()
	}

	log.Debug("[(%s) status] Node Status: %+v\n", status.CRole, status)

	// register our Status struct with RPC
	rpc.Register(status)

	log.Info("[(%s) status] Starting RPC server...\n", status.CRole)

	// fire up an RPC (tcp) server
	l, err := net.Listen("tcp", ":"+strconv.FormatInt(int64(conf.AdvertisePort+1), 10))
	if err != nil {
		log.Error("[(%s) status] Unable to start server!\n%v\n", status.CRole, err)
		return err
	}

	// daemonize the server
	go func(l net.Listener) {

		// accept connections on the RPC server
		for {
			if conn, err := l.Accept(); err != nil {
				log.Error("[(%s) status] RPC server - failed to accept connection!\n%s%n", status.CRole, err.Error())
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
	log.Debug("[(%s) status.Whoami] list.LocalNode() - %+v", status.CRole, list.LocalNode())

	s := &Status{}

	//
	for i := 0; i < 10; i++ {

		// attempt to pull a record from scribble for the current node
		if err := get(list.LocalNode().Name, s); err == nil {
			return s
		} else {
			log.Error("[(%s) status.Whoami] Unable to retrieve record! retrying... (%s)", status.CRole, err)
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

	var conn string

	// find a matching node for the desired 'role'
	for _, m := range list.Members() {
		if m.Name == role {
			conn = fmt.Sprintf("%s:%s", m.Addr, strconv.FormatInt(int64(m.Port+1), 10))
		}
	}

	log.Debug("[(%s) status.Whois] connection - %s", status.CRole, conn)

	// create an RPC client that will connect to the matching node
	client, err := rpc.Dial("tcp", conn)
	if err != nil {
		log.Error("[(%s) status.Whois] RPC Client unable to dial!\n%s\n", status.CRole, err.Error())
		return nil, err
	}

	//
	defer client.Close()

	s := &Status{}

	//
	if err := client.Call("Status.RPCWhois", role, s); err != nil {
		log.Error("[(%s) status.Whois] RPC Client unable to call!\n%s\n", status.CRole, err.Error())
		return nil, err
	}

	return s, nil
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
func Cluster() []*Status {
	log.Debug("[(%s) status.Cluster] Retrieving cluster stats...", status.CRole)

	var members = []*Status{}

	if err := status.RPCCluster("", members); err != nil {
		//
	}

	return members
}

// SetDBRole takes a 'role' string and attempts to set the Status.DBRole, and then
// update the record via scribble
func (s *Status) SetDBRole(role string) {
	log.Debug("[(%s) status.SetDBRole] setting role '%s' on node '%s'\n", status.CRole, role, s.CRole)

	s.DBRole = role

	if err := save(s); err != nil {
		log.Fatal("[(%s) status.SetDBRole] Failed to save status! %s", status.CRole, err)
		panic(err)
	}

	s.UpdatedAt = time.Now()
}

// SetState takes a 'state' string and attempts to set the Status.State, and then
// update the record via scribble
func (s *Status) SetState(state string) {
	log.Debug("[(%s) status.SetState] setting '%s' on '%s'\n", status.CRole, state, s.CRole)

	s.State = state

	if err := save(s); err != nil {
		log.Fatal("[(%s) status.SetDBRole] Failed to save status! %s", status.CRole, err)
		panic(err)
	}
}

// RPCWhois is the response to an RPC call made from Whois requesting the status
// information for the provided 'role'
func (s *Status) RPCWhois(role string, status *Status) error {
	log.Debug("[(%s) status.RPCWhois] '%s' requesting '%s''s status...", status.CRole, s.CRole, role)

	// iterate through each node in memberlist looking for a node whos name matches
	// the desired 'role'
	for _, m := range list.Members() {
		if m.Name == role {

			// attempt to retrieve that Status of the node from scribble
			if err := get(role, status); err != nil {
				log.Error("[(%s) status.RPCWhois] Unable to read '%s'!\n%s\n", status.CRole, role, err.Error())
				return err
			}
		}
	}

	return nil
}

// RPCCluster
func (s *Status) RPCCluster(source string, members []*Status) error {
	log.Debug("[(%s) status.RPCCluster] Requesting cluster stats...", status.CRole)

	cluster := "cluster members - "

	// iterate over all nodes in member list
	for _, m := range list.Members() {

		// retrieve each nodes Status
		s, err := Whois(m.Name)
		if err != nil {
			log.Warn("[(%s) status.Cluster] Failed to retrieve status for '%s'!\n%s\n", status.CRole, m.Name, err)

			// append each status into our slice of statuses
		} else {
			members = append(members, s)
			cluster += fmt.Sprintf("(%s:%s) ", s.CRole, s.Ip)
		}
	}

	log.Info("[(%s) status.Cluster] %s", cluster)

	return nil
}

// Demote is used as a way to 'advise' the current node that it needs to demote
func (s *Status) Demote(source string, status *Status) error {
	log.Debug("[(%s) status.Demote] Advising demote...", status.CRole)

	advice <- "demote"

	return nil
}

// get retrieves a Status from scribble by 'role'
func get(role string, status *Status) error {
	log.Debug("[(%s) status.get] Attempting to get node '%s'", status.CRole, role)

	t := scribble.Transaction{Operation: "read", Collection: "cluster", RecordID: role, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}

// save saves a Status to scribble by 'role'
func save(status *Status) error {
	log.Debug("[(%s) status.save] Attempting to save node '%s'", status.CRole, status.CRole)

	t := scribble.Transaction{Operation: "write", Collection: "cluster", RecordID: status.CRole, Container: &status}
	if err := store.Transact(t); err != nil {
		return err
	}

	return nil
}
