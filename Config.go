package main

type Config struct {
	MasterAuth    string
	ReplicaId     string
	ReplicaOffset int64
}

func defaultConfiguration() Config {
	conf := Config{"", "?", -1}
	return conf
}
