package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
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
	
	// 防止删除系统关键用户
	if isSystemUser(username) {
		http.Error(w, "Cannot delete system user", http.StatusForbidden)
		return
	}
	
	err := deleteSystemUser(username)
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

// getSystemUsers 获取系统用户列表（过滤系统用户）
func getSystemUsers() ([]UserInfo, error) {
	log.Printf("[USER API] Detecting operating system: %s", runtime.GOOS)
	
	if runtime.GOOS == "darwin" {
		// macOS系统
		return getMacOSUsers()
	} else {
		// Linux系统
		return getLinuxUsers()
	}
}

// getLinuxUsers 获取Linux系统用户
func getLinuxUsers() ([]UserInfo, error) {
	log.Printf("[USER API] Executing 'getent passwd' command for Linux...")
	cmd := exec.Command("getent", "passwd")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[USER API] Failed to execute 'getent passwd': %v", err)
		// 尝试备用方法：直接读取/etc/passwd文件
		log.Printf("[USER API] Trying fallback method: reading /etc/passwd...")
		cmd = exec.Command("cat", "/etc/passwd")
		output, err = cmd.Output()
		if err != nil {
			log.Printf("[USER API] Fallback method also failed: %v", err)
			return nil, fmt.Errorf("failed to get user information: %v", err)
		}
	}
	
	return parsePasswdOutput(output)
}

// getMacOSUsers 获取macOS系统用户
func getMacOSUsers() ([]UserInfo, error) {
	log.Printf("[USER API] Getting macOS users using dscl...")
	cmd := exec.Command("dscl", ".", "list", "/Users")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[USER API] Failed to execute dscl command: %v", err)
		return nil, fmt.Errorf("failed to get macOS users: %v", err)
	}
	
	usernames := strings.Split(strings.TrimSpace(string(output)), "\n")
	var users []UserInfo
	
	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		
		// 获取用户详细信息
		userInfo, err := getMacOSUserInfo(username)
		if err != nil {
			log.Printf("[USER API] Failed to get info for user %s: %v", username, err)
			continue
		}
		
		// 过滤系统用户（UID < 500 在macOS中）
		if userInfo.UID < 500 || isMacOSSystemUser(username) {
			log.Printf("[USER API] Skipping macOS system user %s (UID: %d)", username, userInfo.UID)
			continue
		}
		
		log.Printf("[USER API] Adding macOS user: %s (UID: %d)", username, userInfo.UID)
		users = append(users, *userInfo)
	}
	
	log.Printf("[USER API] Found %d non-system macOS users", len(users))
	return users, nil
}

// getMacOSUserInfo 获取macOS用户详细信息
func getMacOSUserInfo(username string) (*UserInfo, error) {
	// 获取UID
	cmd := exec.Command("dscl", ".", "read", "/Users/"+username, "UniqueID")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	uidStr := strings.TrimSpace(strings.Replace(string(output), "UniqueID:", "", 1))
	uid, err := strconv.Atoi(strings.TrimSpace(uidStr))
	if err != nil {
		return nil, err
	}
	
	// 获取GID
	cmd = exec.Command("dscl", ".", "read", "/Users/"+username, "PrimaryGroupID")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}
	gidStr := strings.TrimSpace(strings.Replace(string(output), "PrimaryGroupID:", "", 1))
	gid, err := strconv.Atoi(strings.TrimSpace(gidStr))
	if err != nil {
		return nil, err
	}
	
	// 获取Home目录
	cmd = exec.Command("dscl", ".", "read", "/Users/"+username, "NFSHomeDirectory")
	output, err = cmd.Output()
	homeDir := "/Users/" + username // 默认值
	if err == nil {
		homeDirStr := strings.TrimSpace(strings.Replace(string(output), "NFSHomeDirectory:", "", 1))
		if strings.TrimSpace(homeDirStr) != "" {
			homeDir = strings.TrimSpace(homeDirStr)
		}
	}
	
	// 获取Shell
	cmd = exec.Command("dscl", ".", "read", "/Users/"+username, "UserShell")
	output, err = cmd.Output()
	shell := "/bin/bash" // 默认值
	if err == nil {
		shellStr := strings.TrimSpace(strings.Replace(string(output), "UserShell:", "", 1))
		if strings.TrimSpace(shellStr) != "" {
			shell = strings.TrimSpace(shellStr)
		}
	}
	
	// 获取组名
	groupName, _ := getGroupName(gid)
	if groupName == "" {
		groupName = fmt.Sprintf("gid_%d", gid)
	}
	
	return &UserInfo{
		Username: username,
		UID:      uid,
		GID:      gid,
		Group:    groupName,
		HomeDir:  homeDir,
		Shell:    shell,
	}, nil
}

