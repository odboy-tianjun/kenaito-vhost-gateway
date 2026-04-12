package infra

import (
	"log"
	"xorm.io/xorm"

	"github.com/magiconair/properties"
)

var engine *xorm.Engine

// GetEngine 获取数据库引擎实例
func GetEngine() *xorm.Engine {
	return engine
}

// InitDatabase 初始化数据库连接（由 main 函数显式调用）
func InitDatabase() error {
	// 加载配置文件
	p, err := properties.LoadFile("config.properties", properties.UTF8)
	if err != nil {
		return err
	}

	// 从配置文件中读取数据库连接字符串
	dsn, ok := p.Get("database.dsn")
	if !ok {
		log.Fatal("配置文件中缺少 database.dsn 配置项")
	}

	engine, err = xorm.NewEngine("mysql", dsn)
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
