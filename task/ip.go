package task

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
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
	MaxMS     int
	MinMS     int
	TestAll   = false
	randGen   *rand.Rand
)

type IPRangeList struct {
	Ips           []*net.IPAddr
	unusedIpCount int
	Delays        []IPDelay
}

func InitRandSeed() {
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// CreateData 从IP列表中选择一定数量的IP返回
func CreateData() *IPRangeList {
	ips := loadIPRanges(GetIPv4List())
	return &IPRangeList{
		Ips:           ips,
		unusedIpCount: 0,
		Delays:        []IPDelay{},
	}
}

// loadIPRanges 从CIDR列表中加载IP地址
func loadIPRanges(ipList []string) []*net.IPAddr {
	ranges := newIPRanges()
	for _, ip := range ipList {
		line := strings.TrimSpace(ip) // 去除首尾的空白字符（空格、制表符、换行符等）
		if line == "" {               // 跳过空行
			continue
		}
		ranges.parseCIDR(line) // 解析 IP 段，获得 IP、IP 范围、子网掩码
		if IsIpv4(line) {      // 生成要测速的所有 IPv4 / IPv6 地址（单个/随机/全部）
			ranges.chooseIPv4()
		} else {
			ranges.chooseIPv6()
		}
	}
	return ranges.ips
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
	body, err := io.ReadAll(resp.Body)
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

	if err = json.Unmarshal(body, &data); err != nil || !data.Success {
		fmt.Println("获取在线列表失败，正在使用内置列表")
		return Ipv4Cidr
	}

	fmt.Println("获取在线列表成功，正在使用在线列表")
	return data.Result.IPv4CIDRs
}

// ExcludeInvalid 排除不合格节点
func (p *IPRangeList) ExcludeInvalid() []IPDelay {
	// 初始化一个空IPDelay切片
	var delays []IPDelay
	// 遍历IPRangeList的Delays切片
	for ip, delay := range p.Delays {
		// 如果最大延迟大于0且小于最大延迟
		if MaxMS > 0 && delay.Delay > time.Duration(MaxMS)*time.Millisecond {
			continue
		}
		// 如果最小延迟大于0且大于最小延迟
		if MinMS > 0 && delay.Delay < time.Duration(MinMS)*time.Millisecond {
			continue
		}
		// 将IPDelay的IP和Delay添加到delays切片中
		delays = append(delays, IPDelay{IP: p.Ips[ip], Delay: delay.Delay, DownloadSpeed: 0})
	}
	return delays
}

// SortNodesDesc 按延迟降序排列
func SortNodesDesc(p []IPDelay) []IPDelay {
	sorted := make([]IPDelay, len(p))
	_ = copy(sorted, p)
	// 使用sort.Slice 对切片进行排序
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Delay < sorted[j].Delay
	})
	return sorted
}

// IsIpv4 检查IP地址是否为IPv4
func IsIpv4(ip string) bool {
	return strings.Contains(ip, ".")
}
