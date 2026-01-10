# Nacos配置中心使用指南

## 概述

系统支持通过Nacos配置中心管理所有云服务配置。启动时会先读取本地config.yaml，如果配置了Nacos地址，则会从Nacos加载配置并覆盖本地配置。

## 配置流程

### 1. 配置Nacos连接信息

在`.env`文件中配置Nacos连接信息：

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

### 2. 在Nacos配置中心创建配置

登录Nacos控制台（http://your-nacos-host.com:8848/nacos），创建配置：

**配置信息**：
- Data ID: `xiaohongshu-image`
- Group: `DEFAULT_GROUP`
- 配置格式: `JSON`
- 配置内容: 见下方

## Nacos配置内容示例

```json
{
  "database": {
    "host": "your-mysql-host.com",
    "port": 3306,
    "user": "your-username",
    "password": "your-password",
    "dbname": "xiaohongshu_image",
    "max_open_conns": 100,
    "max_idle_conns": 10,
    "conn_max_lifetime": 3600
  },
  "redis": {
    "host": "your-redis-host.com",
    "port": 6379,
    "password": "your-redis-password",
    "db": 0
  },
  "minio": {
    "endpoint": "your-minio-host.com:9000",
    "access_key": "your-access-key",
    "secret_key": "your-secret-key",
    "bucket": "generated-content",
    "use_ssl": true,
    "region": "us-east-1",
    "presigned_expiry": 3600
  },
  "llm": {
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-your-api-key-here",
    "model": "gpt-4o-mini",
    "timeout": "15s",
    "max_retries": 2
  },
  "smtp": {
    "host": "your-smtp-host.com",
    "port": 587,
    "user": "your-smtp-user",
    "password": "your-smtp-password",
    "from": "noreply@xiaohongshu-image.local"
  },
  "asynq": {
    "redis_addr": "your-redis-host.com:6379",
    "redis_password": "your-redis-password",
    "redis_db": 1,
    "concurrency": 10,
    "queues": {
      "critical": 6,
      "default": 3,
      "low": 1
    }
  }
}
```

## 配置说明

### Database（数据库配置）

| 字段 | 说明 | 示例 |
|------|------|------|
| host | 数据库主机地址 | `rm-xxx.mysql.rds.aliyuncs.com` |
| port | 数据库端口 | `3306` |
| user | 数据库用户名 | `xiaohongshu` |
| password | 数据库密码 | `your-password` |
| dbname | 数据库名称 | `xiaohongshu_image` |
| max_open_conns | 最大连接数 | `100` |
| max_idle_conns | 最大空闲连接数 | `10` |
| conn_max_lifetime | 连接最大生命周期（秒） | `3600` |

### Redis（缓存配置）

| 字段 | 说明 | 示例 |
|------|------|------|
| host | Redis主机地址 | `r-xxx.redis.rds.aliyuncs.com` |
| port | Redis端口 | `6379` |
| password | Redis密码 | `your-password` |
| db | Redis数据库编号 | `0` |

### MinIO（对象存储配置）

| 字段 | 说明 | 示例 |
|------|------|------|
| endpoint | MinIO端点 | `oss-cn-hangzhou.aliyuncs.com` |
| access_key | 访问密钥 | `your-access-key` |
| secret_key | 秘密密钥 | `your-secret-key` |
| bucket | 存储桶名称 | `generated-content` |
| use_ssl | 是否使用SSL | `true` |
| region | 区域 | `us-east-1` |
| presigned_expiry | 签名URL过期时间（秒） | `3600` |

### LLM（大语言模型配置）

| 字段 | 说明 | 示例 |
|------|------|------|
| base_url | API基础URL | `https://api.openai.com/v1` |
| api_key | API密钥 | `sk-your-api-key-here` |
| model | 模型名称 | `gpt-4o-mini` |
| timeout | 超时时间 | `15s` |
| max_retries | 最大重试次数 | `2` |

### SMTP（邮件服务配置）

