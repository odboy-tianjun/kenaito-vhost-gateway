package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"kenaito-vhost-gateway/src/infra"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

var minioClient *minio.Client

// InitMinioClient 初始化 MinIO 客户端
func InitMinioClient() error {
	var err error

	// 从应用配置中获取 OSS 配置
	config := infra.GetAppConfig()

	if "minio" == config.OssType {
		minioClient, err = minio.New(config.OssEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(config.OssAccessKey, config.OssSecretKey, ""),
			Secure: config.OssUseSsl,
		})
		if err != nil {
			return err
		}
		log.Printf("MinIO 客户端初始化成功，Endpoint: %s, Bucket: %s", config.OssEndpoint, config.OssBucket)
	}

	return nil
}

// GetMinioClient 获取 MinIO 客户端实例
func GetMinioClient() *minio.Client {
	return minioClient
}

// CheckDefaultBucketExist 检查默认的存储桶是否已创建
func CheckDefaultBucketExist() bool {
	config := infra.GetAppConfig()
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, config.OssBucket)
	if err != nil {
		log.Fatalf("检查 MinIO 存储桶失败: %v", err)
	}
	if !exists {
		log.Fatalf("MinIO 存储桶 %s 不存在，请先创建", config.OssBucket)
	}
	return exists
}

// UploadDirectoryToMinio 上传本地文件夹到 MinIO 指定路径
// localDir: 本地文件夹路径
// minioPrefix: MinIO 中的目标路径前缀（如 "web/v1.0/"）
func UploadDirectoryToMinio(localDir string, minioPrefix string) error {
	config := infra.GetAppConfig()
	ctx := context.Background()

	// 确保本地目录存在
	info, err := os.Stat(localDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return err
	}

	// 清理路径前缀，确保以 / 结尾
	minioPrefix = strings.TrimPrefix(minioPrefix, "/")
	if minioPrefix != "" && !strings.HasSuffix(minioPrefix, "/") {
		minioPrefix += "/"
	}

	// 递归遍历本地目录并上传文件
	uploadCount := 0
	err = filepath.Walk(localDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localDir, filePath)
		if err != nil {
			return err
		}

		// 构造 MinIO 对象路径
		objectName := minioPrefix + filepath.ToSlash(relPath)

		// 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// 获取文件大小
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}

		// 检测 Content-Type
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// 上传文件到 MinIO
		_, err = minioClient.PutObject(ctx, config.OssBucket, objectName, file, fileInfo.Size(), minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			return err
		}

		uploadCount++
		log.Printf("已上传: %s -> %s/%s", filePath, config.OssBucket, objectName)
		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("上传完成，共上传 %d 个文件到 %s/%s", uploadCount, config.OssBucket, minioPrefix)
	return nil
}
