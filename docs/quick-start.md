# Quick Start Guide

## 5-Minute Quick Start

### Prerequisites
- Docker and Docker Compose installed
- OpenAI API key (or compatible LLM API)
- Nacos configuration center (optional)

### Step 1: Clone and Configure

```bash
git clone <repository-url>
cd xiaohongshu-image

# Copy environment template
cp .env.example .env
```

### Step 2: Configure Environment Variables

Edit `.env` file and choose one of the following two methods:

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
MINIO_USE_SSL=true

# SMTP configuration - external cloud services (e.g., Alibaba Cloud Email, AWS SES)
SMTP_HOST=your-smtp-host.com
SMTP_PORT=587
SMTP_USER=your-smtp-user
SMTP_PASSWORD=your-smtp-password

# LLM configuration
LLM_API_KEY=sk-your-api-key-here
```

### Step 3: Start Application Services

```bash
docker-compose up -d
```

Wait for all services to be healthy (about 30-60 seconds).

### Step 4: Verify Services

```bash
# Check service status
docker-compose ps

# All services should show "Up" status
```

### Step 5: Access the Application

- **Web UI**: http://localhost:31007
- **API Health**: http://localhost:31006/healthz
- **External SMTP Service**: Access based on configuration

### Step 6: Test the Workflow

1. **Open Settings Page**: http://localhost:31007/settings
   - Confirm Connector Mode is "Mock"
   - Click "Run Poll Now"

2. **View Tasks**: http://localhost:31007/tasks
   - Wait for tasks to appear (5-10 seconds)
   - Enable "Auto-refresh" for live updates

3. **Monitor Task Progress**
   - Tasks will progress through: EXTRACTED → SUBMITTED → RUNNING → SUCCEEDED → EMAILED
   - Full cycle takes 10-30 seconds per task

4. **Check Email**: Access based on external SMTP configuration
   - You should see emails with download links
   - Click the link to verify it works

## What Happens Behind the Scenes

1. **PollJob** runs and fetches 6 mock comments
2. **IntentJob** analyzes each comment:
   - 4 comments have clear intent + email → Tasks created
   - 2 comments lack intent/email → Skipped
3. **SubmitJob** submits tasks to MockProvider
4. **StatusJob** polls provider status every 15-30 seconds
5. **EmailJob** sends results when tasks complete

## Mock Data

The system includes 6 pre-configured mock comments:

| # | User | Content | Email | Type | Result |
|---|-------|----------|--------|-------|---------|
| 1 | 测试用户1 | 帮我画一张可爱的猫咪图片 | test1@example.com | Image | ✅ Processed |
| 2 | 测试用户2 | 能生成一个视频吗？主题是海边日落 | contact@demo.com | Video | ✅ Processed |
| 3 | 测试用户3 | 这个笔记真好看！ | - | None | ❌ Skipped |
| 4 | 测试用户4 | AI生成一张赛博朋克风格的图片 | myemail@company.com | Image | ✅ Processed |
| 5 | 测试用户5 | 做个视频，内容是城市夜景 | sendto@user.org | Video | ✅ Processed |
| 6 | 测试用户6 | 出图！风景画，风格是油画 | art@studio.com | Image | ✅ Processed |

## Switching to Real MCP Connector

To use real Xiaohongshu data, configure the MCP Connector.

### Configuration Steps

1. **Get MCP Server Information**
   - MCP server address (or startup command)
   - Authentication information (if required)

2. **Update Settings**
   - Visit http://localhost:31007/settings
   - Change Connector Mode to "MCP"
   - Fill in MCP server URL and authentication info

3. **Configure Real Note**
   - Fill in real Xiaohongshu note URL or ID in Note Target field
   - Example: `https://www.xiaohongshu.com/explore/64a1b2c3d4e5f6`

4. **Important Notes**
   - MCP server must implement `xhs_list_comments` tool
   - Tool parameters: `note_id_or_url`, `cursor`
   - Return format: See [internal/services/xhsconnector/connector.go](internal/services/xhsconnector/connector.go)

### MCP Tool Specification

**Tool Name**: `xhs_list_comments`

**Request Parameters**:
```json
{
  "note_id_or_url": "string",
  "cursor": "string (optional)"
}
```

**Response Format**:
```json
{
  "comments": [
    {
      "comment_id": "string",
      "user_name": "string",
      "content": "string",
      "comment_created_at": "ISO8601 timestamp"
    }
  ],
  "next_cursor": "string",
  "has_more": "boolean"
}
```

## Configuring Provider to Connect to New APIs

The system supports connecting to different generation APIs through configuration without code changes.

### Configuration Example

Configure in the Provider JSON field on the settings page:

