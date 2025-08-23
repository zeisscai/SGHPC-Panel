package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// UserInfo 用户信息结构体
type UserInfo struct {
	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Group    string `json:"group"`
	HomeDir  string `json:"homeDir"`
	Shell    string `json:"shell"`
}

// CreateUserRequest 创建用户请求结构体
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
}

// ChangePasswordRequest 修改密码请求结构体
type ChangePasswordRequest struct {
	Password string `json:"password"`
}

// GenerateKeyRequest 生成SSH密钥请求结构体
type GenerateKeyRequest struct {
	Type    string `json:"type"`
	Comment string `json:"comment"`
}

// HandleGetUsers 获取系统用户列表
func HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("[USER API] Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	
	if r.Method != http.MethodGet {
		log.Printf("[USER API] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	log.Printf("[USER API] Getting system users...")
	users, err := getSystemUsers()
	if err != nil {
		log.Printf("[USER API] Failed to get users: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get users: %v", err), http.StatusInternalServerError)
		return
	}
	
	log.Printf("[USER API] Successfully retrieved %d users", len(users))
	response := map[string]interface{}{
		"users": users,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[USER API] Failed to encode response: %v", err)
	}
}

// HandleCreateUser 创建新用户
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	
	// 验证用户名格式
	if !isValidUsername(req.Username) {
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}
	
	err := createSystemUser(req.Username, req.Password, req.Group)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "User created successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteUser 删除用户
func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从URL路径中提取用户名
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	username := strings.Split(path, "/")[0]
	
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	
	// 防止删除系统关键用户 - 先获取用户信息
	log.Printf("[USER API] Checking if user %s can be deleted...", username)
	userInfo, err := user.Lookup(username)
	if err != nil {
		log.Printf("[USER API] User %s not found: %v", username, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		log.Printf("[USER API] Failed to parse UID for user %s: %v", username, err)
		http.Error(w, "Invalid user UID", http.StatusInternalServerError)
		return
	}
	
	if isSystemUser(username, uid) {
		log.Printf("[USER API] Cannot delete system user %s (UID: %d)", username, uid)
		http.Error(w, "Cannot delete system user", http.StatusForbidden)
		return
	}
	
	err = deleteSystemUser(username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleChangeUserPassword 修改用户密码
func HandleChangeUserPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从URL路径中提取用户名
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "password" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	username := parts[0]
	
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}
	
	err := changeUserPassword(username, req.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to change password: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleGenerateSSHKey 为用户生成SSH密钥
func HandleGenerateSSHKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从URL路径中提取用户名
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "generate-key" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	username := parts[0]
	
	var req GenerateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Type == "" {
		req.Type = "rsa"
	}
	
	err := generateSSHKey(username, req.Type, req.Comment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate SSH key: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "SSH key generated successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// getSystemUsers 获取系统用户列表 (Rocky Linux 9.6)
func getSystemUsers() ([]UserInfo, error) {
	log.Printf("[USER API] Getting system users on Rocky Linux")
	return getLinuxUsers()
}

// getLinuxUsers 获取Rocky Linux用户
func getLinuxUsers() ([]UserInfo, error) {
	log.Printf("[USER API] Executing getent passwd command on Rocky Linux...")
	cmd := exec.Command("getent", "passwd")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[USER API] getent passwd failed: %v, trying to read /etc/passwd", err)
		// 如果getent失败，尝试直接读取/etc/passwd
		output, err = os.ReadFile("/etc/passwd")
		if err != nil {
			log.Printf("[USER API] Failed to read /etc/passwd: %v", err)
			return nil, fmt.Errorf("failed to get user list on Rocky Linux: %v", err)
		}
		log.Printf("[USER API] Successfully read /etc/passwd as fallback")
	}
	
	log.Printf("[USER API] Parsing passwd output...")
	users, err := parsePasswdOutput(output)
	if err != nil {
		return nil, err
	}
	log.Printf("[USER API] Found %d non-system users on Rocky Linux", len(users))
	return users, nil
}







// parsePasswdOutput 解析passwd输出
func parsePasswdOutput(output []byte) ([]UserInfo, error) {
	log.Printf("[USER API] Successfully got passwd data, parsing...")
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var users []UserInfo
	log.Printf("[USER API] Found %d total entries in passwd", len(lines))
	
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			log.Printf("[USER API] Skipping malformed line %d: %s", i+1, line)
			continue
		}
		
		username := parts[0]
		uid, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("[USER API] Invalid UID for user %s: %s", username, parts[2])
			continue
		}
		
		gid, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Printf("[USER API] Invalid GID for user %s: %s", username, parts[3])
			continue
		}
		
		homeDir := parts[5]
		shell := parts[6]
		
		// 过滤系统用户（UID < 1000）和特殊用户
		if uid < 1000 {
			log.Printf("[USER API] Skipping system user %s (UID: %d)", username, uid)
			continue
		}
		
		if isSystemUser(username, uid) {
			log.Printf("[USER API] Skipping known system user: %s (UID: %d)", username, uid)
			continue
		}
		
		// 获取用户组名
		groupName, err := getGroupName(gid)
		if err != nil {
			log.Printf("[USER API] Warning: Failed to get group name for GID %d: %v", gid, err)
			groupName = fmt.Sprintf("gid_%d", gid) // 使用GID作为备用
		}
		
		log.Printf("[USER API] Adding user: %s (UID: %d, GID: %d, Group: %s)", username, uid, gid, groupName)
		users = append(users, UserInfo{
			Username: username,
			UID:      uid,
			GID:      gid,
			Group:    groupName,
			HomeDir:  homeDir,
			Shell:    shell,
		})
	}
	
	log.Printf("[USER API] Filtered to %d non-system users", len(users))
	return users, nil
}

