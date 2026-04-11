package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ServerConfig 单个虚拟主机配置
type ServerConfig struct {
	ServerName  string `json:"server_name"`
	Root        string `json:"root"`
	EnableHTTPS bool   `json:"enable_https"` // 是否启用 HTTPS 并重定向
}

// SSLConfig SSL 证书配置
type SSLConfig struct {
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// Config 总配置
type Config struct {
	Servers     []ServerConfig `json:"servers"`
	SSL         SSLConfig      `json:"ssl"`
	HTTPAddr    string         `json:"http_addr"`
	HTTPSAddr   string         `json:"https_addr"`
	MaxBodySize int64          `json:"max_body_size"`
}

// vhostHandler 虚拟主机处理器（HTTP 和 HTTPS 共用逻辑）
type vhostHandler struct {
	hostRootMap  map[string]string // server_name -> root（有效主机）
	hostHTTPSMap map[string]bool   // server_name -> 是否启用 HTTPS
	invalidHosts map[string]string // server_name -> 错误原因
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

// ServeHTTP 实现 http.Handler 接口（同时用于 HTTP 和 HTTPS）
func (h *vhostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// 无效主机（配置错误）
	if reason, ok := h.invalidHosts[host]; ok {
		log.Printf("请求被拒绝: 域名 %s 无效 (%s)", host, reason)
		http.Error(w, fmt.Sprintf("Host %s is misconfigured", host), http.StatusNotFound)
		return
	}

	// 有效主机
	root, ok := h.hostRootMap[host]
	if !ok {
		http.Error(w, fmt.Sprintf("No such host: %s", host), http.StatusNotFound)
		log.Printf("未匹配域名 %s，返回 404", host)
		return
	}

	// 如果是 HTTP 请求且该域名启用了 HTTPS，则重定向到 HTTPS
	if r.TLS == nil && h.hostHTTPSMap[host] {
		target := "https://" + host + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		log.Printf("HTTP 请求重定向到 HTTPS: %s -> %s", r.Host, target)
		return
	}

	// 正常服务静态文件
	h.serveStatic(w, r, root)
}

func main() {
	configPath := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	// 默认值设置
	if len(cfg.Servers) == 0 {
		log.Fatal("配置中至少需要一个 server")
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":80"
	}
	if cfg.HTTPSAddr == "" {
		cfg.HTTPSAddr = ":443"
	}
	if cfg.MaxBodySize == 0 {
		cfg.MaxBodySize = 5 << 20 // 5MB
	}

	// 构建主机映射
	hostRootMap := make(map[string]string)
	hostHTTPSMap := make(map[string]bool)
	invalidHosts := make(map[string]string)

	for _, srv := range cfg.Servers {
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
		log.Fatal("没有有效的虚拟主机配置，无法启动服务")
	}

	handler := &vhostHandler{
		hostRootMap:  hostRootMap,
		hostHTTPSMap: hostHTTPSMap,
		invalidHosts: invalidHosts,
	}

	// HTTP 服务
	httpServer := &http.Server{
		Addr:           cfg.HTTPAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 固定 1MB
	}
	go func() {
		log.Printf("HTTP 服务启动，监听 %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务失败: %v", err)
		}
	}()

	// 检查是否需要 HTTPS
	needHTTPS := false
	for _, enable := range hostHTTPSMap {
		if enable {
			needHTTPS = true
			break
		}
	}
	if !needHTTPS {
		log.Printf("未启用任何 HTTPS 主机，仅运行 HTTP 服务")
		select {} // 永久阻塞
	}

	if cfg.SSL.CertFile == "" || cfg.SSL.KeyFile == "" {
		log.Fatal("存在启用 HTTPS 的主机，但未配置 ssl.cert_file 和 ssl.key_file")
	}

	// HTTPS 服务
	httpsServer := &http.Server{
		Addr:           cfg.HTTPSAddr,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 固定 1MB
	}
	log.Printf("HTTPS 服务启动，监听 %s", cfg.HTTPSAddr)
	if err := httpsServer.ListenAndServeTLS(cfg.SSL.CertFile, cfg.SSL.KeyFile); err != nil {
		log.Fatalf("HTTPS 服务失败: %v", err)
	}
}
