package server

import (
	"fmt"
	"net"
	"strings"
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

func Handshake(replicaof string) error {
	masterInfo := masterInfoFromParam(replicaof)
	addr := masterInfo.Address()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err)
		return err
	}
	defer conn.Close()
	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	return nil
}
