package main

func main() {
	replica := Replica{}
	replica.Address = "localhost:6379"
	replica.Config = defaultConfiguration()
	replica.Open()
	replica.Close()
}
