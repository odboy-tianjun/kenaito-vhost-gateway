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
	MinioEndpoint           string // MinIO 服务端点
	MinioAccessKey          string // MinIO 访问密钥
	MinioSecretKey          string // MinIO 秘密密钥
	MinioUseSsl             bool   // MinIO 是否使用 SSL
	MinioBucket             string // MinIO 默认存储桶名称
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

	appConfig = &AppConfig{
		DatabaseDsn:             p.GetString("database.dsn", ""),
		DatabaseShowSql:         p.GetBool("database.showSql", false),
		DatabaseMaxIdleConns:    p.GetInt("database.maxIdleConns", 10),
		DatabaseMaxOpenConns:    p.GetInt("database.maxOpenConns", 100),
		DatabaseConnMaxLifetime: p.GetInt("database.connMaxLifetime", 300),
		MinioEndpoint:           p.GetString("minio.endpoint", ""),
		MinioAccessKey:          p.GetString("minio.accessKey", ""),
		MinioSecretKey:          p.GetString("minio.secretKey", ""),
		MinioUseSsl:             p.GetBool("minio.useSsl", false),
		MinioBucket:             p.GetString("minio.bucket", "web-static"),
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
