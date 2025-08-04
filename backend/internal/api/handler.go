package api

import (
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
	"time"

	"panel-tool/internal/services"
	
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
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

// HandleGetSlurmJobs 处理获取SLURM作业状态请求
func HandleGetSlurmJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobs := services.GetSlurmJobs()
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
	
	// 确保目标目录存在
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		http.Error(w, "Unable to create target directory", http.StatusInternalServerError)
		return
	}
	
	// 创建目标文件
	dst, err := os.Create(filepath.Join(targetPath, handler.Filename))
	if err != nil {
		http.Error(w, "Unable to create target file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]interface{}{
		"message": "File uploaded successfully",
		"file": map[string]interface{}{
			"name": handler.Filename,
			"size": handler.Size,
			"path": filepath.Join(targetPath, handler.Filename),
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleFileDownload 处理文件下载请求
func HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	// 获取要下载的文件路径
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
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	
	// 读取并发送文件
	http.ServeFile(w, r, filePath)
}

// HandleFileList 处理文件列表请求
func HandleFileList(w http.ResponseWriter, r *http.Request) {
	// 获取目录路径（默认为当前目录）
	dirPath := r.URL.Query().Get("path")
	if dirPath == "" {
		dirPath = "."
	}
	
	// 检查目录是否存在
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		http.Error(w, "Directory not found", http.StatusNotFound)
		return
	}
	
	if !info.IsDir() {
		http.Error(w, "Path is not a directory", http.StatusBadRequest)
		return
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
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}
		
		file := map[string]interface{}{
			"name":     entry.Name(),
			"type":     "file",
			"size":     fileInfo.Size(),
			"modified": fileInfo.ModTime().Format("2006-01-02 15:04:05"),
		}
		
		if entry.IsDir() {
			file["type"] = "directory"
			file["size"] = 0
		} else {
			// 检查是否可执行
			if fileInfo.Mode()&0111 != 0 {
				file["executable"] = true
			}
		}
		
		// 获取文件权限
		file["permissions"] = fileInfo.Mode().String()
		
		files = append(files, file)
	}
	
	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// HandleFileDelete 处理文件删除请求
func HandleFileDelete(w http.ResponseWriter, r *http.Request) {
	// 只允许DELETE方法
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取要删除的文件路径
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
	
	// 删除文件
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Unable to delete file", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]string{
		"message": "File deleted successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleFilePermissions 处理文件权限修改请求
func HandleFilePermissions(w http.ResponseWriter, r *http.Request) {
	// 只允许PUT方法
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 解析请求体
	var requestData struct {
		Path        string `json:"path"`
		Permissions string `json:"permissions"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 检查必需参数
	if requestData.Path == "" || requestData.Permissions == "" {
		http.Error(w, "Path and permissions are required", http.StatusBadRequest)
		return
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(requestData.Path); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	// 解析权限字符串（例如 "0755" 或 "rwxr-xr-x"）
	var perm os.FileMode
	if strings.HasPrefix(requestData.Permissions, "0") && len(requestData.Permissions) == 4 {
		// 数字格式权限（如 "0755"）
		permValue, err := strconv.ParseUint(requestData.Permissions, 8, 32)
		if err != nil {
			http.Error(w, "Invalid permissions format", http.StatusBadRequest)
			return
		}
		perm = os.FileMode(permValue)
	} else if len(requestData.Permissions) == 9 || len(requestData.Permissions) == 10 {
		// 字符格式权限（如 "rwxr-xr-x"）
		perm = parseSymbolicPermissions(requestData.Permissions)
	} else {
		http.Error(w, "Invalid permissions format", http.StatusBadRequest)
		return
	}
	
	// 修改文件权限
	if err := os.Chmod(requestData.Path, perm); err != nil {
		http.Error(w, "Unable to change file permissions", http.StatusInternalServerError)
		return
	}
	
	// 返回成功响应
	response := map[string]interface{}{
		"message":     "File permissions changed successfully",
		"permissions": perm.String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseSymbolicPermissions 解析符号权限字符串（如 "rwxr-xr-x"）
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
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	
	// 解析请求体
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 获取有效的用户名和密码
	validUsername := os.Getenv("ADMIN_USERNAME")
	validPassword := os.Getenv("ADMIN_PASSWORD")
	
	if validUsername == "" {
		validUsername = "admin"
	}
	
	if validPassword == "" {
		validPassword = "password"
	}
	
	// 验证凭据
	if credentials.Username == validUsername && credentials.Password == validPassword {
		// 登录成功，返回token和用户信息
		response := map[string]interface{}{
			"token": fmt.Sprintf("token_%d", time.Now().Unix()),
			"user": map[string]string{
				"username": credentials.Username,
			},
			"is_default_password": (credentials.Username == "admin" && credentials.Password == "password"),
		}
		json.NewEncoder(w).Encode(response)
	} else {
		// 登录失败
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
	}
}

// HandleChangePassword 处理修改密码请求
func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	
	// 这里应该验证用户身份和权限
	// 为简化起见，我们只实现基本逻辑
	
	var requestData struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 验证当前密码（实际项目中应该从数据库或配置中验证）
	validPassword := os.Getenv("ADMIN_PASSWORD")
	if validPassword == "" {
		validPassword = "password"
	}
	
	if requestData.CurrentPassword != validPassword {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}
	
	// 在实际应用中，这里应该更新密码存储
	// 例如更新数据库或配置文件
	// 为了演示，我们将新密码保存到环境变量中（实际应用中应保存到安全存储中）
	os.Setenv("ADMIN_PASSWORD", requestData.NewPassword)
	
	// 返回成功响应
	response := map[string]string{
		"message": "Password changed successfully",
	}
	json.NewEncoder(w).Encode(response)
}

var upgrader = websocket.Upgrader{
	// 允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息类型常量
const (
	CommandInput   = "input"
	CommandResize  = "resize"
	CommandPing    = "ping"
	CommandLogin   = "login"
)

// ResizeMessage 定义窗口大小调整消息结构
type ResizeMessage struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// WebSocketMessage 定义WebSocket消息结构
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// LoginData 定义登录数据结构
type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// 发送欢迎消息，提示用户登录
	welcomeMsg, _ := json.Marshal(WebSocketMessage{
		Type: "output",
		Data: "Welcome to SGHPC Terminal\r\nPlease login to continue\r\nUsername: ",
	})
	conn.WriteMessage(websocket.TextMessage, welcomeMsg)

	var cmd *exec.Cmd
	var ptmx *os.File

	// 处理来自WebSocket的输入并转发到PTY
	// 登录状态
	loggedIn := false
	expectingUsername := true
	var username, password string

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		// 解析消息
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// 如果未登录，处理登录流程
		if !loggedIn {
			switch msg.Type {
			case CommandInput:
				if data, ok := msg.Data.(string); ok {
					if expectingUsername {
						username = strings.TrimSpace(data)
						expectingUsername = false
						
						// 请求密码
						passwordPrompt, _ := json.Marshal(WebSocketMessage{
							Type: "output",
							Data: "Password: ",
						})
						conn.WriteMessage(websocket.TextMessage, passwordPrompt)
					} else {
						password = strings.TrimSpace(data)
						
						// 验证凭据
						validUsername := os.Getenv("ADMIN_USERNAME")
						validPassword := os.Getenv("ADMIN_PASSWORD")
						
						if validUsername == "" {
							validUsername = "admin"
						}
						
						if validPassword == "" {
							validPassword = "password"
						}
						
						if username == validUsername && password == validPassword {
							loggedIn = true
							
							// 登录成功消息
							successMsg, _ := json.Marshal(WebSocketMessage{
								Type: "output",
								Data: "\r\nLogin successful. Starting terminal...\r\n",
							})
							conn.WriteMessage(websocket.TextMessage, successMsg)
							
							// 创建PTY连接
							cmd = exec.Command("/bin/bash")
							
							// 启动PTY
							ptmx, err = pty.Start(cmd)
							if err != nil {
								log.Printf("Failed to start pty: %v", err)
								errorMsg, _ := json.Marshal(WebSocketMessage{
									Type: "error",
									Data: fmt.Sprintf("Failed to start pty: %v", err),
								})
								conn.WriteMessage(websocket.TextMessage, errorMsg)
								return
							}
							
							// 设置初始窗口大小
							pty.Setsize(ptmx, &pty.Winsize{
								Rows: 30,
								Cols: 120,
							})
							
							// 启动goroutine处理PTY输出并转发到WebSocket
							go func() {
								buf := make([]byte, 1024)
								for {
									n, err := ptmx.Read(buf)
									if err != nil {
										if err != io.EOF {
											log.Printf("Error reading from pty: %v", err)
										}
										return
									}
									
									// 将PTY输出转发到WebSocket
									outputMsg, _ := json.Marshal(WebSocketMessage{
										Type: "output",
										Data: string(buf[:n]),
									})
									err = conn.WriteMessage(websocket.TextMessage, outputMsg)
									if err != nil {
										log.Printf("Error writing to websocket: %v", err)
										return
									}
								}
							}()
						} else {
							// 登录失败，重新提示输入用户名
							failMsg, _ := json.Marshal(WebSocketMessage{
								Type: "output",
								Data: "\r\nLogin failed. Please try again.\r\nUsername: ",
							})
							conn.WriteMessage(websocket.TextMessage, failMsg)
							expectingUsername = true
						}
					}
				}
			case CommandPing:
				// 处理ping消息
				pongMsg, _ := json.Marshal(WebSocketMessage{
					Type: "pong",
					Data: nil,
				})
				conn.WriteMessage(websocket.TextMessage, pongMsg)
			}
			continue
		}

		// 已登录，处理正常终端操作
		switch msg.Type {
		case CommandInput:
			// 处理终端输入
			if data, ok := msg.Data.(string); ok {
				_, err = ptmx.Write([]byte(data))
				if err != nil {
					log.Printf("Write error: %v", err)
					break
				}
			}
		case CommandResize:
			// 处理窗口大小调整
			if data, ok := msg.Data.(map[string]interface{}); ok {
				cols, _ := data["cols"].(float64)
				rows, _ := data["rows"].(float64)
				pty.Setsize(ptmx, &pty.Winsize{
					Rows: uint16(rows),
					Cols: uint16(cols),
				})
			}
		case CommandPing:
			// 处理ping消息
			pongMsg, _ := json.Marshal(WebSocketMessage{
				Type: "pong",
				Data: nil,
			})
			conn.WriteMessage(websocket.TextMessage, pongMsg)
		}
	}
	
	// 清理资源
	if ptmx != nil {
		_ = ptmx.Close()
	}
	if cmd != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
}

// WebSocketWriter 实现io.Writer接口，将数据写入WebSocket连接
type WebSocketWriter struct {
	conn *websocket.Conn
}

func (w *WebSocketWriter) Write(data []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}