package services

import (
	"math/rand"
	"strconv"
	"time"

	"panel-tool/backend/internal/models"
)

// GetSlurmJobs 获取SLURM作业状态
func GetSlurmJobs() []*models.JobModel {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// 模拟一些作业数据
	jobs := make([]*models.JobModel, 5)
	users := []string{"alice", "bob", "charlie"}
	statuses := []string{"pending", "running", "completed", "cancelled", "pending"}
	
	for i := 0; i < 5; i++ {
		jobs[i] = &models.JobModel{
			JobID:          strconv.Itoa(100 + i),
			SubmissionTime: time.Now().Add(-time.Duration(rand.Intn(60)) * time.Minute),
			WaitTime:       strconv.Itoa(rand.Intn(10)) + "m",
			ComputeTime:    strconv.Itoa(rand.Intn(60)) + "m",
			User:           users[rand.Intn(len(users))],
			Status:         statuses[i],
		}
	}
	
	return jobs
}