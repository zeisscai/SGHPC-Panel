package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"math/rand"
	"sync"
	
	"panel-tool/internal/models"
	"panel-tool/internal/utils"
)

// SlurmDeployService Slurm部署服务结构体
type SlurmDeployService struct {
	config         *models.SlurmConfig
	status         *models.SlurmDeploymentStatus
	logs           []string
	configFilePath string
	mu             sync.Mutex
}

// NewSlurmDeployService 创建新的Slurm部署服务实例
func NewSlurmDeployService() *SlurmDeployService {
	return &SlurmDeployService{
		config: &models.SlurmConfig{
			ClusterName:       "hpc_cluster",
			ControlMachine:    "master",
			SlurmUser:         "slurm",
			SlurmUID:          930,
			SlurmGID:          930,
			SlurmctldPort:     6817,
			SlurmdPort:        6818,
			AuthType:          "auth/munge",
			StateSaveLocation: "/var/spool/slurmctld",
			SlurmctldPidFile:  "/var/run/slurmctld.pid",
			SlurmdPidFile:     "/var/run/slurmd.pid",
			SwitchType:        "switch/none",
			MpiDefault:        "pmi2",
			ProctrackType:     "proctrack/cgroup",
			ReturnToService:   1,
			MaxTime:           "INFINITE",
		},
		status: &models.SlurmDeploymentStatus{
			Phase:     "idle",
			Progress:  0,
			Message:   "等待部署指令",
			StartTime: time.Now().Format("2006-01-02 15:04:05"),
		},
		logs: make([]string, 0),
	}
}

// GetConfig 获取当前配置
func (s *SlurmDeployService) GetConfig() *models.SlurmConfig {
	return s.config
}

// UpdateConfig 更新配置
func (s *SlurmDeployService) UpdateConfig(config *models.SlurmConfig) {
	s.config = config
}

// GetStatus 获取部署状态
func (s *SlurmDeployService) GetStatus() *models.SlurmDeploymentStatus {
	return s.status
}

// GetLogs 获取部署日志
func (s *SlurmDeployService) GetLogs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 返回日志的副本
	logsCopy := make([]string, len(s.logs))
	copy(logsCopy, s.logs)
	return logsCopy
}

// CheckSlurmStatus 检查Slurm状态
func (s *SlurmDeployService) CheckSlurmStatus() string {
	// 检查slurmctld是否已安装
	_, err := os.Stat("/usr/sbin/slurmctld")
	if os.IsNotExist(err) {
		return "not_installed"
	}
	
	// 检查slurmctld服务是否正在运行
	cmd := exec.Command("systemctl", "is-active", "slurmctld")
	output, err := cmd.Output()
	if err != nil {
		return "stopped"
	}
	
	status := strings.TrimSpace(string(output))
	if status == "active" {
		return "running"
	}
	
	return "stopped"
}

// addLog 添加日志
func (s *SlurmDeployService) addLog(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	logEntry := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
	s.logs = append(s.logs, logEntry)
	
	// 使用Logger记录日志
	logger := utils.NewLogger()
	logger.Info(logEntry)
}

// generatePassword 生成指定长度的随机密码
func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

