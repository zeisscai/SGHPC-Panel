# Development Documentation

## Project Overview
This project develops a server control panel similar to 1Panel, designed for managing high-performance computing (HPC) servers. It leverages Go for the backend and Vue.js for the frontend，The primary features include a status overview and system management.

## Functional Requirements

### 1. Status Overview Page
- **Management Node Information**: Displays details such as hostname, model, and architecture.
- **Compute Node Information**: Shows hostname, IP address, CPU usage (as integer percentage like 68%), and memory usage (as "used/total" format like "5.4GB/16GB").
- **SLURM Status**: Lists job_id, submission time, wait time, compute time, and submitting user.

### 2. System Management Page
- **Terminal Functionality**: Provides a web-based terminal for server access.
- **File Management**: Supports file upload, download, and permission modification.

## Technology Stack
- **Backend**: Go
- **Frontend**: Vue.js
- **UI Style**: OpenWrt-inspired Bootstrap

## Project File Structure
```
project-root/
├── backend/
│   ├── cmd/
│   │   └── main.go                # Backend entry point
│   ├── internal/
│   │   ├── api/
│   │   │   └── handler.go         # API request handlers
│   │   ├── models/
│   │   │   ├── node.go            # Node data structures
│   │   │   └── job.go             # SLURM job data structures
│   │   ├── services/
│   │   │   ├── node.go            # Node-related business logic
│   │   │   └── slurm.go           # SLURM-related business logic
│   │   └── utils/
│   │       └── logger.go          # Logging utility
│   └── pkg/                       # External packages (if applicable)
├── frontend/
│   ├── src/
│   │   ├── assets/                # Static assets
│   │   ├── components/            # Reusable Vue components
│   │   │   └── Table.vue          # Table component
│   │   ├── views/                 # Page views
│   │   │   ├── Overview.vue       # Status overview page
│   │   │   └── System.vue         # System management page
│   │   ├── router/
│   │   │   └── index.js           # Routing configuration
│   │   ├── store/
│   │   │   └── index.js           # Vuex state management
│   │   ├── api/
│   │   │   └── node.js            # API interaction for nodes
│   │   └── utils/
│   │       └── format.js          # Utility functions (e.g., date formatting)
│   └── public/
│       └── index.html             # Vue app entry HTML
└── README.md                      # Project README
```

## Code File Functional Requirements

### Backend (Go)
- **`/cmd/main.go`**: Entry point; initializes the HTTP server, sets up routing, and invokes API handlers.
- **`/internal/api/handler.go`**: Manages HTTP requests, calls service-layer functions, and returns JSON responses.
- **`/internal/models/node.go`**: Defines `NodeModel` with fields like hostname, IP, CPU usage (percentage), memory usage (percentage).
- **`/internal/models/job.go`**: Defines `JobModel` with fields like job_id, submission time, etc.
- **`/internal/services/node.go`**: Implements logic to retrieve management and compute node data.
- **`/internal/services/slurm.go`**: Implements logic to fetch SLURM job statuses.
- **`/internal/utils/logger.go`**: Provides logging functionality with configurable levels.

### Frontend (Vue)
- **`/src/main.js`**: Vue app entry; initializes Vue instance, mounts router and Vuex.
- **`/src/components/Table.vue`**: Reusable table component for displaying node and SLURM data.
- **`/src/views/Overview.vue`**: Renders the status overview with node and SLURM information. Displays CPU usage as integer percentage and memory usage in "used/total" format (e.g., "5.4GB/16GB").
- **`/src/views/System.vue`**: Implements terminal and file management features.
- **`/src/router/index.js`**: Configures routes mapping to view components.
- **`/src/store/index.js`**: Manages app state (e.g., node data, job statuses) using Vuex.
- **`/src/api/node.js`**: Encapsulates API calls for node and SLURM data.
- **`/src/utils/format.js`**: Offers utility functions like date formatting.

## Code Standards

### Go
- Adhere to [Go's official style guide](https://golang.org/doc/effective_go.html).
- Use lowercase with underscores for variables, functions, and package names (snake_case).
- Comments with `//` or `/* */`.
- Handle errors using the `errors` package; document all functions briefly.

### Vue
- Component names in PascalCase (e.g., `TableComponent`).
- Props in camelCase (e.g., `nodeInfo`).
- Events in kebab-case (e.g., `update-node`).
- Indent with 2 spaces.
- Enforce ESLint for consistency; comment components and functions briefly.

## Function Naming Conventions

### Backend (Go)
- **Handlers**: `Handle[FunctionName]` (e.g., `HandleGetNodeInfo`).
- **Services**: `Get[Resource]` (e.g., `GetNodeInfo`).
- **Models**: `[Resource]Model` (e.g., `NodeModel`).

### Frontend (Vue)
- **API Calls**: `fetch[Resource]` (e.g., `fetchNodeInfo`).
- **Component Methods**: `[action][Resource]` (e.g., `loadNodeInfo`).
- **Vuex Actions/Mutations**: `[ACTION/MUTATION]_[RESOURCE]` (e.g., `FETCH_NODE_INFO`).

## Team Roles
- **Backend Developers**:
  - Build API handlers, business logic, and data models.
  - Implement terminal and file management (upload, download, permissions).
- **Frontend Developers**:
  - Develop Vue components, views, routing, and state management.
  - Integrate with backend APIs and ensure real-time data updates.
  - Design UI following OpenWrt’s Bootstrap style.

## Considerations
- **Backend APIs**: Use RESTful design for easy frontend integration.
- **Real-Time Data**: Implement WebSocket or polling for CPU/memory updates.
- **Terminal**: Consider WebSocket for interactive terminal access.
- **File Management**: Ensure security in file operations to prevent unauthorized access.

## Conclusion
This document provides a clear framework for the project’s structure, requirements, and standards, enabling efficient collaboration and development.