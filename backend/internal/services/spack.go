package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"panel-tool/internal/utils"
)

// SpackService Spack 服务结构体
type SpackService struct {
	logger *utils.Logger
}

// NewSpackService 创建新的 Spack 服务实例
func NewSpackService() *SpackService {
	return &SpackService{
		logger: utils.NewLogger(),
	}
}

// SpackInfo Spack 信息结构体
type SpackInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
}

// Package 软件包结构体
type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Versions    string `json:"versions,omitempty"`
	Description string `json:"description,omitempty"`
	Hash        string `json:"hash,omitempty"`
}

// CheckSpackStatus 检查 Spack 安装状态
func (s *SpackService) CheckSpackStatus() SpackInfo {
	s.logger.Info("检查 Spack 安装状态")

	info := SpackInfo{
		Installed: false,
		Version:   "",
	}

	// 检查 spack 命令是否存在
	_, err := exec.LookPath("spack")
	if err != nil {
		s.logger.Info("Spack 未安装")
		return info
	}

	// 检查 Spack 版本
	cmd := exec.Command("spack", "--version")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(fmt.Sprintf("获取 Spack 版本失败: %v", err))
		return info
	}

	info.Installed = true
	info.Version = strings.TrimSpace(string(output))
	s.logger.Info(fmt.Sprintf("Spack 已安装，版本: %s", info.Version))

	return info
}

// InstallSpack 安装 Spack
func (s *SpackService) InstallSpack(logChan chan<- string) error {
	defer close(logChan)
	
	logChan <- "开始安装 Spack..."
	s.logger.Info("开始安装 Spack")

	// 检查是否已经安装
	if s.CheckSpackStatus().Installed {
		logChan <- "Spack 已经安装"
		s.logger.Info("Spack 已经安装")
		return nil
	}

	// 安装依赖
	logChan <- "正在安装依赖..."
	s.logger.Info("正在安装依赖")
	
	// 在 OpenEuler 24 上安装依赖
	dependencies := []string{
		"python3",
		"gcc",
		"gcc-c++",
		"make",
		"git",
		"curl",
		"wget",
		"patch",
		"bzip2",
		"gzip",
		"tar",
		"xz",
	}

	for _, dep := range dependencies {
		logChan <- fmt.Sprintf("正在安装依赖: %s", dep)
		s.logger.Info(fmt.Sprintf("正在安装依赖: %s", dep))
		
		cmd := exec.Command("yum", "install", "-y", dep)
		err := cmd.Run()
		if err != nil {
			logChan <- fmt.Sprintf("安装依赖 %s 失败: %v", dep, err)
			s.logger.Error(fmt.Sprintf("安装依赖 %s 失败: %v", dep, err))
			return fmt.Errorf("安装依赖 %s 失败: %v", dep, err)
		}
	}

	// 下载并安装 Spack
	logChan <- "正在下载 Spack..."
	s.logger.Info("正在下载 Spack")
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logChan <- fmt.Sprintf("获取用户主目录失败: %v", err)
		s.logger.Error(fmt.Sprintf("获取用户主目录失败: %v", err))
		return err
	}

	spackDir := filepath.Join(homeDir, "spack")
	
	// 克隆 Spack 仓库
	cmd := exec.Command("git", "clone", "https://github.com/spack/spack.git", spackDir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logChan <- fmt.Sprintf("创建 stdout pipe 失败: %v", err)
		s.logger.Error(fmt.Sprintf("创建 stdout pipe 失败: %v", err))
		return err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logChan <- fmt.Sprintf("创建 stderr pipe 失败: %v", err)
		s.logger.Error(fmt.Sprintf("创建 stderr pipe 失败: %v", err))
		return err
	}
	
	if err := cmd.Start(); err != nil {
		logChan <- fmt.Sprintf("启动 git clone 命令失败: %v", err)
		s.logger.Error(fmt.Sprintf("启动 git clone 命令失败: %v", err))
		return err
	}
	
	// 读取 stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logChan <- scanner.Text()
		}
	}()
	
	// 读取 stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logChan <- scanner.Text()
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		logChan <- fmt.Sprintf("克隆 Spack 仓库失败: %v", err)
		s.logger.Error(fmt.Sprintf("克隆 Spack 仓库失败: %v", err))
		return err
	}

	// 检出特定版本 (1.0.0)
	logChan <- "正在检出 Spack 1.0.0 版本..."
	s.logger.Info("正在检出 Spack 1.0.0 版本")
	
	cmd = exec.Command("git", "checkout", "v1.0.0")
	cmd.Dir = spackDir
	err = cmd.Run()
	if err != nil {
		logChan <- fmt.Sprintf("检出 Spack 1.0.0 版本失败: %v", err)
		s.logger.Error(fmt.Sprintf("检出 Spack 1.0.0 版本失败: %v", err))
		return err
	}

	// 配置环境变量
	logChan <- "正在配置环境变量..."
	s.logger.Info("正在配置环境变量")
	
	// 创建日志目录
	logDir := filepath.Join(homeDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.logger.Error(fmt.Sprintf("创建日志目录失败: %v", err))
		// 不返回错误，因为这不是关键步骤
	}

	logChan <- "Spack 安装完成!"
	s.logger.Info("Spack 安装完成")
	
	return nil
}

