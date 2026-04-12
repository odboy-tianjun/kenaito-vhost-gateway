package dataobject

type GlobalConfig struct {
	Id          int    `xorm:"pk autoincr 'id'"`
	HttpAddr    string `xorm:"default ':80' 'http_addr'"`
	HttpsAddr   string `xorm:"default ':443' 'https_addr'"`
	MaxBodySize int    `xorm:"default 5242880 'max_body_size'"`
	CertPem     string `xorm:"text notnull 'cert_pem'"`
	KeyPem      string `xorm:"text notnull 'key_pem'"`
}

func (d *GlobalConfig) TableName() string {
	return "global_config"
}
