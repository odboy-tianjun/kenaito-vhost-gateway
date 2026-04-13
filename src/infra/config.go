package infra

import (
	"log"

	"github.com/magiconair/properties"
)

// AppConfig 应用配置结构体，包含所有从配置文件读取的配置项
type AppConfig struct {
	DatabaseDsn             string // 数据库连接字符串
	DatabaseShowSql         bool   // 是否打印SQL
	DatabaseMaxIdleConns    int    // 最大空闲连接数
	DatabaseMaxOpenConns    int    // 最大打开连接数
	DatabaseConnMaxLifetime int    // 连接的最大生命周期（秒）
	OssType                 string // 默认存储桶类型
	OssBucket               string // 默认存储桶名称
	OssEndpoint             string // MinIO 服务端点
	OssAccessKey            string // MinIO 访问密钥
	OssSecretKey            string // MinIO 秘密密钥
	OssUseSsl               bool   // MinIO 是否使用 SSL
	AdminPort               string // 管理API端口（Controller端口）
}

var appConfig *AppConfig

// LoadAppConfig 加载应用配置文件并缓存结果
func LoadAppConfig() *AppConfig {
	if appConfig != nil {
		return appConfig
	}

	p, err := properties.LoadFile("config.properties", properties.UTF8)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	ossBucket := p.GetString("oss.bucket", "web-static")
	ossType := p.GetString("oss.type", "minio")

	ossEndpoint := p.GetString("oss."+ossType+".endpoint", "")
	ossAccessKey := p.GetString("oss."+ossType+".accessKey", "")
	ossSecretKey := p.GetString("oss."+ossType+".secretKey", "")
	ossUseSsl := p.GetBool("oss."+ossType+".useSsl", false)

	appConfig = &AppConfig{
		DatabaseDsn:             p.GetString("database.dsn", ""),
		DatabaseShowSql:         p.GetBool("database.showSql", false),
		DatabaseMaxIdleConns:    p.GetInt("database.maxIdleConns", 10),
		DatabaseMaxOpenConns:    p.GetInt("database.maxOpenConns", 100),
		DatabaseConnMaxLifetime: p.GetInt("database.connMaxLifetime", 300),
		OssType:                 ossType,
		OssBucket:               ossBucket,
		OssEndpoint:             ossEndpoint,
		OssAccessKey:            ossAccessKey,
		OssSecretKey:            ossSecretKey,
		OssUseSsl:               ossUseSsl,
		AdminPort:               p.GetString("admin.port", ":8080"),
	}

	log.Println("应用配置文件加载成功")
	return appConfig
}

// GetAppConfig 获取已加载的应用配置（必须在 LoadAppConfig 之后调用）
func GetAppConfig() *AppConfig {
	if appConfig == nil {
		log.Fatal("应用配置未初始化，请先调用 LoadAppConfig()")
	}
	return appConfig
}
