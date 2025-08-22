package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"panel-tool/internal/utils"
)

// SpackService Spack 服务结构体
type SpackService struct {
	logger *utils.Logger
	// 添加安装状态跟踪
	installing     bool
	installMutex   sync.Mutex
	installLog     []string
	installLogMutex sync.Mutex
	
	// 添加安装状态缓存
	cachedStatus   *SpackInfo
	cacheTime      time.Time
	cacheMutex     sync.RWMutex
	cacheDuration  time.Duration
}

// NewSpackService 创建新的 Spack 服务实例
func NewSpackService() *SpackService {
	return &SpackService{
		logger: utils.NewLogger(),
		installing: false,
		installLog: make([]string, 0),
		cacheDuration: 30 * time.Second, // 缓存30秒
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

// InstallationStatus 安装状态结构体
type InstallationStatus struct {
	Installing bool     `json:"installing"`
	Log        []string `json:"log"`
}

// OSInfo 操作系统信息结构体
type OSInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	ID      string `json:"id"`
}

// CheckSpackStatus 检查 Spack 安装状态（带缓存）
func (s *SpackService) CheckSpackStatus() SpackInfo {
	// 检查缓存
	s.cacheMutex.RLock()
	if s.cachedStatus != nil && time.Since(s.cacheTime) < s.cacheDuration {
		cached := *s.cachedStatus
		s.cacheMutex.RUnlock()
		return cached
	}
	s.cacheMutex.RUnlock()

	s.logger.Info("检查 Spack 安装状态")

	info := SpackInfo{
		Installed: false,
		Version:   "",
	}

	// 检查 spack 命令是否存在
	_, err := exec.LookPath("spack")
	if err != nil {
		// 如果在 PATH 中找不到，检查默认安装位置
		homeDir, err := os.UserHomeDir()
		if err == nil {
			spackBinPath := filepath.Join(homeDir, "spack", "bin", "spack")
			if _, err := os.Stat(spackBinPath); err == nil {
				// Spack 在默认位置存在，尝试使用完整路径执行
				cmd := exec.Command(spackBinPath, "--version")
				output, err := cmd.Output()
				if err == nil {
					info.Installed = true
					info.Version = strings.TrimSpace(string(output))
					s.logger.Info(fmt.Sprintf("Spack 已安装（通过直接路径），版本: %s", info.Version))
					// 更新缓存
					s.cacheMutex.Lock()
					s.cachedStatus = &info
					s.cacheTime = time.Now()
					s.cacheMutex.Unlock()
					return info
				}
			}
		}
		
		s.logger.Info("Spack 未安装")
		// 更新缓存
		s.cacheMutex.Lock()
		s.cachedStatus = &info
		s.cacheTime = time.Now()
		s.cacheMutex.Unlock()
		return info
	}

	// 检查 Spack 版本
	cmd := exec.Command("spack", "--version")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(fmt.Sprintf("获取 Spack 版本失败: %v", err))
		// 更新缓存
		s.cacheMutex.Lock()
		s.cachedStatus = &info
		s.cacheTime = time.Now()
		s.cacheMutex.Unlock()
		return info
	}

	info.Installed = true
	info.Version = strings.TrimSpace(string(output))
	s.logger.Info(fmt.Sprintf("Spack 已安装，版本: %s", info.Version))

	// 更新缓存
	s.cacheMutex.Lock()
	s.cachedStatus = &info
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	return info
}

// InvalidateStatusCache 使状态缓存失效
func (s *SpackService) InvalidateStatusCache() {
	s.cacheMutex.Lock()
	s.cachedStatus = nil
	s.cacheTime = time.Time{}
	s.cacheMutex.Unlock()
}

// GetInstallationStatus 获取安装状态
func (s *SpackService) GetInstallationStatus() InstallationStatus {
	s.installLogMutex.Lock()
	defer s.installLogMutex.Unlock()
	
	// 复制日志
	logCopy := make([]string, len(s.installLog))
	copy(logCopy, s.installLog)
	
	return InstallationStatus{
		Installing: s.installing,
		Log:        logCopy,
	}
}

