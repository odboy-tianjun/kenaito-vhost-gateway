package controller

import (
	"encoding/json"
	"kenaito-vhost-gateway/src/dal/dataobject"
	"kenaito-vhost-gateway/src/infra"
	serverService "kenaito-vhost-gateway/src/service/server"
	"net/http"

	"xorm.io/xorm"
)

// ServerController 域名管理控制器
type ServerController struct {
	engine        *xorm.Engine
	serverService *serverService.ServerService
}

// NewServerController 创建域名管理控制器实例
func NewServerController() *ServerController {
	return &ServerController{
		engine:        infra.GetEngine(),
		serverService: serverService.NewServerService(),
	}
}

// ListServers 获取域名列表
// POST /api/servers/list
func (c *ServerController) ListServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var servers []dataobject.Server
	err := c.engine.Find(&servers)
	if err != nil {
		Error(w, http.StatusInternalServerError, "查询域名列表失败: "+err.Error())
		return
	}

	Success(w, servers)
}

// GetServer 获取单个域名信息
// POST /api/servers/get
func (c *ServerController) GetServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var req struct {
		Id int `json:"id"` // 域名ID
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if req.Id == 0 {
		Error(w, http.StatusBadRequest, "缺少参数 id")
		return
	}

	server := &dataobject.Server{}
	has, err := c.engine.ID(req.Id).Get(server)
	if err != nil {
		Error(w, http.StatusInternalServerError, "查询域名失败: "+err.Error())
		return
	}

	if !has {
		Error(w, http.StatusNotFound, "域名不存在")
		return
	}

	Success(w, server)
}

// CreateServer 创建域名
// POST /api/servers/create
func (c *ServerController) CreateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var server dataobject.Server
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if server.ServerName == "" {
		Error(w, http.StatusBadRequest, "域名不能为空")
		return
	}

	_, err := c.engine.Insert(&server)
	if err != nil {
		Error(w, http.StatusInternalServerError, "创建域名失败: "+err.Error())
		return
	}

	Success(w, server)
}

// UpdateServer 更新域名
// POST /api/servers/update
func (c *ServerController) UpdateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var server dataobject.Server
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if server.Id == 0 {
		Error(w, http.StatusBadRequest, "缺少参数 id")
		return
	}

	// 检查域名是否存在
	exists, err := c.engine.ID(server.Id).Exist(&dataobject.Server{})
	if err != nil {
		Error(w, http.StatusInternalServerError, "查询域名失败: "+err.Error())
		return
	}

	if !exists {
		Error(w, http.StatusNotFound, "域名不存在")
		return
	}

	_, err = c.engine.ID(server.Id).Cols("server_name", "active_version", "enable_https").Update(&server)
	if err != nil {
		Error(w, http.StatusInternalServerError, "更新域名失败: "+err.Error())
		return
	}

	Success(w, map[string]string{"message": "更新成功"})
}

// DeleteServer 删除域名
// POST /api/servers/delete
func (c *ServerController) DeleteServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var req struct {
		Id int `json:"id"` // 域名ID
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if req.Id == 0 {
		Error(w, http.StatusBadRequest, "缺少参数 id")
		return
	}

	// 检查域名是否存在
	exists, err := c.engine.ID(req.Id).Exist(&dataobject.Server{})
	if err != nil {
		Error(w, http.StatusInternalServerError, "查询域名失败: "+err.Error())
		return
	}

	if !exists {
		Error(w, http.StatusNotFound, "域名不存在")
		return
	}

	_, err = c.engine.ID(req.Id).Delete(&dataobject.Server{})
	if err != nil {
		Error(w, http.StatusInternalServerError, "删除域名失败: "+err.Error())
		return
	}

	Success(w, map[string]string{"message": "删除成功"})
}

// DeployServer 部署应用（上传目录并记录版本）
// POST /api/servers/deploy
func (c *ServerController) DeployServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var req struct {
		LocalDir   string `json:"localDir"`   // 本地目录路径
		ServerName string `json:"serverName"` // 域名
		AppName    string `json:"appName"`    // 应用名称
		AutoSwitch bool   `json:"autoSwitch"` // 是否自动切换版本
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.LocalDir == "" || req.ServerName == "" || req.AppName == "" {
		Error(w, http.StatusBadRequest, "缺少必填参数: localDir, serverName, appName")
		return
	}

	// 执行部署
	err, version := c.serverService.UploadDirWithServer(req.LocalDir, req.ServerName, req.AppName, req.AutoSwitch)
	if err != nil {
		Error(w, http.StatusInternalServerError, "部署失败: "+err.Error())
		return
	}

	Success(w, map[string]interface{}{
		"message": "部署成功",
		"version": version,
	})
}

// SwitchVersion 切换版本
// POST /api/servers/switchVersion
func (c *ServerController) SwitchVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var req struct {
		Id      int    `json:"id"`      // 域名ID
		Version string `json:"version"` // 版本号
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	if req.Id == 0 || req.Version == "" {
		Error(w, http.StatusBadRequest, "缺少必填参数: id, version")
		return
	}

	err := c.serverService.UpdateActiveVersionById(req.Id, req.Version)
	if err != nil {
		Error(w, http.StatusInternalServerError, "切换版本失败: "+err.Error())
		return
	}

	Success(w, map[string]string{"message": "版本切换成功"})
}
