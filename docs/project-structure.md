# Project Structure

```
xiaohongshu-image/
├── cmd/                           # Application entry points
│   ├── api/
│   │   └── main.go               # API server with Gin + Asynq scheduler
│   └── worker/
│       └── main.go               # Worker server with Asynq handlers
│
├── internal/                      # Private application code
│   ├── api/
│   │   └── handler.go           # HTTP handlers (Gin routes)
│   │
│   ├── config/
│   │   └── config.go           # Configuration management (Viper)
│   │
│   ├── db/
│   │   └── database.go          # Database layer (GORM)
│   │
│   ├── models/
│   │   └── models.go           # Data models (Settings, Note, Comment, Task, Delivery, AuditLog)
│   │
│   ├── services/
│   │   ├── xhsconnector/       # Xiaohongshu connector abstraction
│   │   │   ├── connector.go    # Connector interface
│   │   │   ├── mock.go        # Mock connector for testing
│   │   │   └── mcp.go         # MCP connector for real data
│   │   │
│   │   ├── intent/             # Intent recognition service
│   │   │   ├── intent.go      # Rule + LLM based extraction
│   │   │   └── http.go       # HTTP client for LLM API
│   │   │
│   │   ├── provider/           # Generation provider abstraction
│   │   │   ├── provider.go    # Provider interface + UnifiedGenRequest
│   │   │   ├── mock.go        # Mock provider for testing
│   │   │   ├── http.go        # HTTP provider for real APIs
│   │   │   ├── mapper.go      # JSONPath-based request/response mapping
│   │   │   └── storage.go     # Storage interface
│   │   │
│   │   ├── storage/            # Object storage service
│   │   │   └── minio.go       # MinIO implementation
│   │   │
│   │   └── mailer/            # Email service
│   │       └── mailer.go      # SMTP email sending
│   │
│   └── worker/
│       └── worker.go           # Asynq job handlers
│           # PollJob, IntentJob, SubmitJob, StatusJob, EmailJob
│
├── pkg/                          # Public library code
│   └── logger/
│       └── logger.go           # Logging utilities (Zap)
│
├── web/                          # Next.js frontend
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx      # Root layout with navigation
│   │   │   ├── page.tsx        # Home page
│   │   │   ├── settings/
│   │   │   │   └── page.tsx  # Settings page
│   │   │   ├── tasks/
│   │   │   │   ├── page.tsx  # Tasks list page
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx # Task detail page
│   │   │   └── globals.css    # Global styles
│   │   └── lib/
│   │       └── api.ts         # API client (Axios)
│   ├── public/                  # Static assets
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.js
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   └── Dockerfile
│
├── migrations/                   # Database migrations
│   ├── 000001_init.up.sql      # Create tables
│   └── 000001_init.down.sql    # Drop tables
│
├── config/                       # Configuration files
│   └── config.yaml             # Main configuration (Viper)
│
├── docs/                         # Documentation
│   ├── architecture.md          # System architecture
│   ├── deployment.md           # Deployment guide
│   └── acceptance-criteria.md # Acceptance criteria
│
├── .env.example                 # Environment variables template
├── .gitignore                  # Git ignore rules
├── go.mod                      # Go module definition
├── go.sum                      # Go dependencies checksum
├── Makefile                    # Build automation
├── docker-compose.yml           # Docker Compose configuration
├── Dockerfile.api              # API server Dockerfile
├── Dockerfile.worker           # Worker server Dockerfile
└── README.md                   # Project documentation
```

## Key Components

### Backend Services

1. **API Server** (`cmd/api/main.go`)
   - Gin HTTP server
   - RESTful API endpoints
   - Asynq scheduler for periodic polling
   - Serves Next.js frontend

2. **Worker Server** (`cmd/worker/main.go`)
   - Asynq job processor
   - 5 job types: Poll, Intent, Submit, Status, Email
   - Distributed processing support

### Core Services

1. **XHS Connector** (`internal/services/xhsconnector/`)
   - Abstract interface for comment fetching
   - Mock connector for demo/testing
   - MCP connector for real data

2. **Intent Recognition** (`internal/services/intent/`)
   - Two-layer filtering (rules + LLM)
   - Email extraction and validation
   - Confidence scoring

