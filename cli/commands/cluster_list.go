package commands

import (
  "flag"
  "fmt"
  "net/rpc"
  "os"
  "time"
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

// ClusterListCommand satisfies the Command interface for listing a user's apps
type ClusterListCommand struct{}

// Help prints detailed help text for the app list command
func (c *ClusterListCommand) Help() {
  fmt.Printf(`
Description:
  Lists all the members (nodes) in the cluster.

Usage:
  yoke list
  yoke cluster:list

  ex. yoke list
  `)
}

// Run displays select information about all of a user's apps
func (c *ClusterListCommand) Run(opts []string) {

  // flags
  flags := flag.NewFlagSet("flags", flag.ContinueOnError)
  flags.Usage = func() { c.Help() }

  var fHost string
  flags.StringVar(&fHost, "o", "localhost", "")
  flags.StringVar(&fHost, "host", "localhost", "")

  var fPort string
  flags.StringVar(&fPort, "p", "4401", "")
  flags.StringVar(&fPort, "port", "4401", "")

  if err := flags.Parse(opts); err != nil {
    fmt.Println("Failed to parse flags!", err)
  }

  // create an RPC client that will connect to the matching node
  client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%s", fHost, fPort))
  if err != nil {
    fmt.Println("Failed to dial!", err)
    os.Exit(1)
  }

  //
  defer client.Close()

  var members = &[]Status{}

  //
  if err := client.Call("Status.RPCCluster", "", members); err != nil {
    fmt.Println("Failed to call!", err)
    os.Exit(1)
  }


  //
  fmt.Println(`
Cluster Role |   Cluster IP    |       State        |  Postgres Role  |  Postgres Port  |      Last Updated
-----------------------------------------------------------------------------------------------------------------`)
  for _, member := range *members {
    fmt.Printf("%-12s | %-15s | %-18s | %-15s | %-15d | %-25s\n", member.CRole, member.Ip, member.State, member.DBRole, member.PGPort, member.UpdatedAt.Format("01.02.06 (15:04:05) MST"))
  }

  fmt.Println("")
}
