package task

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"net"
	"sync"
	"time"
)

var (
	PingTimes         int
	Routines          int
	TcpPort           int
	TcpConnectTimeOut = time.Second * 1
)

type IPDelay struct {
	IP            *net.IPAddr
	Delay         time.Duration
	DownloadSpeed float64
}

func (p *IPRangeList) Run() *IPRangeList {
	var wg sync.WaitGroup
	ch := make(chan struct{}, Routines)
	fmt.Println("正在测试IP", "协议为:", "TCP", "端口为:", TcpPort, "并发数为:", Routines, "Ping次数为:", PingTimes)
	// 创建进度条
	bar := progressbar.NewOptions(len(p.Ips),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("测试IP中"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)
	for _, ip := range p.Ips {
		wg.Add(1)
		ch <- struct{}{} // 控制最大并发数
		go func(ip *net.IPAddr) {
			defer wg.Done()
			defer func() { <-ch }() // 释放并发控制
			success, duration := TCPing(ip)
			if success && duration > 0 {
				p.Delays = append(p.Delays, IPDelay{IP: ip, Delay: duration, DownloadSpeed: 0})
			}
			err := bar.Add(1)
			if err != nil {
				return
			} // 更新进度条
		}(ip)
	}
	wg.Wait()
	err := bar.Finish()
	if err != nil {
		return nil
	}
	fmt.Printf("调用成功，可用的IP数量: %d\n", len(p.Delays))
	return p
}

// TCPing 通过TCP连接测试IP是否可用
func TCPing(ip *net.IPAddr) (bool, time.Duration) {
	startTime := time.Now()
	var fullAddress string
	if IsIpv4(ip.String()) {
		fullAddress = fmt.Sprintf("%s:%d", ip.String(), TcpPort)
	} else {
		fullAddress = fmt.Sprintf("[%s]:%d", ip.String(), TcpPort)
	}
	conn, err := net.DialTimeout("tcp", fullAddress, TcpConnectTimeOut)
	if err != nil {
		return false, 0
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	duration := time.Since(startTime)
	return true, duration
}
