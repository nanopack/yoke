package commands

import (
  "github.com/pagodabox-tools/cli/ui"
)

// HelpCommand satisfies the Command interface for obtaining user info
type HelpCommand struct{}

// Help prints detailed help text for the user command
func (c *HelpCommand) Help() {
  ui.CPrint(`
Description:
  Prints help text for entire CLI

Usage:
  pagoda
  pagoda -h
  pagoda --help
  pagoda help

  ex. pagoda help
  `)
}

// Run prints out the help text for the entire CLI
func (c *HelpCommand) Run(opts []string) {
  ui.CPrint(`
Description:
  This CLI should provide a few methods to make monitoring your yoke cluster easier

  All commands have a short [-*] and a verbose [--*] option when passing flags.

  You can pass -h, --help, or help to any command to receive detailed information
  about that command.

Usage:
  pagoda (<COMMAND>:<ACTION> OR <ALIAS>) [GLOBAL FLAG] <POSITIONAL> [SUB FLAGS]

Options:
  -h, --help, help
    Run anytime to receive detailed information about a command.

Available Commands:
  list            : List all your applications.
  demote          : Display info about an application.
  `)
}
