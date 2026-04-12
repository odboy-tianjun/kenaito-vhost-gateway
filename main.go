package main

import (
	"crypto/tls"
	"errors"
	handler2 "kenaito-vhost-gateway/src/handler"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"kenaito-vhost-gateway/src/infra"
	"kenaito-vhost-gateway/src/service/config"
	serverService "kenaito-vhost-gateway/src/service/server"
)

func main() {
	// 加载应用配置文件
	infra.LoadAppConfig()

	// 初始化数据库连接
	err := infra.InitDatabase()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 加载全局配置并初始化 MinIO 客户端
	global, err := config.LoadGlobalConfig()
	if err != nil {
		log.Fatalf("加载全局配置失败: %v", err)
	}

	// 创建服务层实例
	srvService := serverService.NewServerService()

	// 创建 HTTP 处理器
	appConfig := infra.GetAppConfig()
	handler := &handler2.VHostHandler{
		MinioClient:   infra.GetMinioClient(),
		Bucket:        appConfig.MinioBucket,
		ServerService: srvService,
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
	httpsCount, err := srvService.CountHttpsServers()
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