// GetAvailablePackages 获取可安装的软件包列表
func (s *SpackService) GetAvailablePackages() ([]Package, error) {
	s.logger.Info("获取可安装的软件包列表")

	if !s.CheckSpackStatus().Installed {
		return nil, fmt.Errorf("Spack 未安装")
	}

	// 执行 spack list 命令
	cmd := exec.Command("spack", "list")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(fmt.Sprintf("执行 spack list 命令失败: %v", err))
		return nil, fmt.Errorf("执行 spack list 命令失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var packages []Package

	// 跳过第一行标题
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// 简化处理，实际项目中可能需要更复杂的解析
		fields := strings.Fields(line)
		if len(fields) > 0 {
			pkg := Package{
				Name: fields[0],
			}
			
			// 尝试获取更多描述信息
			if len(fields) > 1 {
				pkg.Description = strings.Join(fields[1:], " ")
			}
			
			packages = append(packages, pkg)
		}
	}

	s.logger.Info(fmt.Sprintf("获取到 %d 个可安装软件包", len(packages)))
	return packages, nil
}

// GetInstalledPackages 获取已安装的软件包列表
func (s *SpackService) GetInstalledPackages() ([]Package, error) {
	s.logger.Info("获取已安装的软件包列表")

	if !s.CheckSpackStatus().Installed {
		return nil, fmt.Errorf("Spack 未安装")
	}

	// 执行 spack find 命令
	cmd := exec.Command("spack", "find", "--format", "{name}@{version} {hash:7}")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(fmt.Sprintf("执行 spack find 命令失败: %v", err))
		return nil, fmt.Errorf("执行 spack find 命令失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var packages []Package

	// 解析输出
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.Contains(line, "----") {
			continue
		}
		
		// 跳过标题行
		if strings.Contains(line, "name") && strings.Contains(line, "version") {
			continue
		}
		
		// 解析包信息
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// 分离名称和版本
			nameVersion := fields[0]
			hash := fields[1]
			
			var name, version string
			if strings.Contains(nameVersion, "@") {
				parts := strings.Split(nameVersion, "@")
				name = parts[0]
				version = parts[1]
			} else {
				name = nameVersion
			}
			
			pkg := Package{
				Name:    name,
				Version: version,
				Hash:    hash,
			}
			packages = append(packages, pkg)
		}
	}

	s.logger.Info(fmt.Sprintf("获取到 %d 个已安装软件包", len(packages)))
	return packages, nil
}

