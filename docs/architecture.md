# Architecture

## System Overview

The Xiaohongshu Image Generation System is a distributed microservices architecture that automatically detects image/video generation requests from Xiaohongshu comments and delivers results via email.

## Components

### 1. API Service (cmd/api)
- HTTP server built with Gin framework
- RESTful API endpoints for configuration and task management
- Integrates with database, Redis queue, and worker
- Serves the Next.js frontend

### 2. Worker Service (cmd/worker)
- Asynq-based background job processor
- Handles multiple job types:
  - PollJob: Fetch comments from Xiaohongshu
  - IntentJob: Extract generation intent
  - SubmitJob: Submit to generation provider
  - StatusJob: Check generation status
  - EmailJob: Send results via email

### 3. Frontend (web/)
- Next.js 14 with App Router
- React 18 with TypeScript
- Tailwind CSS for styling
- Pages: Home, Settings, Tasks List, Task Detail

### 4. Database (MySQL)
- Stores settings, notes, comments, tasks, deliveries, audit logs
- GORM ORM for database operations
- Migrations with golang-migrate

### 5. Redis
- Asynq job queue
- Rate limiting
- Distributed locks
- Caching

### 6. MinIO
- S3-compatible object storage
- Stores generated content
- Provides presigned URLs for downloads

### 7. Mailhog
- SMTP server for testing
- Web UI to view emails
- No real email delivery in development

## Data Flow

```
1. PollJob (Scheduled)
   └─> Connector.ListComments()
       └─> Save to DB
           └─> Enqueue IntentJob

2. IntentJob
   └─> Rule Filter (keywords + email)
       └─> LLM Extraction
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

## Key Design Decisions

### 1. Asynq for Job Queue
- **Why**: Built on Redis, supports retries, delays, and priorities
- **Benefits**: Reliable, distributed, easy to monitor
- **Alternatives Considered**: Watermill, RabbitMQ

### 2. Idempotency Strategy
- **Comment Level**: `comment_uid` UNIQUE index prevents duplicates
- **Task Level**: `comment_id` UNIQUE ensures one task per comment
- **Job Level**: Asynq jobId ensures jobs are not duplicated
- **Email Level**: Redis rate limiting prevents spam

### 3. Provider Mapping Mechanism
- **Why**: Support multiple APIs without code changes
- **Implementation**: JSONPath-based request/response mapping
- **Benefits**: Easy to add new providers, configuration-driven

### 4. Two-Layer Intent Detection
- **Layer 1**: Rule-based filtering (keywords + email regex)
- **Layer 2**: LLM-based extraction with confidence scoring
- **Why**: Reduces LLM costs, improves accuracy

### 5. Mock Connector & Provider
- **Why**: Enable full-stack testing without external dependencies
- **Benefits**: Fast development, reliable demos, easy CI/CD

## Scalability Considerations

### Horizontal Scaling
- API: Stateless, can scale horizontally
- Worker: Multiple instances can process jobs in parallel
- Database: Read replicas for queries
- Redis: Cluster mode for high availability

### Performance Optimization
- Database connection pooling
- Redis caching for frequent queries
- Batch processing for comments
- Async job processing

### Monitoring
- Structured logging with Zap
- Prometheus metrics (planned)
- Health check endpoints
- Task status tracking

## Security Considerations

### API Security
- Input validation with go-playground/validator
- SQL injection prevention via GORM
- Rate limiting on endpoints

### Data Protection
- Email addresses stored securely
- Presigned URLs with expiration
- No sensitive data in logs

### Authentication (Future)
- JWT-based authentication
- Role-based access control
- API key management

## Deployment Architecture

### Development
```
Docker Compose
├── MySQL (single instance)
├── Redis (single instance)
├── MinIO (single instance)
├── Mailhog (SMTP testing)
├── API (single instance)
├── Worker (single instance)
└── Web (single instance)
```

### Production (Recommended)
```
Kubernetes / Docker Swarm
├── MySQL Cluster (Primary + Replicas)
├── Redis Cluster (Sentinel)
├── MinIO (Distributed mode)
├── SMTP Service (SendGrid/Ses)
├── API (Multiple instances + Load Balancer)
├── Worker (Multiple instances)
├── Web (CDN + Static hosting)
└── Monitoring (Prometheus + Grafana)
```

## Error Handling

### Retry Strategy
- LLM calls: 2 retries with 1s delay
- Provider submit: No retry (idempotent)
- Provider status: 20 retries with exponential backoff (15s, 30s, 60s...)
- Email: No retry (user can request again)

### Dead Letter Queue
- Failed jobs are logged
- Audit trail for debugging
- Manual retry capability (future)

## Testing Strategy

### Unit Tests
- Service layer logic
- Intent extraction
- Provider mapping
- Email validation

### Integration Tests
- Database operations
- Redis queue
- API endpoints

### E2E Tests
- Full workflow with MockConnector
- Email delivery verification
- Task status transitions

## Future Enhancements

1. **Multi-User Support**
   - User authentication
   - Per-user settings
   - Usage quotas

2. **Advanced Features**
   - Comment reply to Xiaohongshu
   - Webhook notifications
   - Batch processing
   - Custom templates

3. **Monitoring & Analytics**
   - Prometheus metrics
   - Grafana dashboards
   - Usage analytics
   - Cost tracking

4. **Performance**
   - GraphQL API
   - WebSocket for real-time updates
   - CDN for static assets
   - Database sharding
