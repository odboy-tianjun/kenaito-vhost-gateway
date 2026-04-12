package dataobject

type Server struct {
	Id            int    `xorm:"pk autoincr 'id'"`
	ServerName    string `xorm:"unique notnull 'server_name'"`
	ActiveVersion string `xorm:"notnull 'active_version'"`
	EnableHttps   bool   `xorm:"default 0 'enable_https'"`
}

func (d *Server) TableName() string {
	return "server"
}
