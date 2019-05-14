package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"redis-g/replica"
)

var (
	app        = kingpin.New("redis-g", "本程序实现了Redis Replication协议, 在运行时本程序将以'slave'的身份连接至源Redis并请求源Redis的数据, 随后来自源Redis的数据经由本程序存入目标Redis内.")
	source     = app.Flag("source", "源Redis的地址, 数据的来源.").PlaceHolder("host:port").Required().Short('s').String()
	sourceAuth = app.Flag("source-auth", "源Redis的密码").PlaceHolder("password").String()
	target     = app.Flag("target", "目标Redis的地址, 数据最终将存入此Redis").PlaceHolder("host:port").Required().Short('t').String()
	targetAuth = app.Flag("target-auth", "目标Redis的密码").PlaceHolder("password").String()
)

func main() {
	if _, err := app.Parse(os.Args[1:]); err != nil {
		app.FatalUsage("%s \r\n", err.Error())
	}

	instance := replica.New(*source, *sourceAuth, *target, *targetAuth)
	instance.Open()
	instance.Close()
}
