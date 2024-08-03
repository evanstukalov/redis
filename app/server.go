package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/redis"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	log.Println("Initializing store!")

	store := store.NewStore()

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, "store", store)

		go handleConnection(ctx, conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
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

		switch strings.ToUpper(args[0]) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			msg := args[1]
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)))
		case "SET":
			key, value := args[1], args[2]

			storeFromContext := ctx.Value("store")

			if storeFromContext != nil {
				if store, ok := storeFromContext.(*store.Store); !ok {
					log.Fatalf("Expected *store.Store, got %T", storeFromContext)
				} else {
					store.Set(key, value)
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
					}
					conn.Write([]byte(fmt.Sprintf("+%s\r\n", value)))
				}
			}

		default:
			conn.Write([]byte("-Error\r\n"))
		}
	}
}
