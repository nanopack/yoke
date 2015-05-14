package commands

import (
	"flag"
	"fmt"
	"net/rpc"
)

// MemberDemoteCommand satisfies the Command interface for demoting a node
type MemberDemoteCommand struct{}

// Help prints detailed help text for the member demote command
func (c *MemberDemoteCommand) Help() {
	fmt.Printf(`
Description:
  Advises a node to 'demote'

Usage:
  cli demote [-m member]

  ex. cli demote -m asdf
  `)
}

// Run attempts to rebuild an app
func (c *MemberDemoteCommand) Run(opts []string) {

	// flags
	flags := flag.NewFlagSet("flags", flag.ContinueOnError)
	flags.Usage = func() { c.Help() }

	var fHost string
	flags.StringVar(&fHost, "h", "localhost", "")
	flags.StringVar(&fHost, "host", "localhost", "")

	var fPort string
	flags.StringVar(&fPort, "p", "4400", "")
	flags.StringVar(&fPort, "port", "4400", "")

	if err := flags.Parse(opts); err != nil {
		fmt.Println("[cli.MemberDemote.Run] Failed to parse flags!", err)
	}

	// create an RPC client that will connect to the matching node
	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%s", fHost, fPort))
	if err != nil {
		fmt.Println("[cli.MemberDemote.Run] Failed to dial!", err)
	}

	//
	defer client.Close()

	//
	if err := client.Call("Status.Demote", "", nil); err != nil {
		fmt.Println("[cli.MemberDemote.Run] Failed to call!", err)
	}

	fmt.Printf("'%s' advised to demote...", fHost)
}