// Deploy 执行Slurm部署
func (s *SlurmDeployService) Deploy() error {
	s.status.Phase = "checking"
	s.status.Progress = 0
	s.status.Message = "开始部署检查"
	s.status.StartTime = time.Now().Format("2006-01-02 15:04:05")
	s.status.ErrorMessage = ""

	s.addLog("开始Slurm部署流程")

	// 1. 环境检查
	if err := s.checkEnvironment(); err != nil {
		s.status.Phase = "failed"
		s.status.ErrorMessage = fmt.Sprintf("环境检查失败: %v", err)
		s.addLog(s.status.ErrorMessage)
		return err
	}

	// 2. 下载源码
	s.status.Phase = "downloading"
	s.status.Progress = 20
	s.status.Message = "正在下载Slurm源码"
	s.addLog("开始下载Slurm源码")

	if err := s.downloadSource(); err != nil {
		s.status.Phase = "failed"
		s.status.ErrorMessage = fmt.Sprintf("源码下载失败: %v", err)
		s.addLog(s.status.ErrorMessage)
		return err
	}

	// 3. 安装依赖
	s.status.Phase = "installing_deps"
	s.status.Progress = 40
	s.status.Message = "正在安装依赖包"
	s.addLog("开始安装依赖包")

	if err := s.installDependencies(); err != nil {
		s.status.Phase = "failed"
		s.status.ErrorMessage = fmt.Sprintf("依赖安装失败: %v", err)
		s.addLog(s.status.ErrorMessage)
		return err
	}

	// 4. 编译安装
	s.status.Phase = "compiling"
	s.status.Progress = 60
	s.status.Message = "正在编译安装Slurm"
	s.addLog("开始编译安装Slurm")

	if err := s.compileAndInstall(); err != nil {
		s.status.Phase = "failed"
		s.status.ErrorMessage = fmt.Sprintf("编译安装失败: %v", err)
		s.addLog(s.status.ErrorMessage)
		return err
	}

	// 5. 配置服务
	s.status.Phase = "configuring"
	s.status.Progress = 80
	s.status.Message = "正在配置Slurm服务"
	s.addLog("开始配置Slurm服务")

	if err := s.configureService(); err != nil {
		s.status.Phase = "failed"
		s.status.ErrorMessage = fmt.Sprintf("服务配置失败: %v", err)
		s.addLog(s.status.ErrorMessage)
		return err
	}

	// 6. 完成部署
	s.status.Phase = "finished"
	s.status.Progress = 100
	s.status.Message = "部署完成"
	s.status.EndTime = time.Now().Format("2006-01-02 15:04:05")
	s.addLog("Slurm部署完成")

	return nil
}

// checkEnvironment 检查部署环境
func (s *SlurmDeployService) checkEnvironment() error {
	s.addLog("检查操作系统版本")

	// 检查是否为OpenEuler发行版
	if _, err := os.Stat("/etc/openeuler-release"); os.IsNotExist(err) {
		// 如果没有openeuler-release文件，尝试检查os-release
		if _, err := os.Stat("/etc/os-release"); os.IsNotExist(err) {
			return fmt.Errorf("当前系统不是OpenEuler或Rocky Linux发行版")
		}
		
		// 检查os-release中是否包含openEuler或Rocky Linux标识
		data, err := os.ReadFile("/etc/os-release")
		if err != nil {
			return fmt.Errorf("无法读取系统版本信息: %v", err)
		}
		
		content := string(data)
		isOpenEuler := strings.Contains(content, "openEuler")
		isRockyLinux := strings.Contains(content, "Rocky Linux")
		
		if !isOpenEuler && !isRockyLinux {
			return fmt.Errorf("当前系统不是OpenEuler或Rocky Linux发行版")
		}
		
		if isOpenEuler {
			// 检查OpenEuler版本号 (要求24以上)
			versionPattern := regexp.MustCompile(`VERSION="?(24\.0[0-9]|24\.[1-9][0-9]*)`)
			if !versionPattern.MatchString(content) {
				return fmt.Errorf("当前OpenEuler系统版本不是24以上，实际版本信息: %s", content)
			}
			s.addLog("检测到系统为OpenEuler发行版，版本符合要求")
		} else if isRockyLinux {
			// 检查Rocky Linux版本号 (要求9.4以上)
			versionPattern := regexp.MustCompile(`VERSION="9\.4|VERSION="[0-9][0-9]\.`)
			if !versionPattern.MatchString(content) {
				// 更精确的匹配
				versionLinePattern := regexp.MustCompile(`VERSION="([0-9]+\.[0-9]+)`)
				matches := versionLinePattern.FindStringSubmatch(content)
				if len(matches) > 1 {
					version := matches[1]
					if version < "9.4" {
						return fmt.Errorf("当前Rocky Linux系统版本不是9.4以上，实际版本: %s", version)
					}
				} else {
					return fmt.Errorf("无法确定Rocky Linux版本，实际版本信息: %s", content)
				}
			}
			s.addLog("检测到系统为Rocky Linux发行版，版本符合要求")
		}
	} else {
		// 检查openeuler-release文件
		data, err := os.ReadFile("/etc/openeuler-release")
		if err != nil {
			return fmt.Errorf("无法读取系统版本信息: %v", err)
		}

		version := strings.TrimSpace(string(data))
		versionPattern := regexp.MustCompile(`(24\.0[0-9]|24\.[1-9][0-9]*)`)
		if !versionPattern.MatchString(version) {
			return fmt.Errorf("当前系统版本不是OpenEuler 24，实际版本信息: %s", version)
		}
		
		s.addLog("检测到系统为OpenEuler发行版，版本符合要求")
	}
	
	// 检查依赖
	s.addLog("检查系统依赖")
	if _, err := exec.LookPath("wget"); err != nil {
		s.addLog("警告: wget未安装，将在依赖安装阶段进行安装")
	}
	
	if _, err := exec.LookPath("yum"); err != nil {
		return fmt.Errorf("系统缺少yum包管理器")
	}
	
	return nil
}

