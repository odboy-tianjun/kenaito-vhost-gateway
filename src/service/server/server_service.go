package server

import (
	"fmt"
	"kenaito-vhost-gateway/src/dal/dataobject"
	"kenaito-vhost-gateway/src/infra"
	"kenaito-vhost-gateway/src/infra/oss"
	"log"
	"time"
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

// UploadDirWithServer 上传目录，并新增ServerVersion记录
// localDir: 本地目录路径
// serverName: 域名（如 example.com）
// appName: 应用名称（如 cutejava-front）
// autoSwitch: 是否自动切换新版本
func (s *ServerService) UploadDirWithServer(localDir string, serverName string, appName string, autoSwitch bool) (error, string) {
	// 查询域名是否存在
	server, err := s.GetServerByName(serverName)
	if err != nil {
		return fmt.Errorf("查询域名失败: %v", err), ""
	}

	// 如果域名不存在，则创建新的 Server 记录
	if server == nil {
		server = &dataobject.Server{
			ServerName:    serverName,
			ActiveVersion: "",
			EnableHttps:   false,
		}
		_, err = s.engine.Insert(server)
		if err != nil {
			return fmt.Errorf("创建域名记录失败: %v", err), ""
		}
		log.Printf("域名记录创建成功: %s", serverName)
	}

	// 生成版本号（yyyyMMddHHmmss格式）
	version := time.Now().Format("20060102150405")

	// 构造 MinIO 路径：appName/version
	minioPrefix := fmt.Sprintf("%s/%s", appName, version)

	// 上传文件到 MinIO
	log.Printf("开始上传目录 %s 到 MinIO: %s", localDir, minioPrefix)
	err = oss.UploadDirectoryToMinio(localDir, minioPrefix)
	if err != nil {
		return fmt.Errorf("上传文件到 MinIO 失败: %v", err), ""
	}

	// 创建版本记录
	serverVersion := &dataobject.ServerVersion{
		ServerName: serverName,
		Version:    version,
		BucketPath: minioPrefix,
	}

	_, err = s.engine.Insert(serverVersion)
	if err != nil {
		return fmt.Errorf("创建版本记录失败: %v", err), ""
	}

	log.Printf("版本记录创建成功: %s -> %s (路径: %s)", serverName, version, minioPrefix)

	if autoSwitch {
		err2 := s.UpdateActiveVersionById(server.Id, version)
		if err2 != nil {
			return err2, ""
		}
	}

	return nil, version
}

// UpdateActiveVersionById 更新域名版本
// id server表id
// version server表active_version
func (s *ServerService) UpdateActiveVersionById(id int, version string) error {
	// 根据id查询server记录
	server := &dataobject.Server{}
	has, err := s.engine.ID(id).Get(server)
	if err != nil {
		return fmt.Errorf("查询域名记录失败: %v", err)
	}
	if !has {
		return fmt.Errorf("域名记录不存在，id: %d", id)
	}

	// 更新active_version字段
	server.ActiveVersion = version
	_, err = s.engine.ID(id).Cols("active_version").Update(server)
	if err != nil {
		return fmt.Errorf("更新域名版本失败: %v", err)
	}

	log.Printf("域名 %s 的版本已更新为: %s", server.ServerName, version)
	return nil
}