// createSystemUser 创建系统用户 (Rocky Linux 9.6)
func createSystemUser(username, password, group string) error {
	log.Printf("[USER API] Creating user: %s with group: %s on Rocky Linux", username, group)
	
	// 检查用户是否已存在
	if _, err := user.Lookup(username); err == nil {
		log.Printf("[USER API] User %s already exists", username)
		return fmt.Errorf("user %s already exists", username)
	}
	
	return createLinuxUser(username, password, group)
}

// createLinuxUser 创建Linux用户
func createLinuxUser(username, password, group string) error {
	log.Printf("[USER API] Creating Linux user: %s", username)
	
	// 构建useradd命令参数
	args := []string{"-m", "-s", "/bin/bash"}
	
	// 处理组参数 - 适配Rocky Linux 9.6
	if group != "" {
		// 检查组是否存在
		log.Printf("[USER API] Checking if group %s exists...", group)
		cmd := exec.Command("getent", "group", group)
		if err := cmd.Run(); err != nil {
			log.Printf("[USER API] Group %s does not exist, creating it...", group)
			// 创建组
			cmd = exec.Command("groupadd", group)
			if output, err := cmd.CombinedOutput(); err != nil {
				log.Printf("[USER API] Failed to create group %s: %v, output: %s", group, err, string(output))
				return fmt.Errorf("failed to create group %s: %v", group, err)
			}
			log.Printf("[USER API] Successfully created group: %s", group)
		}
		args = append(args, "-g", group)
	}
	
	args = append(args, username)
	log.Printf("[USER API] Executing useradd with args: %v", args)
	
	cmd := exec.Command("useradd", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[USER API] Failed to create user %s: %v, output: %s", username, err, string(output))
		return fmt.Errorf("failed to create user %s: %v - %s", username, err, string(output))
	}
	
	log.Printf("[USER API] Successfully created Linux user: %s", username)
	return changeUserPassword(username, password)
}



