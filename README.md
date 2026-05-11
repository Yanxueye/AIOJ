# AIOJ / TerminalOJ

AIOJ is an AI-assisted online judge system for programming practice. It provides problem browsing, code editing, submission judging, submission history, personal learning statistics, and AI-powered learning assistance.

The project contains a Vue 3 frontend and a Go backend. The backend uses RabbitMQ and gRPC to model an asynchronous judging pipeline, closer to a real online judge architecture than a simple synchronous demo.

## Features

- User registration and login with JWT authentication
- Problem list with pagination, keyword search, difficulty filtering, and tag filtering
- Problem detail page with Markdown, LaTeX, and code highlighting support
- Monaco Editor based code editor for C++, Java, Python3, and Go
- Local draft autosave by problem and programming language
- Code submission and asynchronous judging status updates
- Submission history with filtering and sorting
- User profile with solved count, rating, acceptance rate, and learning charts
- AI chat, problem-aware assistant, code diagnosis, and learning support APIs

## Tech Stack

### Frontend

- Vue 3
- Vite
- Vue Router
- Pinia
- Element Plus
- Axios
- Monaco Editor
- ECharts
- marked, KaTeX, highlight.js

### Backend

- Go 1.21
- Gin
- GORM
- MySQL
- RabbitMQ
- gRPC
- JWT
- bcrypt

## Project Structure

```text
AIOJ/
├── backend/
│   ├── cmd/
│   │   ├── server/      # HTTP API service
│   │   └── judger/      # gRPC judging service
│   ├── docker/          # Docker compose and service Dockerfiles
│   ├── internal/
│   │   ├── ai/          # AI service client
│   │   ├── config/      # Configuration loader
│   │   ├── database/    # MySQL initialization and seed data
│   │   ├── handler/     # Gin handlers and routes
│   │   ├── judger/      # Judger client/server logic
│   │   ├── middleware/  # JWT, CORS, recovery, rate limit
│   │   ├── models/      # GORM models and DTOs
│   │   ├── mq/          # RabbitMQ producer and worker
│   │   └── utils/       # Common utilities
│   ├── proto/           # gRPC protocol definitions
│   ├── API.md
│   └── config.yaml
├── frontend/
│   ├── src/
│   │   ├── api/         # Frontend API clients
│   │   ├── components/  # Shared Vue components
│   │   ├── router/      # Vue Router configuration
│   │   ├── stores/      # Pinia stores
│   │   ├── utils/       # Markdown and rendering helpers
│   │   └── views/       # Page views
│   └── package.json
├── PROGRESS.md
└── WORK_SUMMARY.md
```

## Quick Start

### 1. Start Infrastructure

```bash
cd backend
docker compose -f docker/docker-compose.yml up -d mysql rabbitmq
```

RabbitMQ management UI is available at `http://localhost:15672`.

### 2. Start Backend

Open one terminal for the judging service:

```bash
cd backend
go mod tidy
go run ./cmd/judger
```

Open another terminal for the API service:

```bash
cd backend
go run ./cmd/server -config config.yaml
```

The backend API listens on `http://localhost:8080` by default.

### 3. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server is usually available at `http://localhost:5173`.

## Default Account

```text
username: coder_test
password: 123456
```

## Main API Groups

All backend APIs are under `/api`.

- `POST /api/auth/register` - register
- `POST /api/auth/login` - login
- `GET /api/user/profile` - get user profile
- `PUT /api/user/profile` - update user profile
- `GET /api/problems` - problem list
- `GET /api/problems/:id` - problem detail
- `POST /api/submissions` - submit code
- `GET /api/submissions` - submission list
- `GET /api/submissions/:id` - submission detail
- `POST /api/ai/chat` - AI chat
- `POST /api/ai/code-diagnosis` - AI code diagnosis
- `POST /api/ai/knowledge-graph` - AI learning graph

See `backend/API.md` for detailed request and response formats.

## Development Notes

- Frontend production build:

```bash
cd frontend
npm run build
```

- Backend tests:

```bash
cd backend
go test ./...
```

- The AI module can run with a mock implementation when external AI service integration is disabled in `backend/config.yaml`.
- Submission judging uses a queue-based workflow: frontend submission -> API service -> RabbitMQ -> worker -> gRPC judger -> submission result update.

## Documents

- `backend/API.md` - backend API contract
- `backend/PROGRESS.md` - backend development progress
- `PROGRESS.md` - frontend development progress
- `WORK_SUMMARY.md` - recent editor draft autosave improvement summary
