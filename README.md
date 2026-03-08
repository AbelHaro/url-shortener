# URL Shortener

A simple and fast URL shortening service built with Go (backend) and React + TypeScript (frontend).

## Features

- Instant URL shortening
- Easy to remember short codes
- Fast redirects
- RESTful API with Swagger documentation

## Tech Stack

### Backend
- Go with Gin framework
- PostgreSQL database
- GORM ORM

### Frontend
- React 19 + TypeScript
- Vite
- shadcn/ui components
- TanStack Query
- React Router

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL (or use Docker)

### Backend Setup

```bash
cd backend
cp env.example .env
# Configure your .env file
go mod download
go run ./cmd/api
```

The API will be available at `http://localhost:8080`

### Generate Swagger Docs

From `backend/`:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go -d cmd/api,internal/delivery/http,internal/domain -o docs --parseInternal
```

After starting the backend in debug mode, open:

`http://localhost:8080/swagger/index.html`

### Frontend Setup

```bash
cd frontend
cp env.example .env
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173`
