package models

type Node struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`     // master or compute
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Password string `json:"password"`
	Status   string `json:"status"`
}

type DeploymentOptions struct {
	CleanupPrevious bool `json:"cleanupPrevious"`
	UseUSTCRepo     bool `json:"useUSTCRepo"`
}

type DeploymentStatus struct {
	Running bool   `json:"running"`
	Step    string `json:"step"`
	Message string `json:"message"`
}