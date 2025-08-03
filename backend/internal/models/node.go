package models

// NodeModel 定义节点数据结构
type NodeModel struct {
	Hostname     string  `json:"hostname"`
	Model        string  `json:"model"`
	Architecture string  `json:"architecture"`
	IP           string  `json:"ip"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
}

type ManagementNode struct {
	Hostname     string `json:"hostname"`
	Model        string `json:"model"`
	Architecture string `json:"architecture"`
	CPUInfo      string `json:"cpu_info"`
	OSVersion    string `json:"os_version"`
	KernelVersion string `json:"kernel_version"`
	LocalTime    string `json:"local_time"`
	Uptime       string `json:"uptime"`
}