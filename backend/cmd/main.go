package main

import (
	"log"
	"net/http"

	"panel-tool/internal/api"
)

func main() {
	// 设置路由
	http.HandleFunc("/api/management-node", api.HandleGetManagementNode)
	http.HandleFunc("/api/compute-nodes", api.HandleGetComputeNodes)
	http.HandleFunc("/api/slurm-jobs", api.HandleGetSlurmJobs)
	http.HandleFunc("/api/login", api.HandleLogin)
	http.HandleFunc("/api/change-password", api.HandleChangePassword)
	
	// 文件管理相关路由
	http.HandleFunc("/api/file/upload", api.HandleFileUpload)
	http.HandleFunc("/api/file/download", api.HandleFileDownload)
	http.HandleFunc("/api/file/list", api.HandleFileList)
	http.HandleFunc("/api/file/delete", api.HandleFileDelete)
	http.HandleFunc("/api/file/permissions", api.HandleFilePermissions)
	
	// WebSocket终端路由
	http.HandleFunc("/api/ws", api.HandleWebSocket)
	
	// 提供静态文件服务
	http.Handle("/", http.FileServer(http.Dir("./frontend/dist/")))
	
	// 启动服务器
	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}