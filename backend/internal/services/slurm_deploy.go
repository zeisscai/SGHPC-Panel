package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	
	"panel-tool/internal/models"
	"panel-tool/internal/utils"
)

// SlurmDeployService Slurm部署服务结构体
type SlurmDeployService struct {
	config         *models.SlurmConfig
	status         *models.SlurmDeploymentStatus
	logs           []string
	configFilePath string
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
	return s.logs
}

// addLog 添加日志
func (s *SlurmDeployService) addLog(message string) {
	logEntry := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
	s.logs = append(s.logs, logEntry)
	
	// 使用Logger记录日志
	logger := utils.NewLogger()
	logger.Info(logEntry)
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

	// 检查是否为OpenEuler 24
	if _, err := os.Stat("/etc/openeuler-release"); os.IsNotExist(err) {
		return fmt.Errorf("当前系统不是OpenEuler发行版")
	}

	data, err := os.ReadFile("/etc/openeuler-release")
	if err != nil {
		return fmt.Errorf("无法读取系统版本信息: %v", err)
	}

	version := strings.TrimSpace(string(data))
	s.addLog(fmt.Sprintf("检测到系统版本: %s", version))

	if !strings.Contains(version, "24.00") {
		return fmt.Errorf("当前系统版本不是OpenEuler 24，实际版本: %s", version)
	}

	s.addLog("操作系统版本检查通过")

	// 检查必要的命令是否存在
	commands := []string{"yum", "make", "gcc", "rpmbuild"}
	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("缺少必要命令: %s", cmd)
		}
	}

	s.addLog("必要命令检查通过")
	return nil
}

// downloadSource 下载Slurm源码
func (s *SlurmDeployService) downloadSource() error {
	// 这里我们使用一个固定的版本，实际项目中可能需要配置
	slurmVersion := "23.11.4"
	downloadURL := fmt.Sprintf("https://download.schedmd.com/slurm/slurm-%s.tar.bz2", slurmVersion)
	sourceDir := "/tmp/slurm-source"

	// 清理旧的源码目录
	if err := os.RemoveAll(sourceDir); err != nil {
		s.addLog(fmt.Sprintf("警告: 无法清理旧源码目录: %v", err))
	}

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

	s.addLog("Slurm编译安装完成")
	return nil
}

// configureService 配置Slurm服务
func (s *SlurmDeployService) configureService() error {
	// 生成基础配置文件
	s.addLog("正在生成slurm.conf配置文件")
	slurmConf := fmt.Sprintf(`
# Slurm configuration generated by SGHPC-Panel
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

# 计算节点配置
%s
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
		s.generateNodeConfig(),
	)

	// 写入配置文件
	confPath := "/etc/slurm/slurm.conf"
	if err := os.MkdirAll("/etc/slurm", 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	if err := os.WriteFile(confPath, []byte(slurmConf), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 设置配置文件权限
	os.Chown(confPath, 930, 930)
	s.addLog("slurm.conf配置文件生成完成")

	// 启用并启动munge服务
	s.addLog("正在启用并启动munge服务")
	cmd := exec.Command("systemctl", "enable", "munge")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启用munge服务失败: %v, 输出: %s", err, string(output)))
	}

	cmd = exec.Command("systemctl", "start", "munge")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启动munge服务失败: %v, 输出: %s", err, string(output))
	}

	// 启用并启动slurmctld服务
	s.addLog("正在启用并启动slurmctld服务")
	cmd = exec.Command("systemctl", "enable", "slurmctld")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启用slurmctld服务失败: %v, 输出: %s", err, string(output)))
	}

	cmd = exec.Command("systemctl", "start", "slurmctld")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.addLog(fmt.Sprintf("警告: 启动slurmctld服务失败: %v, 输出: %s", err, string(output)))
	}

	s.addLog("Slurm服务配置完成")
	return nil
}

// generateNodeConfig 生成计算节点配置
func (s *SlurmDeployService) generateNodeConfig() string {
	var nodesConfig strings.Builder
	for _, node := range s.config.ComputeNodes {
		nodesConfig.WriteString(fmt.Sprintf("NodeName=%s NodeAddr=%s CPUs=%d State=UNKNOWN\n",
			node, node, 8)) // 这里使用默认的CPU数量，实际应该动态获取
	}

	// 添加分区配置
	nodesConfig.WriteString(fmt.Sprintf("\nPartitionName=debug Nodes=%s Default=YES MaxTime=INFINITE State=UP\n",
		strings.Join(s.config.ComputeNodes, ",")))

	return nodesConfig.String()
}

// ControlSlurmService 控制Slurm服务
func ControlSlurmService(action string) (string, error) {
	var cmd *exec.Cmd
	
	switch action {
	case "start":
		cmd = exec.Command("systemctl", "start", "slurmctld")
	case "stop":
		cmd = exec.Command("systemctl", "stop", "slurmctld")
	case "restart":
		cmd = exec.Command("systemctl", "restart", "slurmctld")
	default:
		return "", fmt.Errorf("无效的操作: %s", action)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v, 输出: %s", err, string(output))
	}
	
	return string(output), nil
}