# AI Gateway - 基于 Go 的 AI 异步调用网关

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 项目简介

AI Gateway 是一个高并发、高可用的 AI 异步调用网关，专门解决 Python/AI 模型服务响应耗时久（5-30s）、I/O 阻塞、并发能力弱等痛点。

## 技术栈

- **Go** - 高性能编程语言
- **Gin** - 轻量级 Web 框架
- **Goroutine + Channel** - 并发编程模型
- **JWT** - 身份认证
- **Redis** - 任务状态存储
- **gRPC** - 微服务通信
- **Docker** - 容器化部署
- **golang.org/x/time/rate** - 令牌桶限流算法

## 核心特性

### 1. 高并发 Web 服务
- 基于 Gin 框架搭建 RESTful API 网关
- 完整的中间件体系：
  - 日志记录中间件
  - 全局错误处理中间件
  - JWT 身份认证中间件
  - 令牌桶限流中间件

### 2. 异步架构设计
- **生产者-消费者模式**：利用 Goroutine + Channel 构建消息队列
- **生产者**：接收 HTTP 请求，封装任务推入 Channel，立即返回任务 ID
- **消费者**：后台 Worker 池从 Channel 获取任务
- **下游调用**：通过 gRPC 调用 Python 模型服务
- **解耦设计**：请求处理与任务执行完全分离

### 3. 高可用保障
- **任务持久化**：Redis 存储任务状态与结果
- **轮询查询**：支持客户端轮询查询任务状态
- **限流保护**：基于 golang.org/x/time/rate 实现令牌桶算法
- **优雅关闭**：支持平滑关闭，确保任务不丢失

### 4. 容器化部署
- **Docker 打包**：一键构建服务镜像
- **Docker Compose**：编排 Gateway、Redis、Model Service
- **快速迭代**：开发、测试、生产环境一致性

## 项目结构

```
AI_Gateway/
├── cmd/
│   └── gateway/          # 主程序入口
│       └── main.go
├── internal/             # 内部包
│   ├── config/          # 配置管理
│   ├── middleware/      # 中间件
│   │   ├── logger.go   # 日志中间件
│   │   ├── error.go    # 错误处理
│   │   ├── jwt.go      # JWT 认证
│   │   └── ratelimit.go # 限流中间件
│   ├── task/            # 任务队列
│   │   └── queue.go    # 生产者-消费者实现
│   ├── redis/           # Redis 存储
│   │   └── storage.go
│   └── grpc/            # gRPC 客户端
│       └── client.go
├── pkg/                 # 公共包
│   ├── logger/         # 日志工具
│   └── response/       # 响应封装
├── Dockerfile          # Docker 构建文件
├── docker-compose.yml  # 服务编排
├── .env.example        # 环境变量示例
├── go.mod              # Go 模块定义
└── README.md           # 项目文档
```

## 快速开始

### 前置要求

- Go 1.24+
- Docker & Docker Compose
- Redis (如果本地运行)

### 本地运行

1. **克隆项目**
```bash
git clone https://github.com/Jacy666/AI_Gateway.git
cd AI_Gateway
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件配置参数
```

4. **启动 Redis**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

5. **运行网关**
```bash
go run cmd/gateway/main.go
```

### Docker 部署

1. **使用 Docker Compose 一键启动**
```bash
docker-compose up -d
```

2. **查看日志**
```bash
docker-compose logs -f gateway
```

3. **停止服务**
```bash
docker-compose down
```

## API 文档

### 认证接口

#### 注册用户
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "test_user",
  "password": "password123",
  "email": "user@example.com"
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "test_user",
  "password": "password123"
}

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "username": "test_user"
    }
  }
}
```

### 任务接口（需要 JWT 认证）

#### 提交任务
```http
POST /api/v1/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "text_classification",
  "payload": {
    "text": "This is a sample text for classification",
    "model": "bert-base"
  }
}

Response:
{
  "code": 201,
  "message": "created",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Task submitted successfully"
  }
}
```

#### 查询任务状态
```http
GET /api/v1/tasks/:task_id
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "text_classification",
    "status": "completed",
    "result": {
      "label": "positive",
      "confidence": 0.95
    },
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:05Z"
  }
}
```

### 健康检查

```http
GET /health

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "healthy",
    "redis": "healthy",
    "grpc": "healthy",
    "timestamp": 1704096000
  }
}
```

## 配置说明

所有配置项可通过环境变量设置：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| GIN_MODE | Gin 运行模式 | debug |
| REDIS_HOST | Redis 主机地址 | localhost |
| REDIS_PORT | Redis 端口 | 6379 |
| JWT_SECRET | JWT 密钥 | your-secret-key |
| JWT_EXPIRE_HOURS | Token 过期时间（小时） | 24 |
| GRPC_MODEL_SERVICE | gRPC 服务地址 | localhost:50051 |
| RATE_LIMIT_RPS | 限流速率（请求/秒） | 100 |
| RATE_LIMIT_BURST | 限流突发大小 | 200 |
| NUM_WORKERS | Worker 数量 | 10 |
| TASK_BUFFER_SIZE | 任务队列缓冲区大小 | 1000 |

## 架构设计

### 异步处理流程

```
客户端 → Gateway (Gin)
           ↓
    [JWT 认证中间件]
           ↓
    [限流中间件]
           ↓
    [提交任务] → Channel (Buffer: 1000)
           ↓
    [返回 Task ID]
           
Worker Pool (10 Goroutines)
    ↓ 从 Channel 获取任务
    ↓ 更新状态: Processing
    ↓ 通过 gRPC 调用 Python 模型服务
    ↓ 更新状态: Completed/Failed
    ↓ 存储结果到 Redis

客户端 → [轮询查询] → Redis → 返回结果
```

### 限流算法

使用 **令牌桶算法** (Token Bucket Algorithm)：
- 每秒产生固定数量的令牌（rate）
- 桶的最大容量为 burst
- 请求消耗令牌，无令牌则拒绝
- 支持突发流量，平滑限流

## 测试

### 单元测试
```bash
go test ./...
```

### 集成测试
```bash
# 启动所有服务
docker-compose up -d

# 运行测试脚本
./scripts/integration_test.sh
```

### API 测试示例

```bash
# 1. 登录获取 Token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 2. 提交任务
TASK_ID=$(curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"text_analysis","payload":{"text":"Hello World"}}' \
  | jq -r '.data.task_id')

# 3. 查询任务状态
curl http://localhost:8080/api/v1/tasks/$TASK_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 性能指标

- **并发处理能力**：10 个 Worker 同时处理任务
- **队列缓冲**：支持 1000 个任务排队
- **限流保护**：默认 100 RPS，突发 200
- **响应时间**：任务提交 < 10ms，立即返回
- **任务处理**：根据模型服务性能（通常 5-30s）

## 生产环境建议

1. **安全性**
   - 修改 JWT_SECRET 为强密码
   - 启用 HTTPS
   - 配置防火墙规则

2. **性能优化**
   - 根据负载调整 NUM_WORKERS
   - 增加 TASK_BUFFER_SIZE 应对突发流量
   - Redis 启用持久化（AOF/RDB）

3. **监控告警**
   - 集成 Prometheus + Grafana
   - 监控队列长度、处理延迟
   - 配置告警规则

4. **高可用**
   - 部署多个 Gateway 实例 + 负载均衡
   - Redis 主从复制或集群模式
   - gRPC 服务多实例部署

## 许可证

MIT License

## 作者

Jacy666

## 贡献

欢迎提交 Issue 和 Pull Request！