3. **Generation Provider** (`internal/services/provider/`)
   - Unified request/response interface
   - Mock provider for demo/testing
   - HTTP provider with JSONPath mapping
   - Configurable for any API

4. **Storage** (`internal/services/storage/`)
   - MinIO S3-compatible storage
   - Presigned URL generation
   - File upload/download

5. **Mailer** (`internal/services/mailer/`)
   - SMTP email sending
   - Template-based emails
   - Error handling

### Frontend

1. **Next.js App** (`web/src/app/`)
   - Home page with overview
   - Settings page for configuration
   - Tasks list with auto-refresh
   - Task detail with full information

2. **API Client** (`web/src/lib/api.ts`)
   - Axios-based HTTP client
   - TypeScript interfaces
   - Error handling

### Infrastructure

1. **Database** (MySQL 8.0)
   - GORM ORM
   - Migrations with golang-migrate
   - 6 tables: settings, notes, comments, tasks, deliveries, audit_logs

2. **Queue** (Redis 7)
   - Asynq job queue
   - Rate limiting
   - Distributed locks

3. **Storage** (MinIO)
   - S3-compatible object storage
   - Presigned URLs for downloads
   - Web console at :9001

4. **Email** (Mailhog)
   - SMTP server for testing
   - Web UI at :8025
   - No real delivery in dev

## Data Flow

```
1. PollJob (Scheduled)
   └─> Connector.ListComments()
       └─> Save to DB (comments table)
           └─> Enqueue IntentJob

2. IntentJob
   └─> Rule Filter (keywords + email)
       └─> LLM Extraction (OpenAI-compatible)
           └─> Create Task (if clear intent)
               └─> Enqueue SubmitJob

3. SubmitJob
   └─> Provider.Submit()
       └─> Save provider_job_id
           └─> Enqueue StatusJob (delayed)

4. StatusJob (Retries with backoff)
   └─> Provider.Status()
       ├─> If SUCCEEDED:
       │   └─> Save result_url
       │       └─> Enqueue EmailJob
       ├─> If FAILED:
       │   └─> Mark task FAILED
       └─> If RUNNING/PENDING:
           └─> Re-enqueue StatusJob

5. EmailJob
   └─> Rate Limit Check
       └─> Mailer.Send()
           └─> Save delivery record
               └─> Mark task EMAILED
```

## Configuration

### Environment Variables

- `SERVER_PORT`: API server port (default: 8080)
- `DATABASE_HOST`: MySQL host
- `DATABASE_PORT`: MySQL port
- `DATABASE_USER`: MySQL user
- `DATABASE_PASSWORD`: MySQL password
- `DATABASE_DBNAME`: Database name
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port
- `MINIO_ENDPOINT`: MinIO endpoint
- `MINIO_ACCESS_KEY`: MinIO access key
- `MINIO_SECRET_KEY`: MinIO secret key
- `LLM_BASE_URL`: LLM API base URL
- `LLM_API_KEY`: LLM API key
- `LLM_MODEL`: LLM model name
- `SMTP_HOST`: SMTP host
- `SMTP_PORT`: SMTP port
- `SMTP_FROM`: SMTP from address
- `ASYNQ_REDIS_ADDR`: Redis address for Asynq

### Configuration File

`config/config.yaml` contains default values that can be overridden by environment variables.

## Deployment

### Development

```bash
docker-compose up -d
```

### Production

- Use Kubernetes or Docker Swarm
- Enable SSL/TLS
- Configure secrets management
- Set up monitoring and alerting
- Use managed services (RDS, ElastiCache, S3)

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

```bash
go test -tags=integration ./...
```

### E2E Tests

```bash
# Start services
docker-compose up -d

# Run E2E tests
go test -tags=e2e ./tests/e2e
```

## Monitoring

### Health Checks

- API: `GET /healthz`
- All services have Docker health checks

### Metrics

- Prometheus metrics at `/metrics`
- Structured logs with Zap
- Task status tracking in database

### Logs

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f worker
```

## Scaling

### Horizontal Scaling

- API: Stateless, can run multiple instances
- Worker: Multiple instances process jobs in parallel
- Database: Use read replicas
- Redis: Use cluster mode

### Vertical Scaling

- Increase Asynq concurrency
- Increase database connection pool
- Increase worker resources (CPU/Memory)