// isMacOSSystemUser 检查是否为macOS系统用户
func isMacOSSystemUser(username string) bool {
	macOSSystemUsers := []string{
		"daemon", "_amavisd", "_appleevents", "_applepay", "_appowner",
		"_ard", "_assetcache", "_astris", "_atsserver", "_avbdeviced",
		"_calendar", "_ces", "_clamav", "_coreaudiod", "_coremedia",
		"_coreml", "_ctkd", "_cvmsroot", "_cvs", "_cyrus", "_devicemgr",
		"_displaypolicyd", "_distnote", "_dovecot", "_dovenull",
		"_dpaudio", "_eppc", "_findmydevice", "_fpsd", "_ftp", "_gamecontrollerd",
		"_hidd", "_iconservices", "_installassistant", "_installer",
		"_jabber", "_kadmin_admin", "_kadmin_changepw", "_krb_anonymous",
		"_krb_changepw", "_krb_kadmin", "_krb_kerberos", "_krb_krbtgt",
		"_krbfast", "_krbtgt", "_launchservicesd", "_lda", "_locationd",
		"_logd", "_lp", "_mailman", "_mbsetupuser", "_mcxalr", "_mdnsresponder",
		"_mysql", "_netbios", "_netstatistics", "_networkd", "_nsurlsessiond",
		"_nsurlstoraged", "_oahd", "_ondemand", "_postfix", "_postgres",
		"_qtss", "_sandbox", "_screensaver", "_scsd", "_securityagent",
		"_serialnumberd", "_softwareupdate", "_spotlight", "_sshd",
		"_svn", "_taskgated", "_teamsserver", "_timezone", "_tokend",
		"_trustevaluationagent", "_unknown", "_update_sharing", "_usbmuxd",
		"_uucp", "_warmd", "_webauthserver", "_windowserver", "_www",
		"_wwwproxy", "_xserverdocs", "nobody", "root",
	}
	
	for _, sysUser := range macOSSystemUsers {
		if username == sysUser {
			return true
		}
	}
	return false
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
		
		if isSystemUser(username) {
			log.Printf("[USER API] Skipping known system user: %s", username)
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

// createSystemUser 创建系统用户
func createSystemUser(username, password, group string) error {
	log.Printf("[USER API] Creating user: %s with group: %s on %s", username, group, runtime.GOOS)
	
	// 检查用户是否已存在
	if _, err := user.Lookup(username); err == nil {
		log.Printf("[USER API] User %s already exists", username)
		return fmt.Errorf("user %s already exists", username)
	}
	
	if runtime.GOOS == "darwin" {
		return createMacOSUser(username, password, group)
	} else {
		return createLinuxUser(username, password, group)
	}
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

// createMacOSUser 创建macOS用户
func createMacOSUser(username, password, group string) error {
	log.Printf("[USER API] Creating macOS user: %s", username)
	
	// 在macOS上，用户创建功能受限，通常需要管理员权限
	// 这里提供一个基本的实现，但在生产环境中可能需要更复杂的处理
	return fmt.Errorf("user creation is not supported on macOS in this demo version. Please use System Preferences to create users manually")
}

// deleteSystemUser 删除系统用户
func deleteSystemUser(username string) error {
	log.Printf("[USER API] Deleting user: %s", username)
	
	// 检查用户是否存在
	if _, err := user.Lookup(username); err != nil {
		log.Printf("[USER API] User %s does not exist", username)
		return fmt.Errorf("user %s does not exist", username)
	}
	
	// 检查是否为系统用户
	if isSystemUser(username) {
		log.Printf("[USER API] Attempted to delete system user: %s", username)
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

// isSystemUser 检查是否为系统用户 - 适配Rocky Linux 9.6
func isSystemUser(username string) bool {
	// Rocky Linux 9.6 系统用户列表
	systemUsers := []string{
		// 基础系统用户
		"root", "bin", "daemon", "adm", "lp", "sync", "shutdown", "halt",
		"mail", "operator", "games", "ftp", "nobody", "systemd-network",
		"dbus", "polkitd", "libstoragemgmt", "cockpit-ws", "cockpit-wsinstance",
		"sssd", "chrony", "systemd-resolve", "tss", "systemd-coredump",
		"systemd-oom", "clevis", "kmod", "systemd-timesync", "systemd-journal-remote",
		
		// 网络和安全相关
		"sshd", "rpc", "rpcuser", "nfsnobody", "unbound", "named",
		"dhcpd", "openvpn", "nm-openvpn", "nm-openconnect",
		
		// 数据库和Web服务
		"mysql", "mariadb", "postgres", "postgresql", "redis", "mongodb",
		"nginx", "apache", "httpd", "www-data", "lighttpd",
		
		// 容器和虚拟化
		"docker", "podman", "libvirt", "qemu", "kvm", "virt",
		"containers", "crio", "kubernetes",
		
		// 监控和日志
		"zabbix", "nagios", "prometheus", "grafana", "elasticsearch",
		"logstash", "kibana", "fluentd", "rsyslog", "syslog",
		
		// HPC和科学计算相关
		"slurm", "munge", "condor", "torque", "pbs", "sge",
		"openmpi", "mpich", "intel", "cuda", "nvidia",
		
		// 其他常见服务
		"postfix", "dovecot", "exim", "sendmail", "bind", "ntp",
		"snmp", "ldap", "openldap", "samba", "winbind", "avahi",
		"cups", "lpadmin", "gdm", "lightdm", "xdm", "pulse", "rtkit",
		"colord", "geoclue", "flatpak", "packagekit", "usbmuxd",
		"tcpdump", "wireshark", "pcap", "nmap",
	}
	
	log.Printf("[USER API] Checking if %s is a system user...", username)
	for _, sysUser := range systemUsers {
		if username == sysUser {
			log.Printf("[USER API] %s is identified as a system user", username)
			return true
		}
	}
	
	log.Printf("[USER API] %s is not a system user", username)
	return false
}