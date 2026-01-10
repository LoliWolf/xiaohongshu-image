# 小红书图片生成系统

一个从小红书评论中自动识别图片/视频生成意图，并通过邮件返回结果的完整SaaS原型系统。

## 功能特性

- 🔍 **评论监控**：自动轮询小红书笔记评论，提取生成意图
- 🤖 **意图识别**：结合规则过滤和LLM智能识别，只处理明确意图
- 🎨 **内容生成**：支持多种生成服务提供商（Mock + HTTP）
- 📧 **邮件投递**：自动将生成结果发送到评论中的邮箱
- 🔄 **任务队列**：基于Asynq的可靠异步任务处理
- 📊 **可观测性**：完整的任务状态跟踪和日志记录
- 🔌 **可扩展**：模块化设计，易于对接新的Connector和Provider

## 技术栈

### 后端
- **语言**：Go 1.21
- **框架**：Gin (HTTP), Asynq (任务队列)
- **数据库**：MySQL 8.0 + GORM
- **缓存/队列**：Redis 7
- **对象存储**：MinIO (S3兼容)
- **日志**：Zap
- **配置**：Viper

### 前端
- **框架**：Next.js 14 + React 18
- **样式**：Tailwind CSS
- **HTTP客户端**：Axios

### 基础设施
- **容器化**：Docker + Docker Compose
- **配置中心**：Nacos
- **邮件测试**：Mailhog

## 项目结构

```
xiaohongshu-image/
├── cmd/
│   ├── api/          # API服务入口
│   └── worker/       # Worker服务入口
├── internal/
│   ├── api/          # HTTP API处理器
│   ├── config/       # 配置管理
│   ├── db/           # 数据库层
│   ├── models/       # 数据模型
│   ├── services/
│   │   ├── xhsconnector/  # 小红书Connector (Mock + MCP)
│   │   ├── intent/         # 意图识别服务
│   │   ├── provider/       # 生成任务Provider (Mock + HTTP + Mapping)
│   │   ├── storage/        # MinIO存储服务
│   │   └── mailer/        # 邮件发送服务
│   └── worker/       # Worker作业处理器
├── pkg/
│   └── logger/       # 日志工具
├── web/             # Next.js前端
├── migrations/       # 数据库迁移脚本
├── config/          # 配置文件
├── docker-compose.yml
└── README.md
```

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- (可选) Go 1.21+ (本地开发)
- (可选) Node.js 18+ (本地开发前端)

### 配置外部云服务

系统已配置为连接外部云服务，无需运行内置的MySQL、Redis、MinIO、Mailhog服务。

1. **克隆项目**
```bash
git clone <repository-url>
cd xiaohongshu-image
```

2. **配置环境变量**
```bash
cp .env.example .env
```

编辑 `.env` 文件，选择以下两种方式之一：

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

