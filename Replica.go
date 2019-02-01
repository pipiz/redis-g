package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

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
	send("PSYNC", replica.Config.ReplicaId, strconv.Itoa(replica.Config.ReplicaOffset))
	reply := receive().(string)
	fmt.Println(reply)
	if strings.HasPrefix(reply, "FULLRESYNC") {
		receive()
		resp := strings.Split(reply, " ")
		replica.Config.ReplicaId = resp[1]
		offset, _ := strconv.Atoi(resp[2])
		replica.Config.ReplicaOffset = offset
	} else if strings.HasPrefix(reply, "CONTINUE") {

	} else if strings.HasPrefix(reply, "NOMASTERLINK") || strings.HasPrefix(reply, "LOADING") {

	}
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
	writer.WriteByte(Star)
	writer.Write(getBytes(argsLen + 1))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteByte(Dollar)
	writer.Write(getBytes(commLen))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteString(command)
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	for i := 0; i < argsLen; i++ {
		argLen := len(args[i])
		writer.WriteByte(Dollar)
		writer.Write(getBytes(argLen))
		writer.WriteByte(Cr)
		writer.WriteByte(Lf)
		writer.WriteString(args[i])
		writer.WriteByte(Cr)
		writer.WriteByte(Lf)
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
		case Plus: // RESP Simple Strings
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, _ = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					return builder.String()
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
		case Minus: // RESP Errors
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, e = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					return builder.String()
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
		case Dollar: // RESP Bulk Strings
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, e = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					break
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
			resp := builder.String()
			var size int
			if !strings.HasPrefix(resp, "EOF:") {
				_size, err := strconv.Atoi(resp)
				size = _size
				if err != nil {
					return -1
				}
			}
			return parseRdb(size)
		case '\n':
		default:
			break
		}
	}
	return ""
}

func connect(address string) {
	conn, e := net.Dial("tcp4", address)
	if e != nil {
		panic(e)
	}
	connection = conn
	writer = bufio.NewWriter(conn)
	reader = bufio.NewReader(conn)
}
