package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"panel-tool/internal/models"
)

// GetSlurmJobs 获取真实的SLURM作业状态
func GetSlurmJobs() []models.JobModel {
	// 检查slurm是否安装
	if _, err := os.Stat("/usr/sbin/slurmctld"); os.IsNotExist(err) {
		// Slurm未安装
		return []models.JobModel{}
	}
	
	// 检查slurmctld服务是否运行
	cmd := exec.Command("systemctl", "is-active", "slurmctld")
	output, err := cmd.Output()
	if err != nil {
		// Slurmctld未运行
		return []models.JobModel{}
	}
	
	status := strings.TrimSpace(string(output))
	if status != "active" {
		// Slurmctld未运行
		return []models.JobModel{}
	}
	
	// 尝试使用squeue命令获取作业信息
	// 使用JSON格式输出便于解析
	cmd = exec.Command("squeue", "--json")
	output, err = cmd.Output()
	if err != nil {
		// 如果JSON格式不支持，使用传统格式
		return getJobsLegacyFormat()
	}
	
	// 解析JSON输出
	var result struct {
		Jobs []struct {
			JobID          string `json:"job_id"`
			Name           string `json:"name"`
			UserName       string `json:"user_name"`
			State          string `json:"state"`
			SubmitTime     string `json:"submit_time"`
			StartTime      string `json:"start_time"`
			RunTime        int    `json:"run_time"`
			TimeLimit      string `json:"time_limit"`
			Partition      string `json:"partition"`
			Priority       int    `json:"priority"`
		} `json:"jobs"`
	}
	
	if err := json.Unmarshal(output, &result); err != nil {
		// JSON解析失败，使用传统格式
		return getJobsLegacyFormat()
	}
	
	var jobs []models.JobModel
	for _, job := range result.Jobs {
		jobModel := models.JobModel{
			JobID: job.JobID,
			User:  job.UserName,
		}
		
		// 转换状态
		switch strings.ToLower(job.State) {
		case "pending", "p":
			jobModel.Status = "pending"
		case "running", "r":
			jobModel.Status = "running"
		default:
			jobModel.Status = strings.ToLower(job.State)
		}
		
		// 解析提交时间
		if job.SubmitTime != "" {
			if t, err := time.Parse("2006-01-02T15:04:05", job.SubmitTime); err == nil {
				jobModel.SubmissionTime = t
			} else {
				jobModel.SubmissionTime = time.Now()
			}
		} else {
			jobModel.SubmissionTime = time.Now()
		}
		
		// 计算等待时间和计算时间
		if jobModel.Status == "pending" {
			jobModel.WaitTime = time.Since(jobModel.SubmissionTime).String()
			jobModel.ComputeTime = "00:00:00"
		} else if jobModel.Status == "running" {
			jobModel.WaitTime = (jobModel.SubmissionTime.Sub(parseStartTime(job.StartTime))).String()
			jobModel.ComputeTime = formatDuration(job.RunTime)
		} else {
			jobModel.WaitTime = "00:00:00"
			jobModel.ComputeTime = "00:00:00"
		}
		
		jobs = append(jobs, jobModel)
	}
	
	return jobs
}

// getJobsLegacyFormat 使用传统格式获取作业信息
func getJobsLegacyFormat() []models.JobModel {
	cmd := exec.Command("squeue", "-h", "-o", "%i|%j|%u|%t|%V|%S|%M")
	output, err := cmd.Output()
	if err != nil {
		// 无法获取作业信息
		return []models.JobModel{}
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		// 没有作业
		return []models.JobModel{}
	}
	
	var jobs []models.JobModel
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) < 7 {
			continue
		}
		
		jobModel := models.JobModel{
			JobID: parts[0],
			User:  parts[2],
		}
		
		// 转换状态
		switch strings.ToLower(parts[3]) {
		case "pending", "p":
			jobModel.Status = "pending"
		case "running", "r":
			jobModel.Status = "running"
		default:
			jobModel.Status = strings.ToLower(parts[3])
		}
		
		// 解析提交时间
		if parts[4] != "" && parts[4] != "Unknown" {
			if t, err := time.Parse("2006-01-02T15:04:05", parts[4]); err == nil {
				jobModel.SubmissionTime = t
			} else {
				jobModel.SubmissionTime = time.Now()
			}
		} else {
			jobModel.SubmissionTime = time.Now()
		}
		
		// 计算等待时间和计算时间
		if jobModel.Status == "pending" {
			jobModel.WaitTime = time.Since(jobModel.SubmissionTime).String()
			jobModel.ComputeTime = "00:00:00"
		} else if jobModel.Status == "running" {
			if parts[5] != "" && parts[5] != "Unknown" {
				if start, err := time.Parse("2006-01-02T15:04:05", parts[5]); err == nil {
					jobModel.WaitTime = start.Sub(jobModel.SubmissionTime).String()
				} else {
					jobModel.WaitTime = "00:00:00"
				}
			} else {
				jobModel.WaitTime = "00:00:00"
			}
			jobModel.ComputeTime = parts[6]
		} else {
			jobModel.WaitTime = "00:00:00"
			jobModel.ComputeTime = "00:00:00"
		}
		
		jobs = append(jobs, jobModel)
	}
	
	return jobs
}

// parseStartTime 解析开始时间
func parseStartTime(startTime string) time.Time {
	if startTime == "" || startTime == "Unknown" {
		return time.Now()
	}
	
	if t, err := time.Parse("2006-01-02T15:04:05", startTime); err == nil {
		return t
	}
	
	return time.Now()
}

// formatDuration 将秒数格式化为时间字符串
func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// ControlSlurmService 控制Slurm服务
func ControlSlurmService(action string) (string, error) {
	var cmd *exec.Cmd
	
	switch action {
	case "start":
		cmd = exec.Command("systemctl", "start", "slurmctld")
	case "stop":
		cmd = exec.Command("systemctl", "stop", "slurmctld")
	case "restart":
		cmd = exec.Command("systemctl", "restart", "slurmctld")
	default:
		return "", &exec.ExitError{}
	}
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}