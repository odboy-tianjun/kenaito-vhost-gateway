# kenaito-vhost-gateway

[English Version](README_EN.md) | [中文版](README.md)

基于 Go 语言开发的多域名 HTTPS 静态文件代理网关，支持从 MinIO 对象存储动态加载静态资源、版本管理、自动部署等功能。

## 特性

- **多虚拟主机**：支持多个域名绑定，每个域名独立配置
- **HTTPS/TLS 支持**：自动 HTTP → HTTPS 重定向
- **MinIO 对象存储**：静态文件存储在 MinIO，支持高可用和扩展
- **版本管理**：每次部署生成唯一版本号（yyyyMMddHHmmss），支持版本回滚
- **SPA 路由回退**：不存在的路径自动返回 `index.html`
- **路径遍历防护**：防止目录穿越攻击
- **管理 API**：提供完整的 RESTful API 进行配置管理和应用部署
- **Nginx 风格日志**：详细的访问日志记录

## 架构说明

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Browser    │ ──────> │   Gateway     │ ──────> │   MinIO     │
│  (example.com)│         │  (Port 80/443)│         │  (Storage)  │
└─────────────┘         └──────────────┘         └─────────────┘
                               ▲
                               │
                        ┌──────┴──────┐
                        │  Admin API   │
                        │  (Port 8080) │
                        └──────────────┘
```

**端口说明：**

- **80/443**：虚拟主机服务端口（HTTP/HTTPS），处理实际的域名访问请求
- **8080**：管理 API 端口，提供配置管理和部署接口

## 快速开始

### 1. 环境要求

- Go 1.25+
- MySQL 5.7+
- MinIO 对象存储

### 2. 配置文件

编辑 `config.properties` 文件：

### 3. 数据库初始化

执行 SQL 脚本创建数据库表：

```bash
mysql -u root -p < doc/kenaito_vhost_gateway_v1.0.4.sql
```

插入初始全局配置：

```sql
INSERT INTO global_config (id, http_addr, https_addr, max_body_size, cert_pem, key_pem)
VALUES (1, ':80', ':443', 5242880, '-----BEGIN CERTIFICATE-----...', '-----BEGIN PRIVATE KEY-----...');
```

### 4. 编译运行

```bash
# 编译
go build -o kenaito-vhost-gateway main.go

# 运行
./kenaito-vhost-gateway
```

启动后会看到：

```
应用配置文件加载成功
数据库连接初始化成功
MinIO 客户端初始化成功，Endpoint: localhost:9000, Bucket: web-static
全局配置加载成功
HTTP 服务启动，监听 :80
管理 API 服务启动，监听 :8080
HTTPS 服务启动，监听 :443
```

## 管理 API 使用

所有接口统一使用 **POST** 方法，请求和响应均为 JSON 格式。

### 1. 全局配置管理

#### 获取全局配置

```bash
curl -X POST http://localhost:8080/api/config/global/get
```

响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "httpAddr": ":80",
    "httpsAddr": ":443",
    "maxBodySize": 5242880,
    "certPem": "-----BEGIN CERTIFICATE-----...",
    "keyPem": "-----BEGIN PRIVATE KEY-----..."
  }
}
```

#### 更新全局配置

```bash
curl -X POST http://localhost:8080/api/config/global/update \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{
    "httpAddr": ":80",
    "httpsAddr": ":443",
    "maxBodySize": 5242880,
    "certPem": "-----BEGIN CERTIFICATE-----...",
    "keyPem": "-----BEGIN PRIVATE KEY-----..."
  }'
```

### 2. 域名管理

#### 获取域名列表

```bash
curl -X POST http://localhost:8080/api/servers/list
```

响应：

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "serverName": "example.com",
      "activeVersion": "20260412093045",
      "enableHttps": true
    }
  ]
}
```

#### 获取单个域名

```bash
curl -X POST http://localhost:8080/api/servers/get \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{"id": 1}'
```

#### 创建域名

```bash
curl -X POST http://localhost:8080/api/servers/create \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{
    "serverName": "example.com",
    "enableHttps": false
  }'
