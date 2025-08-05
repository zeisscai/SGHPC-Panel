package api

import (
	"panel-tool/internal/models"
	"panel-tool/internal/services"
	"encoding/json"
	"net/http"
)

// 全局Slurm部署服务实例
var slurmDeployService *services.SlurmDeployService

func init() {
	slurmDeployService = services.NewSlurmDeployService()
}

// HandleGetSlurmConfig 获取Slurm配置
func HandleGetSlurmConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	config := slurmDeployService.GetConfig()
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateSlurmConfig 更新Slurm配置
func HandleUpdateSlurmConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config models.SlurmConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	slurmDeployService.UpdateConfig(&config)
	
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Configuration updated successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleGetSlurmDeployStatus 获取部署状态
func HandleGetSlurmDeployStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := slurmDeployService.GetStatus()
	json.NewEncoder(w).Encode(status)
}

// HandleStartSlurmDeploy 启动Slurm部署
func HandleStartSlurmDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 在goroutine中执行部署，避免阻塞HTTP请求
	go func() {
		slurmDeployService.Deploy()
	}()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Deployment started successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleGetSlurmDeployLogs 获取部署日志
func HandleGetSlurmDeployLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logs := slurmDeployService.GetLogs()
	json.NewEncoder(w).Encode(logs)
}

// HandleSlurmServiceControl Slurm服务控制
func HandleSlurmServiceControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Action string `json:"action"` // start, stop, restart
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var output string
	var err error
	
	switch request.Action {
	case "start":
		output, err = services.ControlSlurmService("start")
	case "stop":
		output, err = services.ControlSlurmService("stop")
	case "restart":
		output, err = services.ControlSlurmService("restart")
	default:
		http.Error(w, "Invalid action. Supported actions: start, stop, restart", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Service control command executed",
		"output":  output,
	}
	json.NewEncoder(w).Encode(response)
}