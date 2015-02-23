package main
 
import (
 "net"
 "net/rpc"
 "strconv"
 "fmt"
 "github.com/hashicorp/memberlist"
)

type Status struct {
	Role string
	State string
}

type Statuses []Status

func (self *Status) save() {
	
}

func (self *Status) retrieve() {
	self.Role = conf.Role
	self.State = "booting"
}

func (self *Status) updateState(state string) {
	self.State = state
	self.save()
}

func (self *Status) Get(who string, reply *Status) error {
  fmt.Println("whos asking: " + who)
  self.retrieve()
  reply = self
  return nil
}

func (self *Status) Cluster(who string, reply *Statuses) error {
	for _, member := range list.Members() {
		append(reply, GetNodeStatus(member))
	}
	return nil
}

func GetNodeStatus(memberlist.Node) Status {

}

var status *Status

func RpcStart() error {
  status = new(Status)
  rpc.Register(status)
  listener, err := net.Listen("tcp", ":"+ strconv.FormatInt(int64(conf.ClusterPort + 1), 10))
  if err != nil { return err }

  go func(listener net.Listener) {
	  for {
	    if conn, err := listener.Accept(); err != nil {
	      fmt.Println("accept error: " + err.Error())
	    } else {
	      fmt.Printf("new connection established\n")
	      go rpc.ServeConn(conn)
	    }
	  }
  }(listener)
  return nil
}
