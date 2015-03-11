package commands

import (
  "flag"
  "fmt"
  "net/rpc"
)

// MemberDemoteCommand satisfies the Command interface for rebuilding an app
type MemberDemoteCommand struct{}

// Help prints detailed help text for the app rebuild command
func (c *MemberDemoteCommand) Help() {
  fmt.Printf(`
Description:
  Sends a 'demote' suggestion action

Usage:
  yoke demote [-m member]

  ex. pagoda demote -m asdf
  `)
}

// Run attempts to rebuild an app
func (c *MemberDemoteCommand) Run(opts []string) {

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
  }

  //
  defer client.Close()

  //
  if err := client.Call("Status.Demote", "", nil); err != nil {
    fmt.Println("Failed to call!", err)
  }

  fmt.Printf("'%s' advised to demote...", fHost)
}
