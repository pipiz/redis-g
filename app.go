package main

import (
	"flag"
	"redis-g/replica"
)

var (
	masterAddress string
	masterAuth    string
	targetAddress string
	targetAuth    string
)

func init() {
	const (
		defaultSourceAddress  = "localhost:6379"
		sourceDescription     = "数据源Redis的地址\n程序将从这个节点接收数据并发送至目标Redis"
		sourceAuthDescription = "数据源Redis的密码"
		targetDescription     = "目标Redis的地址\n数据将数据发送至此Redis"
		targetAuthDescription = "目标Redis的密码"
	)

	flag.StringVar(&masterAddress, "source", defaultSourceAddress, sourceDescription)
	flag.StringVar(&masterAddress, "s", defaultSourceAddress, sourceDescription)
	flag.StringVar(&masterAuth, "source-auth", "", sourceAuthDescription)
	flag.StringVar(&masterAuth, "sa", "", sourceAuthDescription)
	flag.StringVar(&targetAddress, "target", "", targetDescription)
	flag.StringVar(&targetAddress, "t", "", targetDescription)
	flag.StringVar(&targetAuth, "auth", "", targetAuthDescription)
	flag.StringVar(&targetAuth, "a", "", targetAuthDescription)
}

func main() {
	flag.Parse()

	if "" == targetAddress {
		flag.PrintDefaults()
		return
	}

	instance := replica.New(masterAddress, masterAuth, targetAddress, targetAuth)
	instance.Open()
	instance.Close()
}
