package commands

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Command interface {
	Execute(ctx context.Context, conn net.Conn, config config.Config, args []string)
}

type EchoCommand struct{}

func (c *EchoCommand) Execute(
	ctx context.Context,
	conn net.Conn,
	config config.Config,
	args []string,
) {
	msg := args[1]
	conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)))
}

type PingCommand struct{}

func (c *PingCommand) Execute(
	ctx context.Context,
	conn net.Conn,
	config config.Config,
	args []string,
) {
	conn.Write([]byte("+PONG\r\n"))
}

type SetCommand struct{}

func (c *SetCommand) Execute(
	ctx context.Context,
	conn net.Conn,
	config config.Config,
	args []string,
) {
	key, value := args[1], args[2]

	var px *int

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
}

type GetCommand struct{}

func (c *GetCommand) Execute(
	ctx context.Context,
	conn net.Conn,
	config config.Config,
	args []string,
) {
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
}

type InfoCommand struct{}

func (c *InfoCommand) Execute(
	ctx context.Context,
	conn net.Conn,
	config config.Config,
	args []string,
) {
	switch args[1] {
	case "replication":
		var builder strings.Builder
		builder.Grow(128)

		role := fmt.Sprintf("role:%s", config.Role)
		builder.WriteString(fmt.Sprintf("%s\n", role))

		master_replid := fmt.Sprintf("master_replid:%s", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb")
		builder.WriteString(fmt.Sprintf("%s\n", master_replid))

		master_repl_offset := fmt.Sprintf("master_repl_offset:%d", 0)
		builder.WriteString(
			fmt.Sprintf("%s\n", master_repl_offset),
		)

		result := builder.String()

		finalResult := fmt.Sprintf("$%d\r\n%s\r\n", len(result), result)
		fmt.Println(finalResult)

		conn.Write([]byte(finalResult))

	default:
		conn.Write([]byte("-Error\r\n"))
	}
}

var commands = map[string]Command{
	"PING": &PingCommand{},
	"ECHO": &EchoCommand{},
	"SET":  &SetCommand{},
	"GET":  &GetCommand{},
	"INFO": &InfoCommand{},
}

func HandleCommand(ctx context.Context, conn net.Conn, config config.Config, args []string) {
	cmd, exists := commands[strings.ToUpper(args[0])]
	if !exists {
		conn.Write([]byte("-Error\r\n"))
		return
	}

	cmd.Execute(ctx, conn, config, args)
}
