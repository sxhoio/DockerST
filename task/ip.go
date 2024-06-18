package task

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	Ipv4Cidr = []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	}
	IPCidrApi = "https://api.cloudflare.com/client/v4/ips"
	IsOff     = false
)

type IPRangeList struct {
	Ips           []*net.IPAddr
	unusedIpCount int
	wg            sync.WaitGroup
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CreateData 从IP列表中选择一定数量的IP返回
func CreateData() *IPRangeList {
	ips := loadIPRanges(GetIPv4List())
	return &IPRangeList{
		Ips:           ips,
		unusedIpCount: 0,
	}
}

// loadIPRanges 从CIDR列表中加载IP地址
func loadIPRanges(ciders []string) []*net.IPAddr {
	var ipAddresses []*net.IPAddr
	for _, cidr := range ciders {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			fmt.Printf("解析CIDR %s 时出错: %v\n", cidr, err)
			continue
		}
		// 计算给定IPNet的范围
		ones, _ := ipnet.Mask.Size()
		if ones > 24 {
			fmt.Printf("CIDR %s 小于 /24\n", cidr)
			continue
		}
		numSubnets := 1 << (24 - ones)
		for i := 0; i < numSubnets; i++ {
			ip := generateRandomIP(ipnet, i)
			ipAddresses = append(ipAddresses, ip)
		}
	}
	return ipAddresses
}

// generateRandomIP 在给定IPNet范围内生成随机IP地址
func generateRandomIP(ipnet *net.IPNet, subnetIndex int) *net.IPAddr {
	ip := ipnet.IP.To4()
	if ip == nil {
		return nil
	}
	// 设置第三个字节为子网索引
	ip[2] = ip[2] + byte(subnetIndex)
	// 为最后一个字节生成随机值
	ip[3] = byte(rand.Intn(256))
	return &net.IPAddr{IP: ip}
}

// GetIPv4List 获取IPv4 CIDR列表
func GetIPv4List() []string {
	// 离线变量
	if IsOff {
		return Ipv4Cidr
	}
	// 获取在线IPv4 CIDR列表
	resp, err := http.Get(IPCidrApi)
	if err != nil {
		fmt.Println("获取在线列表失败，正在使用内置列表")
		return Ipv4Cidr
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 读取响应主体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("获取在线列表失败，正在使用内置列表")
		return Ipv4Cidr
	}

	// 解析JSON数据
	var data struct {
		Result struct {
			IPv4CIDRs []string `json:"ipv4_cidrs"`
		} `json:"result"`
		Success bool `json:"success"`
	}

	if err := json.Unmarshal(body, &data); err != nil || !data.Success {
		fmt.Println("获取在线列表失败，正在使用内置列表")
		return Ipv4Cidr
	}

	fmt.Println("获取在线列表成功，正在使用在线列表")
	return data.Result.IPv4CIDRs
}

// IsIpv4 检查IP地址是否为IPv4
func IsIpv4(ip string) bool {
	return net.ParseIP(ip) != nil
}
