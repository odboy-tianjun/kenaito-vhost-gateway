package infra

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

// InitMinioClient 初始化 MinIO 客户端
func InitMinioClient() error {
	var err error

	// 从应用配置中获取 MinIO 配置
	config := GetAppConfig()

	minioClient, err = minio.New(config.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinioAccessKey, config.MinioSecretKey, ""),
		Secure: config.MinioUseSsl,
	})
	if err != nil {
		return err
	}
	log.Printf("MinIO 客户端初始化成功，Endpoint: %s, Bucket: %s", config.MinioEndpoint, config.MinioBucket)
	return nil
}

// GetMinioClient 获取 MinIO 客户端实例
func GetMinioClient() *minio.Client {
	return minioClient
}

// CheckDefaultBucketExist 检查默认的存储桶是否已创建
func CheckDefaultBucketExist() bool {
	config := GetAppConfig()
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, config.MinioBucket)
	if err != nil {
		log.Fatalf("检查 MinIO 存储桶失败: %v", err)
	}
	if !exists {
		log.Fatalf("MinIO 存储桶 %s 不存在，请先创建", config.MinioBucket)
	}
	return exists
}
