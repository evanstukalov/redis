package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/redis"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Config struct {
	Port int
	Role string
}

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	replicaOf := flag.String("replicaof", "", "Replica to another server")

	flag.Parse()

	if *replicaOf == "" {
		*replicaOf = "master"
	} else {
		*replicaOf = "slave"
	}

	config := Config{
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

func handleConnection(ctx context.Context, conn net.Conn, config Config) {
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

		var px *int

		switch strings.ToUpper(args[0]) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			msg := args[1]
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)))
		case "SET":
			key, value := args[1], args[2]

			if len(args) > 3 {
				switch strings.ToUpper(args[3]) {
				case "PX":
					parsedPx, err := strconv.Atoi(args[4])
					if err != nil {
						conn.Write([]byte("px arg in not valid"))
						return
					}
					px = &parsedPx
				}
			}

			storeFromContext := ctx.Value("store")

			if storeFromContext != nil {
				if store, ok := storeFromContext.(*store.Store); !ok {
					log.Fatalf("Expected *store.Store, got %T", storeFromContext)
				} else {

					store.Set(key, value, px)
					conn.Write([]byte("+OK\r\n"))
				}
			}
		case "GET":

			// TODO: fix boilerplate
			// TODO: what can be better than ctx

			key := args[1]

			storeFromContext := ctx.Value("store")

			if storeFromContext != nil {
				if store, ok := storeFromContext.(*store.Store); !ok {
					log.Fatalf("Expected *store.Store, got %T", storeFromContext)
				} else {
					value, err := store.Get(key)
					if err != nil {
						conn.Write([]byte("$-1\r\n"))
					} else {
						conn.Write([]byte(fmt.Sprintf("+%s\r\n", value)))
					}
				}
			}
		case "INFO":

			switch args[1] {
			case "replication":
				info := fmt.Sprintf("role:%s", config.Role)
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(info), info)))
			default:
				conn.Write([]byte("-Error\r\n"))
			}

		default:
			conn.Write([]byte("-Error\r\n"))
		}
	}
}
