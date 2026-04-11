package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

// Server 对应数据库 servers 表
type Server struct {
	Id          int    `xorm:"pk autoincr"`
	ServerName  string `xorm:"unique notnull"`
	Root        string `xorm:"notnull"`
	EnableHttps bool   `xorm:"default 0"`
}

// GlobalConfig 对应数据库 global_config 表（单条记录，id=1）
type GlobalConfig struct {
	Id          int    `xorm:"pk autoincr"`
	HttpAddr    string `xorm:"default ':80'"`
	HttpsAddr   string `xorm:"default ':443'"`
	MaxBodySize int    `xorm:"default 5242880"` // 5MB
	CertPem     string `xorm:"text notnull"`    // 证书 PEM 文本
	KeyPem      string `xorm:"text notnull"`    // 私钥 PEM 文本
}

// vhostHandler 虚拟主机处理器（每次请求直接查数据库）
type vhostHandler struct {
	engine *xorm.Engine // 数据库连接池
}

// serveStatic 提供静态文件服务（含 SPA fallback）
func (h *vhostHandler) serveStatic(w http.ResponseWriter, r *http.Request, root string) {
	start := time.Now()

	cleanPath := filepath.Clean(r.URL.Path)
	if cleanPath == "." {
		cleanPath = "/"
	}
	fullPath := filepath.Join(root, cleanPath)

	// 路径安全检查
	if !strings.HasPrefix(fullPath, filepath.Clean(root)+string(os.PathSeparator)) &&
		fullPath != filepath.Clean(root) {
		http.NotFound(w, r)
		logRequest(r, root, http.StatusNotFound, 0, start)
		return
	}

	info, err := os.Stat(fullPath)
	if err == nil && info.IsDir() {
		indexPath := filepath.Join(fullPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			serveFile(w, r, indexPath)
			logRequest(r, root, http.StatusOK, 0, start)
			return
		}
	} else if err == nil && !info.IsDir() {
		serveFile(w, r, fullPath)
		logRequest(r, root, http.StatusOK, 0, start)
		return
	}

	// SPA fallback
	spaFile := filepath.Join(root, "index.html")
	if _, err := os.Stat(spaFile); err == nil {
		serveFile(w, r, spaFile)
		logRequest(r, root, http.StatusOK, 0, start)
	} else {
		http.NotFound(w, r)
		logRequest(r, root, http.StatusNotFound, 0, start)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	if ctype := mime.TypeByExtension(filepath.Ext(filePath)); ctype != "" {
		w.Header().Set("Content-Type", ctype)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	http.ServeContent(w, r, filepath.Base(filePath), time.Time{}, file)
}

func logRequest(r *http.Request, root string, status int, size int64, start time.Time) {
	duration := time.Since(start)
	log.Printf("%s - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" root=%s %.3f ms",
		r.RemoteAddr,
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.URL.Path,
		r.Proto,
		status,
		size,
		r.Header.Get("Referer"),
		r.Header.Get("User-Agent"),
		root,
		float64(duration.Microseconds())/1000.0,
	)
}

// ServeHTTP 实现 http.Handler 接口（每次请求直接查数据库）
func (h *vhostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// 直接从数据库查询该 host 的配置
	var server Server
	has, err := h.engine.Where("server_name = ?", host).Get(&server)
	if err != nil {
		log.Printf("查询数据库失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !has {
		log.Printf("未匹配域名 %s，返回 404", host)
		http.Error(w, fmt.Sprintf("No such host: %s", host), http.StatusNotFound)
		return
	}

	// 输出读取到的 enable_https 值
	log.Printf("域名 %s: enable_https=%v, 请求协议 TLS=%v", host, server.EnableHttps, r.TLS != nil)

	// 检查根目录是否存在
	if info, err := os.Stat(server.Root); err != nil || !info.IsDir() {
		log.Printf("域名 %s 根目录无效: %s", host, server.Root)
		http.Error(w, fmt.Sprintf("Host %s is misconfigured", host), http.StatusNotFound)
		return
	}

	// HTTP 请求且启用了 HTTPS 则重定向
	if r.TLS == nil && server.EnableHttps {
		target := "https://" + host + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		log.Printf("HTTP 请求重定向到 HTTPS: %s -> %s", r.Host, target)
		return
	}

	h.serveStatic(w, r, server.Root)
}

// loadGlobalConfig 从数据库加载全局配置（仅启动时调用）
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

	// 加载全局配置（证书等）
	global, err := loadGlobalConfig(dsn)
	if err != nil {
		log.Fatalf("加载全局配置失败: %v", err)
	}

	// 创建数据库引擎（连接池），供处理器使用
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		log.Fatalf("创建数据库引擎失败: %v", err)
	}
	// 设置连接池参数
	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(100)
	engine.SetConnMaxLifetime(10 * time.Minute)

	// 可选：打印 SQL（调试用，生产环境可关闭）
	// engine.ShowSQL(true)
	// engine.Logger().SetLevel(log2.LOG_DEBUG)

	// 创建处理器
	handler := &vhostHandler{
		engine: engine,
	}

	// 设置默认值
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

	// 启动 HTTP 服务
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

	// 检查是否有任何主机启用了 HTTPS（通过查询数据库）
	var httpsCount int64
	httpsCount, err = engine.Where("enable_https = ?", true).Count(&Server{})
	if err != nil {
		log.Printf("查询 HTTPS 主机失败: %v", err)
	}
	if httpsCount == 0 {
		log.Printf("未启用任何 HTTPS 主机，仅运行 HTTP 服务")
		select {} // 永久阻塞
	}

	// 从数据库中的 PEM 文本加载证书
	if global.CertPem == "" || global.KeyPem == "" {
		log.Fatal("存在启用 HTTPS 的主机，但数据库 global_config 中 CertPem 或 KeyPem 为空")
	}
	cert, err := tls.X509KeyPair([]byte(global.CertPem), []byte(global.KeyPem))
	if err != nil {
		log.Fatalf("加载证书失败: %v", err)
	}

	// 配置 TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// 启动 HTTPS 服务
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
