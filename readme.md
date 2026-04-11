# kenaito-vhost-gateway

多域名 HTTPS 静态文件代理，支持域名与根目录一对一绑定、SPA 回退、HTTP 自动重定向。

## 特性

- 多虚拟主机：通过 JSON 配置绑定多个 `server_name` 与静态文件根目录
- HTTPS（TLS）服务
- SPA 路由回退：不存在的路径自动返回 `index.html`
- HTTP → HTTPS 301 重定向（可选）
- Nginx 风格访问日志
- 路径遍历防护

## 配置文件示例 (`config.json`)

```json
{
  "servers": [
    {
      "server_name": "example1.odboy.com",
      "root": "/var/www/example1/dist"
    },
    {
      "server_name": "example2.odboy.com",
      "root": "/var/www/example2/dist"
    }
  ],
  "ssl": {
    "cert_file": "/path/to/cert.pem",
    "key_file": "/path/to/key.pem"
  },
  "http_addr": ":80",
  "https_addr": ":443",
  "http_redirect": true,
  "max_body_size": 5242880
}
```

## 快速开始

```shell
# 编译
go build -o kenaito-vhost-gateway main.go

# 运行
./kenaito-vhost-gateway -config config.json
```

## 命令行参数

- config：配置文件路径（默认 config.json）

