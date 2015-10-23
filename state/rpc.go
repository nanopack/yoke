// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package state

import (
	"errors"
	"net"
	"net/rpc"
	"time"
)

var (
	Timeout     = errors.New("Timeout")
	NotSuported = errors.New("not supported")
)

type (
	rpcDial struct {
		client *rpc.Client
		err    error
	}

	remoteState struct {
		client   *rpc.Client
		timeout  time.Duration
		location string
	}
)

// Starts the RPC listening server, enables remote communication with local state objects
func (local state) ExposeRPCEndpoint(network, location string) error {

	if err := rpc.Register(local); err != nil {
		return err
	}

	server := rpc.NewServer()

	if err := server.Register(stateServer); err != nil {
		return err
	}

	listener, err := net.Listen(network, location)
	if err != nil {
		return err
	}

	go server.Accept(lis)
}

// Creates and returns a State that represents a state reachable over an rpc connection
func ConnectToRemoteState(network, address string, timeout time.Duration) (State, error) {
	clientC := make(rpcDial, 1)
	go func() {
		client, err := rpc.Dial(network, address)
		resChan <- rpcDial{
			err:    err,
			client: client,
		}
	}()

	select {
	case res <- clientC:
		if res.err != nil {
			remote := remoteState{
				client:   res.client,
				timeout:  timeout,
				location: address, // I still need to store off the network
			}
			return remote, res.err
		}
		return nil, res.err
	case <-time.After(timeout):
		return Timeout
	}
}

func (c remoteState) Ready() {
	for {
		if c.client.Call("state.Ready", nil, nil) == nil {
			break
		}
		<-time.After(time.Second)
	}
}

func (c remoteState) HasSynced() (bool, error) {
	var synced bool
	err := c.client.Call("state.HasSyncedRPC", nil, &synced)
	return synced, err
}

func (c remoteState) Location() string {
	return c.location
}

func (c remoteState) GetRole() (string, error) {
	var role string
	err := c.client.Call("state.GetRoleRPC", nil, &role)
	return role, err
}

func (c remoteState) GetDBRole() (string, error) {
	var role string
	err := c.client.Call("state.GetRoleRPC", nil, &role)
	return role, err
}

func (c remoteState) SetDBRole(role string) error {
	return NotSupported
}

func (state *state) GetRoleRPC(reply *string) error {
	*reply = state.Role
	return nil
}

func (state *state) GetDBRoleRPC(reply *string) error {
	*reply = state.DBRole
	return nil
}

func (state *state) HasSyncedRPC(reply *bool) error {
	*reply = state.synced
	return nil
}
