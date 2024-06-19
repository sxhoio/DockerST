package task

import (
	"context"
	"fmt"
	"github.com/VividCortex/ewma"
	"github.com/schollz/progressbar/v3"
	"io"
	"net"
	"net/http"
	"sort"
	"time"
)

const (
	bufferSize                     = 1024
	defaultURL                     = "https://cf.xiu2.xyz/url"
	defaultTimeout                 = 10 * time.Second
	defaultDisableDownload         = false
	defaultTestNum                 = 10
	defaultMinSpeed        float64 = 0.0
)

var (
	URL         = defaultURL
	Timeout     = defaultTimeout
	Disable     = defaultDisableDownload
	DownloadNum = defaultTestNum
	MinSpeed    = defaultMinSpeed
)

type BySpeed []IPDelay

func TestDownloadSpeed(p []IPDelay) []IPDelay {
	if Disable {
		return p
	}
	if len(p) <= 0 { // IP数组长度(IP数量) 大于 0 时才会继续下载测速
		fmt.Println("\n[信息] 延迟测速结果 IP 数量为 0，跳过下载测速。")
		return p
	}
	speedSet := make([]IPDelay, 0)
	testNum := DownloadNum
	if len(p) < DownloadNum || MinSpeed > 0 { // 如果IP数组长度(IP数量) 小于下载测速数量（-dn），则次数修正为IP数
		testNum = len(p)
	}
	if testNum < DownloadNum {
		DownloadNum = testNum
	}
	fmt.Printf("开始下载测速（下限：%.2f MB/s, 数量：%d, 队列：%d）\n", MinSpeed, DownloadNum, testNum)
	// 创建进度条
	bar := progressbar.NewOptions(len(p),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("下载测速"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)
	for i := 0; i < testNum; i++ {
		speed := downloadHandler(p[i].IP)
		p[i].DownloadSpeed = speed
		// 在每个 IP 下载测速后，以 [下载速度下限] 条件过滤结果
		if speed >= MinSpeed*1024*1024 {
			p[i].DownloadSpeed = speed
			err := bar.Add(1)
			if err != nil {
				return nil
			}
		} else {
			p = p[:i]
			break
		}
	}
	err := bar.Finish()
	if err != nil {
		return nil
	}
	if len(speedSet) == 0 { // 没有符合速度限制的数据，返回所有测试数据
		speedSet = p
	}
	// 按速度排序
	sort.Sort(BySpeed(speedSet))
	return speedSet
}

func getDialContext(ip *net.IPAddr) func(ctx context.Context, network, address string) (net.Conn, error) {
	var fakeSourceAddr string
	if IsIpv4(ip.String()) {
		fakeSourceAddr = fmt.Sprintf("%s:%d", ip.String(), TcpPort)
	} else {
		fakeSourceAddr = fmt.Sprintf("[%s]:%d", ip.String(), TcpPort)
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, fakeSourceAddr)
	}
}

// return download Speed
func downloadHandler(ip *net.IPAddr) float64 {
	client := &http.Client{
		Transport: &http.Transport{DialContext: getDialContext(ip)},
		Timeout:   Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 { // 限制最多重定向 10 次
				return http.ErrUseLastResponse
			}
			if req.Header.Get("Referer") == defaultURL { // 当使用默认下载测速地址时，重定向不携带 Referer
				req.Header.Del("Referer")
			}
			return nil
		},
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return 0.0
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")

	response, err := client.Do(req)
	if err != nil {
		return 0.0
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
		}
	}(response.Body)
	if response.StatusCode != 200 {
		return 0.0
	}
	timeStart := time.Now()           // 开始时间（当前）
	timeEnd := timeStart.Add(Timeout) // 加上下载测速时间得到的结束时间

	contentLength := response.ContentLength // 文件大小
	buffer := make([]byte, bufferSize)

	var (
		contentRead     int64 = 0
		timeSlice             = Timeout / 100
		timeCounter           = 1
		lastContentRead int64 = 0
	)

	var nextTime = timeStart.Add(timeSlice * time.Duration(timeCounter))
	e := ewma.NewMovingAverage()

	// 循环计算，如果文件下载完了（两者相等），则退出循环（终止测速）
	for contentLength != contentRead {
		currentTime := time.Now()
		if currentTime.After(nextTime) {
			timeCounter++
			nextTime = timeStart.Add(timeSlice * time.Duration(timeCounter))
			e.Add(float64(contentRead - lastContentRead))
			lastContentRead = contentRead
		}
		// 如果超出下载测速时间，则退出循环（终止测速）
		if currentTime.After(timeEnd) {
			break
		}
		bufferRead, err := response.Body.Read(buffer)
		if err != nil {
			if err != io.EOF { // 如果文件下载过程中遇到报错（如 Timeout），且并不是因为文件下载完了，则退出循环（终止测速）
				break
			} else if contentLength == -1 { // 文件下载完成 且 文件大小未知，则退出循环（终止测速），例如：https://speed.cloudflare.com/__down?bytes=200000000 这样的，如果在 10 秒内就下载完成了，会导致测速结果明显偏低甚至显示为 0.00（下载速度太快时）
				break
			}
			// 获取上个时间片
			lastTimeSlice := timeStart.Add(timeSlice * time.Duration(timeCounter-1))
			// 下载数据量 / (用当前时间 - 上个时间片/ 时间片)
			e.Add(float64(contentRead-lastContentRead) / (float64(currentTime.Sub(lastTimeSlice)) / float64(timeSlice)))
		}
		contentRead += int64(bufferRead)
	}
	return e.Value() / (Timeout.Seconds() / 120)
}
func (a BySpeed) Len() int           { return len(a) }
func (a BySpeed) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySpeed) Less(i, j int) bool { return a[i].DownloadSpeed > a[j].DownloadSpeed } // 这里使用 > 使其降序排序
