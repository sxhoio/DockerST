package main

import (
	"DockerST/task"
	"flag"
	"fmt"
)

var (
	VersionPrint bool
	Version      string
	DDownload    bool
)

func init() {
	// checkUpdate()
	// 定义一个字符串类型的命令行标志
	flag.IntVar(&task.TcpPort, "tcp", 443, "TCP端口")
	flag.IntVar(&task.PingTimes, "t", 4, "Ping次数")
	flag.IntVar(&task.Routines, "r", 200, "存活检测并发数检测")
	flag.IntVar(&task.MinMS, "mis", 0, "只输出高于指定平均延迟的 IP")
	flag.IntVar(&task.MaxMS, "mxs", 1000, "只输出低于指定平均延迟的 IP")
	flag.IntVar(&task.DownloadNum, "dn", 10, "下载数量")
	flag.StringVar(&task.DownloadUrl, "url", "https://github.com", "默认文件下载地址")
	flag.BoolVar(&task.TestAll, "ta", false, "测试所有 IP")
	flag.BoolVar(&DDownload, "du", false, "禁止下载")
	flag.BoolVar(&VersionPrint, "v", false, "输出版本")
	flag.BoolVar(&task.IsOff, "om", false, "不下载子网列表")
	flag.Parse()
	if VersionPrint {
		fmt.Println("Version:", Version)
	}
}

func main() {
	if VersionPrint {
		return
	}
	// InitRandSeed 初始化随机数种子
	task.InitRandSeed()
	// 输出版本
	fmt.Printf("# DockerST %s \n", Version)
	pingData := task.CreateData().Run().ExcludeInvalid()
	// 按照延迟排序
	sortedPingData := task.SortNodesDesc(pingData)
	// 仅输出前10个结果
	for i, ipDelay := range sortedPingData {
		if i >= 10 {
			break
		}
		fmt.Printf("IP: %s, 延迟: %s\n", ipDelay.IP.String(), ipDelay.Delay)
	}

	if DDownload {

	}
}
