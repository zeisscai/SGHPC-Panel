package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type NodeConfig struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
}

type DeploymentConfig struct {
	Nodes []NodeConfig `json:"nodes"`
}

type ScriptStatus struct {
	Running   bool   `json:"running"`
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
}

var currentStatus ScriptStatus

func main() {
	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/api/config", getConfig).Methods("GET")
	r.HandleFunc("/api/config", saveConfig).Methods("POST")
	r.HandleFunc("/api/deploy", startDeployment).Methods("POST")
	r.HandleFunc("/api/status", getStatus).Methods("GET")

	// Serve frontend files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	configFile := filepath.Join(".", "deploy.conf")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Return default config
		defaultConfig := DeploymentConfig{
			Nodes: []NodeConfig{
				{Name: "master", IP: "", Password: "", Hostname: ""},
				{Name: "node1", IP: "", Password: "", Hostname: ""},
				{Name: "node2", IP: "", Password: "", Hostname: ""},
				{Name: "node3", IP: "", Password: "", Hostname: ""},
				{Name: "node4", IP: "", Password: "", Hostname: ""},
			},
		}
		json.NewEncoder(w).Encode(defaultConfig)
		return
	}

	// Parse existing config file
	config := parseConfigFile()
	json.NewEncoder(w).Encode(config)
}

