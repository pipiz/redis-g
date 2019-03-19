package main

import (
	. "redis-g/replica"
)

func main() {
	replica := Replica{Master: "localhost:6379"}
	replica.Open()
	replica.Close()
}
