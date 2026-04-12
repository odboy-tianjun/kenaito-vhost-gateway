package infra

import (
	"log"
	"xorm.io/xorm"
)

var engine *xorm.Engine

// GetEngine 获取数据库引擎实例
func GetEngine() *xorm.Engine {
	return engine
}

// InitDatabase 初始化数据库连接（由 main 函数显式调用）
func InitDatabase() error {
	// 从应用配置中获取数据库连接字符串
	config := GetAppConfig()

	var err error
	engine, err = xorm.NewEngine("mysql", config.DatabaseDsn)
	if err != nil {
		return err
	}

	// 设置最大空闲连接数
	engine.SetMaxIdleConns(10)
	// 设置最大打开连接数
	engine.SetMaxOpenConns(100)
	// 设置连接的最大生命周期（秒）
	engine.SetConnMaxLifetime(300)

	// 启用SQL日志（生产环境可关闭）
	engine.ShowSQL(true)

	log.Println("数据库连接初始化成功")
	return nil
}
