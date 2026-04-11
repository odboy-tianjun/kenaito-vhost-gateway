package main

import (
	"crypto/tls"
	"flag"
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
	Id          int64  `xorm:"pk autoincr"`
	ServerName  string `xorm:"unique notnull"`
	Root        string `xorm:"notnull"`
	EnableHTTPS bool   `xorm:"default 0"`
}

// GlobalConfig 对应数据库 global_config 表（单条记录，id=1）
type GlobalConfig struct {
	Id          int64  `xorm:"pk autoincr"`
	HTTPAddr    string `xorm:"default ':80'"`
	HTTPSAddr   string `xorm:"default ':443'"`
	MaxBodySize int64  `xorm:"default 5242880"` // 5MB
	MaxHeaderMB int    `xorm:"default 1"`       // 单位 MB
	CertPEM     string `xorm:"text notnull"`    // 证书 PEM 文本
	KeyPEM      string `xorm:"text notnull"`    // 私钥 PEM 文本
}

// vhostHandler 虚拟主机处理器
type vhostHandler struct {
	hostRootMap  map[string]string // server_name -> root
	hostHTTPSMap map[string]bool   // server_name -> enable_https
	invalidHosts map[string]string // 无效主机（如根目录不存在）
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

// ServeHTTP 实现 http.Handler 接口
func (h *vhostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	if reason, ok := h.invalidHosts[host]; ok {
		log.Printf("请求被拒绝: 域名 %s 无效 (%s)", host, reason)
		http.Error(w, fmt.Sprintf("Host %s is misconfigured", host), http.StatusNotFound)
		return
	}

	root, ok := h.hostRootMap[host]
	if !ok {
		http.Error(w, fmt.Sprintf("No such host: %s", host), http.StatusNotFound)
		log.Printf("未匹配域名 %s，返回 404", host)
		return
	}

	// HTTP 请求且启用了 HTTPS 则重定向
	if r.TLS == nil && h.hostHTTPSMap[host] {
		target := "https://" + host + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		log.Printf("HTTP 请求重定向到 HTTPS: %s -> %s", r.Host, target)
		return
	}

	h.serveStatic(w, r, root)
}

// loadConfigFromDB 从 MySQL 数据库加载全局配置和 servers 列表
func loadConfigFromDB(dsn string) (*GlobalConfig, *vhostHandler, error) {
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer engine.Close()

	// 设置时区等（可选）
	engine.SetConnMaxLifetime(10 * time.Second)

	// 同步表结构
	if err := engine.Sync2(new(Server), new(GlobalConfig)); err != nil {
		return nil, nil, fmt.Errorf("同步表结构失败: %v", err)
	}

	// 读取全局配置（id=1）
	global := &GlobalConfig{}
	has, err := engine.ID(1).Get(global)
	if err != nil {
		return nil, nil, fmt.Errorf("读取全局配置失败: %v", err)
	}
	if !has {
		return nil, nil, fmt.Errorf("数据库中没有全局配置，请先插入一条记录 (id=1)")
	}

	// 读取所有 server
	var servers []Server
	if err := engine.Find(&servers); err != nil {
		return nil, nil, fmt.Errorf("读取 servers 失败: %v", err)
	}
	if len(servers) == 0 {
		return nil, nil, fmt.Errorf("数据库中没有配置任何 server，请先插入")
	}

	// 构建映射
	hostRootMap := make(map[string]string)
	hostHTTPSMap := make(map[string]bool)
	invalidHosts := make(map[string]string)

	for _, srv := range servers {
		if srv.ServerName == "" || srv.Root == "" {
			invalidHosts[srv.ServerName] = "server_name 或 root 为空"
			log.Printf("警告: 域名 %s 配置无效（server_name 或 root 为空）", srv.ServerName)
			continue
		}
		if info, err := os.Stat(srv.Root); err != nil || !info.IsDir() {
			invalidHosts[srv.ServerName] = fmt.Sprintf("根目录不存在或不是目录: %s", srv.Root)
			log.Printf("警告: 域名 %s 无效 - %s", srv.ServerName, invalidHosts[srv.ServerName])
			continue
		}
		hostRootMap[srv.ServerName] = srv.Root
		hostHTTPSMap[srv.ServerName] = srv.EnableHTTPS
		log.Printf("虚拟主机: %s -> %s (HTTPS: %v)", srv.ServerName, srv.Root, srv.EnableHTTPS)
	}

	if len(hostRootMap) == 0 {
		return nil, nil, fmt.Errorf("没有有效的虚拟主机配置")
	}

	handler := &vhostHandler{
		hostRootMap:  hostRootMap,
		hostHTTPSMap: hostHTTPSMap,
		invalidHosts: invalidHosts,
	}
	return global, handler, nil
}

func main() {
	dsn := flag.String("dsn", "", "MySQL 连接字符串，如: user:password@tcp(127.0.0.1:3306)/gateway?charset=utf8mb4&parseTime=True&loc=Local")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("必须通过 -dsn 参数指定 MySQL 连接字符串")
	}

	// 从数据库加载配置
	global, handler, err := loadConfigFromDB(*dsn)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置默认值
	if global.HTTPAddr == "" {
		global.HTTPAddr = ":80"
	}
	if global.HTTPSAddr == "" {
		global.HTTPSAddr = ":443"
	}
	if global.MaxBodySize == 0 {
		global.MaxBodySize = 5 << 20
	}
	if global.MaxHeaderMB == 0 {
		global.MaxHeaderMB = 1
	}
	maxHeaderBytes := global.MaxHeaderMB * 1024 * 1024

	// 启动 HTTP 服务
	httpServer := &http.Server{
		Addr:           global.HTTPAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
	go func() {
		log.Printf("HTTP 服务启动，监听 %s", global.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务失败: %v", err)
		}
	}()

	// 检查是否需要 HTTPS
	needHTTPS := false
	for _, enable := range handler.hostHTTPSMap {
		if enable {
			needHTTPS = true
			break
		}
	}
	if !needHTTPS {
		log.Printf("未启用任何 HTTPS 主机，仅运行 HTTP 服务")
		select {} // 永久阻塞
	}

	// 从数据库中的 PEM 文本加载证书
	if global.CertPEM == "" || global.KeyPEM == "" {
		log.Fatal("存在启用 HTTPS 的主机，但数据库 global_config 中 CertPEM 或 KeyPEM 为空")
	}
	cert, err := tls.X509KeyPair([]byte(global.CertPEM), []byte(global.KeyPEM))
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
		Addr:           global.HTTPSAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
		TLSConfig:      tlsConfig,
	}
	log.Printf("HTTPS 服务启动，监听 %s", global.HTTPSAddr)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("HTTPS 服务失败: %v", err)
	}
}
