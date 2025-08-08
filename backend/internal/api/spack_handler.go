package api

import (
	"panel-tool/internal/services"
	"encoding/json"
	"net/http"
	
	"github.com/gorilla/websocket"
)

// 全局 Spack 服务实例
var spackService *services.SpackService

func init() {
	spackService = services.NewSpackService()
}

// SpackStatusResponse Spack 状态响应结构体
type SpackStatusResponse struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
}

// InstallSpackRequest 安装 Spack 请求结构体
type InstallSpackRequest struct {
	// 可以添加安装选项
}

// InstallPackageRequest 安装软件包请求结构体
type InstallPackageRequest struct {
	PackageName string `json:"package_name"`
	Options     string `json:"options"`
}

// UninstallPackageRequest 卸载软件包请求结构体
type UninstallPackageRequest struct {
	PackageName string `json:"package_name"`
}

// RepositoriesRequest 软件源配置请求结构体
type RepositoriesRequest struct {
	Content string `json:"content"`
}

// HandleGetSpackStatus 获取 Spack 状态
func HandleGetSpackStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	info := spackService.CheckSpackStatus()
	response := SpackStatusResponse{
		Installed: info.Installed,
		Version:   info.Version,
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleGetSpackInstallationStatus 获取 Spack 安装状态
func HandleGetSpackInstallationStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := spackService.GetInstallationStatus()
	json.NewEncoder(w).Encode(status)
}

// HandleInstallSpack 安装 Spack
func HandleInstallSpack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	// 检查是否已经在安装
	status := spackService.GetInstallationStatus()
	if status.Installing {
		response := map[string]interface{}{
			"message": "Spack installation is already in progress",
			"status":  "in_progress",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// 立即返回响应，表示安装已开始
	response := map[string]interface{}{
		"message": "Spack installation started",
		"status":  "started",
	}
	json.NewEncoder(w).Encode(response)
	
	// 在 goroutine 中执行安装过程，不阻塞 HTTP 响应
	go func() {
		// 创建一个 channel 用于传输日志
		logChan := make(chan string, 100) // 带缓冲的 channel
		
		spackService.InstallSpack(logChan)
	}()
}

// HandleGetAvailablePackages 获取可安装的软件包列表
func HandleGetAvailablePackages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	packages, err := spackService.GetAvailablePackages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(packages)
}

// HandleGetInstalledPackages 获取已安装的软件包列表
func HandleGetInstalledPackages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	packages, err := spackService.GetInstalledPackages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(packages)
}

// HandleInstallPackage 安装软件包
func HandleInstallPackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request InstallPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	// 创建一个 channel 用于传输日志
	logChan := make(chan string)
	
	// 在 goroutine 中执行安装过程
	go func() {
		spackService.InstallPackage(request.PackageName, request.Options, logChan)
	}()
	
	// 立即返回响应，表示安装已开始
	response := map[string]interface{}{
		"message":      "Package installation started",
		"package_name": request.PackageName,
		"status":       "started",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleUninstallPackage 卸载软件包
func HandleUninstallPackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request UninstallPackageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	err := spackService.UninstallPackage(request.PackageName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message":      "Package uninstalled successfully",
		"package_name": request.PackageName,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleGetRepositories 获取软件源配置
func HandleGetRepositories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	content, err := spackService.GetRepositories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := map[string]string{
		"content": content,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleSetRepositories 设置软件源配置
func HandleSetRepositories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request RepositoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	err := spackService.SetRepositories(request.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Repositories updated successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleSpackInstallLogs 通过 WebSocket 提供 Spack 安装日志
func HandleSpackInstallLogs(w http.ResponseWriter, r *http.Request) {
	// 升级连接到 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// 创建一个 channel 用于传输日志
	logChan := make(chan string, 100) // 带缓冲的 channel
	
	// 在 goroutine 中执行安装过程
	go func() {
		spackService.InstallSpack(logChan)
	}()
	
	// 发送日志到客户端
	for logEntry := range logChan {
		err := conn.WriteMessage(websocket.TextMessage, []byte(logEntry))
		if err != nil {
			// 客户端断开连接或其他错误
			break
		}
	}
	
	// 发送结束消息
	conn.WriteMessage(websocket.TextMessage, []byte("INSTALL_COMPLETED"))
}

// HandlePackageInstallLogs 通过 WebSocket 提供软件包安装日志
func HandlePackageInstallLogs(w http.ResponseWriter, r *http.Request) {
	// 升级连接到 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// 从查询参数获取包名和选项
	packageName := r.URL.Query().Get("package")
	options := r.URL.Query().Get("options")

	// 创建一个 channel 用于传输日志
	logChan := make(chan string)
	
	// 在 goroutine 中执行安装过程
	go func() {
		spackService.InstallPackage(packageName, options, logChan)
	}()
	
	// 发送日志到客户端
	for logEntry := range logChan {
		err := conn.WriteMessage(websocket.TextMessage, []byte(logEntry))
		if err != nil {
			// 客户端断开连接或其他错误
			break
		}
	}
	
	// 发送结束消息
	conn.WriteMessage(websocket.TextMessage, []byte("INSTALL_COMPLETED"))
}