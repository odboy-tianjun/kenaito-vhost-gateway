package dataobject

type ServerVersion struct {
	Id         int    `xorm:"pk autoincr 'id'" json:"id"`
	ServerName string `xorm:"notnull 'server_name'" json:"serverName"`
	Version    string `xorm:"notnull 'version'" json:"version"`
	BucketPath string `xorm:"notnull 'bucket_path'" json:"bucketPath"`
}

func (d *ServerVersion) TableName() string {
	return "server_version"
}
