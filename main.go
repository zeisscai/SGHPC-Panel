package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

type NodeConfig struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
}

type DeploymentConfig struct {
	Nodes []NodeConfig `json:"nodes"`
}

type ScriptStatus struct {
	Running   bool   `json:"running"`
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
}

var currentStatus ScriptStatus

func main() {
	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/api/config", getConfig).Methods("GET")
	r.HandleFunc("/api/config", saveConfig).Methods("POST")
	r.HandleFunc("/api/deploy", startDeployment).Methods("POST")
	r.HandleFunc("/api/status", getStatus).Methods("GET")

	// Serve frontend files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	configFile := filepath.Join(".", "deploy.conf")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Return default config
		defaultConfig := DeploymentConfig{
			Nodes: []NodeConfig{
				{Name: "master", IP: "", Password: "", Hostname: ""},
				{Name: "node1", IP: "", Password: "", Hostname: ""},
				{Name: "node2", IP: "", Password: "", Hostname: ""},
				{Name: "node3", IP: "", Password: "", Hostname: ""},
				{Name: "node4", IP: "", Password: "", Hostname: ""},
			},
		}
		json.NewEncoder(w).Encode(defaultConfig)
		return
	}

	// Parse existing config file
	// For simplicity, we'll just return an empty config
	config := DeploymentConfig{
		Nodes: []NodeConfig{
			{Name: "master", IP: "", Password: "", Hostname: ""},
			{Name: "node1", IP: "", Password: "", Hostname: ""},
			{Name: "node2", IP: "", Password: "", Hostname: ""},
			{Name: "node3", IP: "", Password: "", Hostname: ""},
			{Name: "node4", IP: "", Password: "", Hostname: ""},
		},
	}
	json.NewEncoder(w).Encode(config)
}

func saveConfig(w http.ResponseWriter, r *http.Request) {
	var config DeploymentConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate deploy.conf file
	configContent := ""
	for _, node := range config.Nodes {
		if node.IP != "" {
			configContent += fmt.Sprintf("[%s]\n", node.Name)
			configContent += fmt.Sprintf("ip = %s\n", node.IP)
			configContent += fmt.Sprintf("password = %s\n", node.Password)
			configContent += fmt.Sprintf("hostname = %s\n\n", node.Hostname)
		}
	}

	err := os.WriteFile(filepath.Join(".", "deploy.conf"), []byte(configContent), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func startDeployment(w http.ResponseWriter, r *http.Request) {
	go runDeploymentScript()
	
	currentStatus = ScriptStatus{
		Running:   true,
		Message:   "Deployment started",
		Completed: false,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func runDeploymentScript() {
	defer func() {
		currentStatus.Running = false
		currentStatus.Completed = true
	}()

	scriptPath := filepath.Join(".", "slurm_install-Rocky9.6-1.4.sh")
	
	// Update status
	currentStatus.Message = "Running deployment script..."
	
	// Execute the script
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Dir = "."
	
	// Capture output
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		currentStatus.Message = fmt.Sprintf("Deployment failed: %v\nOutput: %s", err, string(output))
		return
	}
	
	currentStatus.Message = "Deployment completed successfully"
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(currentStatus)
}