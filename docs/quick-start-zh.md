# 快速开始指南

## 5分钟快速开始

### 前置要求
- 已安装Docker和Docker Compose
- OpenAI API密钥（或兼容的LLM API）
- Nacos配置中心（可选）

### 第1步：克隆和配置

```bash
git clone <repository-url>
cd xiaohongshu-image

# 复制环境变量模板
cp .env.example .env
```

### 第2步：配置环境变量

编辑`.env`文件，选择以下两种方式之一：

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
MINIO_USE_SSL=true

# SMTP配置 - 外部云服务（如阿里云邮件、AWS SES）
SMTP_HOST=your-smtp-host.com
SMTP_PORT=587
SMTP_USER=your-smtp-user
SMTP_PASSWORD=your-smtp-password

# LLM配置
LLM_API_KEY=sk-your-api-key-here
```

### 第3步：启动应用服务

```bash
docker-compose up -d
```

等待所有服务健康（约30-60秒）。

### 第4步：验证服务

```bash
# 检查服务状态
docker-compose ps

# 所有服务应显示"Up"状态
```

### 第5步：访问应用

- **Web UI**: http://localhost:31007
- **API健康检查**: http://localhost:31006/healthz
- **外部SMTP服务**: 根据配置访问

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

## 访问凭据

| 服务 | URL | 用户名 | 密码 |
|----------|------|----------|----------|
| Web UI | http://localhost:31007 | - | - |
| API | http://localhost:31006 | - | - |
| MySQL | ${DATABASE_HOST} | ${DATABASE_USER} | ${DATABASE_PASSWORD} |
| Redis | ${REDIS_HOST} | - | ${REDIS_PASSWORD} |
| MinIO | ${MINIO_ENDPOINT} | ${MINIO_ACCESS_KEY} | ${MINIO_SECRET_KEY} |
| SMTP | ${SMTP_HOST} | ${SMTP_USER} | ${SMTP_PASSWORD} |

## 下一步

- 查看[Nacos配置指南](nacos-config-zh.md)了解如何配置Nacos
- 查看[部署指南](deployment-zh.md)了解生产环境部署
- 查看[系统架构](architecture-zh.md)了解系统设计

## 支持

如遇问题和疑问：
- GitHub Issues: https://github.com/your-org/xiaohongshu-image/issues
- 文档: https://docs.example.com
- 邮箱: support@example.com