```json
[
  {
    "provider_name": "my-custom-api",
    "type": "both",
    "base_url": "https://api.example.com/v1",
    "api_key": "your-api-key",
    "submit_path": "/generate",
    "status_path_template": "/jobs/{id}",
    "headers": {
      "X-Custom-Header": "value"
    },
    "request_mapping": {
      "prompt_text": "$.prompt",
      "style": "$.style",
      "width": "$.width",
      "height": "$.height",
      "duration": "$.duration_sec"
    },
    "response_mapping": {
      "job_id_jsonpath": "$.data.job_id"
    },
    "status_mapping": {
      "status_jsonpath": "$.status",
      "result_url_jsonpath": "$.result.url"
    }
  }
]
```

### Mapping Rules

#### Request Mapping
- Use `$.field` to reference unified request fields
- Supports nested objects and arrays
- Example:
  ```json
  {
    "input": {
      "text": "$.prompt",
      "negative": "$.negative_prompt"
    },
    "settings": {
      "width": "$.width",
      "height": "$.height"
    }
  }
  ```

#### Response Mapping
- Use JSONPath to extract fields from response
- Example:
  ```json
  {
    "job_id_jsonpath": "$.data.id",
    "status_jsonpath": "$.status",
    "result_url_jsonpath": "$.output.download_url"
  }
  ```

### Unified Request Fields

| Field | Type | Description |
|-------|------|-------------|
| request_id | string | Unique request identifier |
| type | image/video | Generation type |
| prompt | string | Generation description |
| negative_prompt | string | Negative prompt (optional) |
| style | string | Style (optional) |
| width | int | Width (optional) |
| height | int | Height (optional) |
| duration_sec | int | Video duration (optional) |
| ratio | string | Aspect ratio (optional) |
| seed | int | Random seed (optional) |
| extra | map | Extra fields |

## Intent Recognition Rules

The system uses two layers of filtering to ensure only clear generation intents are processed.

### Layer 1: Rule-based Filtering

**Keyword Matching**:
- Image keywords: 出图, 生成图, 做图片, 帮我画, AI生成, 来一张, 画一张, 生成一张, 画个, 做个图, 出个图, 生成个, 画一幅, 生成一幅
- Video keywords: 做视频, 生成视频, 做个视频, 生成个视频, 出视频, 来个视频, 做短片, 生成短片, 做个短片

**Email Extraction**:
- Use regex to extract email
- Validate email format
- Take the first one if multiple

### Layer 2: LLM Recognition

**System Prompt**:
```
You are an intent extractor. You can only output JSON, not any explanation, Markdown, or code blocks.
Please determine if there is a clear "generate image/generate video" request from the comment,
and extract the prompt for the generation model,
also extract the email (if exists). You must return has_request=false when uncertain.
```

**Output Format**:
```json
{
  "has_request": boolean,
  "request_type": "image"|"video"|"unknown",
  "prompt": string,
  "email": string|null,
  "confidence": number (0..1),
  "reason": string
}
```

**Clear Intent Determination** (all must be met):
- `has_request = true`
- `request_type` is "image" or "video"
- `prompt` is non-empty and length >= 8
- `email` is valid email
- `confidence >= threshold` (default 0.7)

## Troubleshooting

### Services Won't Start

1. Check if ports are occupied
2. View container logs: `docker-compose logs <service-name>`
3. Confirm dependent services are ready (external cloud services)

### Nacos Connection Failed

1. Check if Nacos address and port are correct
2. Check if network connection is normal
3. Check if Nacos username and password are correct
4. Check if Namespace and Group are correct
5. View logs: `docker-compose logs api`

### Task Status Stuck

1. Check Worker logs: `docker-compose logs worker`
2. Check external Redis queue status
3. Confirm Provider configuration is correct

### Email Not Sent

1. Check external SMTP service
2. View SMTP configuration
3. Check error messages in Worker logs

### LLM Call Failed

1. Confirm API key is correct
2. Check Base URL and Model configuration
3. Check if API quota is sufficient

## Access Credentials

| Service | URL | Username | Password |
|----------|------|----------|----------|
| Web UI | http://localhost:31007 | - | - |
| API | http://localhost:31006 | - | - |
| MySQL | ${DATABASE_HOST} | ${DATABASE_USER} | ${DATABASE_PASSWORD} |
| Redis | ${REDIS_HOST} | - | ${REDIS_PASSWORD} |
| MinIO | ${MINIO_ENDPOINT} | ${MINIO_ACCESS_KEY} | ${MINIO_SECRET_KEY} |
| SMTP | ${SMTP_HOST} | ${SMTP_USER} | ${SMTP_PASSWORD} |

## Next Steps

- Check [Nacos Configuration Guide](nacos-config-zh.md) to learn how to configure Nacos
- Check [Deployment Guide](deployment-zh.md) for production deployment
- Check [System Architecture](architecture-zh.md) to understand system design

## Support

For questions and issues:
- GitHub Issues: https://github.com/your-org/xiaohongshu-image/issues
- Documentation: https://docs.example.com
- Email: support@example.com
