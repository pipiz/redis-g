package main

import (
	"redis-g/replica"
)

func main() {
	newReplica := replica.Replica{Address: "localhost:6379"}
	newReplica.Open()
	newReplica.Close()
}