```

#### 更新域名

```bash
curl -X POST http://localhost:8080/api/servers/update \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{
    "id": 1,
    "serverName": "example.com",
    "activeVersion": "20260412093045",
    "enableHttps": true
  }'
```

#### 删除域名

```bash
curl -X POST http://localhost:8080/api/servers/delete \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{"id": 1}'
```

### 3. 部署管理

#### 上传目录并部署

```bash
curl -X POST http://localhost:8080/api/servers/deploy \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{
    "localDir": "E:\\DevFiles\\dist\\myapp",
    "serverName": "example.com",
    "appName": "myapp",
    "autoSwitch": true
  }'
```

**参数说明：**

- `localDir`: 本地构建产物目录路径
- `serverName`: 域名（如 `example.com`）
- `appName`: 应用名称（用于 MinIO 路径分类）
- `autoSwitch`: 是否自动切换到新版本（true/false）

响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "部署成功",
    "version": "20260412093045"
  }
}
```

**部署流程：**

1. 检查域名是否存在，不存在则自动创建
2. 生成版本号（格式：yyyyMMddHHmmss）
3. 上传本地目录到 MinIO（路径：`appName/version/`）
4. 创建版本记录到 `server_version` 表
5. 如果 `autoSwitch=true`，自动更新域名的活跃版本

#### 切换版本

```bash
curl -X POST http://localhost:8080/api/servers/switchVersion \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{
    "id": 1,
    "version": "20260412090000"
  }'
```

## 项目结构

```
kenaito-vhost-gateway/
├── main.go                          # 应用入口
├── config.properties                # 应用配置文件
├── doc/                             # 数据库脚本
├── src/
│   ├── controller/                  # 控制器层（API 接口）
│   │   ├── global_config_controller.go  # 全局配置控制器
│   │   ├── server_controller.go         # 域名管理控制器
│   │   ├── router.go                      # 路由注册
│   │   └── response.go                    # 统一响应封装
│   ├── dal/dataobject/              # 数据对象层
│   │   ├── global_config_do.go      # 全局配置数据对象
│   │   ├── server_do.go             # 域名数据对象
│   │   └── server_version_do.go     # 版本数据对象
│   ├── handler/                     # 处理器层
│   │   └── vhost_handler.go         # 虚拟主机处理器
│   ├── infra/                       # 基础设施层
│   │   ├── config.go                # 配置管理
│   │   ├── mysql.go                 # 数据库连接
│   │   ├── minio.go                 # MinIO 客户端
│   │   └── log.go                   # 日志工具
│   └── service/                     # 服务层
│       ├── config/                  # 配置服务
│       └── server/                  # 域名服务
└── readme.md
```

## 版本迭代

- **v1.0.0**
    - 基础多域名静态文件代理
    - JSON 配置文件

- **v1.0.1**
    - MySQL 数据库配置支持

- **v1.0.3**
    - MinIO 对象存储集成
    - 多版本域名映射

- **v1.0.4-release**
    - 版本自动切换
    - 管理 API 接口

- **v1.0.5-gray**
    - 支持基于请求头的灰度版本

## 常见问题

### 1. 如何修改管理 API 端口？

编辑 `config.properties` 文件：

```properties
admin.port=:8080
```

### 2. 如何修改虚拟主机端口？

通过 API 更新全局配置：

```bash
curl -X POST http://localhost:8080/api/config/global/update \
  -H "Content-Type: application/json;charset=utf8" \
  -d '{
    "httpAddr": ":8080",
    "httpsAddr": ":8443",
    ...其他配置
  }'
```

### 3. 如何回滚到历史版本？

```bash
# 查看历史版本
curl -X POST http://localhost:8080/api/servers/get -d '{"id": 1}'

# 切换到指定版本
curl -X POST http://localhost:8080/api/servers/switchVersion \
  -d '{"id": 1, "version": "20260411080000"}'
```

### 4. MinIO 存储桶不存在怎么办？

程序启动时会自动检查存储桶，如果不存在会报错。请先在 MinIO 控制台创建存储桶：

```bash
# 使用 mc 命令行工具
mc mb myminio/web-static
```

## License

Apache 2.0