func saveConfig(w http.ResponseWriter, r *http.Request) {
	var config DeploymentConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate deploy.conf file
	configContent := ""
	for _, node := range config.Nodes {
		if node.IP != "" {
			configContent += fmt.Sprintf("[%s]\n", node.Name)
			configContent += fmt.Sprintf("ip = %s\n", node.IP)
			configContent += fmt.Sprintf("password = %s\n", node.Password)
			configContent += fmt.Sprintf("hostname = %s\n\n", node.Hostname)
		}
	}

	err := os.WriteFile(filepath.Join(".", "deploy.conf"), []byte(configContent), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func startDeployment(w http.ResponseWriter, r *http.Request) {
	go runDeployment()
	
	currentStatus = ScriptStatus{
		Running:   true,
		Message:   "Deployment started",
		Completed: false,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func runDeployment() {
	defer func() {
		currentStatus.Running = false
		currentStatus.Completed = true
	}()

	// Load configuration
	config := parseConfigFile()
	if len(config.Nodes) == 0 {
		currentStatus.Message = "No nodes configured"
		return
	}

	// Update status
	currentStatus.Message = "Running deployment process..."

	// Execute the deployment steps directly
	if err := deploySlurm(config); err != nil {
		currentStatus.Message = fmt.Sprintf("Deployment failed: %v", err)
		return
	}
	
	currentStatus.Message = "Deployment completed successfully"
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(currentStatus)
}

// Helper functions to replace the shell script functionality

func parseConfigFile() DeploymentConfig {
	configFile := filepath.Join(".", "deploy.conf")
	content, err := os.ReadFile(configFile)
	if err != nil {
		return DeploymentConfig{
			Nodes: []NodeConfig{
				{Name: "master", IP: "", Password: "", Hostname: ""},
				{Name: "node1", IP: "", Password: "", Hostname: ""},
				{Name: "node2", IP: "", Password: "", Hostname: ""},
				{Name: "node3", IP: "", Password: "", Hostname: ""},
				{Name: "node4", IP: "", Password: "", Hostname: ""},
			},
		}
	}

	config := DeploymentConfig{}
	lines := strings.Split(string(content), "\n")
	currentNode := NodeConfig{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// New node section
			if currentNode.Name != "" {
				config.Nodes = append(config.Nodes, currentNode)
			}
			nodeName := strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			currentNode = NodeConfig{Name: nodeName}
		} else if eqIdx := strings.Index(line, "="); eqIdx > 0 {
			key := strings.TrimSpace(line[:eqIdx])
			value := strings.TrimSpace(line[eqIdx+1:])
			
			switch key {
			case "ip":
				currentNode.IP = value
			case "password":
				currentNode.Password = value
			case "hostname":
				currentNode.Hostname = value
			}
		}
	}

	// Add the last node
	if currentNode.Name != "" {
		config.Nodes = append(config.Nodes, currentNode)
	}

	return config
}

func deploySlurm(config DeploymentConfig) error {
	// Filter out unconfigured nodes
	var nodes []NodeConfig
	for _, node := range config.Nodes {
		if node.IP != "" {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes configured")
	}

	// Step 1: Setup SSH connections
	currentStatus.Message = "Setting up SSH connections..."
	if err := setupSSH(nodes); err != nil {
		return fmt.Errorf("SSH setup failed: %v", err)
	}

	// Step 2: Configure hosts files
	currentStatus.Message = "Configuring hosts files..."
	if err := configureHosts(nodes); err != nil {
		return fmt.Errorf("hosts configuration failed: %v", err)
	}

	// Step 3: Setup hostnames
	currentStatus.Message = "Setting up hostnames..."
	if err := setupHostnames(nodes); err != nil {
		return fmt.Errorf("hostname setup failed: %v", err)
	}

	// Step 4: Install base packages
	currentStatus.Message = "Installing base packages..."
	if err := installBasePackages(nodes); err != nil {
		return fmt.Errorf("base package installation failed: %v", err)
	}

	// Step 5: Setup Munge
	currentStatus.Message = "Setting up Munge authentication..."
	if err := setupMunge(nodes); err != nil {
		return fmt.Errorf("Munge setup failed: %v", err)
	}

	// Step 6: Setup MariaDB (on master only)
	currentStatus.Message = "Setting up MariaDB..."
	if err := setupMariaDB(nodes[0]); err != nil {
		return fmt.Errorf("MariaDB setup failed: %v", err)
	}

	// Step 7: Install Slurm packages
	currentStatus.Message = "Installing Slurm packages..."
	if err := installSlurmPackages(nodes); err != nil {
		return fmt.Errorf("Slurm package installation failed: %v", err)
	}

	// Step 8: Configure Slurm
	currentStatus.Message = "Configuring Slurm..."
	if err := configureSlurm(nodes); err != nil {
		return fmt.Errorf("Slurm configuration failed: %v", err)
	}

	// Step 9: Start Slurm services
	currentStatus.Message = "Starting Slurm services..."
	if err := startSlurmServices(nodes); err != nil {
		return fmt.Errorf("Slurm service start failed: %v", err)
	}

	// Step 10: Verify installation
	currentStatus.Message = "Verifying installation..."
	if err := verifyInstallation(nodes[0]); err != nil {
		return fmt.Errorf("installation verification failed: %v", err)
	}

	return nil
}

func setupSSH(nodes []NodeConfig) error {
	// Generate SSH key pair if not exists
	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	if _, err := os.Stat(filepath.Join(sshDir, "id_rsa")); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", filepath.Join(sshDir, "id_rsa"))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to generate SSH key: %v", err)
		}
	}

	// Configure passwordless SSH for each node
	for _, node := range nodes {
		if node.IP == "" || node.Password == "" {
			continue
		}

		cmd := exec.Command("sshpass", "-p", node.Password, "ssh-copy-id", "-o", "StrictHostKeyChecking=no", "root@"+node.IP)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to setup SSH for %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func configureHosts(nodes []NodeConfig) error {
	// Build hosts entries
	var hostsEntries strings.Builder
	for _, node := range nodes {
		if node.IP != "" && node.Hostname != "" {
			hostsEntries.WriteString(fmt.Sprintf("%s %s\n", node.IP, node.Hostname))
		}
	}

	// Update hosts file on each node
	hostsContent := hostsEntries.String()
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		// Create script to update hosts file
		script := fmt.Sprintf(`
cp /etc/hosts /etc/hosts.backup
sed -i '/# Added by Slurm installation script/d' /etc/hosts
echo -e '%s# Added by Slurm installation script' >> /etc/hosts
`, hostsContent)

		cmd := exec.Command("ssh", "root@"+node.IP, script)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure hosts on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func setupHostnames(nodes []NodeConfig) error {
	for _, node := range nodes {
		if node.IP == "" || node.Hostname == "" {
			continue
		}

		cmd := exec.Command("ssh", "root@"+node.IP, fmt.Sprintf("hostnamectl set-hostname %s", node.Hostname))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set hostname on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func installBasePackages(nodes []NodeConfig) error {
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		script := `
if ! dnf list installed epel-release &>/dev/null; then
    rpm --import https://dl.fedoraproject.org/pub/epel/RPM-GPG-KEY-EPEL-9
    dnf install -y https://mirrors.ustc.edu.cn/epel/epel-release-latest-9.noarch.rpm
else 
    echo 'epel-release已安装'
fi

dnf config-manager --set-enabled crb
packages=(
"libjwt"
"libjwt-devel"
)
for pkg in "${packages[@]}"; do
    if ! rpm -q "$pkg" &> /dev/null; then
        echo "正在安装 $pkg..."
        dnf install -y "$pkg"
    else
        echo "$pkg 已安装"
    fi
done
echo '基础依赖包安装完成'
`

		cmd := exec.Command("ssh", "root@"+node.IP, script)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install base packages on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func setupMunge(nodes []NodeConfig) error {
	// Install munge on all nodes
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		script := `
if ! dnf list installed munge &>/dev/null; then
    dnf install -y munge munge-libs munge-devel
    echo 'munge 已安装'
else
    echo 'munge 已安装，跳过'
fi

# 确保munge用户存在
if ! id munge &>/dev/null; then
    useradd -r -s /usr/sbin/nologin munge
    echo '已创建munge用户'
fi
`
		cmd := exec.Command("ssh", "root@"+node.IP, script)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install munge on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	// Generate munge key on master
	master := nodes[0]
	if master.IP == "" {
		return fmt.Errorf("master node not configured")
	}

	cmd := exec.Command("ssh", "root@"+master.IP, `
if [ ! -f /etc/munge/munge.key ]; then
    /usr/sbin/create-munge-key -r
    echo '已生成Munge密钥'
else
    echo 'Munge密钥已存在，跳过'
fi
`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate munge key on master: %v", err)
	}

	// Distribute munge key to all nodes
	cmd = exec.Command("scp", "root@"+master.IP+":/etc/munge/munge.key", "/tmp/munge.key")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy munge key from master: %v", err)
	}

	for _, node := range nodes {
		if node.IP == "" || node.Name == "master" {
			continue
		}

		cmd := exec.Command("scp", "/tmp/munge.key", "root@"+node.IP+":/etc/munge/munge.key")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy munge key to %s (%s): %v", node.Name, node.IP, err)
		}

		cmd = exec.Command("ssh", "root@"+node.IP, "chown munge:munge /etc/munge/munge.key && chmod 400 /etc/munge/munge.key")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set munge key permissions on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	// Set permissions on master
	cmd = exec.Command("ssh", "root@"+master.IP, "chown munge:munge /etc/munge/munge.key && chmod 400 /etc/munge/munge.key")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set munge key permissions on master: %v", err)
	}

	// Start munge service on all nodes
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		cmd := exec.Command("ssh", "root@"+node.IP, `
# 确保munge用户存在后再启动服务
if ! id munge &>/dev/null; then
    useradd -r -s /usr/sbin/nologin munge
fi

systemctl enable munge
systemctl start munge
echo 'Munge服务已启动'
`)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start munge on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func setupMariaDB(master NodeConfig) error {
	if master.IP == "" {
		return fmt.Errorf("master node not configured")
	}

	// Install and start MariaDB
	cmd := exec.Command("ssh", "root@"+master.IP, `
systemctl enable mariadb
systemctl start mariadb
`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start MariaDB: %v", err)
	}

	// Wait for MariaDB to start
	time.Sleep(5 * time.Second)

	// Setup MariaDB security
	// In a real implementation, we would generate random passwords here
	mysqlRootPass := "root_password"
	slurmDbPass := "slurm_password"

	script := fmt.Sprintf(`
mysql -u root << EOF
-- 设置root密码
ALTER USER 'root'@'localhost' IDENTIFIED BY '%s';
-- 删除匿名用户
DELETE FROM mysql.user WHERE User='';
-- 禁止root远程登录
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
-- 删除test数据库
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\\\\_%%';
-- 重新加载权限表
FLUSH PRIVILEGES;
EOF
`, mysqlRootPass)

	cmd = exec.Command("ssh", "root@"+master.IP, script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to secure MariaDB: %v", err)
	}

	// Create Slurm database and user
	script = fmt.Sprintf(`
mysql -u root -p"%s" << EOF
CREATE DATABASE IF NOT EXISTS slurm_acct_db;
CREATE USER IF NOT EXISTS 'slurm'@'localhost' IDENTIFIED BY '%s';
GRANT ALL PRIVILEGES ON slurm_acct_db.* TO 'slurm'@'localhost';
FLUSH PRIVILEGES;
EOF
`, mysqlRootPass, slurmDbPass)

	cmd = exec.Command("ssh", "root@"+master.IP, script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Slurm database: %v", err)
	}

	return nil
}

func installSlurmPackages(nodes []NodeConfig) error {
	// Install OpenHPC repo on all nodes
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		cmd := exec.Command("ssh", "root@"+node.IP, `
if [ -f "/etc/yum.repos.d/OpenHPC.repo" ]; then
    echo 'OpenHPC仓库已配置，跳过'
    exit 0
fi

# 安装OpenHPC仓库
dnf install -y http://repos.openhpc.community/OpenHPC/3/EL_9/x86_64/ohpc-release-3-1.el9.x86_64.rpm
echo 'OpenHPC仓库配置完成'
`)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install OpenHPC repo on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	// Install Slurm packages
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		var script string
		if node.Name == "master" {
			script = `
# 先尝试安装依赖
dnf install -y libjwt || true

dnf install -y ohpc-slurm-server slurm-ohpc slurm-devel-ohpc slurm-example-configs-ohpc slurm-slurmctld-ohpc slurm-slurmdbd-ohpc slurm-slurmd-ohpc mariadb-server mariadb
echo '主节点Slurm包安装完成'
`
		} else {
			script = `
# 先尝试安装依赖
dnf install -y libjwt || true

dnf install -y ohpc-slurm-client slurm-ohpc slurm-slurmd-ohpc
echo '计算节点Slurm包安装完成'
`
		}

		cmd := exec.Command("ssh", "root@"+node.IP, script)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install Slurm packages on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func configureSlurm(nodes []NodeConfig) error {
	// Create Slurm user and directories on all nodes
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		script := `
# 创建Slurm用户
if ! id slurm &>/dev/null; then
    useradd -r -s /bin/false -d /var/lib/slurm slurm
    echo '已创建slurm用户'
else
    echo 'slurm用户已存在，跳过'
fi

# 创建必要目录
mkdir -p /var/spool/slurm/ctld
mkdir -p /var/spool/slurm/d
mkdir -p /var/log/slurm
mkdir -p /etc/slurm  # 添加这行确保配置目录存在

chown slurm:slurm /var/spool/slurm/ctld
chown slurm:slurm /var/spool/slurm/d
chown slurm:slurm /var/log/slurm
`
		cmd := exec.Command("ssh", "root@"+node.IP, script)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create Slurm user/dirs on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	// Get hardware info from nodes
	nodeInfo := make(map[string]map[string]string)
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		// Get CPU count
		cmd := exec.Command("ssh", "root@"+node.IP, "nproc")
		cpuOut, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get CPU count for %s (%s): %v", node.Name, node.IP, err)
		}
		cpuCount := strings.TrimSpace(string(cpuOut))

		// Get memory
		cmd = exec.Command("ssh", "root@"+node.IP, "free -m | grep '^Mem:' | awk '{print $2}'")
		memOut, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get memory for %s (%s): %v", node.Name, node.IP, err)
		}
		memMB := strings.TrimSpace(string(memOut))

		nodeInfo[node.Name] = map[string]string{
			"cpus":   cpuCount,
			"memory": memMB,
		}
	}

	// Generate slurm.conf on master
	master := nodes[0]
	if master.IP == "" {
		return fmt.Errorf("master node not configured")
	}

	// Build node configurations
	var nodeConfigs strings.Builder
	var partitionNodes strings.Builder

	for i, node := range nodes {
		if node.IP == "" {
			continue
		}

		cpus := nodeInfo[node.Name]["cpus"]
		memory := nodeInfo[node.Name]["memory"]
		
		// Reserve 90% of memory for Slurm
		slurmMemory := 0
		fmt.Sscanf(memory, "%d", &slurmMemory)
		slurmMemory = slurmMemory * 90 / 100

		nodeConfigs.WriteString(fmt.Sprintf("NodeName=%s CPUs=%s RealMemory=%d State=UNKNOWN\n", node.Hostname, cpus, slurmMemory))
		
		if i > 0 {
			partitionNodes.WriteString(",")
		}
		partitionNodes.WriteString(node.Hostname)
	}

	// Create slurm.conf
	slurmConf := fmt.Sprintf(`
# slurm.conf file generated by automated script
ClusterName=cluster
ControlMachine=%s
ControlAddr=%s

SlurmUser=slurm
SlurmdUser=root
SlurmctldPort=6817
SlurmdPort=6818
AuthType=auth/munge
StateSaveLocation=/var/spool/slurm/ctld
SlurmdSpoolDir=/var/spool/slurm/d
SwitchType=switch/none
MpiDefault=none
SlurmctldPidFile=/var/run/slurm/slurmctld.pid
SlurmdPidFile=/var/run/slurm/slurmd.pid
ProctrackType=proctrack/pgid
ReturnToService=1
SlurmctldTimeout=120
SlurmdTimeout=300
InactiveLimit=0
MinJobAge=300
KillWait=30
MaxJobCount=10000
Waittime=0

# SCHEDULING
SchedulerType=sched/backfill
SelectType=select/cons_tres
SelectTypeParameters=CR_Core

# LOGGING AND ACCOUNTING
AccountingStorageType=accounting_storage/slurmdbd
AccountingStoreFlags=job_comment
JobCompType=jobcomp/none
JobAcctGatherFrequency=30
JobAcctGatherType=jobacct_gather/linux
SlurmctldDebug=info
SlurmctldLogFile=/var/log/slurm/slurmctld.log
SlurmdDebug=info
SlurmdLogFile=/var/log/slurm/slurmd.log

# NODES
%s

# PARTITIONS
PartitionName=compute Nodes=%s Default=YES MaxTime=INFINITE State=UP
`, master.Hostname, master.Hostname, nodeConfigs.String(), partitionNodes.String())

	cmd := exec.Command("ssh", "root@"+master.IP, fmt.Sprintf("cat > /etc/slurm/slurm.conf << 'EOF'\n%s\nEOF", slurmConf))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create slurm.conf: %v", err)
	}

	// Create slurmdbd.conf on master
	// In a real implementation, we would use the actual generated passwords
	slurmDbPass := "slurm_password"
	slurmdbdConf := fmt.Sprintf(`
AuthType=auth/munge
AuthInfo=/var/run/munge/munge.socket.2
DbdAddr=localhost
DbdHost=localhost
SlurmUser=slurm
DebugLevel=verbose
LogFile=/var/log/slurm/slurmdbd.log
PidFile=/var/run/slurm/slurmdbd.pid
StorageType=accounting_storage/mysql
StorageHost=localhost
StoragePass=%s
StorageUser=slurm
StorageLoc=slurm_acct_db
`, slurmDbPass)

	cmd = exec.Command("ssh", "root@"+master.IP, fmt.Sprintf("cat > /etc/slurm/slurmdbd.conf << 'EOF'\n%s\nEOF", slurmdbdConf))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create slurmdbd.conf: %v", err)
	}

	cmd = exec.Command("ssh", "root@"+master.IP, "chmod 600 /etc/slurm/slurmdbd.conf && chown slurm:slurm /etc/slurm/slurmdbd.conf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set slurmdbd.conf permissions: %v", err)
	}

	// Distribute slurm.conf to compute nodes
	for _, node := range nodes[1:] { // Skip master
		if node.IP == "" {
			continue
		}

		cmd := exec.Command("ssh", "root@"+node.IP, "mkdir -p /etc/slurm")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create slurm dir on %s (%s): %v", node.Name, node.IP, err)
		}

		cmd = exec.Command("scp", "root@"+master.IP+":/etc/slurm/slurm.conf", "root@"+node.IP+":/etc/slurm/slurm.conf")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy slurm.conf to %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func startSlurmServices(nodes []NodeConfig) error {
	// Start services on master
	master := nodes[0]
	if master.IP == "" {
		return fmt.Errorf("master node not configured")
	}

	cmd := exec.Command("ssh", "root@"+master.IP, `
systemctl enable slurmdbd
systemctl enable slurmctld

systemctl start slurmdbd
sleep 5
systemctl start slurmctld

echo '主节点Slurm服务已启动'
`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Slurm services on master: %v", err)
	}

	// Start slurmd on all nodes
	for _, node := range nodes {
		if node.IP == "" {
			continue
		}

		cmd := exec.Command("ssh", "root@"+node.IP, `
systemctl enable slurmd
systemctl start slurmd
echo '节点Slurm服务已启动'
`)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start slurmd on %s (%s): %v", node.Name, node.IP, err)
		}
	}

	return nil
}

func verifyInstallation(master NodeConfig) error {
	if master.IP == "" {
		return fmt.Errorf("master node not configured")
	}

	// Check if Slurm commands exist
	cmd := exec.Command("ssh", "root@"+master.IP, `
if ! command -v sinfo &> /dev/null; then
    # 尝试安装slurm包来修复命令
    dnf install -y slurm-ohpc --skip-broken || true
fi

if ! command -v sinfo &> /dev/null; then
    echo '错误: sinfo命令未找到'
    exit 1
fi

if ! command -v sbatch &> /dev/null; then
    echo '错误: sbatch命令未找到'
    exit 1
fi

if ! command -v squeue &> /dev/null; then
    echo '错误: squeue命令未找到'
    exit 1
fi

if ! command -v scontrol &> /dev/null; then
    echo '错误: scontrol命令未找到'
    exit 1
fi
`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Slurm commands not found: %v", err)
	}

	// Check service status
	cmd = exec.Command("ssh", "root@"+master.IP, `
if ! systemctl is-active --quiet slurmctld-ohpc; then
    echo '错误: slurmctld服务未运行'
    exit 1
fi

if ! systemctl is-active --quiet slurmdbd-ohpc; then
    echo '错误: slurmdbd服务未运行'
    exit 1
fi
`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Slurm services not running: %v", err)
	}

	// Wait for services to fully start
	time.Sleep(15 * time.Second)

	// Test job submission
	cmd = exec.Command("ssh", "root@"+master.IP, `
echo '#!/bin/bash' > test_job.sh
echo '#SBATCH --job-name=test' >> test_job.sh
echo '#SBATCH --output=test.out' >> test_job.sh
echo '#SBATCH --error=test.err' >> test_job.sh
echo '#SBATCH --ntasks=1' >> test_job.sh
echo 'srun hostname' >> test_job.sh

job_id=$(sbatch test_job.sh 2>&1 | grep -o '[0-9]*')

if [[ -z "$job_id" ]]; then
    echo '错误: 作业提交失败'
    exit 1
fi

echo '作业提交成功，作业ID: '$job_id

# 等待作业完成
sleep 10

# 检查作业状态
job_state=$(squeue -j $job_id -h -o '%%T' 2>/dev/null)
if [[ -n "$job_state" ]]; then
    echo '作业仍在队列中，状态: '$job_state
else
    echo '作业已完成或不存在于队列中'
fi
`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("job submission test failed: %v", err)
	}

	return nil
}