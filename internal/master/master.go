package master

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/clients"
	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/redis"
)

func AcceptConnections(l net.Listener, connChan chan<- net.Conn, errChan chan<- error) {
	for {
		conn, err := l.Accept()
		if err != nil {

			errChan <- err
			return
		}

		connChan <- conn
	}
}

func ReadFromConnection(ctx context.Context, conn net.Conn, config config.Config) {
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

		HandleCommand(ctx, conn, config, args)
	}
}

func HandleCommand(ctx context.Context, conn net.Conn, config config.Config, args []string) {
	cmd, exists := commands.Commands[strings.ToUpper(args[0])]
	if !exists {
		conn.Write([]byte("-Error\r\n"))
		return
	}

	cmd.Execute(ctx, conn, config, args)

	for _, command := range commands.Propagated {
		if command == args[0] {
			go SendCommandAllClients(ctx, conn, args)
		}
	}
}

func SendCommandAllClients(ctx context.Context, conn net.Conn, args []string) {
	clientsFromContext := ctx.Value("clients")
	if clientsFromContext != nil {
		if clients, ok := clientsFromContext.(*clients.Clients); !ok {
			log.Fatalf("Expected *master.Clients, got %T", clientsFromContext)
		} else {
			for clientConn := range clients.Clients {
				clientConn.Write([]byte(redis.ConvertToRESP(args)))
				fmt.Println("Propagated command :", args, conn.RemoteAddr())
			}
		}
	}
}
