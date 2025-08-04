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

// GetComputeNodes 获取计算节点信息（暂时仍使用模拟数据，实际环境中应从集群管理系统获取）
func GetComputeNodes() []models.NodeModel {
	// 注意：在实际HPC环境中，这些信息应该从集群管理系统（如Slurm）获取
	// 这里暂时保留模拟数据，直到有真实的集群环境可以连接
	nodes := []models.NodeModel{
		{
			Hostname:     "compute-01",
			IP:           "192.168.1.11",
			CPUUsage:     68,
			MemoryUsage:  33.75, // 5.4GB/16GB as percentage
		},
		{
			Hostname:     "compute-02",
			IP:           "192.168.1.12",
			CPUUsage:     24,
			MemoryUsage:  19.375, // 3.1GB/16GB as percentage
		},
		{
			Hostname:     "compute-03",
			IP:           "192.168.1.13",
			CPUUsage:     87,
			MemoryUsage:  80, // 12.8GB/16GB as percentage
		},
		{
			Hostname:     "compute-04",
			IP:           "192.168.1.14",
			CPUUsage:     5,
			MemoryUsage:  7.5, // 1.2GB/16GB as percentage
		},
	}

	return nodes
}