package redis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	DELIM         = "\r\n"
	EMPTYRDBSTORE = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
)

func ConvertToRESP(cmd []string) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("*%d\r\n", len(cmd)))

	for _, arg := range cmd {
		buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	return buffer.String()
}

func parseLen(data []byte) (int, error) {
	n := strings.Trim(string(data), "\r\n")
	return strconv.Atoi(n)
}

func UnpackInput(r *bufio.Reader) ([]string, error) {
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