```env
# 不设置NACOS_ADDR，系统使用本地配置
NACOS_ADDR=

# 数据库配置 - 外部云服务（如阿里云RDS、AWS RDS）
DATABASE_HOST=your-mysql-host.com
DATABASE_PORT=3306
DATABASE_USER=your-username
DATABASE_PASSWORD=your-password
DATABASE_DBNAME=xiaohongshu_image

# Redis配置 - 外部云服务（如阿里云Redis、AWS ElastiCache）
REDIS_HOST=your-redis-host.com
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# MinIO配置 - 外部云服务（如阿里云OSS、AWS S3）
MINIO_ENDPOINT=your-minio-host.com:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=generated-content
MINIO_USE_SSL=true
MINIO_REGION=us-east-1

# SMTP配置 - 外部云服务（如阿里云邮件、AWS SES）
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

3. **启动应用服务**
```bash
docker-compose up -d
```

等待所有服务健康（约30-60秒）。

4. **验证服务**
```bash
docker-compose ps
```

所有服务应显示"Up"状态。

5. **访问应用**
- Web UI: http://localhost:31007
- API: http://localhost:31006/healthz
- 外部SMTP服务: 根据配置访问

## 使用MockConnector跑通全链路

系统默认使用MockConnector，无需小红书登录即可测试完整流程。

### 演示步骤

1. **访问设置页面**：http://localhost:31007/settings
   - 确认Connector Mode为"Mock"
   - 查看其他配置是否正确

2. **手动触发轮询**：点击"Run Poll Now"按钮
   - 系统会拉取Mock评论
   - 包含邮箱和生成关键词的评论会自动创建任务

3. **查看任务列表**：http://localhost:31007/tasks
   - 可以看到新创建的任务
   - 任务状态会自动更新（每5秒刷新）

4. **查看任务详情**：点击任务ID
   - 查看完整的任务信息
   - 包括原始评论、意图识别结果、生成状态等

5. **检查邮件**：根据外部SMTP配置访问
   - 任务完成后会自动发送邮件
   - 邮件包含生成结果的下载链接

### Mock数据说明

MockConnector内置了以下测试评论：

| 用户 | 内容 | 邮箱 | 类型 |
|------|------|------|------|
| 测试用户1 | 帮我画一张可爱的猫咪图片 | test1@example.com | 图片 |
| 测试用户2 | 能生成一个视频吗？主题是海边日落 | contact@demo.com | 视频 |
| 测试用户3 | 这个笔记真好看！ | - | 无意图 |
| 测试用户4 | AI生成一张赛博朋克风格的图片 | myemail@company.com | 图片 |
| 测试用户5 | 做个视频，内容是城市夜景 | sendto@user.org | 视频 |
| 测试用户6 | 出图！风景画，风格是油画 | art@studio.com | 图片 |

## 切换到Real MCP Connector

如需使用真实的小红书数据，需要配置MCP Connector。

### 配置步骤

1. **获取MCP服务器信息**
   - MCP服务器地址（或启动命令）
   - 认证信息（如果需要）

2. **修改设置**
   - 访问 http://localhost:31007/settings
   - 将Connector Mode改为"MCP"
   - 填写MCP服务器URL和认证信息

3. **配置真实笔记**
   - 在Note Target字段填写真实的小红书笔记URL或ID
   - 例如：`https://www.xiaohongshu.com/explore/64a1b2c3d4e5f6`

4. **注意事项**
   - MCP服务器需要实现`xhs_list_comments`工具
   - 工具参数：`note_id_or_url`, `cursor`
   - 返回格式：见[internal/services/xhsconnector/connector.go](internal/services/xhsconnector/connector.go)

### MCP工具规范

**工具名**：`xhs_list_comments`

**请求参数**：
```json
{
  "note_id_or_url": "string",
  "cursor": "string (optional)"
}
```

**响应格式**：
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

## 配置Provider对接新API

系统支持通过配置对接不同的生成API，无需修改代码。

### 配置示例

在设置页面的Provider JSON字段中配置：

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

### Mapping规则

#### Request Mapping
- 使用`$.field`引用统一请求的字段
- 支持嵌套对象和数组
- 示例：
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
- 使用JSONPath提取响应中的字段
- 示例：
  ```json
  {
    "job_id_jsonpath": "$.data.id",
    "status_jsonpath": "$.status",
    "result_url_jsonpath": "$.output.download_url"
  }
  ```

### 统一请求字段

| 字段 | 类型 | 说明 |
|------|------|------|
| request_id | string | 请求唯一标识 |
| type | image/video | 生成类型 |
| prompt | string | 生成描述 |
| negative_prompt | string | 负面提示词（可选） |
| style | string | 风格（可选） |
| width | int | 宽度（可选） |
| height | int | 高度（可选） |
| duration_sec | int | 视频时长（可选） |
| ratio | string | 宽高比（可选） |
| seed | int | 随机种子（可选） |
| extra | map | 扩展字段 |

## 意图识别规则

系统使用两层过滤确保只处理明确的生成意图。

### 第一层：规则过滤

**关键词匹配**：
- 图片关键词：出图、生成图、做图片、帮我画、AI生成、来一张、画一张、生成一张、画个、做个图、出个图、生成个、画一幅、生成一幅
- 视频关键词：做视频、生成视频、做个视频、生成个视频、出视频、来个视频、做短片、生成短片、做个短片

**邮箱提取**：
- 使用正则表达式提取邮箱
- 验证邮箱格式有效性
- 多个邮箱时取第一个

