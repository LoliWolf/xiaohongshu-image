# 系统架构

## 系统概览

小红书图片生成系统是一个分布式微服务架构，自动从小红书评论中检测图片/视频生成请求，并通过邮件返回结果。

## 组件

### 1. API服务 (cmd/api)
- 基于Gin框架构建的HTTP服务器
- 配置和任务管理的RESTful API端点
- 集成数据库、Redis队列和Worker
- 服务Next.js前端

### 2. Worker服务 (cmd/worker)
- 基于Asynq的后台作业处理器
- 处理多种作业类型：
  - PollJob: 从小红书获取评论
  - IntentJob: 提取生成意图
  - SubmitJob: 提交给生成提供商
  - StatusJob: 检查生成状态
  - EmailJob: 通过邮件发送结果

### 3. 前端 (web/)
- Next.js 14 + App Router
- React 18 + TypeScript
- Tailwind CSS样式
- 页面：首页、设置、任务列表、任务详情

### 4. 数据库 (MySQL)
- 存储设置、笔记、评论、任务、投递记录、审计日志
- GORM ORM进行数据库操作
- 使用golang-migrate进行迁移

### 5. Redis
- Asynq作业队列
- 限流
- 分布式锁
- 缓存

### 6. MinIO
- S3兼容的对象存储
- 存储生成的内容
- 提供下载的签名URL

### 7. Mailhog
- 用于测试的SMTP服务器
- 查看邮件的Web UI
- 开发环境不进行真实邮件投递

## 数据流

```
1. PollJob (定时任务)
   └─> Connector.ListComments()
       └─> 保存到数据库
           └─> 入队 IntentJob

2. IntentJob
   └─> 规则过滤（关键词 + 邮箱）
       └─> LLM提取
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

## 关键设计决策

### 1. 使用Asynq作为作业队列
- **原因**：基于Redis，支持重试、延迟和优先级
- **优势**：可靠、分布式、易于监控
- **考虑的替代方案**：Watermill、RabbitMQ

### 2. 幂等性策略
- **评论级别**：`comment_uid`唯一索引防止重复
- **任务级别**：`comment_id`唯一确保每条评论一个任务
- **作业级别**：Asynq jobId确保作业不重复
- **邮件级别**：Redis限流防止垃圾邮件

### 3. Provider映射机制
- **原因**：支持多个API而无需代码更改
- **实现**：基于JSONPath的请求/响应映射
- **优势**：易于添加新提供商、配置驱动

### 4. 两层意图检测
- **第一层**：基于规则的过滤（关键词 + 邮箱正则）
- **第二层**：基于LLM的提取和置信度评分
- **原因**：降低LLM成本，提高准确性

### 5. Mock Connector和Provider
- **原因**：无需外部依赖即可进行全栈测试
- **优势**：快速开发、可靠演示、易于CI/CD

## 可扩展性考虑

### 水平扩展
- API：无状态，可水平扩展
- Worker：多个实例可并行处理作业
- 数据库：读副本用于查询
- Redis：集群模式实现高可用

### 性能优化
- 数据库连接池
- 频繁查询的Redis缓存
- 评论批处理
- 异步作业处理

### 监控
- 使用Zap的结构化日志
- Prometheus指标（计划中）
- 健康检查端点
- 任务状态跟踪

## 安全考虑

### API安全
- 使用go-playground/validator进行输入验证
- 通过GORM防止SQL注入
- 端点限流

### 数据保护
- 邮箱地址安全存储
- 带有过期时间的签名URL
- 日志中无敏感数据

### 认证（未来）
- 基于JWT的认证
- 基于角色的访问控制
- API密钥管理

## 部署架构

### 开发环境
```
Docker Compose
├── MySQL（单实例）
├── Redis（单实例）
├── MinIO（单实例）
├── Mailhog（SMTP测试）
├── API（单实例）
├── Worker（单实例）
└── Web（单实例）
```

### 生产环境（推荐）
```
Kubernetes / Docker Swarm
├── MySQL集群（主 + 副本）
├── Redis集群（哨兵）
├── MinIO（分布式模式）
├── SMTP服务（SendGrid/Ses）
├── API（多实例 + 负载均衡）
├── Worker（多实例）
├── Web（CDN + 静态托管）
└── 监控（Prometheus + Grafana）
```

## 错误处理

### 重试策略
- LLM调用：2次重试，1秒延迟
- Provider提交：不重试（幂等）
- Provider状态：20次重试，指数退避（15s, 30s, 60s...）
- 邮件：不重试（用户可以再次请求）

### 死信队列
- 失败的作业被记录
- 用于调试的审计跟踪
- 手动重试功能（未来）

## 测试策略

### 单元测试
- 服务层逻辑
- 意图提取
- Provider映射
- 邮箱验证

### 集成测试
- 数据库操作
- Redis队列
- API端点

### 端到端测试
- 使用MockConnector的完整工作流
- 邮件投递验证
- 任务状态转换

## 未来增强

1. **多用户支持**
   - 用户认证
   - 每用户设置
   - 使用配额

2. **高级功能**
   - 回复小红书评论
   - Webhook通知
   - 批处理
   - 自定义模板

3. **监控与分析**
   - Prometheus指标
   - Grafana仪表板
   - 使用分析
   - 成本跟踪

4. **性能**
   - GraphQL API
   - WebSocket实时更新
   - CDN静态资源
   - 数据库分片
