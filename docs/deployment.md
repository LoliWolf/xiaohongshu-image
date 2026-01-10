# Deployment Guide

## Prerequisites
- Docker 20.10+
- Docker Compose 2.0+
- Minimum 4GB RAM
- Minimum 20GB disk space

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/your-org/xiaohongshu-image.git
cd xiaohongshu-image
```

### 2. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` and choose one of the following two methods:

#### Method 1: Using Nacos Configuration Center (Recommended)

```env
# Nacos configuration center
NACOS_ADDR=your-nacos-host.com
NACOS_PORT=8848
NACOS_NAMESPACE=public
NACOS_GROUP=DEFAULT_GROUP
NACOS_DATA_ID=xiaohongshu-image
NACOS_USERNAME=nacos
NACOS_PASSWORD=nacos
```

Then create configuration in Nacos console (JSON format):
```json
{
  "database": {
    "host": "your-mysql-host.com",
    "port": 3306,
    "user": "your-username",
    "password": "your-password",
    "dbname": "xiaohongshu_image"
  },
  "redis": {
    "host": "your-redis-host.com",
    "port": 6379,
    "password": "your-redis-password"
  },
  "minio": {
    "endpoint": "your-minio-host.com:9000",
    "access_key": "your-access-key",
    "secret_key": "your-secret-key"
  },
  "llm": {
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-your-api-key-here",
    "model": "gpt-4o-mini"
  },
  "smtp": {
    "host": "your-smtp-host.com",
    "port": 587,
    "user": "your-smtp-user",
    "password": "your-smtp-password"
  },
  "asynq": {
    "redis_addr": "your-redis-host.com:6379",
    "redis_password": "your-redis-password"
  }
}
```

#### Method 2: Without Nacos (Local Configuration)

Edit `.env` and set external cloud service configuration:

```env
# Don't set NACOS_ADDR, system uses local configuration
NACOS_ADDR=

# Database configuration - external cloud services (e.g., Alibaba Cloud RDS, AWS RDS)
DATABASE_HOST=your-mysql-host.com
DATABASE_PORT=3306
DATABASE_USER=your-username
DATABASE_PASSWORD=your-password
DATABASE_DBNAME=xiaohongshu_image

# Redis configuration - external cloud services (e.g., Alibaba Cloud Redis, AWS ElastiCache)
REDIS_HOST=your-redis-host.com
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# MinIO configuration - external cloud services (e.g., Alibaba Cloud OSS, AWS S3)
MINIO_ENDPOINT=your-minio-host.com:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=generated-content
MINIO_USE_SSL=true
MINIO_REGION=us-east-1

# SMTP configuration - external cloud services (e.g., Alibaba Cloud Email, AWS SES)
SMTP_HOST=your-smtp-host.com
SMTP_PORT=587
SMTP_USER=your-smtp-user
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=noreply@xiaohongshu-image.local

# LLM configuration
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-your-api-key-here
LLM_MODEL=gpt-4o-mini
```

### 3. Start Application Services

```bash
docker-compose up -d
```

Wait for all services to be healthy (about 30-60 seconds).

### 4. Verify Services

```bash
docker-compose ps
```

All services should show "Up" status.

### 5. Access Application

- Web UI: http://localhost:31007
- API: http://localhost:31006/healthz
- External SMTP Service: Access based on configuration

## Service Details

### Application Services

#### API Service
- **Port**: 31006 (external access) / 8080 (internal)
- **Health Check**: http://localhost:31006/healthz
- **Metrics**: http://localhost:31006/metrics

#### Worker Service
- **No exposed ports**
- **Processes jobs from external Redis queue**

#### Web Service
- **Port**: 31007 (external access) / 3000 (internal)
- **Next.js App Router**

### External Cloud Services

#### MySQL
- **Configuration**: Via environment variables DATABASE_HOST, etc.
- **Database Name**: xiaohongshu_image
- **Recommendation**: Use cloud database services (e.g., RDS, Alibaba Cloud RDS)

#### Redis
- **Configuration**: Via environment variables REDIS_HOST, etc.
- **Recommendation**: Use cloud Redis services (e.g., ElastiCache, Alibaba Cloud Redis)

#### MinIO
- **Configuration**: Via environment variables MINIO_ENDPOINT, etc.
- **Recommendation**: Use cloud object storage (e.g., AWS S3, Alibaba Cloud OSS)
- **Storage Bucket**: generated-content

#### SMTP
- **Configuration**: Via environment variables SMTP_HOST, etc.
- **Recommendation**: Use cloud email services (e.g., SendGrid, Alibaba Cloud Email)

## Production Deployment

### Security Considerations

1. **Use Strong Passwords**
   - Database passwords
   - Redis passwords
   - MinIO access keys
   - SMTP passwords

2. **Enable SSL/TLS**
   - Use HTTPS for API
   - Enable SSL for MinIO
   - Use secure SMTP (e.g., port 587 + TLS)

3. **Network Isolation**
   - Use private networks
   - Only expose necessary ports
   - Configure firewall rules

