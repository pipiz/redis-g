package replica

import (
	"bufio"
	"fmt"
	"net"
	. "redis-g/command"
	"redis-g/io"
	. "redis-g/utils/numbers"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Replica struct {
	Master        string
	MasterAuth    string
	replicaId     string
	replicaOffset int64
}

var status string
var connection net.Conn
var reader *io.MyReader
var writer *bufio.Writer

func (replica *Replica) Open() {
	connect(replica.Master)
	if replica.MasterAuth != "" {
		auth(replica.MasterAuth)
	}
	replica.sync()
}

func (replica *Replica) Close() {
	if connection != nil {
		connection.Close()
	}
}

func (replica *Replica) sync() {
	go handleCommand()
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
			replica.addReplOffset(length)
		})
		arr, ok := reply.([][]byte)
		if ok {
			commandName := string(arr[0])
			if commandName == "REPLCONF" && string(arr[1]) == "GETACK" {
				replica.startHeartbeat()
			} else if commandName != "PING" {
				commChan <- Command{Name: commandName, Args: arr[1:]}
			}
		}
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

func auth(password string) {
	send("AUTH", password)
}

func send(command string, args ...string) {
	commLen := len(command)
	argsLen := len(args)
	writer.WriteByte(Star)
	writer.Write(ToBytes(argsLen + 1))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteByte(Dollar)
	writer.Write(ToBytes(commLen))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteString(command)
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	for i := 0; i < argsLen; i++ {
		argLen := len(args[i])
		writer.WriteByte(Dollar)
		writer.Write(ToBytes(argLen))
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
