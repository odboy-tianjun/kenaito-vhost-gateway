package infra

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"kenaito-vhost-gateway/src/dal/dataobject"
	"log"
)

var minioClient *minio.Client

// InitMinioClient 初始化 MinIO 客户端
func InitMinioClient() error {
	var err error

	properties := GetAppProperties()

	minioClient, err = minio.New(properties.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(properties.MinioAccessKey, properties.MinioSecretKey, ""),
		Secure: properties.MinioUseSsl,
	})
	if err != nil {
		return err
	}
	log.Printf("MinIO 客户端初始化成功，Endpoint: %s, Bucket: %s", properties.MinioEndpoint, properties.MinioBucket)
	return nil
}

// GetMinioClient 获取 MinIO 客户端实例
func GetMinioClient() *minio.Client {
	return minioClient
}

// CheckDefaultBucketExist 检查默认的存储桶是否已创建
func CheckDefaultBucketExist(global *dataobject.GlobalConfig) bool {
	properties := GetAppProperties()
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, properties.MinioBucket)
	if err != nil {
		log.Fatalf("检查 MinIO 存储桶失败: %v", err)
	}
	if !exists {
		log.Fatalf("MinIO 存储桶 %s 不存在，请先创建", properties.MinioBucket)
	}
	return exists
}
