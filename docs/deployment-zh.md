# 部署指南

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 最小4GB内存
- 最小20GB磁盘空间

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/your-org/xiaohongshu-image.git
cd xiaohongshu-image
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑`.env`并选择以下两种方式之一：

#### 方式1：使用Nacos配置中心（推荐）

```env
# Nacos配置中心
NACOS_ADDR=your-nacos-host.com
NACOS_PORT=8848
NACOS_NAMESPACE=public
NACOS_GROUP=DEFAULT_GROUP
NACOS_DATA_ID=xiaohongshu-image
NACOS_USERNAME=nacos
NACOS_PASSWORD=nacos
```

然后在Nacos控制台创建配置（JSON格式）：
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

#### 方式2：不使用Nacos（本地配置）

编辑`.env`并设置外部云服务配置：

```env
# 不设置NACOS_ADDR，系统使用本地配置
NACOS_ADDR=

# 数据库配置 - 外部云服务
DATABASE_HOST=your-mysql-host.com
DATABASE_PORT=3306
DATABASE_USER=your-username
DATABASE_PASSWORD=your-password
DATABASE_DBNAME=xiaohongshu_image

# Redis配置 - 外部云服务
REDIS_HOST=your-redis-host.com
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# MinIO配置 - 外部云服务
MINIO_ENDPOINT=your-minio-host.com:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=generated-content
MINIO_USE_SSL=true
MINIO_REGION=us-east-1

# SMTP配置 - 外部云服务
SMTP_HOST=your-smtp-host.com
SMTP_PORT=587
SMTP_USER=your-smtp-user
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=noreply@xiaohongshu-image.local

# LLM配置
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-your-api-key-here
LLM_MODEL=gpt-4o-mini
```

### 3. 启动应用服务

```bash
docker-compose up -d
```

等待所有服务健康（约30-60秒）。

### 4. 验证服务

```bash
docker-compose ps
```

所有服务应显示"Up"状态。

### 5. 访问应用

- Web UI: http://localhost:31007
- API: http://localhost:31006/healthz
- 外部SMTP服务: 根据配置访问

## 服务详情

### 应用服务

