package replica

import (
	"bufio"
	"github.com/gomodule/redigo/redis"
	"log"
	"net"
	"os"
	"redis-g/command"
	"redis-g/io"
	"redis-g/utils/numbers"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type replica struct {
	Source        string
	SourceAuth    string
	Target        string
	TargetAuth    string
	replicaId     string
	replicaOffset int64
}

func New(source string, sourceAuth string, target string, targetAuth string) *replica {
	if source == "" || target == "" {
		return nil
	}
	return &replica{Source: source, SourceAuth: sourceAuth, Target: target, TargetAuth: targetAuth}
}

var (
	status   string
	source   net.Conn
	reader   *io.Reader
	writer   *bufio.Writer
	target   redis.Conn
	commChan = make(chan command.Command, 65535)
	logger   = log.New(os.Stdout, "", log.LstdFlags)
)

func (replica *replica) Open() {
	replica.connectSource()
	replica.sendMetadata()
	replica.connectTarget()
	replica.sync()
}

func (replica *replica) Close() {
	if source != nil {
		e := source.Close()
		logger.Println(e)
	}
	if target != nil {
		e := target.Close()
		logger.Println(e)
	}
	close(commChan)
}

func (replica *replica) sync() {
	go handleCommand()
	logger.Printf("PSYNC %s %d\n", replica.getReplId(), replica.getReplOffset())
	send("PSYNC", replica.getReplId(), strconv.Itoa(replica.getReplOffset()))
	reply := replyStr()
	logger.Println(reply)
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
				commChan <- command.New(commandName, arr[1:])
			}
		}
	}
}

func (replica *replica) trySync(reply string) (mode string) {
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

func (replica *replica) startHeartbeat() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			offset := strconv.Itoa(replica.getReplOffset())
			send("REPLCONF", "ACK", offset)
		}
	}()
	logger.Println("heartbeat started")
}

func (replica *replica) addReplOffset(length int) {
	atomic.AddInt64(&replica.replicaOffset, int64(length))
}

func (replica *replica) setReplOffset(offset int) {
	atomic.StoreInt64(&replica.replicaOffset, int64(offset))
}

func (replica *replica) getReplOffset() int {
	offset := atomic.LoadInt64(&replica.replicaOffset)
	if offset == 0 {
		return -1
	}
	return int(offset)
}

func (replica *replica) setReplId(replId string) {
	replica.replicaId = replId
}

func (replica *replica) getReplId() (replId string) {
	if replica.replicaId == "" {
		replId = "?"
	}
	return replId
}

func send(command string, args ...string) {
	commLen := len(command)
	argsLen := len(args)
	writer.WriteByte(Star)
	writer.Write(numbers.ToBytes(argsLen + 1))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteByte(Dollar)
	writer.Write(numbers.ToBytes(commLen))
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	writer.WriteString(command)
	writer.WriteByte(Cr)
	writer.WriteByte(Lf)
	for i := 0; i < argsLen; i++ {
		argLen := len(args[i])
		writer.WriteByte(Dollar)
		writer.Write(numbers.ToBytes(argLen))
		writer.WriteByte(Cr)
		writer.WriteByte(Lf)
		writer.WriteString(args[i])
		writer.WriteByte(Cr)
		writer.WriteByte(Lf)
	}
	writer.Flush()
}

func (replica *replica) connectSource() {
	logger.Println("Connecting", replica.Source)
	conn, err := net.Dial("tcp4", replica.Source)
	if err != nil {
		logger.Fatalln("Connection failed:", err.Error())
	}
	writer = bufio.NewWriter(conn)
	reader = &io.Reader{Input: bufio.NewReader(conn)}

	// 检查是否需要认证
	send("PING")
	reply := replyStr()
	if strings.HasPrefix(reply, "NOAUTH") {
		if replica.SourceAuth != "" {
			send("AUTH", replica.SourceAuth)
			reply := replyStr()
			if "OK" != reply {
				logger.Fatalln(reply)
			}
		} else {
			panic("请通过参数提供主Redis的密码信息")
		}
	}

	logger.Println("Connected")
	source = conn
	status = "CONNECTED"
}

func (replica *replica) connectTarget() {
	dialOptions := make([]redis.DialOption, 0)
	if replica.TargetAuth != "" {
		password := redis.DialPassword(replica.TargetAuth)
		dialOptions = append(dialOptions, password)
	}
	var err error
	target, err = redis.Dial("tcp4", replica.Target, dialOptions...)
	if err != nil {
		panic(err)
	}
}

func (replica *replica) sendMetadata() {
	addr := strings.Split(source.LocalAddr().String(), ":")
	logger.Println("REPLCONF listening-port:", addr[1])
	send("REPLCONF", "listening-port", addr[1])
	reply := replyStr()
	if "OK" != reply {
		logger.Panic(reply)
	}
}

func handleCommand() {
	for {
		theCommand := <-commChan
		target.Do(theCommand.Name(), theCommand.Args()...)
	}
}
