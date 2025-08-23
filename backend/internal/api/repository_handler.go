package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RepositoryStatus 软件源状态结构
type RepositoryStatus struct {
	Type       string `json:"type"`       // default, ustc, custom
	URL        string `json:"url"`        // 当前软件源URL
	Accessible bool   `json:"accessible"` // 是否可访问
}

// RepositoryResponse 通用响应结构
type RepositoryResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SwitchRepositoryRequest 切换软件源请求结构
type SwitchRepositoryRequest struct {
	Type string `json:"type"` // default, ustc
}

// CustomRepositoryRequest 自定义软件源请求结构
type CustomRepositoryRequest struct {
	URL string `json:"url"`
}

// TestRepositoryRequest 测试软件源请求结构
type TestRepositoryRequest struct {
	URL string `json:"url"`
}

// 预定义的软件源配置
var repositoryConfigs = map[string]map[string]string{
	"ustc": {
		"baseurl": "https://mirrors.ustc.edu.cn",
		"name":    "USTC Mirror",
	},
	"default": {
		"baseurl": "", // 系统默认
		"name":    "Default Repository",
	},
}

// HandleCleanRepositoryCache 处理清理软件包缓存请求
func HandleCleanRepositoryCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 执行缓存清理命令
	err := cleanRepositoryCache()
	if err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: fmt.Sprintf("缓存清理失败: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := RepositoryResponse{
		Success: true,
		Message: "缓存清理完成",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleSwitchRepository 处理切换软件源请求
func HandleSwitchRepository(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SwitchRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: "请求参数解析失败",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证软件源类型
	if req.Type != "default" && req.Type != "ustc" {
		response := RepositoryResponse{
			Success: false,
			Message: "不支持的软件源类型",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 执行软件源切换
	err := switchRepository(req.Type)
	if err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: fmt.Sprintf("软件源切换失败: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := RepositoryResponse{
		Success: true,
		Message: fmt.Sprintf("已切换到%s软件源", getRepositoryDisplayName(req.Type)),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleGetRepositoryStatus 处理获取软件源状态请求
func HandleGetRepositoryStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取当前软件源状态
	status, err := getRepositoryStatus()
	if err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: fmt.Sprintf("获取软件源状态失败: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := RepositoryResponse{
		Success: true,
		Message: "获取状态成功",
		Data:    status,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleSetCustomRepository 处理设置自定义软件源请求
func HandleSetCustomRepository(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CustomRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: "请求参数解析失败",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证URL格式
	if !isValidURL(req.URL) {
		response := RepositoryResponse{
			Success: false,
			Message: "无效的URL格式",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 设置自定义软件源
	err := setCustomRepository(req.URL)
	if err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: fmt.Sprintf("设置自定义软件源失败: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := RepositoryResponse{
		Success: true,
		Message: "自定义软件源设置成功",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleTestRepository 处理测试软件源连接请求
func HandleTestRepository(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := RepositoryResponse{
			Success: false,
			Message: "请求参数解析失败",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 测试软件源连接
	accessible := testRepositoryConnection(req.URL)

	response := RepositoryResponse{
		Success: true,
		Message: "测试完成",
		Data: map[string]bool{
			"accessible": accessible,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// cleanRepositoryCache 清理软件包缓存
func cleanRepositoryCache() error {
	// 执行 dnf clean all
	cmd := exec.Command("dnf", "clean", "all")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 dnf clean all 失败: %v", err)
	}

	// 执行 dnf makecache
	cmd = exec.Command("dnf", "makecache")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 dnf makecache 失败: %v", err)
	}

	return nil
}

// switchRepository 切换软件源
func switchRepository(repoType string) error {
	repoDir := "/etc/yum.repos.d"
	backupDir := "/etc/yum.repos.d.backup"

	// 创建备份目录
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 备份当前repo文件
	if err := backupRepoFiles(repoDir, backupDir); err != nil {
		return fmt.Errorf("备份repo文件失败: %v", err)
	}

	switch repoType {
	case "default":
		return restoreDefaultRepositories(repoDir)
	case "ustc":
		return setupUSTCRepositories(repoDir)
	default:
		return fmt.Errorf("不支持的软件源类型: %s", repoType)
	}
}

// getRepositoryStatus 获取当前软件源状态
func getRepositoryStatus() (*RepositoryStatus, error) {
	status := &RepositoryStatus{
		Type:       "default",
		URL:        "",
		Accessible: false,
	}

	// 检查当前使用的软件源类型
	repoType, url := detectCurrentRepository()
	status.Type = repoType
	status.URL = url

	// 测试连接性
	if url != "" {
		status.Accessible = testRepositoryConnection(url)
	} else {
		// 对于默认源，测试系统默认的repo连接
		status.Accessible = testDefaultRepositoryConnection()
	}

	return status, nil
}

// setCustomRepository 设置自定义软件源
func setCustomRepository(baseURL string) error {
	repoDir := "/etc/yum.repos.d"
	backupDir := "/etc/yum.repos.d.backup"

	// 创建备份目录
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 备份当前repo文件
	if err := backupRepoFiles(repoDir, backupDir); err != nil {
		return fmt.Errorf("备份repo文件失败: %v", err)
	}

	// 生成自定义repo配置
	return generateCustomRepositoryConfig(repoDir, baseURL)
}

// 辅助函数

// backupRepoFiles 备份repo文件
func backupRepoFiles(sourceDir, backupDir string) error {
	files, err := filepath.Glob(filepath.Join(sourceDir, "*.repo"))
	if err != nil {
		return err
	}

	for _, file := range files {
		basename := filepath.Base(file)
		backupPath := filepath.Join(backupDir, basename)
		
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		
		if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// restoreDefaultRepositories 恢复默认软件源
func restoreDefaultRepositories(repoDir string) error {
	// 删除当前repo文件
	files, err := filepath.Glob(filepath.Join(repoDir, "*.repo"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	// 重新安装默认repo包（如果可能）
	cmd := exec.Command("dnf", "reinstall", "-y", "centos-release")
	cmd.Run() // 忽略错误，因为可能不是CentOS系统

	return nil
}

// setupUSTCRepositories 设置中科大镜像源
func setupUSTCRepositories(repoDir string) error {
	// 删除现有repo文件
	files, err := filepath.Glob(filepath.Join(repoDir, "*.repo"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	// 创建USTC镜像源配置
	return generateUSTCRepositoryConfig(repoDir)
}

// generateUSTCRepositoryConfig 生成中科大镜像源配置
func generateUSTCRepositoryConfig(repoDir string) error {
	config := `[base]
name=CentOS-$releasever - Base - mirrors.ustc.edu.cn
baseurl=https://mirrors.ustc.edu.cn/centos/$releasever/BaseOS/$basearch/os/
gpgcheck=1
gpgkey=https://mirrors.ustc.edu.cn/centos/RPM-GPG-KEY-CentOS-Official

[appstream]
name=CentOS-$releasever - AppStream - mirrors.ustc.edu.cn
baseurl=https://mirrors.ustc.edu.cn/centos/$releasever/AppStream/$basearch/os/
gpgcheck=1
gpgkey=https://mirrors.ustc.edu.cn/centos/RPM-GPG-KEY-CentOS-Official

[extras]
name=CentOS-$releasever - Extras - mirrors.ustc.edu.cn
baseurl=https://mirrors.ustc.edu.cn/centos/$releasever/extras/$basearch/os/
gpgcheck=1
gpgkey=https://mirrors.ustc.edu.cn/centos/RPM-GPG-KEY-CentOS-Official
`

	filePath := filepath.Join(repoDir, "ustc.repo")
	return ioutil.WriteFile(filePath, []byte(config), 0644)
}

// generateCustomRepositoryConfig 生成自定义软件源配置
func generateCustomRepositoryConfig(repoDir, baseURL string) error {
	// 确保URL以/结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	config := fmt.Sprintf(`[custom-base]
name=Custom Repository - Base
baseurl=%scentos/$releasever/BaseOS/$basearch/os/
gpgcheck=0
enabled=1

[custom-appstream]
name=Custom Repository - AppStream
baseurl=%scentos/$releasever/AppStream/$basearch/os/
gpgcheck=0
enabled=1

[custom-extras]
name=Custom Repository - Extras
baseurl=%scentos/$releasever/extras/$basearch/os/
gpgcheck=0
enabled=1
`, baseURL, baseURL, baseURL)

	filePath := filepath.Join(repoDir, "custom.repo")
	return ioutil.WriteFile(filePath, []byte(config), 0644)
}

// detectCurrentRepository 检测当前使用的软件源
func detectCurrentRepository() (string, string) {
	repoDir := "/etc/yum.repos.d"
	files, err := filepath.Glob(filepath.Join(repoDir, "*.repo"))
	if err != nil {
		return "default", ""
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "mirrors.ustc.edu.cn") {
			return "ustc", "https://mirrors.ustc.edu.cn"
		}
		if strings.Contains(contentStr, "custom") {
			// 尝试提取自定义URL
			lines := strings.Split(contentStr, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "baseurl=") {
					url := strings.TrimPrefix(line, "baseurl=")
					// 提取基础URL
					if idx := strings.Index(url, "/centos"); idx > 0 {
						url = url[:idx]
					}
					return "custom", url
				}
			}
		}
	}

	return "default", ""
}

// testRepositoryConnection 测试软件源连接
func testRepositoryConnection(url string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// testDefaultRepositoryConnection 测试默认软件源连接
func testDefaultRepositoryConnection() bool {
	// 尝试执行 dnf check-update 来测试连接
	cmd := exec.Command("dnf", "check-update", "--quiet")
	err := cmd.Run()
	// dnf check-update 返回100表示有更新可用，这也算成功
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == 100
		}
		return false
	}
	return true
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// getRepositoryDisplayName 获取软件源显示名称
func getRepositoryDisplayName(repoType string) string {
	switch repoType {
	case "default":
		return "默认"
	case "ustc":
		return "中科大镜像"
	default:
		return "自定义"
	}
}