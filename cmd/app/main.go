package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/clients"
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/master"
	"github.com/codecrafters-io/redis-starter-go/internal/slave"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	replicaOf := flag.String("replicaof", "", "Replica to another server")

	flag.Parse()

	config := config.Config{
		Port:             *port,
		MasterReplId:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		MasterReplOffset: 0,
	}

	storeObj := store.NewStore()
	expiredCollector := store.NewExpiredCollector(storeObj)
	defer expiredCollector.Stop()
	clients := clients.NewClients()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "store", storeObj)
	ctx = context.WithValue(ctx, "clients", clients)

	address := fmt.Sprintf("0.0.0.0:%d", config.Port)

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port ", config.Port)
		os.Exit(1)
	}
	defer l.Close()

	connChan := make(chan net.Conn)
	errChan := make(chan error)

	if *replicaOf == "" {
		config.Role = "master"
	} else {
		config.Role = "slave"
		if masterConn, err := slave.Handshakes(*replicaOf, config); err != nil {
			log.Fatalln("There is was error in handshakes with master : ", err)
		} else {
			go slave.ReadFromConnection(ctx, masterConn, config)
		}
	}

	go master.AcceptConnections(l, connChan, errChan)
	go expiredCollector.Tick()

	for {
		select {
		case conn := <-connChan:

			go master.ReadFromConnection(ctx, conn, config)

		case err := <-errChan:
			fmt.Println("Error accepting connection", err.Error())

		}
	}
}
