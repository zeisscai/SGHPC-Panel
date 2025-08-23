package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	users, err := getSystemUsers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get users: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"users": users,
	}
	
	json.NewEncoder(w).Encode(response)
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
	cmd := exec.Command("getent", "passwd")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var users []UserInfo
	
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}
		
		username := parts[0]
		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])
		homeDir := parts[5]
		shell := parts[6]
		
		// 过滤系统用户（UID < 1000）和特殊用户
		if uid < 1000 || isSystemUser(username) {
			continue
		}
		
		// 获取用户组名
		groupName, _ := getGroupName(gid)
		
		users = append(users, UserInfo{
			Username: username,
			UID:      uid,
			GID:      gid,
			Group:    groupName,
			HomeDir:  homeDir,
			Shell:    shell,
		})
	}
	
	return users, nil
}

// createSystemUser 创建系统用户
func createSystemUser(username, password, group string) error {
	// 创建用户
	args := []string{"-m", "-s", "/bin/bash"}
	if group != "" {
		args = append(args, "-g", group)
	}
	args = append(args, username)
	
	cmd := exec.Command("useradd", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	
	// 设置密码
	return changeUserPassword(username, password)
}

// deleteSystemUser 删除系统用户
func deleteSystemUser(username string) error {
	cmd := exec.Command("userdel", "-r", username)
	return cmd.Run()
}

// changeUserPassword 修改用户密码
func changeUserPassword(username, password string) error {
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	return cmd.Run()
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

// isSystemUser 检查是否为系统用户
func isSystemUser(username string) bool {
	systemUsers := []string{
		"root", "daemon", "bin", "sys", "sync", "games", "man", "lp",
		"mail", "news", "uucp", "proxy", "www-data", "backup", "list",
		"irc", "gnats", "nobody", "systemd-network", "systemd-resolve",
		"syslog", "messagebus", "_apt", "lxd", "uuidd", "dnsmasq",
		"landscape", "pollinate", "sshd", "mysql", "postgres", "redis",
		"mongodb", "nginx", "apache", "docker",
	}
	
	for _, sysUser := range systemUsers {
		if username == sysUser {
			return true
		}
	}
	
	return false
}