4. **Environment Variable Management**
   - Never commit `.env` file
   - Use secret management (e.g., Docker Secrets, Kubernetes Secrets)
   - Regularly rotate API keys

### Scaling

#### API Scaling

```yaml
api:
  deploy:
    replicas: 3
  resources:
    limits:
      cpus: '1'
      memory: 1G
```

#### Worker Scaling

```yaml
worker:
  deploy:
    replicas: 5
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

#### Database Scaling
- Use MySQL read/write splitting
- Configure read replicas
- Implement connection pooling

#### Redis Scaling
- Use Redis cluster
- Configure persistence (AOF/RDB)
- Set max memory policy

### Monitoring

#### Health Checks

Application services include health checks:

```bash
# Check all services
docker-compose ps

# Check specific service
docker-compose exec api curl http://localhost:8080/healthz
```

#### Logs

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f worker

# View last 100 lines
docker-compose logs --tail=100 api
```

#### Metrics

- API exposes Prometheus metrics at `/metrics`
- Integrate with Prometheus + Grafana
- Set up alerts for:
  - High error rate
  - Queue depth
  - Response time

### Backup

#### Database Backup

```bash
# Use backup tools provided by cloud database
# Example: AWS RDS automatic backup
```

#### MinIO Backup

```bash
# Use backup tools provided by cloud object storage
# Example: AWS S3 version control, lifecycle policies
```

#### Redis Backup

```bash
# Use backup tools provided by cloud Redis
# Example: ElastiCache automatic backup
```

## Troubleshooting

### Services Won't Start

1. Check logs:
   ```bash
   docker-compose logs <service-name>
   ```

2. Check resource usage:
   ```bash
   docker stats
   ```

3. Check port conflicts:
   ```bash
   netstat -tuln | grep LISTEN
   ```

4. Check external service connections:
   ```bash
   # Test database connection
   docker-compose exec api mysql -h ${DATABASE_HOST} -u ${DATABASE_USER} -p${DATABASE_PASSWORD} -e "SELECT 1"
   
   # Test Redis connection
   docker-compose exec api redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping
   ```

### Nacos Connection Failed

1. Check if Nacos address and port are correct
2. Check if network connection is normal
3. Check if Nacos username and password are correct
4. Check if Namespace and Group are correct
5. View logs: `docker-compose logs api`

### Database Connection Issues

1. Verify database is accessible:
   ```bash
   mysql -h ${DATABASE_HOST} -u ${DATABASE_USER} -p
   ```

2. Check connection string configuration

3. Verify database exists:
   ```bash
   mysql -h ${DATABASE_HOST} -u ${DATABASE_USER} -p -e "SHOW DATABASES;"
   ```

4. Run database migrations:
   ```bash
   docker-compose exec api migrate -path /migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_DBNAME}" up
   ```

### Redis Connection Issues

1. Verify Redis is accessible:
   ```bash
   redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping
   ```

2. Check queue depth:
   ```bash
   redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} LLEN asynq:queues:default
   ```

### MinIO Connection Issues

1. Verify MinIO is accessible:
   ```bash
   curl http://${MINIO_ENDPOINT}/minio/health/live
   ```

2. Check bucket exists:
   ```bash
   mc ls ${MINIO_ENDPOINT}/${MINIO_BUCKET}
   ```

### Worker Not Processing Jobs

1. Check worker logs:
   ```bash
   docker-compose logs worker
   ```

2. Verify Redis connection

3. Check queue depth:
   ```bash
   redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} KEYS "asynq:*"
   ```

### Email Not Sent

1. Check external SMTP service logs
2. Verify SMTP configuration in settings
3. Check error messages in worker logs

## Updates and Maintenance

### Update Application

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build
docker-compose up -d
```

### Database Migrations

```bash
# Apply new migrations
docker-compose exec api migrate -path /migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_DBNAME}" up
```

### Dependency Updates

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update Node dependencies
cd web
npm update
```

## Performance Tuning

### Database

```yaml
# Configure in cloud database console
max_connections: 500
```

### Redis

```yaml
# Configure in cloud Redis console
maxmemory: 256mb
maxmemory-policy: allkeys-lru
```

### Worker

Adjust concurrency in environment variables:
```env
ASYNQ_CONCURRENCY=20
```

## Cost Optimization

1. **Use cloud service auto-scaling** (production)
2. **Auto-scale workers based on queue depth**
3. **Compress stored content**
4. **Regularly clean old data**
5. **Use CDN for static resources**

## Cloud Service Recommendations

### Database
- AWS RDS
- Alibaba Cloud RDS
- Tencent Cloud Database
- Google Cloud SQL

### Redis
- AWS ElastiCache
- Alibaba Cloud Redis
- Tencent Cloud Redis
- Google Cloud Memorystore

### Object Storage
- AWS S3
- Alibaba Cloud OSS
- Tencent Cloud COS
- Google Cloud Storage

### Email Services
- AWS SES
- SendGrid
- Alibaba Cloud Email
- Tencent Cloud SES

## Support

For questions and issues:
- GitHub Issues: https://github.com/your-org/xiaohongshu-image/issues
- Documentation: https://docs.example.com
- Email: support@example.com