### 第二层：LLM识别

**System Prompt**：
```
你是一个意图抽取器。你只能输出 JSON，不能输出任何解释、Markdown、代码块。
请从评论中判断是否存在明确的"生成图片/生成视频"请求，并抽取用于生成模型的 prompt，
同时抽取邮箱（如果存在）。不确定时必须返回 has_request=false。
```

**输出格式**：
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

**明确意图判定**（必须全部满足）：
- `has_request = true`
- `request_type` 为 "image" 或 "video"
- `prompt` 非空且长度 >= 8
- `email` 为有效邮箱
- `confidence >= threshold`（默认0.7）

## API文档

### 健康检查
```
GET /healthz
```

### 获取设置
```
GET /api/settings
```

### 更新设置
```
PUT /api/settings
Content-Type: application/json

{
  "connector_mode": "mock|mcp",
  "note_target": "string",
  "polling_interval_sec": 120,
  "llm_base_url": "string",
  "llm_api_key": "string",
  "llm_model": "string",
  "intent_threshold": 0.7,
  "smtp_host": "string",
  "smtp_port": 1025,
  "smtp_from": "string",
  "provider_json": "string"
}
```

### 手动触发轮询
```
POST /api/poll/run
```

### 获取任务列表
```
GET /api/tasks?limit=100&offset=0
```

### 获取任务详情
```
GET /api/tasks/:id
```

### 下载文件
```
GET /api/files/:key
```

## 数据库模型

### settings
系统配置表，单行记录。

### notes
笔记跟踪表，记录轮询状态和游标。

### comments
评论表，存储从小红书拉取的评论。

### tasks
任务表，记录生成任务的状态和结果。

### deliveries
邮件投递表，记录邮件发送状态。

### audit_logs
审计日志表，记录系统事件。

## 开发指南

### 本地开发（后端）

```bash
# 安装依赖
go mod download

# 运行API服务
go run cmd/api/main.go

# 运行Worker服务
go run cmd/worker/main.go
```

### 本地开发（前端）

```bash
cd web

# 安装依赖
npm install

# 运行开发服务器
npm run dev
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/services/intent
```

### 数据库迁移

```bash
# 应用迁移
migrate -path migrations -database "mysql://root:rootpassword@tcp(localhost:3306)/xiaohongshu_image" up

# 回滚迁移
migrate -path migrations -database "mysql://root:rootpassword@tcp(localhost:3306)/xiaohongshu_image" down
```

## 故障排查

### 服务无法启动

1. 检查端口是否被占用
2. 查看容器日志：`docker-compose logs <service-name>`
3. 确认依赖服务已就绪（外部云服务）

### Nacos连接失败

1. 检查Nacos地址和端口是否正确
2. 检查网络连接是否正常
3. 检查Nacos用户名和密码是否正确
4. 检查Namespace和Group是否正确
5. 查看日志：`docker-compose logs api`

### 任务状态卡住

1. 检查Worker日志：`docker-compose logs worker`
2. 查看外部Redis队列状态
3. 确认Provider配置正确

### 邮件未发送

1. 检查外部SMTP服务
2. 查看SMTP配置
3. 检查Worker日志中的错误信息

### LLM调用失败

1. 确认API密钥正确
2. 检查Base URL和Model配置
3. 查看API额度是否充足

## 性能优化

- **并发控制**：调整Asynq的concurrency参数
- **队列优先级**：使用critical/default/low队列
- **缓存策略**：Redis缓存频繁访问的数据
- **数据库索引**：确保关键字段有索引

## 安全建议

- 不要在生产环境使用默认密码
- 使用HTTPS保护API
- 限制API访问频率
- 定期更新依赖包
- 启用日志审计

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题，请通过以下方式联系：
- 提交GitHub Issue
- 发送邮件至：support@example.com

## 相关文档

- [快速开始指南](docs/quick-start-zh.md) - 5分钟快速上手
- [系统架构](docs/architecture-zh.md) - 架构设计详解
- [部署指南](docs/deployment-zh.md) - 生产环境部署
- [项目结构](docs/project-structure-zh.md) - 代码结构说明
- [验收标准](docs/acceptance-criteria-zh.md) - 功能验收清单
