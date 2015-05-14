package commands

import (
	"fmt"
)

// HelpCommand satisfies the Command interface for obtaining user info
type HelpCommand struct{}

// Help prints detailed help text for the user command
func (c *HelpCommand) Help() {
	fmt.Println(`
Description:
  Prints help text for entire CLI

Usage:
  cli
  cli help
  cli --help

  ex. cli help
  `)
}

// Run prints out the help text for the entire CLI
func (c *HelpCommand) Run(opts []string) {
	fmt.Println(`
Description:
  This CLI should provide a few methods to make monitoring your yoke cluster easier

  All commands have a short [-*] and a verbose [--*] option when passing flags.

  You can pass help, or --help to any command to receive detailed information
  about that command.

Usage:
  cli (<COMMAND>:<ACTION> OR <ALIAS>) [GLOBAL FLAG] <POSITIONAL> [SUB FLAGS]

Options:
  help, --help
    Run anytime to receive detailed information about a command.

Available Commands:
  list   : Returns status information for all nodes in the cluster
  demote : Advises a node to 'demote'
  `)
}
