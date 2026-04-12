package server

import (
	"fmt"
	"kenaito-vhost-gateway/src/dal/dataobject"
	"kenaito-vhost-gateway/src/infra"
	"xorm.io/xorm"
)

// ServerService 域名服务，提供域名相关的业务逻辑
type ServerService struct {
	engine *xorm.Engine
}

// NewServerService 创建域名服务实例
func NewServerService() *ServerService {
	return &ServerService{
		engine: infra.GetEngine(),
	}
}

// GetServerByName 根据域名查询服务器配置
func (s *ServerService) GetServerByName(serverName string) (*dataobject.Server, error) {
	var server dataobject.Server
	has, err := s.engine.Where("server_name = ?", serverName).Get(&server)
	if err != nil {
		return nil, fmt.Errorf("查询域名配置失败: %v", err)
	}
	if !has {
		return nil, nil
	}
	return &server, nil
}

// GetServerVersion 根据域名和版本号查询版本信息
func (s *ServerService) GetServerVersion(serverName, version string) (*dataobject.ServerVersion, error) {
	var serverVersion dataobject.ServerVersion
	has, err := s.engine.Where("server_name = ? AND version = ?", serverName, version).Get(&serverVersion)
	if err != nil {
		return nil, fmt.Errorf("查询版本配置失败: %v", err)
	}
	if !has {
		return nil, nil
	}
	return &serverVersion, nil
}

// CountHttpsServers 统计启用 HTTPS 的域名数量
func (s *ServerService) CountHttpsServers() (int64, error) {
	count, err := s.engine.Where("enable_https = ?", true).Count(&dataobject.Server{})
	if err != nil {
		return 0, fmt.Errorf("查询 HTTPS 主机失败: %v", err)
	}
	return count, nil
}
