package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const CR = '\r'
const LF = '\n'
const STAR = '*'
const DOLLAR = '$'
const PLUS = '+'
const MINUS = '-'
const COLON = ':'

type Replica struct {
	Address string
	Config  Config
}

var connection net.Conn
var reader *bufio.Reader
var writer *bufio.Writer

func (replica *Replica) Open() {
	connect(replica.Address)
	if replica.Config.MasterAuth != "" {
		auth(replica.Config.MasterAuth)
	}
	ping()
	localAddr := connection.LocalAddr()
	addr := strings.Split(localAddr.String(), ":")
	sendSlavePort(addr[1])
	sendSlaveIp(addr[0])
	sendSlaveCapa("eof")
	sendSlaveCapa("psync2")
	sync(replica)
}

func (replica *Replica) Close() {
	if connection != nil {
		connection.Close()
	}
}

func sync(replica *Replica) {
	fmt.Printf("PSYNC %s %d\n", replica.Config.ReplicaId, replica.Config.ReplicaOffset)
	send("PSYNC", replica.Config.ReplicaId, string(replica.Config.ReplicaOffset))
	reply := receive().(string)
	fmt.Println(reply)
	if strings.HasPrefix(reply, "FULLRESYNC") {
		parseDump()
		resp := strings.Split(reply, " ")
		replica.Config.ReplicaId = resp[1]
		offset, _ := strconv.Atoi(resp[2])
		replica.Config.ReplicaOffset = int64(offset)
	} else if strings.HasPrefix(reply, "CONTINUE") {

	} else if strings.HasPrefix(reply, "NOMASTERLINK") || strings.HasPrefix(reply, "LOADING") {

	}
}

func parseDump() {
	// TODO
}

func sendSlaveCapa(command string) {
	fmt.Printf("REPLCONF capa %s\n", command)
	send("REPLCONF", "capa", command)
	reply := receive().(string)
	fmt.Println(reply)
	if "OK" == reply {
		return
	}
}

func sendSlaveIp(ip string) {
	fmt.Printf("REPLCONF ip-address %s\n", ip)
	send("REPLCONF", "ip-address", ip)
	reply := receive()
	fmt.Println(reply)
	if "OK" == reply {
		return
	}
}

func sendSlavePort(port string) {
	fmt.Printf("REPLCONF listening-port %s\n", port)
	send("REPLCONF", "listening-port", port)
	reply := receive()
	fmt.Println(reply)
	if "OK" == reply {
		return
	}
}

func ping() {
	send("PING")
	reply := receive().(string)
	if "PONG" == reply {
		return
	}
}

func auth(password string) {
	send("AUTH", password)
}

func send(command string, args ...string) {
	commLen := len(command)
	argsLen := len(args)
	writer.WriteByte(STAR)
	writer.Write(getBytes(argsLen + 1))
	writer.WriteByte(CR)
	writer.WriteByte(LF)
	writer.WriteByte(DOLLAR)
	writer.Write(getBytes(commLen))
	writer.WriteByte(CR)
	writer.WriteByte(LF)
	writer.WriteString(command)
	writer.WriteByte(CR)
	writer.WriteByte(LF)
	for i := 0; i < argsLen; i++ {
		argLen := len(args[i])
		writer.WriteByte(DOLLAR)
		writer.Write(getBytes(argLen))
		writer.WriteByte(CR)
		writer.WriteByte(LF)
		writer.WriteString(args[i])
		writer.WriteByte(CR)
		writer.WriteByte(LF)
	}
	writer.Flush()
}

func getBytes(i int) []byte {
	return []byte(strconv.Itoa(i))
}

func receive() interface{} {
	for {
		b, _ := reader.ReadByte()
		switch b {
		case PLUS: // RESP Simple Strings
			var builder strings.Builder
			for byt, _ := reader.ReadByte(); byt != CR; {
				builder.WriteByte(byt)
				byt, _ = reader.ReadByte()
			}
			if byt, _ := reader.ReadByte(); byt == LF {
				return builder.String()
			}
		case MINUS: // RESP Errors
			var builder strings.Builder
			for byt, _ := reader.ReadByte(); byt != CR; {
				builder.WriteByte(byt)
				byt, _ = reader.ReadByte()
			}
			if byt, _ := reader.ReadByte(); byt == LF {
				return builder.String()
			}
		}
	}
	return nil
}

func connect(address string) {
	conn, e := net.Dial("tcp", address)
	if e != nil {
		panic(e)
	}
	connection = conn
	writer = bufio.NewWriter(conn)
	reader = bufio.NewReader(conn)
}
