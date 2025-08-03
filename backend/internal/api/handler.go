package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"panel-tool/backend/internal/services"
	
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
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
	// 实现文件上传逻辑
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

// HandleFileDownload 处理文件下载请求
func HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	// 实现文件下载逻辑
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File downloaded successfully"))
}

// HandleFilePermissions 处理文件权限修改请求
func HandleFilePermissions(w http.ResponseWriter, r *http.Request) {
	// 实现文件权限修改逻辑
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File permissions changed successfully"))
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

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// 直接连接到本地SSH服务器
	config := &ssh.ClientConfig{
		User: os.Getenv("SSH_USER"),
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Getenv("SSH_PASSWORD")),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 如果环境变量未设置，使用默认值
	if config.User == "" {
		// 尝试从系统中获取当前用户
		if currentUser := os.Getenv("USER"); currentUser != "" {
			config.User = currentUser
		} else {
			config.User = "user"
		}
	}
	
	// 尝试使用SSH密钥认证
	keyPath := os.Getenv("SSH_KEY_PATH")
	if keyPath == "" {
		// 默认密钥路径
		keyPath = os.Getenv("HOME") + "/.ssh/id_rsa"
	}
	
	// 尝试读取SSH密钥
	key, err := os.ReadFile(keyPath)
	if err != nil {
		// 如果无法读取密钥，则使用密码认证
		if os.Getenv("SSH_PASSWORD") != "" {
			config.Auth = []ssh.AuthMethod{
				ssh.Password(os.Getenv("SSH_PASSWORD")),
			}
		} else {
			// 使用默认密码
			config.Auth = []ssh.AuthMethod{
				ssh.Password("password"),
			}
		}
	} else {
		// 使用SSH密钥认证
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// 如果密钥解析失败，回退到密码认证
			if os.Getenv("SSH_PASSWORD") != "" {
				config.Auth = []ssh.AuthMethod{
					ssh.Password(os.Getenv("SSH_PASSWORD")),
				}
			} else {
				config.Auth = []ssh.AuthMethod{
					ssh.Password("password"),
				}
			}
		} else {
			config.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		}
	}

	// 连接到本地SSH服务器
	client, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		log.Printf("Failed to dial: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to connect to local SSH server: %v\r\nPlease make sure SSH server is running on localhost:22", err)))
		return
	}
	defer client.Close()

	// 创建SSH会话
	session, err := client.NewSession()
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to create SSH session: %v", err)))
		return
	}
	defer session.Close()

	// 设置会话
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Printf("Failed to create stdin pipe: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to create stdin pipe: %v", err)))
		return
	}
	
	session.Stdout = &WebSocketWriter{conn}
	session.Stderr = &WebSocketWriter{conn}
	
	// 请求PTY
	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		log.Printf("Failed to request pty: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to request PTY: %v", err)))
		return
	}

	// 启动shell
	if err := session.Shell(); err != nil {
		log.Printf("Failed to start shell: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to start shell: %v", err)))
		return
	}

	// 处理来自WebSocket的输入
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		// 将输入发送到SSH会话
		if _, err := stdin.Write(message); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
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