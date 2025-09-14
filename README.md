# TaskHub

A modern task management application with Go backend, React frontend, and production-ready deployment infrastructure.

## Overview

- **Backend**: Go-based REST API with SQLite database
- **Frontend**: React web application
- **Deployment**: Docker Compose for local development, Kubernetes for production
- **CI/CD**: GitHub Actions with OIDC authentication

## Quick Start

### Local Development (Docker Compose)

```bash
# Start the application
docker-compose up --build

# Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080/api/v1/health
```

### Kubernetes Deployment

**Using Kustomize:**
```bash
# Deploy to staging
kubectl apply -k k8s/overlays/staging

# Deploy to production  
kubectl apply -k k8s/overlays/production
```

## API Endpoints

- `GET /api/v1/health` - Health check
- `GET /api/v1/tasks` - List tasks
- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks/:id` - Get task by ID

## Development

**Backend:**
```bash
cd backend
go mod download
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm start
```
