package api

import (
	"panel-tool/internal/models"
	"panel-tool/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// HandleGetManagementNode 处理获取管理节点信息请求
func HandleGetManagementNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	node := services.GetManagementNode()
	json.NewEncoder(w).Encode(node)
}

// HandleGetComputeNodes 处理获取计算节点信息请求
func HandleGetComputeNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	nodes := services.GetComputeNodes()
	json.NewEncoder(w).Encode(nodes)
}

// HandleGetSlurmJobs 获取Slurm作业信息
func HandleGetSlurmJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 检查Slurm是否已安装
	_, err := exec.LookPath("sinfo")
	if err != nil {
		// Slurm未安装，返回空数组
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	
	// 执行squeue命令获取作业信息
	cmd := exec.Command("squeue", "--all", "--states=all", "--format=%i|%j|%u|%t|%M|%l|%N", "--noheader")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Failed to execute squeue command", http.StatusInternalServerError)
		return
	}
	
	// 解析输出
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var jobs []models.JobModel
	
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 7 {
			continue
		}
		
		//jobID, _ := strconv.Atoi(parts[0])
		job := models.JobModel{
			JobID:       parts[0],
			User:        parts[2],
			Status:      parts[3],
			ComputeTime: parts[4],
			WaitTime:    parts[5],
		}
		jobs = append(jobs, job)
	}
	
	json.NewEncoder(w).Encode(jobs)
}

// HandleFileUpload 处理文件上传请求
func HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	// 设置最大内存大小为32MB
	r.ParseMultipartForm(32 << 20)
	
	// 获取上传的文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// 获取目标路径（默认为当前目录）
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "./uploads"
	}
	
	// 创建目标目录（如果不存在）
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		http.Error(w, "Unable to create target directory", http.StatusInternalServerError)
		return
	}
	
	// 创建目标文件
	targetFile, err := os.Create(filepath.Join(targetPath, handler.Filename))
	if err != nil {
		http.Error(w, "Unable to create target file", http.StatusInternalServerError)
		return
	}
	defer targetFile.Close()
	
	// 将上传的文件内容复制到目标文件
	_, err = io.Copy(targetFile, file)
	if err != nil {
		http.Error(w, "Unable to save uploaded file", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "File uploaded successfully",
		"file":    handler.Filename,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleFileDownload 处理文件下载请求
func HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	// 获取文件路径
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	// 设置响应头
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	w.Header().Set("Content-Type", "application/octet-stream")
	
	// 读取并发送文件内容
	http.ServeFile(w, r, filePath)
}

// HandleFileList 处理文件列表请求
func HandleFileList(w http.ResponseWriter, r *http.Request) {
	// 获取目录路径
	dirPath := r.URL.Query().Get("path")
	if dirPath == "" {
		dirPath = "."
	}
	
	// 读取目录内容
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "Unable to read directory", http.StatusInternalServerError)
		return
	}
	
	// 构造响应数据
	var files []map[string]interface{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		file := map[string]interface{}{
			"name":    entry.Name(),
			"is_dir":  entry.IsDir(),
			"size":    info.Size(),
			"mod_time": info.ModTime(),
			"mode":    info.Mode().String(),
		}
		files = append(files, file)
	}
	
	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// HandleFileDelete 处理文件删除请求
