# 验收标准

本文档概述小红书图片生成系统的验收标准。

## 1. MockConnector完整工作流

### 标准
- [ ] 系统能"拉取"新评论并存储到数据库
- [ ] comments表显示轮询后的新条目
- [ ] 可通过API手动触发轮询
- [ ] 轮询按计划自动运行

### 验证步骤

1. 启动所有服务：
   ```bash
   docker-compose up -d
   ```

2. 访问设置页面：http://localhost:3000/settings
   - 确认Connector模式为"mock"

3. 点击"Run Poll Now"按钮

4. 检查数据库：
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT * FROM comments;"
   ```

5. 验证评论已存储，数据正确：
   - comment_uid
   - user_name
   - content
   - ingested_at

**预期结果**：数据库中应存储6条mock评论。

---

## 2. 包含邮箱+明确关键词的评论

### 标准
- [ ] IntentJob通过并创建任务
- [ ] 任务状态为EXTRACTED
- [ ] 邮箱被正确提取
- [ ] Prompt被正确提取
- [ ] 置信度分数 >= 阈值

### 验证步骤

1. 确保轮询已运行（见AC #1）

2. 访问任务页面：http://localhost:3000/tasks

3. 验证为以下评论创建了任务：
   - 邮箱地址
   - 生成关键词（出图、生成图、做视频等）

4. 点击任务查看详情

5. 验证任务详情：
   - 状态：EXTRACTED
   - 邮箱：有效邮箱地址
   - Prompt：从评论中提取
   - 置信度：>= 0.7
   - 请求类型：image或video

**预期结果**：应创建4个任务（评论1、2、4、5、6）。

---

## 3. MockProvider生成

### 标准
- [ ] 任务状态在10-30秒内变为SUCCEEDED
- [ ] result_url被填充
- [ ] result_url可访问（MinIO签名链接或/api/files）
- [ ] Provider作业ID被存储

### 验证步骤

1. 在任务页面监控一个任务（启用自动刷新）

2. 等待状态从EXTRACTED → SUBMITTED → RUNNING → SUCCEEDED

3. 点击任务查看详情

4. 验证：
   - 状态：SUCCEEDED
   - Provider名称：mock
   - Provider作业ID：mock_job_*
   - 结果URL：http://localhost:9000/... (MinIO URL)

5. 点击结果URL验证其可访问

**预期结果**：任务在10-30秒内完成，具有有效的结果URL。

---

## 4. 邮件投递

### 标准
- [ ] Mailhog/SMTP收到包含result_url的邮件
- [ ] 任务状态变为EMAILED
- [ ] deliveries表记录SENT状态
- [ ] 邮件内容包含：
  - 请求类型（图片/视频）
  - Prompt
  - 下载链接
  - 过期通知

### 验证步骤

1. 等待任务达到SUCCEEDED状态

2. 访问Mailhog：http://localhost:8025

3. 验证收到邮件：
   - 检查收件箱
   - 验证发件人：noreply@xiaohongshu-image.local
   - 验证收件人匹配评论邮箱

4. 打开邮件并验证内容：
   - 主题：您的图片生成结果已就绪 / 您的视频生成结果已就绪
   - 正文包含prompt
   - 正文包含下载链接
   - 正文提到1小时过期

5. 检查任务详情页面：
   - 状态：EMAILED
   - 投递部分显示SENT状态
   - 发送时间戳已填充

6. 检查数据库：
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT * FROM deliveries;"
   ```

**预期结果**：Mailhog中收到正确内容的邮件，任务状态为EMAILED。

---

## 5. 幂等性

### 标准
- [ ] 相同的comment_uid不创建重复任务
- [ ] 相同的comment_uid不发送重复邮件
- [ ] 数据库约束防止重复

### 验证步骤

1. 多次点击"Run Poll Now"

2. 检查comments表：
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT COUNT(*) FROM comments;"
   ```

3. 检查tasks表：
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT COUNT(*) FROM tasks;"
   ```