// deleteSystemUser 删除系统用户
func deleteSystemUser(username string) error {
	log.Printf("[USER API] Deleting user: %s", username)
	
	// 检查用户是否存在并获取用户信息
	userInfo, err := user.Lookup(username)
	if err != nil {
		log.Printf("[USER API] User %s does not exist: %v", username, err)
		return fmt.Errorf("user %s does not exist", username)
	}
	
	// 获取UID用于系统用户检查
	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		log.Printf("[USER API] Failed to parse UID for user %s: %v", username, err)
		return fmt.Errorf("invalid user UID for %s", username)
	}
	
	// 检查是否为系统用户
	if isSystemUser(username, uid) {
		log.Printf("[USER API] Attempted to delete system user: %s (UID: %d)", username, uid)
		return fmt.Errorf("cannot delete system user: %s", username)
	}
	
	cmd := exec.Command("userdel", "-r", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[USER API] Failed to delete user %s: %v, output: %s", username, err, string(output))
		return fmt.Errorf("failed to delete user %s: %v - %s", username, err, string(output))
	}
	
	log.Printf("[USER API] Successfully deleted user: %s", username)
	return nil
}

// changeUserPassword 修改用户密码
func changeUserPassword(username, password string) error {
	log.Printf("[USER API] Changing password for user: %s", username)
	
	// 检查用户是否存在
	if _, err := user.Lookup(username); err != nil {
		log.Printf("[USER API] User %s does not exist", username)
		return fmt.Errorf("user %s does not exist", username)
	}
	
	// 使用chpasswd命令设置密码
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[USER API] Failed to change password for user %s: %v, output: %s", username, err, string(output))
		return fmt.Errorf("failed to change password for user %s: %v - %s", username, err, string(output))
	}
	
	log.Printf("[USER API] Successfully changed password for user: %s", username)
	return nil
}

// generateSSHKey 为用户生成SSH密钥
func generateSSHKey(username, keyType, comment string) error {
	// 获取用户信息
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}
	
	sshDir := fmt.Sprintf("%s/.ssh", u.HomeDir)
	keyPath := fmt.Sprintf("%s/id_%s", sshDir, keyType)
	
	// 创建.ssh目录
	cmd := exec.Command("mkdir", "-p", sshDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %v", err)
	}
	
	// 生成密钥
	args := []string{"-t", keyType, "-f", keyPath, "-N", ""}
	if comment != "" {
		args = append(args, "-C", comment)
	}
	
	cmd = exec.Command("ssh-keygen", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate SSH key: %v", err)
	}
	
	// 设置正确的权限和所有者
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	
	// 设置.ssh目录权限
	syscall.Chmod(sshDir, 0700)
	syscall.Chown(sshDir, uid, gid)
	
	// 设置密钥文件权限
	syscall.Chmod(keyPath, 0600)
	syscall.Chown(keyPath, uid, gid)
	syscall.Chmod(keyPath+".pub", 0644)
	syscall.Chown(keyPath+".pub", uid, gid)
	
	return nil
}

// getGroupName 根据GID获取组名
func getGroupName(gid int) (string, error) {
	cmd := exec.Command("getent", "group", strconv.Itoa(gid))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	parts := strings.Split(strings.TrimSpace(string(output)), ":")
	if len(parts) > 0 {
		return parts[0], nil
	}
	
	return "", fmt.Errorf("group not found")
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	// 用户名只能包含字母、数字、下划线和连字符，且以字母开头
	matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_-]*$", username)
	return matched && len(username) <= 32
}

