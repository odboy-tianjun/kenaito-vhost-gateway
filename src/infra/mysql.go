package infra

import (
	"log"
	"time"
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

	// 设置数据库连接池参数
	engine.SetMaxIdleConns(config.DatabaseMaxIdleConns)
	engine.SetMaxOpenConns(config.DatabaseMaxOpenConns)
	engine.SetConnMaxLifetime(time.Duration(config.DatabaseConnMaxLifetime) * time.Second)

	// 启用SQL日志（生产环境建议关闭）
	engine.ShowSQL(config.DatabaseShowSql)

	log.Println("数据库连接初始化成功")
	return nil
}