func HandleFileDelete(w http.ResponseWriter, r *http.Request) {
	// 获取文件路径
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}
	
	// 删除文件或目录
	err := os.RemoveAll(filePath)
	if err != nil {
		http.Error(w, "Unable to delete file or directory", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "File or directory deleted successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleFilePermissions 处理文件权限修改请求
func HandleFilePermissions(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var requestData struct {
		Path        string `json:"path"`
		Permissions string `json:"permissions"` // 八进制权限字符串，例如 "755"
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 检查必需字段
	if requestData.Path == "" || requestData.Permissions == "" {
		http.Error(w, "Path and permissions are required", http.StatusBadRequest)
		return
	}
	
	// 解析权限字符串
	permValue, err := strconv.ParseUint(requestData.Permissions, 8, 32)
	if err != nil {
		http.Error(w, "Invalid permissions format", http.StatusBadRequest)
		return
	}
	
	// 修改文件权限
	err = os.Chmod(requestData.Path, os.FileMode(permValue))
	if err != nil {
		http.Error(w, "Unable to change file permissions", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "File permissions updated successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleFileMkdir 处理创建目录请求
func HandleFileMkdir(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var requestData struct {
		Path string `json:"path"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 检查必需字段
	if requestData.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}
	
	// 创建目录
	err := os.MkdirAll(requestData.Path, 0755)
	if err != nil {
		http.Error(w, "Unable to create directory", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "Directory created successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleFileRename 处理文件重命名请求
func HandleFileRename(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var requestData struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 检查必需字段
	if requestData.OldPath == "" || requestData.NewPath == "" {
		http.Error(w, "Old path and new path are required", http.StatusBadRequest)
		return
	}
	
	// 重命名文件或目录
	err := os.Rename(requestData.OldPath, requestData.NewPath)
	if err != nil {
		http.Error(w, "Unable to rename file or directory", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "File or directory renamed successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// parseSymbolicPermissions 解析符号权限字符串（例如 "rwxr-xr-x"）
func parseSymbolicPermissions(permStr string) os.FileMode {
	var perm os.FileMode
	
	// 处理用户权限位
	if len(permStr) > 0 && permStr[0] == 'r' {
		perm |= 0400
	}
	if len(permStr) > 1 && permStr[1] == 'w' {
		perm |= 0200
	}
	if len(permStr) > 2 && permStr[2] == 'x' {
		perm |= 0100
	} else if len(permStr) > 2 && permStr[2] == 's' {
		perm |= 0100 | os.ModeSetuid
	} else if len(permStr) > 2 && permStr[2] == 'S' {
		perm |= os.ModeSetuid
	}
	
	// 处理组权限位
	if len(permStr) > 3 && permStr[3] == 'r' {
		perm |= 040
	}
	if len(permStr) > 4 && permStr[4] == 'w' {
		perm |= 020
	}
	if len(permStr) > 5 && permStr[5] == 'x' {
		perm |= 010
	} else if len(permStr) > 5 && permStr[5] == 's' {
		perm |= 010 | os.ModeSetgid
	} else if len(permStr) > 5 && permStr[5] == 'S' {
		perm |= os.ModeSetgid
	}
	
	// 处理其他用户权限位
	if len(permStr) > 6 && permStr[6] == 'r' {
		perm |= 04
	}
	if len(permStr) > 7 && permStr[7] == 'w' {
		perm |= 02
	}
	if len(permStr) > 8 && permStr[8] == 'x' {
		perm |= 01
	} else if len(permStr) > 8 && permStr[8] == 't' {
		perm |= 01 | os.ModeSticky
	} else if len(permStr) > 8 && permStr[8] == 'T' {
		perm |= os.ModeSticky
	}
	
	return perm
}

// HandleLogin 处理用户登录请求
func HandleLogin(w http.ResponseWriter, r *http.Request, authService *services.AuthService) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	
	// 解析请求体
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 使用认证服务验证用户凭据
	if authService.AuthenticateUser(credentials.Username, credentials.Password) {
		// 登录成功，返回token和用户信息
		response := map[string]interface{}{
			"token": fmt.Sprintf("token_%d", time.Now().Unix()),
			"user": map[string]string{
				"username": credentials.Username,
			},
			"is_default_password": (credentials.Username == "admin" && credentials.Password == "password"),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// 登录失败
	log.Printf("Authentication failed for user: %s", credentials.Username)
	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
}

// HandleChangePassword 处理修改密码请求
func HandleChangePassword(w http.ResponseWriter, r *http.Request, authService *services.AuthService) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	
	var requestData struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 获取当前用户（在实际应用中应该从token中获取）
	// 这里为了简化，我们假设是admin用户
	username := "admin"
	
	// 使用认证服务修改密码
	if err := authService.ChangePassword(username, requestData.CurrentPassword, requestData.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "Password changed successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// TerminalSession 表示一个终端会话
type TerminalSession struct {
	ID        string
	Conn      *websocket.Conn
	PTY       *os.File
	Cmd       *exec.Cmd
	Username  string
	Password  string
	AuthState int
	LastActive time.Time
	Mutex     sync.RWMutex
}

const (
	AuthNone = iota
	AuthPending
	AuthSuccess
)

// TerminalSessionManager 管理所有终端会话
type TerminalSessionManager struct {
	Sessions map[string]*TerminalSession
	Mutex    sync.RWMutex
}

var sessionManager = &TerminalSessionManager{
	Sessions: make(map[string]*TerminalSession),
}

// closeSession 关闭会话并清理资源
func closeSession(session *TerminalSession) {
	session.Mutex.Lock()
	defer session.Mutex.Unlock()
	
	if session.PTY != nil {
		session.PTY.Close()
	}
	
	if session.Cmd != nil && session.Cmd.Process != nil {
		session.Cmd.Process.Kill()
	}
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接到WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	
	// 生成会话ID
	sessionID := uuid.New().String()
	
	// 创建新的终端会话
	session := &TerminalSession{
		ID:         sessionID,
		Conn:       conn,
		AuthState:  AuthNone,
		LastActive: time.Now(),
	}
	
	// 添加会话到管理器
	sessionManager.Mutex.Lock()
	sessionManager.Sessions[sessionID] = session
	sessionManager.Mutex.Unlock()
	
	// 发送欢迎消息
	welcomeMsg, _ := json.Marshal(map[string]interface{}{
		"type": "output",
		"data": fmt.Sprintf("\r\n欢迎使用终端！请输入服务器账户和密码进行登录。\r\n会话ID: %s\r\n\r\n请输入用户名: ", sessionID),
	})
	conn.WriteMessage(websocket.TextMessage, welcomeMsg)
	
	// 监听WebSocket消息
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// 连接已关闭，清理会话
			sessionManager.Mutex.Lock()
			if s, exists := sessionManager.Sessions[sessionID]; exists {
				closeSession(s)
				delete(sessionManager.Sessions, sessionID)
			}
			sessionManager.Mutex.Unlock()
			break
		}
		
		// 只处理文本消息
		if messageType != websocket.TextMessage {
			continue
		}
		
		// 解析消息
		var msg struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}
		
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		
		// 更新最后活动时间
		session.Mutex.Lock()
		session.LastActive = time.Now()
		session.Mutex.Unlock()
		
		// 根据消息类型处理
		switch msg.Type {
		case "input":
			// 处理用户输入
			if data, ok := msg.Data.(string); ok {
				switch session.AuthState {
				case AuthNone:
					// 获取用户名
					session.Mutex.Lock()
					session.Username = strings.TrimSpace(data)
					session.AuthState = AuthPending
					session.Mutex.Unlock()
					
					// 请求密码
					passwordPrompt, _ := json.Marshal(map[string]interface{}{
						"type": "output",
						"data": "\r\n请输入密码: ",
					})
					conn.WriteMessage(websocket.TextMessage, passwordPrompt)
					
				case AuthPending:
					// 获取密码
					session.Mutex.Lock()
					session.Password = strings.TrimSpace(data)
					username := session.Username
					password := session.Password
					session.Mutex.Unlock()
					
					// 验证用户名和密码
					processMsg, _ := json.Marshal(map[string]interface{}{
						"type": "output",
						"data": "\r\n正在验证用户凭据...\r\n",
					})
					conn.WriteMessage(websocket.TextMessage, processMsg)
					
					// 验证用户凭据
					// 注意：这里应该使用实际的认证服务
					if username == "admin" && password == "password" {
						// 认证成功
						sessionManager.Mutex.RLock()
						if s, exists := sessionManager.Sessions[sessionID]; exists {
							s.Mutex.Lock()
							s.AuthState = AuthSuccess
							s.Mutex.Unlock()
							
							// 启动PTY并处理输出
							go startPTYAndHandleOutput(s)
						}
						sessionManager.Mutex.RUnlock()
					} else {
						// 认证失败
						errorMsg, _ := json.Marshal(map[string]interface{}{
							"type": "output",
							"data": "\r\n认证失败: 用户名或密码错误\r\n请重新输入用户名: ",
						})
						conn.WriteMessage(websocket.TextMessage, errorMsg)
						
						// 重置认证状态
						sessionManager.Mutex.RLock()
						if s, exists := sessionManager.Sessions[sessionID]; exists {
							s.Mutex.Lock()
							s.AuthState = AuthNone
							s.Username = ""
							s.Password = ""
							s.Mutex.Unlock()
						}
						sessionManager.Mutex.RUnlock()
					}
					
				case AuthSuccess:
					// 已认证，直接转发输入到PTY
					sessionManager.Mutex.RLock()
					if s, exists := sessionManager.Sessions[sessionID]; exists && s.PTY != nil {
						_, err = s.PTY.Write([]byte(data))
						if err != nil {
							log.Printf("Write error: %v", err)
						}
					}
					sessionManager.Mutex.RUnlock()
				}
			}
			
		case "resize":
			// 处理窗口大小调整
			if session.AuthState == AuthSuccess {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					cols, _ := data["cols"].(float64)
					rows, _ := data["rows"].(float64)
					
					sessionManager.Mutex.RLock()
					if s, exists := sessionManager.Sessions[sessionID]; exists && s.PTY != nil {
						pty.Setsize(s.PTY, &pty.Winsize{
							Rows: uint16(rows),
							Cols: uint16(cols),
						})
					}
					sessionManager.Mutex.RUnlock()
				}
			}
			
		case "ping":
			// 处理心跳请求
			pongMsg, _ := json.Marshal(map[string]interface{}{
				"type": "pong",
				"data": time.Now().UnixNano() / int64(time.Millisecond),
			})
			conn.WriteMessage(websocket.TextMessage, pongMsg)
		}
	}
}

// 验证用户凭据
func authenticateUser(username, password string) bool {
	// 简单的验证逻辑，实际项目中应该查询数据库或使用其他认证机制
	return username == "admin" && password == "password"
}

// 启动PTY并处理输出的函数
func startPTYAndHandleOutput(session *TerminalSession) {
	// 发送登录成功消息
	successMsg, _ := json.Marshal(map[string]interface{}{
		"type": "output",
		"data": fmt.Sprintf("\r\n登录成功! 欢迎 %s\r\n\r\n", session.Username),
	})
	
	// 获取WebSocket连接
	session.Mutex.Lock()
	conn := session.Conn
	session.Mutex.Unlock()
	
	if conn == nil {
		log.Printf("Error: WebSocket connection is nil")
		return
	}
	
	conn.WriteMessage(websocket.TextMessage, successMsg)
	
	// 启动一个bash shell
	cmd := exec.Command("/bin/bash")
	
	// 创建伪终端
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Error starting pty: %v", err)
		errorMsg, _ := json.Marshal(map[string]interface{}{
			"type": "output",
			"data": fmt.Sprintf("\r\n启动终端失败: %v\r\n", err),
		})
		conn.WriteMessage(websocket.TextMessage, errorMsg)
		return
	}
	
	// 更新会话信息
	session.Mutex.Lock()
	session.PTY = ptmx
	session.Cmd = cmd
	session.Mutex.Unlock()

	// 设置初始窗口大小
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: 30,
		Cols: 120,
	})

	// 处理PTY输出并转发到WebSocket
	buf := make([]byte, 1024)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from pty: %v", err)
			}
			// 发送EOF消息到前端
			eofMsg, _ := json.Marshal(map[string]interface{}{
				"type": "output",
				"data": "\r\n会话已结束.\r\n",
			})
			conn.WriteMessage(websocket.TextMessage, eofMsg)

			// 清理会话
			sessionManager.Mutex.Lock()
			if s, exists := sessionManager.Sessions[session.ID]; exists {
				closeSession(s)
				delete(sessionManager.Sessions, session.ID)
			}
			sessionManager.Mutex.Unlock()
			return
		}
		
		// 更新最后活动时间
		session.Mutex.Lock()
		session.LastActive = time.Now()
		session.Mutex.Unlock()
		
		// 将PTY输出转发到WebSocket
		outputMsg, _ := json.Marshal(map[string]interface{}{
			"type": "output",
			"data": string(buf[:n]),
		})
		err = conn.WriteMessage(websocket.TextMessage, outputMsg)
		if err != nil {
			log.Printf("Error writing to websocket: %v", err)
			
			// 清理会话
			sessionManager.Mutex.Lock()
			if s, exists := sessionManager.Sessions[session.ID]; exists {
				closeSession(s)
				delete(sessionManager.Sessions, session.ID)
			}
			sessionManager.Mutex.Unlock()
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	// 允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息类型常量
const (
	// CommandInput represents input command
	CommandInput = "input"
	// CommandResize represents resize command
	CommandResize = "resize"
	// CommandPing represents ping command
	CommandPing = "ping"
)