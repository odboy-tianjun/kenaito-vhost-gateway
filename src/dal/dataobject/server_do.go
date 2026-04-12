package dataobject

type Server struct {
	Id            int    `xorm:"pk autoincr 'id'" json:"id"`                     // ID
	ServerName    string `xorm:"unique notnull 'server_name'" json:"serverName"` // 域名
	ActiveVersion string `xorm:"notnull 'active_version'" json:"activeVersion"`  // 当前激活的版本
	EnableHttps   bool   `xorm:"default 0 'enable_https'" json:"enableHttps"`    // 是否启用HTTPS
}

func (d *Server) TableName() string {
	return "server"
}
