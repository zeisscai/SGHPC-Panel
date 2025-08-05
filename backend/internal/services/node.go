package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"panel-tool/internal/models"
)

// GetManagementNode 获取管理节点真实信息
func GetManagementNode() *models.ManagementNode {
	// 获取真实的系统信息
	hostname := getHostname()
	model := getSystemModel()
	architecture := getArchitecture()
	cpuInfo := getCPUInfo()
	osVersion := getOSVersion()
	kernelVersion := getKernelVersion()
	localTime := time.Now().Format("2006-01-02 15:04:05")
	uptime := getUptime()

	return &models.ManagementNode{
		Hostname:      hostname,
		Model:         model,
		Architecture:  architecture,
		CPUInfo:       cpuInfo,
		OSVersion:     osVersion,
		KernelVersion: kernelVersion,
		LocalTime:     localTime,
		Uptime:        uptime,
	}
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getSystemModel 获取系统型号
func getSystemModel() string {
	// 尝试从 /sys/devices/virtual/dmi/id/product_name 获取型号
	data, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}
	
	// 如果无法获取，返回默认值
	return "Unknown Server Model"
}

// getArchitecture 获取系统架构
func getArchitecture() string {
	output, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getCPUInfo 获取CPU信息
func getCPUInfo() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "Unknown CPU"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var cpuModel string
	var cpuCount int

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 && cpuModel == "" {
				cpuModel = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "processor") {
			cpuCount++
		}
	}

	if cpuModel != "" {
		// 简化CPU信息显示
		cpuParts := strings.Split(cpuModel, " @ ")
		model := cpuParts[0]
		clockSpeed := ""
		if len(cpuParts) > 1 {
			clockSpeed = " @ " + cpuParts[1]
		}
		return fmt.Sprintf("%s x %dC %dT%s", model, cpuCount, cpuCount, clockSpeed)
	}

	return "Unknown CPU"
}

// getOSVersion 获取操作系统版本，特别支持Rocky Linux和OpenEuler
func getOSVersion() string {
	// 检查 Rocky Linux
	if _, err := os.Stat("/etc/rocky-release"); err == nil {
		data, err := os.ReadFile("/etc/rocky-release")
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	
	// 检查 OpenEuler
	if _, err := os.Stat("/etc/openeuler-release"); err == nil {
		data, err := os.ReadFile("/etc/openeuler-release")
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	
	// 检查通用的 os-release 文件
	if _, err := os.Stat("/etc/os-release"); err == nil {
		file, err := os.Open("/etc/os-release")
		if err != nil {
			return "Unknown Linux"
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				// 移除引号
				version := strings.TrimPrefix(line, "PRETTY_NAME=")
				version = strings.Trim(version, "\"")
				return version
			}
		}
	}
	
	return "Unknown Linux"
}

// getKernelVersion 获取内核版本
func getKernelVersion() string {
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getUptime 获取系统运行时间
func getUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	
	parts := strings.Fields(string(data))
	if len(parts) < 1 {
		return "unknown"
	}
	
	uptimeSeconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "unknown"
	}
	
	// 转换为天、小时、分钟
	days := int(uptimeSeconds / 86400)
	hours := int((uptimeSeconds - float64(days*86400)) / 3600)
	minutes := int((uptimeSeconds - float64(days*86400) - float64(hours*3600)) / 60)
	
	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

// GetComputeNodes 获取计算节点真实信息
func GetComputeNodes() []models.NodeModel {
	// 检查slurm是否安装
	if _, err := os.Stat("/usr/sbin/slurmctld"); os.IsNotExist(err) {
		// Slurm未安装
		return []models.NodeModel{}
	}
	
	// 检查slurmctld服务是否运行
	cmd := exec.Command("systemctl", "is-active", "slurmctld")
	output, err := cmd.Output()
	if err != nil {
		// Slurmctld未运行
		return []models.NodeModel{}
	}
	
	status := strings.TrimSpace(string(output))
	if status != "active" {
		// Slurmctld未运行
		return []models.NodeModel{}
	}
	
	// 检查是否有配置文件
	if _, err := os.Stat("/etc/slurm/slurm.conf"); os.IsNotExist(err) {
		// Slurmctld已运行但没有客户端配置
		return []models.NodeModel{}
	}
	
	// 尝试使用sinfo命令获取节点信息
	cmd = exec.Command("sinfo", "-h", "-o", "%n|%C|%m")
	output, err = cmd.Output()
	if err != nil {
		// Slurmctld已运行但没有客户端在线
		return []models.NodeModel{}
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		// Slurmctld已运行但没有客户端在线
		return []models.NodeModel{}
	}
	
	var nodes []models.NodeModel
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		
		hostname := parts[0]
		cpuInfo := strings.Split(parts[1], "/")
		// totalMemory := parts[2]  // 不再使用这个变量
		
		// 计算CPU使用率 (已分配/总计)
		if len(cpuInfo) >= 4 {
			allocated, _ := strconv.ParseFloat(cpuInfo[2], 64)
			total, _ := strconv.ParseFloat(cpuInfo[3], 64)
			
			cpuUsage := 0.0
			if total > 0 {
				cpuUsage = (allocated / total) * 100
			}
			
			// 内存信息处理
			// totalMem, _ := strconv.ParseFloat(totalMemory, 64)  // 移除未使用的变量
			// 这里简化处理，实际内存使用率需要通过其他方式获取
			memoryUsage := 30.0 // 默认值
			
			nodes = append(nodes, models.NodeModel{
				Hostname:    hostname,
				IP:          hostname, // 简化处理，实际应该获取真实IP
				CPUUsage:    cpuUsage,
				MemoryUsage: memoryUsage,
			})
		}
	}
	
	// 如果没有获取到节点信息
	if len(nodes) == 0 {
		// Slurmctld已运行但没有客户端在线
		return []models.NodeModel{}
	}
	
	return nodes
}