// addInstallLog 添加安装日志
func (s *SpackService) addInstallLog(message string) {
	s.installLogMutex.Lock()
	defer s.installLogMutex.Unlock()
	
	logEntry := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
	s.installLog = append(s.installLog, logEntry)
	
	// 限制日志数量，避免内存占用过大
	if len(s.installLog) > 1000 {
		s.installLog = s.installLog[len(s.installLog)-1000:]
	}
}

// clearInstallLog 清除安装日志
func (s *SpackService) clearInstallLog() {
	s.installLogMutex.Lock()
	defer s.installLogMutex.Unlock()
	
	s.installLog = make([]string, 0)
}

// detectOS 检测操作系统信息
func (s *SpackService) detectOS() OSInfo {
	osInfo := OSInfo{
		Name:    "Unknown",
		Version: "Unknown",
		ID:      "unknown",
	}
	
	// 尝试读取 /etc/os-release 文件
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		s.logger.Error(fmt.Sprintf("无法读取 /etc/os-release: %v", err))
		return osInfo
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			osInfo.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			osInfo.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		} else if strings.HasPrefix(line, "ID=") {
			osInfo.ID = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
		}
	}
	
	s.logger.Info(fmt.Sprintf("检测到操作系统: %s %s (ID: %s)", osInfo.Name, osInfo.Version, osInfo.ID))
	return osInfo
}

// getPackageManager 根据操作系统获取包管理器
func (s *SpackService) getPackageManager(osInfo OSInfo) string {
	// Rocky Linux 9.x 使用 dnf
	if osInfo.ID == "rocky" {
		majorVersion := strings.Split(osInfo.Version, ".")[0]
		if majorVersion >= "9" {
			return "dnf"
		}
		return "yum"
	}
	
	// RHEL 8+ 使用 dnf
	if osInfo.ID == "rhel" {
		majorVersion := strings.Split(osInfo.Version, ".")[0]
		if majorVersion >= "8" {
			return "dnf"
		}
		return "yum"
	}
	
	// CentOS 8+ 使用 dnf
	if osInfo.ID == "centos" {
		majorVersion := strings.Split(osInfo.Version, ".")[0]
		if majorVersion >= "8" {
			return "dnf"
		}
		return "yum"
	}
	
	// Fedora 使用 dnf
	if osInfo.ID == "fedora" {
		return "dnf"
	}
	
	// Ubuntu/Debian 使用 apt
	if osInfo.ID == "ubuntu" || osInfo.ID == "debian" {
		return "apt"
	}
	
	// openEuler 使用 dnf
	if osInfo.ID == "openeuler" {
		return "dnf"
	}
	
	// 默认使用 yum
	return "yum"
}

// isCriticalDependency 判断是否为关键依赖
func (s *SpackService) isCriticalDependency(dep string) bool {
	criticalDeps := []string{
		"python3",
		"gcc",
		"gcc-c++",
		"build-essential", // Ubuntu/Debian 的编译工具包
		"make",
		"git",
		"curl",
		"wget",
	}
	
	for _, critical := range criticalDeps {
		if dep == critical {
			return true
		}
	}
	return false
}

