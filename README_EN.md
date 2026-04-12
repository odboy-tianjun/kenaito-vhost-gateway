# kenaito-vhost-gateway

A multi-domain HTTPS static file proxy gateway developed in Go, supporting dynamic loading of static resources from MinIO object storage, version management, and automated deployment.

## Features

- **Multi-Virtual Host**: Support multiple domain bindings with independent configurations
- **HTTPS/TLS Support**: Automatic HTTP → HTTPS redirection
- **MinIO Object Storage**: Static files stored in MinIO for high availability and scalability
- **Version Management**: Each deployment generates a unique version number (yyyyMMddHHmmss) with rollback support
- **SPA Route Fallback**: Non-existent paths automatically return `index.html`
- **Path Traversal Protection**: Prevent directory traversal attacks
- **Management API**: Complete RESTful API for configuration management and application deployment
- **Nginx-Style Logging**: Detailed access log recording

## Architecture

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

**Port Description:**
- **80/443**: Virtual host service ports (HTTP/HTTPS) for handling actual domain access requests
- **8080**: Management API port for configuration management and deployment interfaces

## Quick Start

### 1. Requirements

- Go 1.25+
- MySQL 5.7+
- MinIO Object Storage

### 2. Configuration

Edit the `config.properties` file:

```properties
# Database Configuration
database.dsn=root:password@tcp(localhost:3306)/kenaito_vhost_gateway?charset=utf8mb4&parseTime=True&loc=Local
database.showSql=true
database.maxIdleConns=10
database.maxOpenConns=100
database.connMaxLifetime=300

# MinIO Configuration
minio.endpoint=localhost:9000
minio.accessKey=minioadmin
minio.secretKey=minioadmin
minio.useSsl=false
minio.bucket=web-static

# Management API Port
admin.port=:8080
```

### 3. Database Initialization

Execute SQL scripts to create database tables:

```bash
mysql -u root -p < doc/kenaito_vhost_gateway_v1.0.3.sql
```

Insert initial global configuration:

```sql
INSERT INTO global_config (id, http_addr, https_addr, max_body_size, cert_pem, key_pem) 
VALUES (1, ':80', ':443', 5242880, '-----BEGIN CERTIFICATE-----...', '-----BEGIN PRIVATE KEY-----...');
```

### 4. Build and Run

```bash
# Build
go build -o kenaito-vhost-gateway main.go

# Run
./kenaito-vhost-gateway
```

After startup, you will see:
```
Application configuration loaded successfully
Database connection initialized successfully
MinIO client initialized successfully, Endpoint: localhost:9000, Bucket: web-static
Global configuration loaded successfully
HTTP service started, listening on :80
Management API service started, listening on :8080
HTTPS service started, listening on :443
```

## Management API Usage

All interfaces use the **POST** method uniformly, with JSON format for both requests and responses.

### 1. Global Configuration Management

#### Get Global Configuration

```bash
curl -X POST http://localhost:8080/api/config/global/get
```

Response:
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

#### Update Global Configuration

```bash
curl -X POST http://localhost:8080/api/config/global/update \
  -H "Content-Type: application/json" \
  -d '{
    "httpAddr": ":80",
    "httpsAddr": ":443",
    "maxBodySize": 5242880,
    "certPem": "-----BEGIN CERTIFICATE-----...",
    "keyPem": "-----BEGIN PRIVATE KEY-----..."
  }'
```

### 2. Domain Management

#### Get Domain List

```bash
curl -X POST http://localhost:8080/api/servers/list
```

Response:
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

#### Get Single Domain

```bash
curl -X POST http://localhost:8080/api/servers/get \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

#### Create Domain

```bash
curl -X POST http://localhost:8080/api/servers/create \
  -H "Content-Type: application/json" \
  -d '{
    "serverName": "example.com",
    "enableHttps": false
  }'
```

#### Update Domain

```bash
curl -X POST http://localhost:8080/api/servers/update \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "serverName": "example.com",
    "activeVersion": "20260412093045",
    "enableHttps": true
  }'
```

#### Delete Domain

```bash
curl -X POST http://localhost:8080/api/servers/delete \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

### 3. Deployment Management

#### Upload Directory and Deploy

```bash
curl -X POST http://localhost:8080/api/servers/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "localDir": "/path/to/dist/myapp",
    "serverName": "example.com",
    "appName": "myapp",
    "autoSwitch": true
  }'
```

**Parameters:**
- `localDir`: Local build output directory path
- `serverName`: Domain name (e.g., `example.com`)
- `appName`: Application name (used for MinIO path classification)
- `autoSwitch`: Whether to automatically switch to the new version (true/false)

Response:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Deployment successful",
    "version": "20260412093045"
  }
}
```

**Deployment Process:**
1. Check if domain exists, create automatically if not
2. Generate version number (format: yyyyMMddHHmmss)
3. Upload local directory to MinIO (path: `appName/version/`)
4. Create version record in `server_version` table
5. If `autoSwitch=true`, automatically update domain's active version

#### Switch Version

```bash
curl -X POST http://localhost:8080/api/servers/switchVersion \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "version": "20260412090000"
  }'
```

## Project Structure

```
kenaito-vhost-gateway/
├── main.go                          # Application entry point
├── config.properties                # Application configuration file
├── doc/                             # Database scripts
├── src/
│   ├── controller/                  # Controller layer (API endpoints)
│   │   ├── global_config_controller.go  # Global config controller
│   │   ├── server_controller.go         # Domain management controller
│   │   ├── router.go                      # Route registration
│   │   └── response.go                    # Unified response wrapper
│   ├── dal/dataobject/              # Data Access Layer objects
│   │   ├── global_config_do.go      # Global config data object
│   │   ├── server_do.go             # Domain data object
│   │   └── server_version_do.go     # Version data object
│   ├── handler/                     # Handler layer
│   │   └── vhost_handler.go         # Virtual host handler
│   ├── infra/                       # Infrastructure layer
│   │   ├── config.go                # Configuration management
│   │   ├── mysql.go                 # Database connection
│   │   ├── minio.go                 # MinIO client
│   │   └── log.go                   # Logging utilities
│   └── service/                     # Service layer
│       ├── config/                  # Configuration service
│       └── server/                  # Domain service
└── readme.md
```

## Version History

- **v1.0.0**
  - Basic multi-domain static file proxy
  - JSON configuration file
  
- **v1.0.1**
  - MySQL database configuration support
  
- **v1.0.3**
  - MinIO object storage integration
  - Multi-version domain mapping
  - Automatic version switching
  - Management API endpoints
  - Externalized configuration (properties file)

## FAQ

### 1. How to change the Management API port?

Edit the `config.properties` file:
```properties
admin.port=:9090
```

### 2. How to change the virtual host ports?

Update global configuration via API:
```bash
curl -X POST http://localhost:8080/api/config/global/update \
  -H "Content-Type: application/json" \
  -d '{
    "httpAddr": ":8080",
    "httpsAddr": ":8443",
    ...other configurations
  }'
```

### 3. How to rollback to a historical version?

```bash
# View historical versions
curl -X POST http://localhost:8080/api/servers/get -d '{"id": 1}'

# Switch to specified version
curl -X POST http://localhost:8080/api/servers/switchVersion \
  -d '{"id": 1, "version": "20260411080000"}'
```

### 4. What if MinIO bucket doesn't exist?

The program will automatically check the bucket at startup and report an error if it doesn't exist. Please create the bucket in MinIO console first:
```bash
# Using mc command line tool
mc mb myminio/web-static
```

## License

MIT
