package config

import (
	"fmt"
	"kenaito-vhost-gateway/src/dal/dataobject"
	"kenaito-vhost-gateway/src/infra"
	"log"
)

var globalConfig *dataobject.GlobalConfig

// LoadGlobalConfig 加载全局配置并初始化 MinIO 客户端
func LoadGlobalConfig() (*dataobject.GlobalConfig, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	engine := infra.GetEngine()
	if engine == nil {
		return nil, fmt.Errorf("数据库引擎未初始化，请先调用 infra.InitDatabase()")
	}

	global := &dataobject.GlobalConfig{}
	has, err := engine.ID(1).Get(global)
	if err != nil {
		return nil, fmt.Errorf("读取全局配置失败: %v", err)
	}
	if !has {
		return nil, fmt.Errorf("数据库中没有全局配置，请先插入一条记录 (id=1)")
	}

	// 保存全局配置
	globalConfig = global

	// 初始化 MinIO 客户端
	err = infra.InitMinioClient()
	if err != nil {
		return nil, fmt.Errorf("初始化 MinIO 客户端失败: %v", err)
	}

	// 检查存储桶是否存在
	infra.CheckDefaultBucketExist()

	log.Println("全局配置加载成功")
	return global, nil
}

// GetGlobalConfig 获取全局配置（不重新加载）
func GetGlobalConfig() *dataobject.GlobalConfig {
	return globalConfig
}
