package controller

import (
	"net/http"
)

// Router 路由器，负责注册和分发HTTP请求
type Router struct {
	globalConfigCtrl *GlobalConfigController
	serverCtrl       *ServerController
}

// NewRouter 创建路由器实例
func NewRouter() *Router {
	return &Router{
		globalConfigCtrl: NewGlobalConfigController(),
		serverCtrl:       NewServerController(),
	}
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// 所有接口统一使用 POST 方法
	if req.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	// API 路由匹配
	if path == "/api/config/global/get" {
		r.globalConfigCtrl.GetGlobalConfig(w, req)
		return
	}

	if path == "/api/config/global/update" {
		r.globalConfigCtrl.UpdateGlobalConfig(w, req)
		return
	}

	if path == "/api/servers/list" {
		r.serverCtrl.ListServers(w, req)
		return
	}

	if path == "/api/servers/get" {
		r.serverCtrl.GetServer(w, req)
		return
	}

	if path == "/api/servers/create" {
		r.serverCtrl.CreateServer(w, req)
		return
	}

	if path == "/api/servers/update" {
		r.serverCtrl.UpdateServer(w, req)
		return
	}

	if path == "/api/servers/delete" {
		r.serverCtrl.DeleteServer(w, req)
		return
	}

	if path == "/api/servers/deploy" {
		r.serverCtrl.DeployServer(w, req)
		return
	}

	if path == "/api/servers/switchVersion" {
		r.serverCtrl.SwitchVersion(w, req)
		return
	}

	// 未匹配的路由，交给其他处理器处理
	http.NotFound(w, req)
}
