// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package commands

import "github.com/spf13/cobra"

//
var (

	//
	YokeCmd = &cobra.Command{
		Use:   "yoke",
		Short: "",
		Long:  ``,
	}

	// subcommands
	clusterCmd = &cobra.Command{Use: "cluster", Short: "", Long: ``}
	memberCmd  = &cobra.Command{Use: "member", Short: "", Long: ``}

	// flags
	fHost string //
	fPort string //
)

// init creates the list of available nanobox commands and sub commands
func init() {

	// persistent flags
	YokeCmd.PersistentFlags().StringVarP(&fHost, "host", "H", "localhost", "")
	YokeCmd.PersistentFlags().StringVarP(&fPort, "port", "p", "4400", "")

	//
	YokeCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterListCmd)

	//
	YokeCmd.AddCommand(memberCmd)
	memberCmd.AddCommand(memberDemoteCmd)
}