#### API服务
- **端口**: 31006 (外部访问) / 8080 (内部）
- **健康检查**: http://localhost:31006/healthz
- **指标**: http://localhost:31006/metrics

#### Worker服务
- **无暴露端口**
- **从外部Redis队列处理作业**

#### Web服务
- **端口**: 31007 (外部访问) / 3000 (内部）
- **Next.js App Router**

### 外部云服务

#### MySQL
- **配置**: 通过环境变量DATABASE_HOST等配置
- **数据库名**: xiaohongshu_image
- **建议**: 使用云数据库服务（如RDS、阿里云RDS）

#### Redis
- **配置**: 通过环境变量REDIS_HOST等配置
- **建议**: 使用云Redis服务（如ElastiCache、阿里云Redis）

#### MinIO
- **配置**: 通过环境变量MINIO_ENDPOINT等配置
- **建议**: 使用云对象存储（如AWS S3、阿里云OSS）
- **存储桶**: generated-content

#### SMTP
- **配置**: 通过环境变量SMTP_HOST等配置
- **建议**: 使用云邮件服务（如SendGrid、阿里云邮件）

## 生产环境部署

### 安全考虑

1. **使用强密码**
   - 数据库密码
   - Redis密码
   - MinIO访问密钥
   - SMTP密码

2. **启用SSL/TLS**
   - 对API使用HTTPS
   - 为MinIO启用SSL
   - 使用安全SMTP（如587端口 + TLS）

3. **网络隔离**
   - 使用私有网络
   - 仅暴露必要端口
   - 配置防火墙规则

4. **环境变量管理**
   - 永不提交`.env`文件
   - 使用密钥管理（如Docker Secrets、Kubernetes Secrets）
   - 定期轮换API密钥

### 扩展

#### API扩展

```yaml
api:
  deploy:
    replicas: 3
  resources:
    limits:
      cpus: '1'
      memory: 1G
```

#### Worker扩展

```yaml
worker:
  deploy:
    replicas: 5
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

#### 数据库扩展

- 使用MySQL读写分离
- 配置读副本
- 实现连接池

#### Redis扩展

- 使用Redis集群
- 配置持久化（AOF/RDB）
- 设置最大内存策略

### 监控

#### 健康检查

应用服务包含健康检查：

```bash
# 检查所有服务
docker-compose ps

# 检查特定服务
docker-compose exec api curl http://localhost:8080/healthz
```

#### 日志

```bash
# 查看所有日志
docker-compose logs -f

# 查看特定服务
docker-compose logs -f worker

# 查看最近100行
docker-compose logs --tail=100 api
```

#### 指标

- API在`/metrics`暴露Prometheus指标
- 与Prometheus + Grafana集成
- 设置以下告警：
  - 高错误率
  - 队列深度
  - 响应时间

### 备份

#### 数据库备份

```bash
# 使用云数据库提供的备份工具
# 例如：AWS RDS自动备份
```

#### MinIO备份

```bash
# 使用云对象存储提供的备份工具
# 例如：AWS S3版本控制、生命周期策略
```

#### Redis备份

```bash
# 使用云Redis提供的备份工具
# 例如：ElastiCache自动备份
```

## 故障排查

### 服务无法启动

1. 检查日志：
   ```bash
   docker-compose logs <service-name>
   ```

2. 检查资源使用：
   ```bash
   docker stats
   ```

3. 检查端口冲突：
   ```bash
   netstat -tuln | grep LISTEN
   ```

4. 检查外部服务连接：
   ```bash
   # 测试数据库连接
   docker-compose exec api mysql -h ${DATABASE_HOST} -u ${DATABASE_USER} -p${DATABASE_PASSWORD} -e "SELECT 1"
   
   # 测试Redis连接
   docker-compose exec api redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping
   ```

### Nacos连接失败

1. 检查Nacos地址和端口是否正确
2. 检查网络连接是否正常
3. 检查Nacos用户名和密码是否正确
4. 检查Namespace和Group是否正确
5. 查看日志：`docker-compose logs api`

### 数据库连接问题

1. 验证数据库可访问：
   ```bash
   mysql -h ${DATABASE_HOST} -u ${DATABASE_USER} -p
   ```

2. 检查连接字符串配置

3. 验证数据库存在：
   ```bash
   mysql -h ${DATABASE_HOST} -u ${DATABASE_USER} -p -e "SHOW DATABASES;"
   ```

4. 运行数据库迁移：
   ```bash
   docker-compose exec api migrate -path /migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_DBNAME}" up
   ```

### Redis连接问题

1. 验证Redis可访问：
   ```bash
   redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping
   ```

2. 检查队列深度：
   ```bash
   redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} LLEN asynq:queues:default
   ```

### MinIO连接问题

1. 验证MinIO可访问：
   ```bash
   curl http://${MINIO_ENDPOINT}/minio/health/live
   ```

2. 检查存储桶存在：
   ```bash
   mc ls ${MINIO_ENDPOINT}/${MINIO_BUCKET}
   ```

### Worker不处理作业

1. 检查worker日志：
   ```bash
   docker-compose logs worker
   ```

2. 验证Redis连接

3. 检查队列深度：
   ```bash
   redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} KEYS "asynq:*"
   ```

### 邮件未发送

1. 检查SMTP服务日志
2. 验证SMTP配置在设置中
3. 检查worker日志中的错误信息

## 更新和维护

### 更新应用

```bash
# 拉取最新代码
git pull origin main

# 重新构建并重启
docker-compose down
docker-compose build
docker-compose up -d
```

### 数据库迁移

```bash
# 应用新迁移
docker-compose exec api migrate -path /migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_DBNAME}" up
```

### 依赖更新

```bash
# 更新Go依赖
go get -u ./...
go mod tidy

# 更新Node依赖
cd web
npm update
```

## 性能调优

### 数据库

```yaml
# 在云数据库控制台配置
max_connections: 500
```

### Redis

```yaml
# 在云Redis控制台配置
maxmemory: 256mb
maxmemory-policy: allkeys-lru
```

### Worker

调整并发在环境变量中：
```env
ASYNQ_CONCURRENCY=20
```

## 成本优化

1. **使用云服务自动扩展**（生产环境）
2. **根据队列深度自动扩展Worker**
3. **压缩存储内容**
4. **定期清理旧数据**
5. **对静态资源使用CDN**

## 云服务推荐

### 数据库
- AWS RDS
- 阿里云RDS
- 腾讯云数据库
- Google Cloud SQL

### Redis
- AWS ElastiCache
- 阿里云Redis
- 腾讯云Redis
- Google Cloud Memorystore

### 对象存储
- AWS S3
- 阿里云OSS
- 腾讯云COS
- Google Cloud Storage

### 邮件服务
- AWS SES
- SendGrid
- 阿里云邮件
- 腾讯云SES

## 支持

如遇问题和疑问：
- GitHub Issues: https://github.com/your-org/xiaohongshu-image/issues
- 文档: https://docs.example.com
- 邮箱: support@example.com
