package task

import (
	"fmt"
	"net"
	"time"
)

var (
	TcpPort           = 443
	PingTimes         = 4
	Routines          = 200
	TcpConnectTimeOut = time.Second * 1
)

func (p *IPRangeList) Run() *IPRangeList {
	ch := make(chan struct{}, Routines)
	for _, ip := range p.Ips {
		p.wg.Add(1)
		ch <- struct{}{} // 控制最大并发数
		go func(ip *net.IPAddr) {
			defer p.wg.Done()
			defer func() { <-ch }() // 释放并发控制
			success, duration := TCPing(ip)
			if success {
				p.unusedIpCount++
			} else {
				// 删除
			}
			fmt.Printf("IP: %s, Success: %t, Duration: %v\n", ip.String(), success, duration)
		}(ip)
	}
	// 多线程执行
	return p
}

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
	defer conn.Close()
	duration := time.Since(startTime)
	return true, duration
}
