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
- **Backend**: Go 1.24.5
- **Frontend**: Vue.js 3, Vuetify 3
- **UI Style**: Vuetify Material Design
- **WebSocket**: gorilla/websocket
- **Authentication**: JWT based authentication

## Project File Structure
```
project-root/
├── backend/
│   ├── cmd/
│   │   └── main.go                # Backend entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler.go         # API request handlers
│   │   │   └── middleware.go      # Authentication middleware
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
│   │   │   ├── System.vue         # System management page
│   │   │   ├── Terminal.vue       # Terminal functionality
│   │   │   ├── FileManagement.vue # File management
│   │   │   ├── Login.vue          # Login page
│   │   │   └── Settings.vue       # Settings page
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

## Development Environment Setup

### Prerequisites
- Go 1.24.5 or higher
- Node.js and npm

### Setup Steps
1. Clone the repository
2. Install Go dependencies:
   ```
   cd backend
   go mod tidy
   ```
3. Install frontend dependencies:
   ```
   cd frontend
   npm install
   ```

### Running the Application
- Start backend server:
  ```
  go run backend/cmd/main.go
  ```
- Start frontend development server:
  ```
  cd frontend
  npm run serve
  ```

### Building for Production
- Build backend:
  ```
  go build -o backend/bin/panel backend/cmd/main.go
  ```
- Build frontend:
  ```
  cd frontend
  npm run build
  ```

## Code File Functional Requirements

### Backend (Go)
- **`/cmd/main.go`**: Entry point; initializes the HTTP server, sets up routing, and invokes API handlers.
- **`/internal/api/handler.go`**: Manages HTTP requests, calls service-layer functions, and returns JSON responses.
- **`/internal/api/middleware.go`**: Implements authentication middleware for securing endpoints.
- **`/internal/models/node.go`**: Defines `NodeModel` with fields like hostname, IP, CPU usage (percentage), memory usage (percentage).
- **`/internal/models/job.go`**: Defines `JobModel` with fields like job_id, submission time, etc.
- **`/internal/services/node.go`**: Implements logic to retrieve management and compute node data.
- **`/internal/services/slurm.go`**: Implements logic to fetch SLURM job statuses.
- **`/internal/utils/logger.go`**: Provides logging functionality with configurable levels.

### Frontend (Vue)
- **`/src/main.js`**: Vue app entry; initializes Vue instance, mounts router and Vuex, configures Vuetify.
- **`/src/App.vue`**: Main application component with layout structure.
- **`/src/components/Table.vue`**: Reusable table component for displaying node and SLURM data.
- **`/src/views/Overview.vue`**: Renders the status overview with node and SLURM information. Displays CPU usage as integer percentage and memory usage in "used/total" format (e.g., "5.4GB/16GB").
- **`/src/views/System.vue`**: Implements system management features with nested routes.
- **`/src/views/Terminal.vue`**: Provides web-based terminal functionality.
- **`/src/views/FileManagement.vue`**: Implements file management features.
- **`/src/views/Login.vue`**: Handles user authentication.
- **`/src/views/Settings.vue`**: Provides user settings functionality.
- **`/src/router/index.js`**: Configures routes mapping to view components.
- **`/src/store/index.js`**: Manages app state (e.g., node data, job statuses) using Vuex.
- **`/src/api/node.js`**: Encapsulates API calls for node and SLURM data.
- **`/src/utils/format.js`**: Offers utility functions like date formatting.

## Code Standards

### Go
- Adhere to [Go's official style guide](https://golang.org/doc/effective_go.html).
- Use camelCase for variables, functions, and package names.
- Comments with `//` or `/* */`.
- Handle errors using the `errors` package; document all functions briefly.
- Use `fmt` for formatting and `log` for logging.

### Vue
- Component names in PascalCase (e.g., `TableComponent`).
- Props in camelCase (e.g., `nodeInfo`).
- Events in kebab-case (e.g., `update-node`).
- Indent with 2 spaces.
- Enforce ESLint for consistency; comment components and functions briefly.

## API Endpoints

### Authentication
- `POST /api/login` - User login
- `POST /api/change-password` - Change user password

### Node Information
- `GET /api/management-node` - Get management node information
- `GET /api/compute-nodes` - Get compute nodes information

### SLURM Jobs
- `GET /api/slurm-jobs` - Get SLURM job statuses

### File Management
- `POST /api/file/upload` - Upload a file
- `GET /api/file/download` - Download a file
- `PUT /api/file/permissions` - Change file permissions

### WebSocket
- `GET /api/ws` - WebSocket connection for real-time updates

## Function Naming Conventions

### Backend (Go)
- **Handlers**: `Handle[FunctionName]` (e.g., `HandleGetNodeInfo`).
- **Services**: `Get[Resource]` or `Process[Action]` (e.g., `GetNodeInfo`, `ProcessFileUpload`).
- **Middleware**: `[Action]Middleware` (e.g., `AuthMiddleware`).

### Frontend (Vue)
- **Components**: PascalCase (e.g., `NodeTable`).
- **Methods**: camelCase (e.g., `fetchNodeData`).
- **Computed Properties**: camelCase (e.g., `formattedMemory`).