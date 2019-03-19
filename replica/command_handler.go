package replica

import (
	"github.com/gomodule/redigo/redis"
	. "redis-g/command"
)

var commChan = make(chan Command, 65535)

var conn, e = redis.Dial("tcp", "localhost:6400")

func handleCommand() {
	for {
		command := <-commChan
		conn.Do(command.Name, command.GetArgs()...)
	}
}
