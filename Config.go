package main

type Config struct {
	MasterAuth    string
	ReplicaId     string
	ReplicaOffset int
}

func defaultConfiguration() Config {
	conf := Config{"", "?", -1}
	return conf
}
