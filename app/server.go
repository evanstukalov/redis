package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const DELIM = "\r\n"

func main() {
	fmt.Println("Logs from your program will appear here!")

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

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		r := bufio.NewReader(conn)

		args, err := unpackInput(r)
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
		default:
			conn.Write([]byte("-Error\r\n"))
		}

	}
}

func parseLen(data []byte) (int, error) {
	n := strings.Trim(string(data), "\r\n")
	return strconv.Atoi(n)
}

func unpackInput(r *bufio.Reader) ([]string, error) {
	args := make([]string, 0, 64)

	firstLine, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var commandLen int

	if firstLine[0] == '*' {
		commandLen, err = parseLen(firstLine[1:])
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < commandLen; i++ {

		line, err := r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		var argLen int

		if line[0] == '$' {
			argLen, err = parseLen(line[1:])
			if err != nil {
				fmt.Println("Error: ", err)
				return nil, err
			}
		}

		if argLen == 0 {
			continue
		}

		buf := make([]byte, argLen+len(DELIM))

		if _, err = io.ReadFull(r, buf); err != nil {
			fmt.Println("Error: ", err)
			return nil, err
		}

		arg := strings.Trim(string(buf), DELIM)
		args = append(args, arg)
	}

	return args, nil
}
