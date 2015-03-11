package commands

// Commands represents a map of all the available commands that the Pagoda Box
// CLI can run
var Commands map[string]Command

// Command represents a Pagoda Box CLI command. Every command must have a Help()
// and Run() function
type Command interface {
  Help()             // Prints the help text associated with this command
  Run(opts []string) // Houses the logic that will be run upon calling this command
}

// init builds the list of available Pagoda Box CLI commands
func init() {

  // the map of all available commands the Pagoda Box CLI can run
  Commands = map[string]Command{
    "help":             &HelpCommand{},
    "list":             &ClusterListCommand{},
    "cluster:list":     &ClusterListCommand{},
    "demote":           &MemberDemoteCommand{},
    "member:demote":    &MemberDemoteCommand{},
  }
}
