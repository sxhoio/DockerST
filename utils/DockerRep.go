package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	DefaultDockerUrl = "http://docker.sxh.workers.dev"
)

func (s DownloadSpeedSet) DockerSet() {
	// 选择最优的节点
	bestIP := s[0].IP
	bestSpeed := s[0].DownloadSpeed
	// 自动优选节点 最高速度为 0 时，不进行优选
	if bestSpeed == 0 {
		fmt.Println("\n[信息] 未找到最优节点，跳过优选节点。")
		return
	} else {
		fmt.Println("\n[信息] 最优节点：", bestIP, " 速度：", bestSpeed, "MB/s")
	}
	// 输出结果
	fmt.Println("\n[信息] 开始写入 hosts 文件...")
	err := WriteHosts(bestIP.String())
	if err != nil {
		fmt.Println("\n[错误] 写入 hosts 文件失败：", err)
		return
	}

	err = SetDockerAccelerator(DefaultDockerUrl)
	if err != nil {
		fmt.Println("\n[错误] 设置 Docker 加速器失败：", err)
	}
}

func WriteHosts(bestIP string) error {
	// 获取当前系统
	system := runtime.GOOS
	switch system {
	case "windows":
		return writeHostsWindows(bestIP)
	case "darwin":
		return writeHostsMac(bestIP)
	case "linux":
		return writeHostsLinux(bestIP)
	default:
		return fmt.Errorf("不支持的操作系统：%s", system)
	}
}

func writeHostsWindows(bestIP string) error {
	hostsFilePath := "C:\\Windows\\System32\\drivers\\etc\\hosts"
	return writeHostsFile(hostsFilePath, bestIP)
}

func writeHostsMac(bestIP string) error {
	hostsFilePath := "/etc/hosts"
	return writeHostsFile(hostsFilePath, bestIP)
}

func writeHostsLinux(bestIP string) error {
	hostsFilePath := "/etc/hosts"
	return writeHostsFile(hostsFilePath, bestIP)
}

func writeHostsFile(hostsFilePath, bestIP string) error {
	// 检查文件权限
	file, err := os.OpenFile(hostsFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("无法打开 hosts 文件: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// 读取文件内容
	var lines []string
	scanner := bufio.NewScanner(file)
	inBlock := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "# DockerST Start") {
			inBlock = true
		}
		if !inBlock {
			lines = append(lines, line)
		}
		if strings.Contains(line, "# DockerST End") {
			inBlock = false
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 hosts 文件出错: %w", err)
	}

	// 重写文件内容
	file, err = os.Create(hostsFilePath)
	if err != nil {
		return fmt.Errorf("无法重写 hosts 文件: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, _ = fmt.Fprintln(writer, line)
	}
	_, _ = fmt.Fprintln(writer, "# DockerST Start")
	domain := strings.Split(strings.Split(DefaultDockerUrl, "//")[1], "/")[0]
	_, _ = fmt.Fprintln(writer, bestIP+" "+domain)
	_, _ = fmt.Fprintln(writer, "# DockerST End")

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("写入 hosts 文件出错: %w", err)
	}

	fmt.Println("\n[信息] 成功写入 hosts 文件。")
	return nil
}

func SetDockerAccelerator(dockerUrl string) error {
	system := runtime.GOOS
	var err error

	switch system {
	case "windows":
		err = setDockerAcceleratorWindows(dockerUrl)
	case "darwin":
		err = setDockerAcceleratorMac(dockerUrl)
	case "linux":
		err = setDockerAcceleratorLinux(dockerUrl)
	default:
		err = fmt.Errorf("不支持的操作系统：%s", system)
	}

	if err != nil {
		return fmt.Errorf("设置 Docker 加速器失败: %w", err)
	}
	fmt.Println("\n[信息] Docker 加速器已设置为：", dockerUrl)
	return nil
}

func setDockerAcceleratorWindows(dockerUrl string) error {
	configPath := os.Getenv("USERPROFILE") + "\\.docker\\daemon.json"
	return updateDockerConfig(configPath, dockerUrl)
}

func setDockerAcceleratorMac(dockerUrl string) error {
	configPath := os.Getenv("HOME") + "/.docker/daemon.json"
	return updateDockerConfig(configPath, dockerUrl)
}

func setDockerAcceleratorLinux(dockerUrl string) error {
	configPath := "/etc/docker/daemon.json"
	return updateDockerConfig(configPath, dockerUrl)
}

func updateDockerConfig(configPath, dockerUrl string) error {
	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("无法打开 Docker 配置文件: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var config map[string]interface{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		config = make(map[string]interface{})
	}

	config["registry-mirrors"] = []string{dockerUrl}

	file, err = os.Create(configPath)
	if err != nil {
		return fmt.Errorf("无法重写 Docker 配置文件: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("写入 Docker 配置文件出错: %w", err)
	}

	fmt.Println("\n[信息] 成功更新 Docker 配置文件。")
	return nil
}