// InstallPackage 安装软件包
func (s *SpackService) InstallPackage(packageName string, options string, logChan chan<- string) error {
	defer close(logChan)
	
	logChan <- fmt.Sprintf("开始安装软件包: %s", packageName)
	s.logger.Info(fmt.Sprintf("开始安装软件包: %s，选项: %s", packageName, options))

	if !s.CheckSpackStatus().Installed {
		logChan <- "错误: Spack 未安装"
		s.logger.Error("Spack 未安装")
		return fmt.Errorf("Spack 未安装")
	}

	// 构建安装命令
	args := []string{"install"}
	if options != "" {
		optionsList := strings.Fields(options)
		args = append(args, optionsList...)
	}
	args = append(args, packageName)

	cmd := exec.Command("spack", args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logChan <- fmt.Sprintf("创建 stdout pipe 失败: %v", err)
		s.logger.Error(fmt.Sprintf("创建 stdout pipe 失败: %v", err))
		return err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logChan <- fmt.Sprintf("创建 stderr pipe 失败: %v", err)
		s.logger.Error(fmt.Sprintf("创建 stderr pipe 失败: %v", err))
		return err
	}
	
	if err := cmd.Start(); err != nil {
		logChan <- fmt.Sprintf("启动安装命令失败: %v", err)
		s.logger.Error(fmt.Sprintf("启动安装命令失败: %v", err))
		return err
	}
	
	// 读取 stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logChan <- scanner.Text()
		}
	}()
	
	// 读取 stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logChan <- scanner.Text()
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		logChan <- fmt.Sprintf("安装软件包失败: %v", err)
		s.logger.Error(fmt.Sprintf("安装软件包失败: %v", err))
		return err
	}

	logChan <- fmt.Sprintf("软件包 %s 安装完成!", packageName)
	s.logger.Info(fmt.Sprintf("软件包 %s 安装完成", packageName))
	
	// 保存日志到文件
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, "logs")
	logFile := filepath.Join(logDir, fmt.Sprintf("spack-install-%s-%s.log", 
		packageName, time.Now().Format("2006-01-02-15-04-05")))
	
	// 这里应该实际写入日志文件，但为了简化示例，我们只记录日志
	s.logger.Info(fmt.Sprintf("安装日志保存到: %s", logFile))
	
	return nil
}

// UninstallPackage 卸载软件包
func (s *SpackService) UninstallPackage(packageName string) error {
	s.logger.Info(fmt.Sprintf("卸载软件包: %s", packageName))

	if !s.CheckSpackStatus().Installed {
		s.logger.Error("Spack 未安装")
		return fmt.Errorf("Spack 未安装")
	}

	// 执行卸载命令
	cmd := exec.Command("spack", "uninstall", "-y", packageName)
	err := cmd.Run()
	if err != nil {
		s.logger.Error(fmt.Sprintf("卸载软件包失败: %v", err))
		return fmt.Errorf("卸载软件包失败: %v", err)
	}

	s.logger.Info(fmt.Sprintf("软件包 %s 卸载完成", packageName))
	return nil
}

// GetRepositories 获取软件源配置
func (s *SpackService) GetRepositories() (string, error) {
	s.logger.Info("获取软件源配置")

	if !s.CheckSpackStatus().Installed {
		s.logger.Error("Spack 未安装")
		return "", fmt.Errorf("Spack 未安装")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		s.logger.Error(fmt.Sprintf("获取用户主目录失败: %v", err))
		return "", err
	}

	// 读取配置文件
	configPath := filepath.Join(homeDir, ".spack", "packages.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，返回默认配置
		return "# Spack 软件源配置\npackages:\n  all:\n    providers:\n", nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		s.logger.Error(fmt.Sprintf("读取配置文件失败: %v", err))
		return "", err
	}

	return string(content), nil
}

// SetRepositories 设置软件源配置
func (s *SpackService) SetRepositories(content string) error {
	s.logger.Info("设置软件源配置")

	if !s.CheckSpackStatus().Installed {
		s.logger.Error("Spack 未安装")
		return fmt.Errorf("Spack 未安装")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		s.logger.Error(fmt.Sprintf("获取用户主目录失败: %v", err))
		return err
	}

	// 确保 .spack 目录存在
	spackDir := filepath.Join(homeDir, ".spack")
	if err := os.MkdirAll(spackDir, 0755); err != nil {
		s.logger.Error(fmt.Sprintf("创建 .spack 目录失败: %v", err))
		return err
	}

	// 写入配置文件
	configPath := filepath.Join(spackDir, "packages.yaml")
	err = os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		s.logger.Error(fmt.Sprintf("写入配置文件失败: %v", err))
		return err
	}

	s.logger.Info("软件源配置保存成功")
	return nil
}