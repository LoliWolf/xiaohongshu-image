# 项目结构

```
xiaohongshu-image/
├── cmd/                           # 应用程序入口点
│   ├── api/
│   │   └── main.go               # API服务器（Gin + Asynq调度器）
│   └── worker/
│       └── main.go               # Worker服务器（Asynq处理器）
│
├── internal/                      # 私有应用代码
│   ├── api/
│   │   └── handler.go           # HTTP处理器（Gin路由）
│   │
│   ├── config/
│   │   └── config.go           # 配置管理（Viper）
│   │
│   ├── db/
│   │   └── database.go          # 数据库层（GORM）
│   │
│   ├── models/
│   │   └── models.go           # 数据模型（Settings, Note, Comment, Task, Delivery, AuditLog）
│   │
│   ├── services/
│   │   ├── xhsconnector/       # 小红书connector抽象
│   │   │   ├── connector.go    # Connector接口
│   │   │   ├── mock.go        # Mock connector用于测试
│   │   │   └── mcp.go         # MCP connector用于真实数据
│   │   │
│   │   ├── intent/             # 意图识别服务
│   │   │   ├── intent.go      # 规则 + LLM提取
│   │   │   └── http.go       # LLM API的HTTP客户端
│   │   │
│   │   ├── provider/           # 生成provider抽象
│   │   │   ├── provider.go    # Provider接口 + UnifiedGenRequest
│   │   │   ├── mock.go        # Mock provider用于测试
│   │   │   ├── http.go        # HTTP provider用于真实API
│   │   │   ├── mapper.go      # 基于JSONPath的请求/响应映射
│   │   │   └── storage.go     # 存储接口
│   │   │
│   │   ├── storage/            # 对象存储服务
│   │   │   └── minio.go       # MinIO实现
│   │   │
│   │   └── mailer/            # 邮件服务
│   │       └── mailer.go      # SMTP邮件发送
│   │
│   └── worker/
│       └── worker.go           # Asynq作业处理器
│           # PollJob, IntentJob, SubmitJob, StatusJob, EmailJob
│
├── pkg/                          # 公共库代码
│   └── logger/
│       └── logger.go           # 日志工具（Zap）
│
├── web/                          # Next.js前端
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx      # 根布局和导航
│   │   │   ├── page.tsx        # 首页
│   │   │   ├── settings/
│   │   │   │   └── page.tsx  # 设置页面
│   │   │   ├── tasks/
│   │   │   │   ├── page.tsx  # 任务列表页面
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx # 任务详情页面
│   │   │   └── globals.css    # 全局样式
│   │   └── lib/
│   │       └── api.ts         # API客户端（Axios）
│   ├── public/                  # 静态资源
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.js
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   └── Dockerfile
│
├── migrations/                   # 数据库迁移
│   ├── 000001_init.up.sql      # 创建表
│   └── 000001_init.down.sql    # 删除表
│
├── config/                       # 配置文件
│   └── config.yaml             # 主配置（Viper）
│
├── docs/                         # 文档
│   ├── architecture-zh.md    # 系统架构（中文）
│   ├── deployment-zh.md      # 部署指南（中文）
│   ├── acceptance-criteria-zh.md # 验收标准（中文）
│   ├── quick-start-zh.md     # 快速开始（中文）
│   └── project-structure.md # 项目结构（英文）
│
├── .env.example                 # 环境变量模板
├── .gitignore                  # Git忽略规则
├── go.mod                      # Go模块定义
├── go.sum                      # Go依赖校验
├── Makefile                    # 构建自动化
├── docker-compose.yml           # Docker Compose配置
├── Dockerfile.api              # API服务器Dockerfile
├── Dockerfile.worker           # Worker服务器Dockerfile
├── README.md                   # 项目文档（英文）
└── README-zh.md               # 项目文档（中文）
```

## 关键组件

### 后端服务

1. **API服务器** (`cmd/api/main.go`)
   - Gin HTTP服务器
   - RESTful API端点
   - 用于定期轮询的Asynq调度器
   - 服务Next.js前端

2. **Worker服务器** (`cmd/worker/main.go`)
   - Asynq作业处理器
   - 5种作业类型：Poll, Intent, Submit, Status, Email
   - 支持分布式处理

### 核心服务

1. **XHS Connector** (`internal/services/xhsconnector/`)
   - 评论获取的抽象接口
   - 用于演示/测试的Mock connector
   - 用于真实数据的MCP connector

2. **意图识别** (`internal/services/intent/`)
   - 两层过滤（规则 + LLM）
   - 邮箱提取和验证
   - 置信度评分

