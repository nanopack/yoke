// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package commands

import (
	"fmt"
	"net/rpc"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
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

//
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "Returns status information for all nodes in the cluster",
	Long:  ``,

	Run: clusterList,
}

// clusterList displays select information about all of the nodes in a cluster
func clusterList(ccmd *cobra.Command, args []string) {

	// create an RPC client that will connect to the designated node
	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%s", fHost, fPort))
	if err != nil {
		fmt.Println("[cli.ClusterList.run] Failed to dial!", err)
		os.Exit(1)
	}
	defer client.Close()

	// issue a request to the designated node for the status of the cluster
	var members = &[]Status{}
	if err := client.Call("Status.RPCCluster", "", members); err != nil {
		fmt.Println("[cli.ClusterList.run] Failed to call!", err)
		os.Exit(1)
	}

	//
	fmt.Println(`
Cluster Role |   Cluster IP    |     State     |    Status    |  Postgres Role  |  Postgres Port  |      Last Updated
---------------------------------------------------------------------------------------------------------------------------`)
	for _, member := range *members {

		state := "--"
		status := "running"

		//
		if subMatch := regexp.MustCompile(`^\((.*)\)(.*)$`).FindStringSubmatch(member.State); subMatch != nil {
			state = subMatch[1]
			status = subMatch[2]
		}

		//
		fmt.Printf("%-12s | %-15s | %-13s | %-12s | %-15s | %-15d | %-25s\n", member.CRole, member.Ip, state, status, member.DBRole, member.PGPort, member.UpdatedAt.Format("01.02.06 (15:04:05) MST"))
	}

	fmt.Println("")
}
