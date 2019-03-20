package main

import (
	"redis-g/replica"
)

func main() {
	instance := replica.New("localhost:6379", "123456", "localhost:6400", "")
	instance.Open()
	instance.Close()
}
