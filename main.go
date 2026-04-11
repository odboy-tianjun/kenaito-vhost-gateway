package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"xorm.io/xorm"
)

// Server 对应 servers 表
type Server struct {
	Id            int    `xorm:"pk autoincr"`
	ServerName    string `xorm:"unique notnull"`
	ActiveVersion string `xorm:"default 'v1'"`
	EnableHttps   bool   `xorm:"default 0"`
}

// ServerVersion 对应 server_versions 表（无 root 字段）
type ServerVersion struct {
	Id         int    `xorm:"pk autoincr"`
	ServerName string `xorm:"notnull"`
	Version    string `xorm:"notnull"`
	BucketPath string `xorm:"notnull"`
}

// GlobalConfig 对应 global_config 表
type GlobalConfig struct {
	Id             int    `xorm:"pk autoincr"`
	HttpAddr       string `xorm:"default ':80'"`
	HttpsAddr      string `xorm:"default ':443'"`
	MaxBodySize    int    `xorm:"default 5242880"`
	CertPem        string `xorm:"text notnull"`
	KeyPem         string `xorm:"text notnull"`
	MinioEndpoint  string `xorm:"notnull"`
	MinioAccessKey string `xorm:"notnull"`
	MinioSecretKey string `xorm:"notnull"`
	MinioUseSsl    bool   `xorm:"default 0"`
	MinioBucket    string `xorm:"notnull"`
}

// vhostHandler 虚拟主机处理器
type vhostHandler struct {
	engine      *xorm.Engine
	minioClient *minio.Client
	bucket      string
}

// serveFromMinIO 从 MinIO 读取文件
func (h *vhostHandler) serveFromMinIO(w http.ResponseWriter, r *http.Request, bucketPath string) {
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

	obj, err := h.minioClient.GetObject(ctx, h.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			// SPA fallback: 返回 index.html
			fallbackPath := fmt.Sprintf("%s/index.html", basePath)
			fallbackObj, fallbackErr := h.minioClient.GetObject(ctx, h.bucket, fallbackPath, minio.GetObjectOptions{})
			if fallbackErr != nil {
				log.Printf("文件不存在且 fallback 失败: %s, err: %v", objectPath, fallbackErr)
				http.NotFound(w, r)
				logRequest(r, bucketPath, http.StatusNotFound, 0, start)
				return
			}
			defer fallbackObj.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			size, _ := io.Copy(w, fallbackObj)
			logRequest(r, bucketPath, http.StatusOK, size, start)
			return
		}
		log.Printf("读取 MinIO 对象失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logRequest(r, bucketPath, http.StatusInternalServerError, 0, start)
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
	logRequest(r, bucketPath, http.StatusOK, size, start)
}

func logRequest(r *http.Request, prefix string, status int, size int64, start time.Time) {
	duration := time.Since(start)
	log.Printf("%s - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" prefix=%s %.3f ms",
		r.RemoteAddr,
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.URL.Path,
		r.Proto,
		status,
		size,
		r.Header.Get("Referer"),
		r.Header.Get("User-Agent"),
		prefix,
		float64(duration.Microseconds())/1000.0,
	)
}

// ServeHTTP 实现 http.Handler 接口
func (h *vhostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// 1. 查询域名主表获取 active_version
	var server Server
	has, err := h.engine.Where("server_name = ?", host).Get(&server)
	if err != nil {
		log.Printf("查询域名配置失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !has {
		log.Printf("未匹配域名 %s，返回 404", host)
		http.Error(w, fmt.Sprintf("No such host: %s", host), http.StatusNotFound)
		return
	}

	// HTTP -> HTTPS 重定向
	if r.TLS == nil && server.EnableHttps {
		target := "https://" + host + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		log.Printf("HTTP 请求重定向到 HTTPS: %s -> %s", r.Host, target)
		return
	}

	// 2. 根据域名和 active_version 查询版本表获取 bucket_path
	var version ServerVersion
	has, err = h.engine.Where("server_name = ? AND version = ?", host, server.ActiveVersion).Get(&version)
	if err != nil {
		log.Printf("查询版本配置失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !has {
		log.Printf("域名 %s 的版本 %s 不存在", host, server.ActiveVersion)
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	// 3. 从 MinIO 服务静态文件
	h.serveFromMinIO(w, r, version.BucketPath)
}

// loadGlobalConfig 加载全局配置
func loadGlobalConfig(dsn string) (*GlobalConfig, error) {
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer engine.Close()

	global := &GlobalConfig{}
	has, err := engine.ID(1).Get(global)
	if err != nil {
		return nil, fmt.Errorf("读取全局配置失败: %v", err)
	}
	if !has {
		return nil, fmt.Errorf("数据库中没有全局配置，请先插入一条记录 (id=1)")
	}
	return global, nil
}

func main() {
	dsn := "root:kd123456789@tcp(219.151.187.115:3306)/kenaito_vhost_gateway?charset=utf8mb4&parseTime=True&loc=Local"

	global, err := loadGlobalConfig(dsn)
	if err != nil {
		log.Fatalf("加载全局配置失败: %v", err)
	}

	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		log.Fatalf("创建数据库引擎失败: %v", err)
	}
	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(100)
	engine.SetConnMaxLifetime(10 * time.Minute)

	// 初始化 MinIO 客户端
	minioClient, err := minio.New(global.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(global.MinioAccessKey, global.MinioSecretKey, ""),
		Secure: global.MinioUseSsl,
	})
	if err != nil {
		log.Fatalf("初始化 MinIO 客户端失败: %v", err)
	}
	log.Printf("MinIO 客户端初始化成功，Endpoint: %s, Bucket: %s", global.MinioEndpoint, global.MinioBucket)

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, global.MinioBucket)
	if err != nil {
		log.Fatalf("检查 MinIO 存储桶失败: %v", err)
	}
	if !exists {
		log.Fatalf("MinIO 存储桶 %s 不存在，请先创建", global.MinioBucket)
	}

	handler := &vhostHandler{
		engine:      engine,
		minioClient: minioClient,
		bucket:      global.MinioBucket,
	}

	httpAddr := global.HttpAddr
	if httpAddr == "" {
		httpAddr = ":80"
	}
	httpsAddr := global.HttpsAddr
	if httpsAddr == "" {
		httpsAddr = ":443"
	}
	if global.MaxBodySize == 0 {
		global.MaxBodySize = 5 << 20
	}
	maxHeaderBytes := 1 * 1024 * 1024

	httpServer := &http.Server{
		Addr:           httpAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
	go func() {
		log.Printf("HTTP 服务启动，监听 %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务失败: %v", err)
		}
	}()

	// 检查是否有启用 HTTPS 的域名
	var httpsCount int64
	httpsCount, err = engine.Where("enable_https = ?", true).Count(&Server{})
	if err != nil {
		log.Printf("查询 HTTPS 主机失败: %v", err)
	}
	if httpsCount == 0 {
		log.Printf("未启用任何 HTTPS 主机，仅运行 HTTP 服务")
		select {}
	}

	if global.CertPem == "" || global.KeyPem == "" {
		log.Fatal("存在启用 HTTPS 的主机，但数据库 global_config 中 CertPem 或 KeyPem 为空")
	}
	cert, err := tls.X509KeyPair([]byte(global.CertPem), []byte(global.KeyPem))
	if err != nil {
		log.Fatalf("加载证书失败: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	httpsServer := &http.Server{
		Addr:           httpsAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
		TLSConfig:      tlsConfig,
	}
	log.Printf("HTTPS 服务启动，监听 %s", httpsAddr)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("HTTPS 服务失败: %v", err)
	}
}
