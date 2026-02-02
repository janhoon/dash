# Dash - Monitoring Dashboard

A Grafana-like monitoring dashboard built with Vue.js, Go, and Prometheus.

## Tech Stack

- **Frontend:** Vue.js 3 (Composition API + TypeScript)
- **Backend:** Go API
- **Database:** PostgreSQL (metadata storage)
- **Data Source:** Prometheus

## Features (Planned)

- Dashboard CRUD operations
- Panel system with 12-column grid layout
- Time range picker with presets and custom ranges
- Prometheus data source integration
- PromQL query editor
- Line chart visualizations (ECharts)
- Auto-refresh at configurable intervals
- Drag-and-drop dashboard layout

## Development

### Prerequisites

- Node.js 18+
- Go 1.21+
- Docker and Docker Compose

### Setup

1. Start the infrastructure services:
   ```bash
   docker-compose up -d
   ```

2. Start the backend API:
   ```bash
   cd backend
   go run ./cmd/api
   ```
   The API will be available at http://localhost:8080

3. Start the frontend dev server:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   The frontend will be available at http://localhost:5173

### Running Tests

Frontend:
```bash
cd frontend
npm run type-check
npm run test
```

Backend:
```bash
cd backend
go test ./...
```

### API Endpoints

- `GET /api/health` - Health check endpoint

## Project Structure

```
dash/
├── frontend/           # Vue.js 3 application
│   ├── src/
│   └── package.json
├── backend/            # Go API
│   ├── cmd/api/        # Application entrypoint
│   ├── internal/       # Private application code
│   │   ├── handlers/   # HTTP handlers
│   │   ├── models/     # Data models
│   │   └── db/         # Database connection and migrations
│   └── pkg/            # Public packages
├── agent/              # Ralph agent for automated development
├── docker-compose.yml  # PostgreSQL + Prometheus services
└── README.md
```
