package handler

import (
	"context"
	"fmt"
	"io"
	"kenaito-vhost-gateway/src/infra"
	"kenaito-vhost-gateway/src/service/server"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// VHostHandler 虚拟主机处理器
type VHostHandler struct {
	MinioClient   *minio.Client
	Bucket        string
	ServerService *server.ServerService // 域名服务
}

// serveFromMinIO 从 MinIO 读取文件
func (h *VHostHandler) serveFromMinIO(w http.ResponseWriter, r *http.Request, bucketPath string) {
	start := time.Now()
	ctx := context.Background()

	// 构造对象路径: bucketPath/请求路径
	basePath := strings.TrimPrefix(bucketPath, "/")
	requestPath := strings.TrimPrefix(r.URL.Path, "/")
	var objectPath string
	if requestPath == "" {
		objectPath = fmt.Sprintf("%s/index.html", basePath)
	} else {
		objectPath = fmt.Sprintf("%s/%s", basePath, requestPath)
	}

	obj, err := h.MinioClient.GetObject(ctx, h.Bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			// SPA fallback: 返回 index.html
			fallbackPath := fmt.Sprintf("%s/index.html", basePath)
			fallbackObj, fallbackErr := h.MinioClient.GetObject(ctx, h.Bucket, fallbackPath, minio.GetObjectOptions{})
			if fallbackErr != nil {
				log.Printf("文件不存在且 fallback 失败: %s, err: %v", objectPath, fallbackErr)
				http.NotFound(w, r)
				infra.LogRequest(r, bucketPath, http.StatusNotFound, 0, start)
				return
			}
			defer fallbackObj.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			size, _ := io.Copy(w, fallbackObj)
			infra.LogRequest(r, bucketPath, http.StatusOK, size, start)
			return
		}
		log.Printf("读取 MinIO 对象失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		infra.LogRequest(r, bucketPath, http.StatusInternalServerError, 0, start)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err == nil {
		ext := filepath.Ext(objectPath)
		if ctype := mime.TypeByExtension(ext); ctype != "" {
			w.Header().Set("Content-Type", ctype)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		_ = stat // 可用于设置 ETag 等
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	size, _ := io.Copy(w, obj)
	infra.LogRequest(r, bucketPath, http.StatusOK, size, start)
}

// ServeHTTP 实现 http.Handler 接口
func (h *VHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// 1. 查询域名主表获取 active_version
	currentServer, err := h.ServerService.GetServerByName(host)
	if err != nil {
		log.Printf("查询域名配置失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if currentServer == nil {
		log.Printf("未匹配域名 %s，返回 404", host)
		http.Error(w, fmt.Sprintf("No such host: %s", host), http.StatusNotFound)
		return
	}

	// HTTP -> HTTPS 重定向
	if r.TLS == nil && currentServer.EnableHttps {
		target := "https://" + host + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		log.Printf("HTTP 请求重定向到 HTTPS: %s -> %s", r.Host, target)
		return
	}

	// 2. 根据域名和 active_version 查询版本表获取 bucket_path
	version, err := h.ServerService.GetServerVersion(host, currentServer.ActiveVersion)
	if err != nil {
		log.Printf("查询版本配置失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if version == nil {
		log.Printf("域名 %s 的版本 %s 不存在", host, currentServer.ActiveVersion)
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	// 3. 从 MinIO 服务静态文件
	h.serveFromMinIO(w, r, version.BucketPath)
}