| 字段 | 说明 | 示例 |
|------|------|------|
| host | SMTP主机地址 | `smtpdm.aliyun.com` |
| port | SMTP端口 | `587` |
| user | SMTP用户名 | `your-smtp-user` |
| password | SMTP密码 | `your-password` |
| from | 发件人地址 | `noreply@xiaohongshu-image.local` |

### Asynq（任务队列配置）

| 字段 | 说明 | 示例 |
|------|------|------|
| redis_addr | Redis地址 | `your-redis-host.com:6379` |
| redis_password | Redis密码 | `your-password` |
| redis_db | Redis数据库编号 | `1` |
| concurrency | 并发数 | `10` |
| queues | 队列优先级配置 | `{"critical": 6, "default": 3, "low": 1}` |

## 配置优先级

1. **Nacos配置**（最高优先级）
   - 从Nacos配置中心读取的配置会覆盖本地配置

2. **环境变量**（中等优先级）
   - 通过`.env`文件设置的环境变量会覆盖config.yaml中的默认值

3. **config.yaml**（最低优先级）
   - 本地配置文件中的默认值

## 使用场景

### 场景1：开发环境

不配置Nacos，使用本地config.yaml和环境变量：

```env
# 不设置NACOS_ADDR，系统使用本地配置
NACOS_ADDR=

# 使用环境变量覆盖本地配置
DATABASE_HOST=localhost
REDIS_HOST=localhost
```

### 场景2：测试环境

配置Nacos，从配置中心读取测试环境配置：

```env
# 配置测试环境Nacos
NACOS_ADDR=test-nacos.example.com
NACOS_NAMESPACE=test
NACOS_GROUP=DEFAULT_GROUP
```

### 场景3：生产环境

配置Nacos，从配置中心读取生产环境配置：

```env
# 配置生产环境Nacos
NACOS_ADDR=prod-nacos.example.com
NACOS_NAMESPACE=prod
NACOS_GROUP=DEFAULT_GROUP
```

## 配置更新

### 动态更新（未来支持）

计划支持配置动态更新，无需重启服务：
- 监听Nacos配置变更
- 热更新配置
- 重新初始化相关服务

### 当前实现

当前需要重启服务才能应用新配置：
1. 在Nacos控制台更新配置
2. 重启API和Worker服务
3. 系统重新加载配置

## 故障排查

### Nacos连接失败

**现象**：日志显示"Warning: failed to load config from Nacos"

**解决方案**：
1. 检查Nacos地址和端口是否正确
2. 检查网络连接是否正常
3. 检查Nacos用户名和密码是否正确
4. 检查Namespace和Group是否正确

### 配置内容为空

**现象**：日志显示"config content is empty"

**解决方案**：
1. 检查Data ID是否正确
2. 检查配置是否已发布
3. 检查配置格式是否为JSON

### 配置格式错误

**现象**：日志显示"failed to unmarshal Nacos config"

**解决方案**：
1. 检查JSON格式是否正确
2. 使用JSON验证工具验证配置
3. 检查字段名称是否正确

## 最佳实践

1. **敏感信息管理**
   - 不要在代码中硬编码敏感信息
   - 使用Nacos管理所有敏感配置
   - 定期轮换密钥和密码

2. **环境隔离**
   - 使用不同的Namespace隔离环境
   - 使用不同的Group区分应用
   - 使用有意义的Data ID

3. **配置版本管理**
   - 记录配置变更历史
   - 使用版本号管理配置
   - 重要变更前备份配置

4. **监控和告警**
   - 监控配置加载状态
   - 配置加载失败时发送告警
   - 定期检查配置有效性

## 相关文档

- [Nacos官方文档](https://nacos.io/zh-cn/docs/what-is-nacos.html)
- [Nacos Go SDK](https://github.com/nacos-group/nacos-sdk-go)
- [配置管理最佳实践](https://nacos.io/zh-cn/docs/configuration-management.html)
