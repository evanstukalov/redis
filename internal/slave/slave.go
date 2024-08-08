package slave

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/commands"
	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/redis"
)

type MasterInfo struct {
	Host string
	Port string
}

func (m MasterInfo) Address() string {
	return fmt.Sprintf("%s:%s", m.Host, m.Port)
}

func masterInfoFromParam(replicaOf string) MasterInfo {
	data := strings.Split(replicaOf, " ")
	return MasterInfo{
		Host: data[0],
		Port: data[1],
	}
}

func sendMessage(conn net.Conn, message string) error {
	if _, err := conn.Write([]byte(message)); err != nil {
		return err
	}
	readAnswer(conn)
	return nil
}

func readAnswer(
	conn net.Conn,
) {
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}
	fmt.Println(message)
}

func ConnectMaster(replicaof string, config config.Config) (net.Conn, error) {
	masterInfo := masterInfoFromParam(replicaof)
	addr := masterInfo.Address()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err)
		return nil, err
	}
	return conn, nil
}

func Handshakes(conn net.Conn, config config.Config) error {
	if err := sendMessage(conn, "*1\r\n$4\r\nPING\r\n"); err != nil {
		return err
	}
	if err := sendMessage(
		conn,
		fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%d\r\n", config.Port),
	); err != nil {
		return err
	}
	if err := sendMessage(conn, "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"); err != nil {
		return err
	}
	if err := sendMessage(conn, "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"); err != nil {
		return err
	}
	// fmt.Println("Waiting for response after PSYNC")
	// readAnswer(conn)
	fmt.Println("Handshakes with master is over")

	return nil
}

func ReadFromConnection(ctx context.Context, conn net.Conn, config config.Config) {
	fmt.Println("Starting consuming commands from master")
	defer conn.Close()

	for {
		r := bufio.NewReader(conn)

		args, err := redis.UnpackInput(r)
		if err != nil {
			break
		}

		fmt.Println("New command from master :", args)

		if len(args) == 0 {
			break
		}

		go HandleCommand(ctx, conn, config, args)
	}
}

func HandleCommand(ctx context.Context, conn net.Conn, config config.Config, args []string) {
	cmd, exists := commands.Commands[strings.ToUpper(args[0])]
	if !exists {
		conn.Write([]byte("-Error\r\n"))
		return
	}

	cmd.Execute(ctx, conn, config, args)
}
