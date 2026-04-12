package infra

import (
	"github.com/magiconair/properties"
	"log"
)

type AppConfig struct {
	DatabaseDsn    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSsl    bool
	MinioBucket    string
}

func GetProperties() *properties.Properties {
	p, err := properties.LoadFile("config.properties", properties.UTF8)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	return p
}

func GetAppProperties() AppConfig {
	p := GetProperties()
	appConfig := AppConfig{}
	appConfig.DatabaseDsn = p.GetString("database.dsn", "")
	appConfig.MinioEndpoint = p.GetString("minio.endpoint", "")
	appConfig.MinioAccessKey = p.GetString("minio.accessKey", "")
	appConfig.MinioSecretKey = p.GetString("minio.secretKey", "")
	appConfig.MinioUseSsl = p.GetBool("minio.useSsl", false)
	appConfig.MinioBucket = p.GetString("minio.bucket", "web-static")
	return appConfig
}