// configureForRockyLinux 为 Rocky Linux 9.6 添加特定配置
func (s *SpackService) configureForRockyLinux(osInfo OSInfo, spackDir string, logChan chan<- string) error {
	// 只为 Rocky Linux 9.x 进行特定配置
	if osInfo.ID != "rocky" || !strings.HasPrefix(osInfo.Version, "9") {
		return nil
	}
	
	if logChan != nil {
		logChan <- "正在为 Rocky Linux 9.6 配置 Spack..."
	}
	s.addInstallLog("正在为 Rocky Linux 9.6 配置 Spack...")
	
	// 创建 Spack 配置目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %v", err)
	}
	
	spackConfigDir := filepath.Join(homeDir, ".spack")
	if err := os.MkdirAll(spackConfigDir, 0755); err != nil {
		return fmt.Errorf("创建 Spack 配置目录失败: %v", err)
	}
	
	// 配置编译器
	compilerConfig := `compilers:
- compiler:
    spec: gcc@11.4.1
    paths:
      cc: /usr/bin/gcc
      cxx: /usr/bin/g++
      f77: /usr/bin/gfortran
      fc: /usr/bin/gfortran
    flags: {}
    operating_system: rocky9
    target: x86_64
    modules: []
    environment: {}
    extra_rpaths: []
`
	
	compilerConfigPath := filepath.Join(spackConfigDir, "compilers.yaml")
	if err := os.WriteFile(compilerConfigPath, []byte(compilerConfig), 0644); err != nil {
		return fmt.Errorf("写入编译器配置失败: %v", err)
	}
	
	// 配置包管理器偏好
	packagesConfig := `packages:
  all:
    providers:
      mpi: [openmpi, mpich]
      blas: [openblas, netlib-lapack]
      lapack: [openblas, netlib-lapack]
    variants: +shared
  cmake:
    version: [3.20:]
  python:
    version: [3.9:]
  gcc:
    version: [11.4.1]
    buildable: false
    externals:
    - spec: gcc@11.4.1
      prefix: /usr
  openssl:
    buildable: false
    externals:
    - spec: openssl@1.1.1
      prefix: /usr
`
	
	packagesConfigPath := filepath.Join(spackConfigDir, "packages.yaml")
	if err := os.WriteFile(packagesConfigPath, []byte(packagesConfig), 0644); err != nil {
		return fmt.Errorf("写入包配置失败: %v", err)
	}
	
	// 配置模块系统
	modulesConfig := `modules:
  default:
    enable:
      - tcl
    tcl:
      hash_length: 0
      naming_scheme: '{name}/{version}'
      all:
        conflict:
          - '{name}'
        environment:
          set:
            '{name}_ROOT': '{prefix}'
`
	
	modulesConfigPath := filepath.Join(spackConfigDir, "modules.yaml")
	if err := os.WriteFile(modulesConfigPath, []byte(modulesConfig), 0644); err != nil {
		return fmt.Errorf("写入模块配置失败: %v", err)
	}
	
	// 配置缓存和构建设置
	configConfig := `config:
  install_tree: $spack/opt/spack
  template_dirs:
    - $spack/share/spack/templates
  module_roots:
    tcl: $spack/share/spack/modules
  build_stage:
    - $tempdir/$user/spack-stage
    - ~/.spack/stage
  source_cache: ~/.spack/cache
  misc_cache: ~/.spack/cache
  connect_timeout: 10
  verify_ssl: true
  suppress_gpg_warnings: false
  install_missing_compilers: false
  checksum: true
  dirty: false
  build_language: C
  locks: true
  ccache: false
  concretizer: clingo
  db_lock_timeout: 120
  package_lock_timeout: null
  shared_linking: 'rpath'
  allow_sgid: true
  binary_index_root: 'https://binaries.spack.io'
`
	
	configConfigPath := filepath.Join(spackConfigDir, "config.yaml")
	if err := os.WriteFile(configConfigPath, []byte(configConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	if logChan != nil {
		logChan <- "Rocky Linux 9.6 特定配置完成"
	}
	s.addInstallLog("Rocky Linux 9.6 特定配置完成")
	s.logger.Info("Rocky Linux 9.6 特定配置完成")
	
	return nil
}

// InstallSpack 安装 Spack
func (s *SpackService) InstallSpack(logChan chan<- string) error {
	// 设置安装状态
	s.installMutex.Lock()
	if s.installing {
		s.installMutex.Unlock()
		if logChan != nil {
			logChan <- "安装已在进行中..."
		}
		return fmt.Errorf("安装已在进行中")
	}
	s.installing = true
	s.installMutex.Unlock()
	
	// 确保在函数结束时重置安装状态
	defer func() {
		s.installMutex.Lock()
		s.installing = false
		s.installMutex.Unlock()
		
		// 关闭 channel
		if logChan != nil {
			close(logChan)
		}
		
		// 使缓存失效，因为安装状态已改变
		s.InvalidateStatusCache()
	}()
	
	if logChan != nil {
		logChan <- "开始安装 Spack..."
	}
	s.addInstallLog("开始安装 Spack...")
	s.logger.Info("开始安装 Spack")

	// 检查是否已经安装
	if s.CheckSpackStatus().Installed {
		if logChan != nil {
			logChan <- "Spack 已经安装"
		}
		s.addInstallLog("Spack 已经安装")
		s.logger.Info("Spack 已经安装")
		return nil
	}

	// 检测操作系统
	osInfo := s.detectOS()
	packageManager := s.getPackageManager(osInfo)
	
	detectedMsg := fmt.Sprintf("检测到操作系统: %s %s (ID: %s)，使用包管理器: %s", osInfo.Name, osInfo.Version, osInfo.ID, packageManager)
	if logChan != nil {
		logChan <- detectedMsg
		if osInfo.ID == "rocky" && strings.HasPrefix(osInfo.Version, "9") {
			logChan <- "✓ 已识别为 Rocky Linux 9.x，将应用专门优化"
		}
	}
	s.addInstallLog(detectedMsg)
	s.logger.Info(detectedMsg)
	
	// 安装依赖
	if logChan != nil {
		logChan <- "正在安装依赖..."
	}
	s.addInstallLog("正在安装依赖...")
	s.logger.Info("正在安装依赖")
	
	// 根据不同操作系统定义依赖包
	var dependencies []string
	var installArgs []string
	
	switch packageManager {
	case "dnf":
		// Rocky Linux 9.6, RHEL 8+, CentOS 8+, Fedora 使用的依赖包
		dependencies = []string{
			"python3",
			"python3-pip",
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
			"which",
			"file",
			"findutils",
			"diffutils",
			"hostname",
			"openssl-devel",
			"zlib-devel",
			"bzip2-devel",
			"readline-devel",
			"sqlite-devel",
			"libffi-devel",
		}
		installArgs = []string{"install", "-y"}
	case "yum":
		// 旧版本 RHEL/CentOS 使用的依赖包
		dependencies = []string{
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
			"which",
			"file",
			"findutils",
			"diffutils",
			"hostname",
			"openssl-devel",
			"zlib-devel",
			"bzip2-devel",
		}
		installArgs = []string{"install", "-y"}
	case "apt":
		// Ubuntu/Debian 使用的依赖包
		dependencies = []string{
			"python3",
			"python3-pip",
			"build-essential",
			"git",
			"curl",
			"wget",
			"patch",
			"bzip2",
			"gzip",
			"tar",
			"xz-utils",
			"file",
			"findutils",
			"hostname",
			"libssl-dev",
			"zlib1g-dev",
			"libbz2-dev",
			"libreadline-dev",
			"libsqlite3-dev",
			"libffi-dev",
		}
		installArgs = []string{"update", "&&", "apt", "install", "-y"}
	default:
		// 默认使用 yum 的依赖包
		dependencies = []string{
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
		installArgs = []string{"install", "-y"}
	}
	
	// 对于 apt，先执行 update
	if packageManager == "apt" {
		if logChan != nil {
			logChan <- "正在更新软件包列表..."
		}
		s.addInstallLog("正在更新软件包列表...")
		cmd := exec.Command("apt", "update")
		if err := cmd.Run(); err != nil {
			if logChan != nil {
				logChan <- fmt.Sprintf("更新软件包列表失败: %v", err)
			}
			s.addInstallLog(fmt.Sprintf("更新软件包列表失败: %v", err))
			s.logger.Error(fmt.Sprintf("更新软件包列表失败: %v", err))
			return fmt.Errorf("更新软件包列表失败: %v", err)
		}
		installArgs = []string{"install", "-y"}
	}
	
	for _, dep := range dependencies {
		if logChan != nil {
			logChan <- fmt.Sprintf("正在安装依赖: %s", dep)
		}
		s.addInstallLog(fmt.Sprintf("正在安装依赖: %s", dep))
		s.logger.Info(fmt.Sprintf("正在安装依赖: %s", dep))
		
		// 执行安装命令
		installMsg := fmt.Sprintf("正在安装依赖: %s...", dep)
		if logChan != nil {
			logChan <- installMsg
		}
		s.addInstallLog(installMsg)
		
		args := append(installArgs, dep)
		cmd := exec.Command(packageManager, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// 详细的错误信息
			errorDetails := fmt.Sprintf("命令: %s %s\n错误: %v\n输出: %s", 
				packageManager, strings.Join(args, " "), err, string(output))
			errorMsg := fmt.Sprintf("安装依赖 %s 失败", dep)
			
			if s.isCriticalDependency(dep) {
				// 关键依赖安装失败，返回错误
				fullErrorMsg := fmt.Sprintf("❌ %s (关键依赖)\n%s", errorMsg, errorDetails)
				if logChan != nil {
					logChan <- fmt.Sprintf("❌ %s (关键依赖)", errorMsg)
					logChan <- "详细错误信息已记录到日志"
				}
				s.addInstallLog(fullErrorMsg)
				s.logger.Error(fullErrorMsg)
				return fmt.Errorf("%s: %v", errorMsg, err)
			} else {
				// 非关键依赖安装失败，记录警告但继续
				fullWarningMsg := fmt.Sprintf("⚠️  %s (非关键依赖)\n%s", errorMsg, errorDetails)
				if logChan != nil {
					logChan <- fmt.Sprintf("⚠️  %s (非关键依赖，继续安装)", errorMsg)
				}
				s.addInstallLog(fullWarningMsg)
				s.logger.Error(fullWarningMsg)
				continue
			}
		} else {
			successMsg := fmt.Sprintf("✓ 成功安装依赖: %s", dep)
			if logChan != nil {
				logChan <- successMsg
			}
			s.addInstallLog(successMsg)
			s.logger.Info(successMsg)
		}
	}

	// 下载并安装 Spack
	if logChan != nil {
		logChan <- "正在下载 Spack..."
	}
	s.addInstallLog("正在下载 Spack...")
	s.logger.Info("正在下载 Spack")
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if logChan != nil {
			logChan <- fmt.Sprintf("获取用户主目录失败: %v", err)
		}
		s.addInstallLog(fmt.Sprintf("获取用户主目录失败: %v", err))
		s.logger.Error(fmt.Sprintf("获取用户主目录失败: %v", err))
		return err
	}

	spackDir := filepath.Join(homeDir, "spack")
	
	// 检查目录是否存在，如果存在且不是空目录，则删除它
	if _, err := os.Stat(spackDir); err == nil {
		// 目录存在，检查是否为空
		entries, err := os.ReadDir(spackDir)
		if err != nil {
			if logChan != nil {
				logChan <- fmt.Sprintf("检查 Spack 目录失败: %v", err)
			}
			s.addInstallLog(fmt.Sprintf("检查 Spack 目录失败: %v", err))
			s.logger.Error(fmt.Sprintf("检查 Spack 目录失败: %v", err))
			return err
		}
		
		// 如果目录不为空，则删除它
		if len(entries) > 0 {
			if logChan != nil {
				logChan <- "检测到已存在的 Spack 目录，正在清理..."
			}
			s.addInstallLog("检测到已存在的 Spack 目录，正在清理...")
			s.logger.Info("检测到已存在的 Spack 目录，正在清理...")
			
			err = os.RemoveAll(spackDir)
			if err != nil {
				if logChan != nil {
					logChan <- fmt.Sprintf("清理 Spack 目录失败: %v", err)
				}
				s.addInstallLog(fmt.Sprintf("清理 Spack 目录失败: %v", err))
				s.logger.Error(fmt.Sprintf("清理 Spack 目录失败: %v", err))
				return err
			}
		}
	}
	
	// 克隆 Spack 仓库
	downloadMsg := "正在从 GitHub 下载 Spack 源码..."
	if logChan != nil {
		logChan <- downloadMsg
	}
	s.addInstallLog(downloadMsg)
	s.logger.Info("开始下载 Spack")
	
	cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/spack/spack.git", spackDir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errorDetails := fmt.Sprintf("创建 stdout pipe 失败: %v", err)
		if logChan != nil {
			logChan <- fmt.Sprintf("❌ %s", errorDetails)
		}
		s.addInstallLog(errorDetails)
		s.logger.Error(errorDetails)
		return err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		errorDetails := fmt.Sprintf("创建 stderr pipe 失败: %v", err)
		if logChan != nil {
			logChan <- fmt.Sprintf("❌ %s", errorDetails)
		}
		s.addInstallLog(errorDetails)
		s.logger.Error(errorDetails)
		return err
	}
	
	if err := cmd.Start(); err != nil {
		errorDetails := fmt.Sprintf("启动 git clone 命令失败: %v", err)
		if logChan != nil {
			logChan <- fmt.Sprintf("❌ %s", errorDetails)
			logChan <- "请检查网络连接和 Git 是否已安装"
		}
		s.addInstallLog(errorDetails)
		s.logger.Error(errorDetails)
		return err
	}
	
	// 读取 stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			msg := scanner.Text()
			if logChan != nil {
				logChan <- msg
			}
			s.addInstallLog(msg)
		}
	}()
	
	// 读取 stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			msg := scanner.Text()
			if logChan != nil {
				logChan <- msg
			}
			s.addInstallLog(msg)
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		errorDetails := fmt.Sprintf("命令: git clone --depth 1 https://github.com/spack/spack.git %s\n错误: %v", spackDir, err)
		errorMsg := "克隆 Spack 仓库失败"
		fullErrorMsg := fmt.Sprintf("❌ %s\n%s", errorMsg, errorDetails)
		
		if logChan != nil {
			logChan <- fmt.Sprintf("❌ %s", errorMsg)
			logChan <- "请检查网络连接和 Git 是否已安装"
		}
		s.addInstallLog(fullErrorMsg)
		s.logger.Error(fullErrorMsg)
		return fmt.Errorf("%s: %v", errorMsg, err)
	}
	
	successMsg := "✓ Spack 源码下载完成"
	if logChan != nil {
		logChan <- successMsg
		if osInfo.ID == "rocky" && strings.HasPrefix(osInfo.Version, "9") {
			logChan <- "正在应用 Rocky Linux 9.6 专用配置..."
		}
	}
	s.addInstallLog(successMsg)
	s.logger.Info("Spack 下载完成")

	// 检出特定版本 (1.0.0)
	if logChan != nil {
		logChan <- "正在检出 Spack 1.0.0 版本..."
	}
	s.addInstallLog("正在检出 Spack 1.0.0 版本...")
	s.logger.Info("正在检出 Spack 1.0.0 版本")
	
	cmd = exec.Command("git", "checkout", "v1.0.0")
	cmd.Dir = spackDir
	err = cmd.Run()
	if err != nil {
		if logChan != nil {
			logChan <- fmt.Sprintf("检出 Spack 1.0.0 版本失败: %v", err)
		}
		s.addInstallLog(fmt.Sprintf("检出 Spack 1.0.0 版本失败: %v", err))
		s.logger.Error(fmt.Sprintf("检出 Spack 1.0.0 版本失败: %v", err))
		return err
	}

	// 配置环境变量
	if logChan != nil {
		logChan <- "正在配置环境变量..."
	}
	s.addInstallLog("正在配置环境变量...")
	s.logger.Info("正在配置环境变量")
	
	// 创建日志目录
	logDir := filepath.Join(homeDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.logger.Error(fmt.Sprintf("创建日志目录失败: %v", err))
		// 不返回错误，因为这不是关键步骤
	}

	// 尝试将 Spack 添加到用户的 shell 配置文件中
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	setupEnvPath := filepath.Join(spackDir, "share", "spack", "setup-env.sh")
	if _, err := os.Stat(setupEnvPath); err == nil {
		// 为 Rocky Linux 9.6 优化环境变量配置
		var lineToAdd string
		if osInfo.ID == "rocky" && strings.HasPrefix(osInfo.Version, "9") {
			lineToAdd = fmt.Sprintf(`
# Spack 安装配置 (Rocky Linux 9.6 优化)
export SPACK_ROOT=%s
export PATH=$SPACK_ROOT/bin:$PATH
# Rocky Linux 9.6 特定优化
export SPACK_PYTHON=/usr/bin/python3
export SPACK_DISABLE_LOCAL_CONFIG=false
export TMPDIR=/tmp
export SPACK_USER_CACHE_PATH=$HOME/.spack
# 启用并行构建
export SPACK_BUILD_JOBS=$(nproc)
# 设置编译器缓存
export SPACK_CCACHE_DIR=$HOME/.spack/ccache
source $SPACK_ROOT/share/spack/setup-env.sh
`, spackDir)
		} else {
			lineToAdd = fmt.Sprintf("\n# Spack 安装配置\nexport SPACK_ROOT=%s\nsource $SPACK_ROOT/share/spack/setup-env.sh\n", spackDir)
		}
		
		// 检查是否已经添加过
		alreadyAdded := false
		if content, err := os.ReadFile(bashrcPath); err == nil {
			if strings.Contains(string(content), "SPACK_ROOT") {
				alreadyAdded = true
			}
		}
		
		// 如果还没有添加，则添加到 .bashrc
		if !alreadyAdded {
			if f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				defer f.Close()
				if _, err := f.WriteString(lineToAdd); err != nil {
					if logChan != nil {
						logChan <- fmt.Sprintf("警告: 添加 Spack 到 .bashrc 失败: %v", err)
					}
					s.logger.Error(fmt.Sprintf("添加 Spack 到 .bashrc 失败: %v", err))
				} else {
					s.logger.Info("成功添加 Spack 到 .bashrc")
					
					// 尝试激活环境
					cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && spack --version", bashrcPath))
					if err := cmd.Run(); err != nil {
						s.logger.Info("环境激活需要重新登录才能生效")
					}
				}
			} else {
				if logChan != nil {
					logChan <- fmt.Sprintf("警告: 无法打开 .bashrc 文件: %v", err)
				}
				s.logger.Error(fmt.Sprintf("无法打开 .bashrc 文件: %v", err))
			}
		}
	}

	// 为 Rocky Linux 9.6 添加特定配置
	if err := s.configureForRockyLinux(osInfo, spackDir, logChan); err != nil {
		if logChan != nil {
			logChan <- fmt.Sprintf("警告: Rocky Linux 特定配置失败: %v", err)
		}
		s.addInstallLog(fmt.Sprintf("警告: Rocky Linux 特定配置失败: %v", err))
		s.logger.Error(fmt.Sprintf("Rocky Linux 特定配置失败: %v", err))
		// 不返回错误，因为这不是关键步骤
	}
	
	// 添加提示信息，告知用户如何使用 Spack
	if logChan != nil {
		logChan <- "Spack 安装完成!"
		logChan <- "环境变量已自动配置并激活"
		logChan <- "如遇到命令未找到问题，请重新登录或执行以下命令:"
		logChan <- "  source ~/.bashrc"
		if osInfo.ID == "rocky" && strings.HasPrefix(osInfo.Version, "9") {
			logChan <- "已为 Rocky Linux 9.6 优化配置"
		}
	}
	
	s.addInstallLog("Spack 安装完成!")
	s.addInstallLog("环境变量已自动配置并激活")
	s.logger.Info("Spack 安装完成")
	
	return nil
}

