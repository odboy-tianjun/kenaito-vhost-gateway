# kenaito-vhost-gateway

多域名 HTTPS 静态文件代理，支持域名与根目录一对一绑定、SPA 回退、HTTP 自动重定向。

## 特性

- 多虚拟主机：通过 JSON 配置绑定多个 `server_name` 与静态文件根目录
- HTTPS（TLS）服务
- SPA 路由回退：不存在的路径自动返回 `index.html`
- HTTP → HTTPS 301 重定向（可选）
- Nginx 风格访问日志
- 路径遍历防护

## 快速开始

```shell
# 编译
go build -o kenaito-vhost-gateway main.go

# 运行
./kenaito-vhost-gateway -config config.json
```

## 版本迭代

- v1.0.0
    - 配置文件
- v1.0.1
    - mysql数据库可配置
- v1.0.3
    - mysql数据库可配置
    - 支持minio存储
    - 支持多版本域名映射