// downloadSource 下载Slurm源码
func (s *SlurmDeployService) downloadSource() error {
	const (
		slurmVersion = "23.11.4"
		downloadURL  = "https://download.schedmd.com/slurm/slurm-23.11.4.tar.bz2"
		sourceDir    = "/tmp/slurm_source"
	)

	s.addLog(fmt.Sprintf("准备下载Slurm %s", slurmVersion))

	// 创建源码目录
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("创建源码目录失败: %v", err)
	}

	// 下载源码包
	s.addLog(fmt.Sprintf("正在下载 %s", downloadURL))
	cmd := exec.Command("wget", "-O", filepath.Join(sourceDir, fmt.Sprintf("slurm-%s.tar.bz2", slurmVersion)), downloadURL)
	cmd.Dir = sourceDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("下载源码失败: %v, 输出: %s", err, string(output))
	}

	// 解压源码包
	s.addLog("正在解压源码包")
	cmd = exec.Command("tar", "-xjf", fmt.Sprintf("slurm-%s.tar.bz2", slurmVersion))
	cmd.Dir = sourceDir

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("解压源码失败: %v, 输出: %s", err, string(output))
	}

	s.configFilePath = filepath.Join(sourceDir, fmt.Sprintf("slurm-%s", slurmVersion))
	s.addLog("源码下载和解压完成")
	return nil
}

// installDependencies 安装编译依赖
func (s *SlurmDeployService) installDependencies() error {
	// 更新yum缓存
	s.addLog("正在更新yum缓存")
	cmd := exec.Command("yum", "makecache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.addLog(fmt.Sprintf("警告: 更新yum缓存失败: %v, 输出: %s", err, string(output)))
	}

	// 安装编译依赖
	s.addLog("正在安装编译依赖包")
	dependencies := []string{
		"rpm-build", "gcc", "make", "munge", "munge-devel", "python3",
		"hwloc", "hwloc-devel", "libssh2-devel", "mariadb-devel", "readline-devel",
		"lua", "lua-devel", "man2html", "numactl", "numactl-devel", "perl",
		"python3-devel", "pam-devel", "perl-ExtUtils-MakeMaker",
	}

	cmd = exec.Command("yum", append([]string{"install", "-y"}, dependencies...)...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装依赖包失败: %v, 输出: %s", err, string(output))
	}

	s.addLog("依赖包安装完成")
	return nil
}

