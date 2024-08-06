package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/config"
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

func sendMessage(conn net.Conn, message string, ch chan<- struct{}) error {
	ch <- struct{}{}
	if _, err := conn.Write([]byte(message)); err != nil {
		return err
	}
	return nil
}

func readAnswer(
	conn net.Conn,
	ch <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for range ch {

		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}
		fmt.Println(message)
	}
}

func Handshakes(replicaof string, config config.Config) error {
	masterInfo := masterInfoFromParam(replicaof)
	addr := masterInfo.Address()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err)
		return err
	}
	defer conn.Close()
	ch := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go readAnswer(conn, ch, &wg)

	if err := sendMessage(conn, "*1\r\n$4\r\nPING\r\n", ch); err != nil {
		return err
	}
	if err := sendMessage(
		conn,
		fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%d\r\n", config.Port), ch,
	); err != nil {
		return err
	}
	if err := sendMessage(conn, "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n", ch); err != nil {
		return err
	}
	if err := sendMessage(conn, "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n", ch); err != nil {
		return err
	}

	close(ch)
	wg.Wait()

	return nil
}