// isSystemUser 检查是否为Rocky Linux 9.6系统用户（增强错误处理）
func isSystemUser(username string, uid int) bool {
	log.Printf("[USER API] Checking if user %s (UID: %d) is a Rocky Linux 9.6 system user", username, uid)
	
	// Rocky Linux 9.6中，UID < 1000的用户通常是系统用户
	if uid < 1000 {
		log.Printf("[USER API] User %s is a system user (UID %d < 1000) on Rocky Linux 9.6", username, uid)
		return true
	}
	
	// Rocky Linux 9.6 完整系统用户列表
	systemUsers := []string{
		// 核心系统用户
		"root", "bin", "daemon", "adm", "lp", "sync", "shutdown", "halt",
		"mail", "operator", "games", "ftp", "nobody", "systemd-network",
		"dbus", "polkitd", "libstoragemgmt", "cockpit-ws", "cockpit-wsinstance",
		"sssd", "chrony", "systemd-resolve", "tss", "systemd-coredump",
		"systemd-oom", "clevis", "kmod", "systemd-timesync", "systemd-journal-remote",
		"systemd-journal-upload", "systemd-journal-gateway", "systemd-bus-proxy",
		
		// 网络和安全相关
		"sshd", "rpc", "rpcuser", "nfsnobody", "unbound", "named",
		"dhcpd", "openvpn", "nm-openvpn", "nm-openconnect", "NetworkManager",
		"firewalld", "iptables", "fail2ban", "denyhosts",
		
		// 数据库和Web服务
		"mysql", "mariadb", "postgres", "postgresql", "redis", "mongodb",
		"nginx", "apache", "httpd", "www-data", "lighttpd", "tomcat",
		"jetty", "glassfish", "wildfly", "jboss",
		
		// 容器和虚拟化
		"docker", "podman", "libvirt", "qemu", "kvm", "virt",
		"containers", "crio", "kubernetes", "k8s", "openshift",
		
		// 监控和日志
		"zabbix", "nagios", "prometheus", "grafana", "elasticsearch",
		"logstash", "kibana", "fluentd", "rsyslog", "syslog", "journald",
		"collectd", "telegraf", "influxdb", "chronograf", "kapacitor",
		
		// HPC和科学计算相关（Rocky Linux常用于HPC环境）
		"slurm", "munge", "condor", "torque", "pbs", "sge", "lsf",
		"openmpi", "mpich", "intel", "cuda", "nvidia", "amd",
		"ganglia", "lustre", "beegfs", "gpfs", "ceph", "gluster",
		
		// 邮件服务
		"postfix", "dovecot", "exim", "sendmail", "cyrus", "courier",
		"amavis", "clamav", "spamassassin", "milter",
		
		// 目录服务和认证
		"ldap", "openldap", "samba", "winbind", "krb5", "kerberos",
		"freeipa", "389ds", "dirsrv", "radiusd", "freeradius",
		
		// 其他常见服务
		"bind", "ntp", "snmp", "avahi", "cups", "lpadmin",
		"gdm", "lightdm", "xdm", "pulse", "rtkit", "colord",
		"geoclue", "flatpak", "packagekit", "usbmuxd", "bluetooth",
		"tcpdump", "wireshark", "pcap", "nmap", "mrtg", "cacti",
		
		// 备份和存储
		"bacula", "amanda", "rsync", "rsyncd", "nfs", "smb", "cifs",
		"iscsi", "multipath", "lvm", "mdadm",
		
		// 开发和构建工具
		"git", "svn", "cvs", "jenkins", "gitlab", "nexus", "artifactory",
		"maven", "gradle", "ant", "make", "cmake", "autotools",
	}
	
	for _, sysUser := range systemUsers {
		if username == sysUser {
			log.Printf("[USER API] User %s is identified as a known Rocky Linux 9.6 system user", username)
			return true
		}
	}
	
	// 检查系统用户命名模式（Rocky Linux特有）
	if strings.HasPrefix(username, "systemd-") {
		log.Printf("[USER API] User %s matches systemd service user pattern", username)
		return true
	}
	
	if strings.HasPrefix(username, "_") {
		log.Printf("[USER API] User %s matches underscore system user pattern", username)
		return true
	}
	
	if strings.HasSuffix(username, "$") {
		log.Printf("[USER API] User %s matches machine account pattern", username)
		return true
	}
	
	// 检查是否为服务账户模式
	if strings.Contains(username, "svc-") || strings.Contains(username, "service-") {
		log.Printf("[USER API] User %s matches service account pattern", username)
		return true
	}
	
	log.Printf("[USER API] User %s is not a system user on Rocky Linux 9.6", username)
	return false
}