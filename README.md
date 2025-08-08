# SGHPC-Panel（开发中）

SGHPC-Panel 是一个用于管理和监控高性能计算（HPC）集群的 Web 面板工具。该面板提供了对计算节点、作业调度（如 Slurm）、文件管理和终端操作的集中可视化控制。

## 功能特性

- **节点管理**: 查看和管理 HPC 集群中的计算节点状态
- **作业调度**: 与 Slurm 集成，查看、提交和管理作业
- **文件管理**: 提供 Web 界面进行文件浏览、上传和下载
- **终端访问**: 提供基于 Web 的终端访问功能
- **系统设置**: 配置系统参数和用户权限

## 技术栈

- **后端**: Go 1.24.5
- **前端**: Vue.js 3, Vuetify 3
- **实时通信**: WebSocket
- **构建工具**: npm, go build

## 快速开始

### 环境要求

- Go 1.24.5 或更高版本
- Node.js (支持 npm)

### 后端运行

```bash
go run backend/cmd/main.go
```

### 前端运行

```bash
cd frontend
npm install
npm run serve
```

### 构建

- 后端构建: `go build -o backend/bin/panel backend/cmd/main.go`
- 前端构建: `cd frontend && npm run build`

## 许可证

本项目采用 [GPL-3.0 License](LICENSE.txt) 授权。