3. **生成Provider** (`internal/services/provider/`)
   - 统一的请求/响应接口
   - 用于演示/测试的Mock provider
   - 带JSONPath映射的HTTP provider
   - 可配置以支持任何API

4. **存储** (`internal/services/storage/`)
   - MinIO S3兼容存储
   - 签名URL生成
   - 文件上传/下载

5. **邮件** (`internal/services/mailer/`)
   - SMTP邮件发送
   - 基于模板的邮件
   - 错误处理

### 前端

1. **Next.js应用** (`web/src/app/`)
   - 带概览的首页
   - 配置页面
   - 带自动刷新的任务列表
   - 带完整信息的任务详情

2. **API客户端** (`web/src/lib/api.ts`)
   - 基于Axios的HTTP客户端
   - TypeScript接口
   - 错误处理

### 基础设施

1. **数据库** (MySQL 8.0)
   - GORM ORM
   - 使用golang-migrate的迁移
   - 6个表：settings, notes, comments, tasks, deliveries, audit_logs

2. **队列** (Redis 7)
   - Asynq作业队列
   - 限流
   - 分布式锁

3. **存储** (MinIO)
   - S3兼容的对象存储
   - 下载的签名URL
   - Web控制台在:9001

4. **邮件** (Mailhog)
   - 用于测试的SMTP服务器
   - Web UI在:8025
   - 开发环境不进行真实投递

## 数据流

```
1. PollJob（定时任务）
   └─> Connector.ListComments()
       └─> 保存到DB（comments表）
           └─> 入队 IntentJob

2. IntentJob
   └─> 规则过滤（关键词 + 邮箱）
       └─> LLM提取（OpenAI兼容）
           └─> 创建任务（如果意图明确）
               └─> 入队 SubmitJob

3. SubmitJob
   └─> Provider.Submit()
       └─> 保存provider_job_id
           └─> 入队 StatusJob（延迟）

4. StatusJob（带退避重试）
   └─> Provider.Status()
       ├─> 如果SUCCEEDED:
       │   └─> 保存result_url
       │       └─> 入队 EmailJob
       ├─> 如果FAILED:
       │   └─> 标记任务为FAILED
       └─> 如果RUNNING/PENDING:
           └─> 重新入队 StatusJob

5. EmailJob
   └─> 限流检查
       └─> Mailer.Send()
           └─> 保存投递记录
               └─> 标记任务为EMAILED
```

## 配置

### 环境变量

- `SERVER_PORT`: API服务器端口（默认：8080）
- `DATABASE_HOST`: MySQL主机
- `DATABASE_PORT`: MySQL端口
- `DATABASE_USER`: MySQL用户
- `DATABASE_PASSWORD`: MySQL密码
- `DATABASE_DBNAME`: 数据库名称
- `REDIS_HOST`: Redis主机
- `REDIS_PORT`: Redis端口
- `MINIO_ENDPOINT`: MinIO端点
- `MINIO_ACCESS_KEY`: MinIO访问密钥
- `MINIO_SECRET_KEY`: MinIO秘密密钥
- `LLM_BASE_URL`: LLM API基础URL
- `LLM_API_KEY`: LLM API密钥
- `LLM_MODEL`: LLM模型名称
- `SMTP_HOST`: SMTP主机
- `SMTP_PORT`: SMTP端口
- `SMTP_FROM`: SMTP发件人地址
- `ASYNQ_REDIS_ADDR`: Asynq的Redis地址

### 配置文件

`config/config.yaml`包含可以被环境变量覆盖的默认值。

## 部署

### 开发环境

```bash
docker-compose up -d
```

### 生产环境

- 使用Kubernetes或Docker Swarm
- 启用SSL/TLS
- 配置密钥管理
- 设置监控和告警
- 使用托管服务（RDS, ElastiCache, S3）

## 测试

### 单元测试

```bash
go test ./...
```

### 集成测试

```bash
go test -tags=integration ./...
```

### 端到端测试

```bash
# 启动服务
docker-compose up -d

# 运行E2E测试
go test -tags=e2e ./tests/e2e
```

## 监控

### 健康检查

- API: `GET /healthz`
- 所有服务都有Docker健康检查

### 指标

- `/metrics`的Prometheus指标
- 使用Zap的结构化日志
- 数据库中的任务状态跟踪

### 日志

```bash
# 查看所有日志
docker-compose logs -f

# 查看特定服务
docker-compose logs -f worker
```

## 扩展

### 水平扩展

- API：无状态，可运行多个实例
- Worker：多个实例并行处理作业
- 数据库：使用读副本
- Redis：使用集群模式

### 垂直扩展

- 增加Asynq并发
- 增加数据库连接池
- 增加worker资源（CPU/内存）
