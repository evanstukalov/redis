package clients

import (
	"net"
	"sync"
)

type Clients struct {
	Clients map[net.Conn]struct{}
	Mutex   sync.Mutex
}

func NewClients() *Clients {
	return &Clients{
		Clients: make(map[net.Conn]struct{}),
	}
}

func (cl *Clients) Set(client net.Conn) {
	cl.Mutex.Lock()
	defer cl.Mutex.Unlock()

	cl.Clients[client] = struct{}{}
}