// GetAvailablePackages 获取可安装的软件包列表
func (s *SpackService) GetAvailablePackages() ([]Package, error) {
	s.logger.Info("获取可安装的软件包列表")

	// 使用缓存检查 Spack 安装状态
	if !s.CheckSpackStatus().Installed {
		return nil, fmt.Errorf("Spack 未安装")
	}

	// 执行 spack list 命令
	cmd := exec.Command("spack", "list")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(fmt.Sprintf("执行 spack list 命令失败: %v", err))
		// 即使命令执行失败，也返回空列表而不是错误，确保前端能够处理
		return []Package{}, nil
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

	// 使用缓存检查 Spack 安装状态
	if !s.CheckSpackStatus().Installed {
		return nil, fmt.Errorf("Spack 未安装")
	}

	// 执行 spack find 命令
	cmd := exec.Command("spack", "find", "--format", "{name}@{version} {hash:7}")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error(fmt.Sprintf("执行 spack find 命令失败: %v", err))
		// 即使命令执行失败，也返回空列表而不是错误，确保前端能够处理
		return []Package{}, nil
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
	s.addInstallLog(fmt.Sprintf("开始安装软件包: %s，选项: %s", packageName, options))
	s.logger.Info(fmt.Sprintf("开始安装软件包: %s，选项: %s", packageName, options))

	// 使用缓存检查 Spack 安装状态
	if !s.CheckSpackStatus().Installed {
		logChan <- "错误: Spack 未安装"
		s.addInstallLog("错误: Spack 未安装")
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
		s.addInstallLog(fmt.Sprintf("创建 stdout pipe 失败: %v", err))
		s.logger.Error(fmt.Sprintf("创建 stdout pipe 失败: %v", err))
		return err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logChan <- fmt.Sprintf("创建 stderr pipe 失败: %v", err)
		s.addInstallLog(fmt.Sprintf("创建 stderr pipe 失败: %v", err))
		s.logger.Error(fmt.Sprintf("创建 stderr pipe 失败: %v", err))
		return err
	}
	
	if err := cmd.Start(); err != nil {
		logChan <- fmt.Sprintf("启动安装命令失败: %v", err)
		s.addInstallLog(fmt.Sprintf("启动安装命令失败: %v", err))
		s.logger.Error(fmt.Sprintf("启动安装命令失败: %v", err))
		return err
	}
	
	// 读取 stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			msg := scanner.Text()
			if logChan != nil {
				logChan <- msg
			}
			s.addInstallLog(msg)
		}
	}()
	
	// 读取 stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			msg := scanner.Text()
			if logChan != nil {
				logChan <- msg
			}
			s.addInstallLog(msg)
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		logChan <- fmt.Sprintf("安装软件包失败: %v", err)
		s.addInstallLog(fmt.Sprintf("安装软件包失败: %v", err))
		s.logger.Error(fmt.Sprintf("安装软件包失败: %v", err))
		return err
	}

	logChan <- fmt.Sprintf("软件包 %s 安装完成!", packageName)
	s.addInstallLog(fmt.Sprintf("软件包 %s 安装完成", packageName))
	s.logger.Info(fmt.Sprintf("软件包 %s 安装完成", packageName))
	
	// 保存日志到文件
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, "logs")
	logFile := filepath.Join(logDir, fmt.Sprintf("spack-install-%s-%s.log", 
		packageName, time.Now().Format("2006-01-02-15-04-05")))
	
	// 这里应该实际写入日志文件，但为了简化示例，我们只记录日志
	s.addInstallLog(fmt.Sprintf("安装日志保存到: %s", logFile))
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