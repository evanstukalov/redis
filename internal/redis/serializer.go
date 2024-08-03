package redis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const DELIM = "\r\n"

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
