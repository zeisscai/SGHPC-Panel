package models

// SlurmConfig 定义Slurm配置参数
type SlurmConfig struct {
	ClusterName     string   `json:"cluster_name"`
	ControlMachine  string   `json:"control_machine"`
	ControlAddr     string   `json:"control_addr"`
	SlurmUser       string   `json:"slurm_user"`
	SlurmUID        int      `json:"slurm_uid"`
	SlurmGID        int      `json:"slurm_gid"`
	ComputeNodes    []string `json:"compute_nodes"`
	SlurmctldPort   int      `json:"slurmctld_port"`
	SlurmdPort      int      `json:"slurmd_port"`
	AuthType        string   `json:"auth_type"`
	StateSaveLocation string `json:"state_save_location"`
	SlurmctldPidFile string `json:"slurmctld_pidfile"`
	SlurmdPidFile   string   `json:"slurmd_pidfile"`
	SwitchType      string   `json:"switch_type"`
	MpiDefault      string   `json:"mpi_default"`
	ProctrackType   string   `json:"proctrack_type"`
	ReturnToService int      `json:"return_to_service"`
	MaxTime         string   `json:"max_time"`
}

// SlurmDeploymentStatus 定义部署状态
type SlurmDeploymentStatus struct {
	Phase        string `json:"phase"`         // 部署阶段: checking, downloading, compiling, configuring, finished, failed
	Progress     int    `json:"progress"`      // 部署进度 0-100
	Message      string `json:"message"`       // 当前状态消息
	ErrorMessage string `json:"error_message"` // 错误消息（如果有）
	StartTime    string `json:"start_time"`    // 部署开始时间
	EndTime      string `json:"end_time"`      // 部署结束时间
}