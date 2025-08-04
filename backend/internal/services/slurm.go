package services

import (
	"time"
	"panel-tool/internal/models"
)

// GetSlurmJobs 获取SLURM作业状态
func GetSlurmJobs() []models.JobModel {
	// 模拟数据
	jobs := []models.JobModel{
		{
			JobID:          "12345",
			SubmissionTime: time.Now().Add(-2 * time.Hour),
			WaitTime:       "00:02:30",
			ComputeTime:    "01:25:10",
			User:           "researcher1",
			Status:         "running",
		},
		{
			JobID:          "12346",
			SubmissionTime: time.Now().Add(-1 * time.Hour),
			WaitTime:       "00:00:00",
			ComputeTime:    "00:45:22",
			User:           "researcher2",
			Status:         "running",
		},
		{
			JobID:          "12347",
			SubmissionTime: time.Now().Add(-15 * time.Minute),
			WaitTime:       "00:15:45",
			ComputeTime:    "00:00:00",
			User:           "researcher3",
			Status:         "pending",
		},
	}

	return jobs
}