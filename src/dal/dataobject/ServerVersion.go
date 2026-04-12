package dataobject

type ServerVersion struct {
	Id         int    `xorm:"pk autoincr 'id'"`
	ServerName string `xorm:"notnull 'server_name'"`
	Version    string `xorm:"notnull 'version'"`
	BucketPath string `xorm:"notnull 'bucket_path'"`
}

func (d *ServerVersion) TableName() string {
	return "server_version"
}