4. 检查deliveries表：
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT COUNT(*) FROM deliveries;"
   ```

5. 验证重复轮询后计数不增加

**预期结果**：多次轮询后计数保持相同。

---

## 6. 错误可见性

### 标准
- [ ] LLM/Provider故障将任务状态设置为FAILED
- [ ] 错误消息被记录
- [ ] 队列不无限重试
- [ ] 错误在任务详情中可见
- [ ] 审计日志记录错误事件

### 验证步骤

#### 场景1：LLM故障

1. 在设置中设置无效的LLM API密钥
2. 触发轮询
3. 检查任务状态
4. 验证错误消息可见

#### 场景2：Provider故障

1. 配置无效的Provider URL
2. 等待任务处理
3. 检查任务状态
4. 验证错误消息可见

#### 场景3：超过最大重试次数

1. 监控持续失败的任务
2. 等待20次状态检查
3. 验证任务状态变为FAILED
4. 验证错误："max retries exceeded"

**预期结果**：所有失败都被正确记录并在UI中可见。

---

## 7. 配置管理

### 标准
- [ ] 可通过API获取设置
- [ ] 可通过API更新设置
- [ ] 更改在重启后持久化
- [ ] UI反映当前设置

### 验证步骤

1. 访问设置页面：http://localhost:3000/settings

2. 修改设置（例如，polling_interval_sec）

3. 点击"Save Settings"

4. 刷新页面并验证更改持久化

5. 重启服务：
   ```bash
   docker-compose restart api worker
   ```

6. 再次访问设置页面

7. 验证设置仍然更新

**预期结果**：设置正确持久化。

---

## 8. 真实MCP Connector（可选）

### 标准
- [ ] 可切换到MCP模式
- [ ] 可配置MCP服务器URL
- [ ] 系统连接到MCP服务器
- [ ] 从真实源获取评论

### 验证步骤

1. 访问设置页面

2. 将Connector模式更改为"mcp"

3. 输入MCP服务器URL

4. 保存设置

5. 触发轮询

6. 检查MCP连接日志

**预期结果**：系统连接到MCP服务器并获取真实评论。

---

## 9. Provider映射

### 标准
- [ ] 可配置自定义provider
- [ ] 请求映射正确工作
- [ ] 响应映射正确工作
- [ ] 状态映射正确工作

### 验证步骤

1. 访问设置页面

2. 在Provider JSON中配置自定义provider：

```json
[
  {
    "provider_name": "test-provider",
    "type": "both",
    "base_url": "https://httpbin.org",
    "api_key": "",
    "submit_path": "/post",
    "status_path_template": "/anything/{id}",
    "headers": {},
    "request_mapping": {
      "data": {
        "prompt_text": "$.prompt",
        "type": "$.type"
      }
    },
    "response_mapping": {
      "job_id_jsonpath": "$.json.data"
    },
    "status_mapping": {
      "status_jsonpath": "$.status",
      "result_url_jsonpath": "$.url"
    }
  }
]
```

3. 保存设置

4. 触发轮询以创建任务

5. 监控任务处理

**预期结果**：Provider映射正确转换请求/响应。

---

## 10. 性能和可扩展性

### 标准
- [ ] 系统处理100+并发评论
- [ ] Worker并行处理作业
- [ ] 数据库查询高效
- [ ] 无内存泄漏

### 验证步骤

1. 生成100条mock评论

2. 触发轮询

3. 监控系统资源：
   ```bash
   docker stats
   ```

4. 检查任务处理时间

5. 验证所有任务完成

**预期结果**：系统在无性能下降的情况下处理负载。

---

## 总结清单

### 核心功能
- [ ] MockConnector工作流端到端工作
- [ ] 意图识别正确工作
- [ ] 任务创建和处理工作
- [ ] 邮件投递工作
- [ ] 幂等性得到维护

### 错误处理
- [ ] 错误在UI中可见
- [ ] 错误被正确记录
- [ ] 重试按预期工作
- [ ] 强制执行最大重试限制

### 配置
- [ ] 可以更新设置
- [ ] 设置正确持久化
- [ ] Provider映射工作
- [ ] 可配置MCP connector

### 性能
- [ ] 系统处理预期负载
- [ ] 响应时间可接受
- [ ] 资源使用高效
- [ ] 无内存泄漏

---

## 签字

**测试人员**：_______________ **日期**：_______________

**审核人员**：_______________ **日期**：_______________

**批准**：☐ 是  ☐ 否  ☐ 有变更
