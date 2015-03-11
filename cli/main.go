package main

import (
	"fmt"
	"os"

	"github.com/pagodabox-tools/yoke/cli/commands"
)

//
func main() {
	// command line args w/o program
	args := os.Args[1:]

	// if only program is run, print help by default
	if len(args) <= 0 {
		help()
	}

	// parse command line args; it's safe to assume that args[0] is the command we
	// want to run, or one of our 'shortcut' flags that we'll catch before trying
	// to run the command.
	command := args[0]

	// check for 'global' commands
	switch command {

	// check for help shortcuts
	case "--help", "help":
		help()

	// we didn't find a 'shortcut' flag, so we'll continue parsing the remaining
	// args looking for a command to run.
	default:

		// if we find a valid command we run it
		if val, ok := commands.Commands[command]; ok {

			// args[1:] will be our remaining subcommand or flags after the intial command.
			// This value could also be 0 if running an alias command.
			opts := args[1:]

			//
			if len(opts) >= 1 {
				switch opts[0] {

				// Check for help shortcuts on commands
				case "--help", "help":
					commands.Commands[command].Help()
					os.Exit(0)
				}
			}

			// run the command
			val.Run(opts)

			// no valid command found
		} else {
			fmt.Printf("'%s' is not a valid command. Type 'cli' for available commands and usage.\n", command)
			os.Exit(1)
		}
	}
}

// help
func help() {
	cmd := commands.Commands["help"]
	cmd.Run(nil)
	os.Exit(0)
}
