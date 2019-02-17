package replica

import (
	"bufio"
	"fmt"
	"net"
	"redis-g/io"
	"redis-g/util"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Replica struct {
	Address       string
	MasterAuth    string
	replicaId     string
	replicaOffset int64
}

var status string
var connection net.Conn
var reader *io.MyReader
var writer *bufio.Writer

func (replica *Replica) Open() {
	connect(replica.Address)
	if replica.MasterAuth != "" {
		auth(replica.MasterAuth)
	}
	sendPing()
	addr := strings.Split(connection.LocalAddr().String(), ":")
	sendSlavePort(addr[1])
	sendSlaveIp(addr[0])
	sendSlaveCapa("eof")
	sendSlaveCapa("psync2")
	replica.sync()
}

func (replica *Replica) Close() {
	if connection != nil {
		connection.Close()
	}
}

func (replica *Replica) sync() {
	fmt.Printf("PSYNC %s %d\n", replica.getReplId(), replica.getReplOffset())
	send("PSYNC", replica.getReplId(), strconv.Itoa(replica.getReplOffset()))
	reply := replyStr()
	fmt.Println(reply)
	mode := replica.trySync(reply)
	switch mode {
	case "PSYNC":
		replica.startHeartbeat()
	case "SYNC_LATER":
		return
	}
	for status == "CONNECTED" {
		reply := parse(func(length int) {
			// TODO offset与master有差距
			replica.addReplOffset(length)
		}).([]byte)
		// TODO 处理master返回的数据
		fmt.Println(string(reply))
	}
}

func (replica *Replica) trySync(reply string) (mode string) {
	if strings.HasPrefix(reply, "FULLRESYNC") {
		parseDump()
		resp := strings.Split(reply, " ")
		replica.setReplId(resp[1])
		offset, _ := strconv.Atoi(resp[2])
		replica.addReplOffset(offset)
		mode = "PSYNC"
	} else if strings.HasPrefix(reply, "CONTINUE") {
		mode = "PSYNC"
	} else if strings.HasPrefix(reply, "NOMASTERLINK") || strings.HasPrefix(reply, "LOADING") {
		mode = "SYNC_LATER"
	}
	return mode
}

func (replica *Replica) startHeartbeat() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			offset := strconv.Itoa(replica.getReplOffset())
			send("REPLCONF", "ACK", offset)
			fmt.Println("heartbeat. curr offset: ", offset)
		}
	}()
	fmt.Println("heartbeat started")
}

func (replica *Replica) addReplOffset(length int) {
	atomic.AddInt64(&replica.replicaOffset, int64(length))
}

func (replica *Replica) setReplOffset(offset int) {
	atomic.StoreInt64(&replica.replicaOffset, int64(offset))
}

func (replica *Replica) getReplOffset() int {
	offset := atomic.LoadInt64(&replica.replicaOffset)
	if offset == 0 {
		return -1
	}
	return int(offset)
}

func (replica *Replica) setReplId(replId string) {
	replica.replicaId = replId
}

func (replica *Replica) getReplId() (replId string) {
	if replica.replicaId == "" {
		replId = "?"
	}
	return replId
}

func sendSlaveCapa(command string) {
	fmt.Printf("REPLCONF capa %s\n", command)
	send("REPLCONF", "capa", command)
	reply := replyStr()
	fmt.Println(reply)
	if "OK" == reply {
		return
	}
}

func sendSlaveIp(ip string) {
	fmt.Printf("REPLCONF ip-address %s\n", ip)
	send("REPLCONF", "ip-address", ip)
	reply := replyStr()
	fmt.Println(reply)
	if "OK" == reply {
		return
	}
}

func sendSlavePort(port string) {
	fmt.Printf("REPLCONF listening-port %s\n", port)
	send("REPLCONF", "listening-port", port)
	reply := replyStr()
	fmt.Println(reply)
	if "OK" == reply {
		return
	}
}

func sendPing() {
	send("PING")
	reply := replyStr()
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
	writer.Write(util.ToBytes(argsLen + 1))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteByte(Dollar)
	writer.Write(util.ToBytes(commLen))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteString(command)
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	for i := 0; i < argsLen; i++ {
		argLen := len(args[i])
		writer.WriteByte(Dollar)
		writer.Write(util.ToBytes(argLen))
		writer.WriteByte(Cr)
		writer.WriteByte(Lf)
		writer.WriteString(args[i])
		writer.WriteByte(Cr)
		writer.WriteByte(Lf)
	}
	writer.Flush()
}

func connect(address string) {
	conn, e := net.Dial("tcp4", address)
	if e != nil {
		panic(e)
	}
	connection = conn
	writer = bufio.NewWriter(conn)
	reader = &io.MyReader{Input: bufio.NewReader(conn)}
	status = "CONNECTED"
}
