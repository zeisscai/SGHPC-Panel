package models

import "time"

// JobModel 定义SLURM作业数据结构
type JobModel struct {
	JobID          string    `json:"job_id"`
	SubmissionTime time.Time `json:"submission_time"`
	WaitTime       string    `json:"wait_time"`
	ComputeTime    string    `json:"compute_time"`
	User           string    `json:"user"`
	Status         string    `json:"status"`
}