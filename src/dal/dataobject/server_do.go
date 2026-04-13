package dataobject

type Server struct {
	Id              int    `xorm:"pk autoincr 'id'" json:"id"`                         // ID
	ServerName      string `xorm:"unique notnull 'server_name'" json:"serverName"`     // 域名
	EnableHttps     bool   `xorm:"default 0 'enable_https'" json:"enableHttps"`        // 是否启用HTTPS
	ActiveVersion   string `xorm:"notnull 'active_version'" json:"activeVersion"`      // 当前激活的版本
	GrayVersion     string `xorm:"notnull 'gray_version'" json:"grayVersion"`          // 灰度版本
	GrayHeaderKey   string `xorm:"notnull 'gray_header_key'" json:"grayHeaderKey"`     // 灰度key
	GrayHeaderValue string `xorm:"notnull 'gray_header_value'" json:"grayHeaderValue"` // 灰度value
}

func (d *Server) TableName() string {
	return "server"
}
