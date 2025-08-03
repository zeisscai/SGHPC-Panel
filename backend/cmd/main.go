package main

import (
	"log"
	"net/http"

	"panel-tool/backend/internal/api"
)

func main() {
	// 设置路由
	http.HandleFunc("/api/management-node", api.AuthMiddleware(api.HandleGetManagementNode))
	http.HandleFunc("/api/compute-nodes", api.AuthMiddleware(api.HandleGetComputeNodes))
	http.HandleFunc("/api/slurm-jobs", api.AuthMiddleware(api.HandleGetSlurmJobs))
	http.HandleFunc("/api/file/upload", api.AuthMiddleware(api.HandleFileUpload))
	http.HandleFunc("/api/file/download", api.AuthMiddleware(api.HandleFileDownload))
	http.HandleFunc("/api/file/permissions", api.AuthMiddleware(api.HandleFilePermissions))
	http.HandleFunc("/api/change-password", api.AuthMiddleware(api.HandleChangePassword))
	http.HandleFunc("/api/login", api.HandleLogin) // 登录端点不需要认证
	http.HandleFunc("/api/ws", api.AuthMiddleware(api.HandleWebSocket))

	// 启动HTTP服务器
	log.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}