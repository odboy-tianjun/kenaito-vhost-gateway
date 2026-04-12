package controller

import (
	"encoding/json"
	"net/http"

	"kenaito-vhost-gateway/src/dal/dataobject"
	"kenaito-vhost-gateway/src/infra"

	"xorm.io/xorm"
)

// GlobalConfigController 全局配置控制器
type GlobalConfigController struct {
	engine *xorm.Engine
}

// NewGlobalConfigController 创建全局配置控制器实例
func NewGlobalConfigController() *GlobalConfigController {
	return &GlobalConfigController{
		engine: infra.GetEngine(),
	}
}

// GetGlobalConfig 获取全局配置
// POST /api/config/global/get
func (c *GlobalConfigController) GetGlobalConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	config := &dataobject.GlobalConfig{}
	has, err := c.engine.ID(1).Get(config)
	if err != nil {
		Error(w, http.StatusInternalServerError, "查询配置失败: "+err.Error())
		return
	}

	if !has {
		Error(w, http.StatusNotFound, "全局配置不存在")
		return
	}

	Success(w, config)
}

// UpdateGlobalConfig 更新全局配置
// POST /api/config/global
func (c *GlobalConfigController) UpdateGlobalConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "只支持 POST 请求")
		return
	}

	var config dataobject.GlobalConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		Error(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 检查配置是否存在
	has, err := c.engine.ID(1).Get(&dataobject.GlobalConfig{})
	if err != nil {
		Error(w, http.StatusInternalServerError, "查询配置失败: "+err.Error())
		return
	}

	if has {
		// 更新现有配置
		_, err = c.engine.ID(1).Cols("http_addr", "https_addr", "max_body_size", "cert_pem", "key_pem").Update(&config)
		if err != nil {
			Error(w, http.StatusInternalServerError, "更新配置失败: "+err.Error())
			return
		}
	} else {
		// 插入新配置
		config.Id = 1
		_, err = c.engine.Insert(&config)
		if err != nil {
			Error(w, http.StatusInternalServerError, "创建配置失败: "+err.Error())
			return
		}
	}

	Success(w, map[string]string{"message": "配置更新成功"})
}
