# SGHPC Panel

SGHPC Panel is a web-based management interface for HPC (High-Performance Computing) clusters.

## Features

- Node Management: View and manage compute nodes in the HPC cluster
- Job Scheduling: Integration with Slurm to view, submit, and manage jobs
- File Management: Web interface for browsing, uploading, and downloading files
- Terminal Access: Web-based terminal access
- System Configuration: Configure system parameters and user permissions

## Tech Stack

- Frontend: Vue 3, Vuetify 3
- Backend: Go
- Real-time Communication: WebSocket

## Project Setup

### Backend

```bash
go run backend/cmd/main.go
```

### Frontend

This project has been migrated from Vue CLI to Vite for better performance and Vue 3 support.

Install dependencies:

```bash
cd frontend
npm install
```

#### Development

```bash
cd frontend
npm run dev
```

The frontend will be served at http://localhost:3000 with proxy to backend at http://localhost:8080.

#### Build

```bash
cd frontend
npm run build
```

The built files will be in the `dist` directory.

## Architecture

The project follows a client-server architecture with a separation of frontend and backend:

- Frontend: Vue 3 + Vuetify 3 application (in the `frontend` directory)
- Backend: Go application (in the `backend` directory)

The frontend communicates with the backend through REST API and WebSocket connections.