// compileAndInstall 编译和安装Slurm
func (s *SlurmDeployService) compileAndInstall() error {
	// 创建slurm用户和组
	s.addLog("正在创建slurm用户和组")
	cmd := exec.Command("getent", "group", "slurm")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("groupadd", "-g", "930", "slurm")
		if output, err := cmd.CombinedOutput(); err != nil {
			s.addLog(fmt.Sprintf("警告: 创建slurm组失败: %v, 输出: %s", err, string(output)))
		}
	}

	cmd = exec.Command("getent", "passwd", "slurm")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("useradd", "-u", "930", "-g", "slurm", "slurm")
		if output, err := cmd.CombinedOutput(); err != nil {
			s.addLog(fmt.Sprintf("警告: 创建slurm用户失败: %v, 输出: %s", err, string(output)))
		}
	}

	// 创建必要的目录
	dirs := []string{"/var/spool/slurmctld", "/var/log/slurm"}
	for _, dir := range dirs {
		s.addLog(fmt.Sprintf("正在创建目录: %s", dir))
		if err := os.MkdirAll(dir, 0755); err != nil {
			s.addLog(fmt.Sprintf("警告: 创建目录失败 %s: %v", dir, err))
		}
		os.Chown(dir, 930, 930) // 设置所有者为slurm用户
	}

	// 配置、编译和安装
	s.addLog("正在配置编译环境")
	configureCmd := exec.Command("./configure", "--prefix=/usr", "--sysconfdir=/etc/slurm")
	configureCmd.Dir = s.configFilePath

	output, err := configureCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("配置失败: %v, 输出: %s", err, string(output))
	}

	s.addLog("正在编译Slurm")
	makeCmd := exec.Command("make")
	makeCmd.Dir = s.configFilePath

	output, err = makeCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("编译失败: %v, 输出: %s", err, string(output))
	}

	s.addLog("正在安装Slurm")
	installCmd := exec.Command("make", "install")
	installCmd.Dir = s.configFilePath

	output, err = installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装失败: %v, 输出: %s", err, string(output))
	}

	s.addLog("编译和安装完成")
	return nil
}

// configureService 配置Slurm服务
func (s *SlurmDeployService) configureService() error {
	s.addLog("开始配置Slurm服务")
	
	// 生成数据库密码
	dbPassword := generatePassword(10)
	s.addLog(fmt.Sprintf("生成数据库密码: %s", dbPassword))
	s.addLog("重要提醒: 请保存好数据库密码，20秒后将继续安装")
	
	// 等待20秒
	time.Sleep(20 * time.Second)
	
	// 配置munge
	s.addLog("配置Munge服务")
	if err := s.configureMunge(); err != nil {
		return fmt.Errorf("配置Munge失败: %v", err)
	}
	
	// 配置数据库
	s.addLog("配置Slurm数据库")
	if err := s.configureDatabase(dbPassword); err != nil {
		return fmt.Errorf("配置数据库失败: %v", err)
	}
	
	// 创建基础配置文件
	s.addLog("创建Slurm配置文件")
	if err := s.createSlurmConfig(); err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	
	// 启用并启动服务
	s.addLog("启用并启动Slurm服务")
	cmd := exec.Command("systemctl", "enable", "slurmctld")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启用slurmctld服务失败: %v, 输出: %s", err, string(output)))
	}
	
	cmd = exec.Command("systemctl", "start", "slurmctld")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启动slurmctld服务失败: %v, 输出: %s", err, string(output)))
	}
	
	s.addLog("服务配置完成")
	return nil
}

// configureMunge 配置Munge服务
func (s *SlurmDeployService) configureMunge() error {
	// 生成munge密钥
	s.addLog("生成Munge密钥")
	cmd := exec.Command("dd", "if=/dev/urandom", "of=/etc/munge/munge.key", "bs=1", "count=1024")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("生成munge密钥失败: %v, 输出: %s", err, string(output))
	}
	
	// 设置权限
	cmd = exec.Command("chown", "munge:munge", "/etc/munge/munge.key")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("设置munge密钥权限失败: %v, 输出: %s", err, string(output))
	}
	
	cmd = exec.Command("chmod", "0600", "/etc/munge/munge.key")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("设置munge密钥权限失败: %v, 输出: %s", err, string(output))
	}
	
	// 启动munge服务
	s.addLog("启动Munge服务")
	cmd = exec.Command("systemctl", "enable", "munge")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启用munge服务失败: %v, 输出: %s", err, string(output)))
	}
	
	cmd = exec.Command("systemctl", "start", "munge")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启动munge服务失败: %v, 输出: %s", err, string(output))
	}
	
	return nil
}

