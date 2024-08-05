package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/redis"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	replicaOf := flag.String("replicaof", "", "Replica to another server")

	flag.Parse()

	if *replicaOf == "" {
		*replicaOf = "master"
	} else {
		*replicaOf = "slave"
	}

	config := config.Config{
		Port: *port,
		Role: *replicaOf,
	}

	storeObj := store.NewStore()
	expiredCollector := store.NewExpiredCollector(storeObj)
	defer expiredCollector.Stop()

	address := fmt.Sprintf("0.0.0.0:%d", config.Port)

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	connChan := make(chan net.Conn)
	errChan := make(chan error)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {

				errChan <- err
				return
			}

			connChan <- conn
		}
	}()

	go func() {
		for {
			select {
			case <-expiredCollector.Ticker.C:
				expiredCollector.Collect()
			}
		}
	}()

	for {
		select {
		case conn := <-connChan:

			ctx := context.Background()
			ctx = context.WithValue(ctx, "store", storeObj)

			go handleConnection(ctx, conn, config)

		case err := <-errChan:
			fmt.Println("Error accepting connection", err.Error())

		}
	}
}

func handleConnection(ctx context.Context, conn net.Conn, config config.Config) {
	defer conn.Close()

	for {
		r := bufio.NewReader(conn)

		args, err := redis.UnpackInput(r)
		if err != nil {
			break
		}

		if len(args) == 0 {
			break
		}

		commands.HandleCommand(ctx, conn, config, args)
	}
}
