package main

import (
	"DockerST/task"
	"flag"
	"fmt"
)

var (
	VersionPrint bool
	Version      string
)

func init() {
	// checkUpdate()
	// 定义一个字符串类型的命令行标志
	flag.IntVar(&task.TcpPort, "p", 443, "TCP端口")
	flag.IntVar(&task.PingTimes, "t", 4, "Ping次数")
	flag.IntVar(&task.Routines, "r", 200, "并发数")
	flag.BoolVar(&VersionPrint, "v", false, "输出版本")
	flag.BoolVar(&task.IsOff, "om", false, "是否为离线模式")
	flag.Parse()
	if VersionPrint {
		fmt.Println("Version:", Version)
	}
}

func main() {
	if VersionPrint {
		return
	}
	// 输出版本
	fmt.Printf("# DockerST %s \n", Version)
	_ = task.CreateData().Run()
}

func WriteHost(domain string, ip string) {

}

func checkUpdate() {
}