// configureDatabase 配置数据库
func (s *SlurmDeployService) configureDatabase(password string) error {
	s.addLog("安装MariaDB")
	cmd := exec.Command("yum", "install", "-y", "mariadb-server", "mariadb-devel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装MariaDB失败: %v, 输出: %s", err, string(output))
	}
	
	s.addLog("启动MariaDB服务")
	cmd = exec.Command("systemctl", "enable", "mariadb")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启用mariadb服务失败: %v, 输出: %s", err, string(output)))
	}
	
	cmd = exec.Command("systemctl", "start", "mariadb")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启动mariadb服务失败: %v, 输出: %s", err, string(output))
	}
	
	// 创建slurm数据库和用户
	s.addLog("创建Slurm数据库和用户")
	dbCommands := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS slurm_acct_db;"),
		fmt.Sprintf("CREATE USER IF NOT EXISTS 'slurm'@'localhost' IDENTIFIED BY '%s';", password),
		fmt.Sprintf("GRANT ALL ON slurm_acct_db.* TO 'slurm'@'localhost';"),
		"FLUSH PRIVILEGES;",
	}
	
	for _, command := range dbCommands {
		cmd = exec.Command("mysql", "-e", command)
		if output, err := cmd.CombinedOutput(); err != nil {
			s.addLog(fmt.Sprintf("警告: 执行数据库命令失败: %s, 输出: %s", command, string(output)))
		}
	}
	
	return nil
}

// createSlurmConfig 创建基础Slurm配置文件
func (s *SlurmDeployService) createSlurmConfig() error {
	configContent := fmt.Sprintf(`# Slurm配置文件
ClusterName=%s
ControlMachine=%s
ControlAddr=%s
SlurmUser=%s
SlurmctldPort=%d
SlurmdPort=%d
AuthType=%s
StateSaveLocation=%s
SlurmctldPidFile=%s
SlurmdPidFile=%s
SwitchType=%s
MpiDefault=%s
ProctrackType=%s
ReturnToService=%d
MaxTime=%s

# 节点配置
NodeName=%s CPUs=4 State=UNKNOWN
PartitionName=debug Nodes=%s Default=YES MaxTime=INFINITE State=UP
`,
		s.config.ClusterName,
		s.config.ControlMachine,
		s.config.ControlAddr,
		s.config.SlurmUser,
		s.config.SlurmctldPort,
		s.config.SlurmdPort,
		s.config.AuthType,
		s.config.StateSaveLocation,
		s.config.SlurmctldPidFile,
		s.config.SlurmdPidFile,
		s.config.SwitchType,
		s.config.MpiDefault,
		s.config.ProctrackType,
		s.config.ReturnToService,
		s.config.MaxTime,
		s.config.ControlMachine,
		s.config.ControlMachine,
	)
	
	// 写入配置文件
	if err := os.WriteFile("/etc/slurm/slurm.conf", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("写入slurm.conf失败: %v", err)
	}
	
	// 创建cgroup配置
	cgroupContent := `# Cgroup配置
CgroupAutomount=yes
CgroupReleaseAgentDir="/etc/slurm/cgroup"
ConstrainCores=yes
ConstrainRAMSpace=yes
`
	
	if err := os.MkdirAll("/etc/slurm/cgroup", 0755); err != nil {
		s.addLog(fmt.Sprintf("警告: 创建cgroup目录失败: %v", err))
	}
	
	if err := os.WriteFile("/etc/slurm/cgroup.conf", []byte(cgroupContent), 0644); err != nil {
		s.addLog(fmt.Sprintf("警告: 写入cgroup.conf失败: %v", err))
	}
	
	return nil
}