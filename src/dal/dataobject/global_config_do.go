package dataobject

type GlobalConfig struct {
	Id          int    `xorm:"pk autoincr 'id'" json:"id"`                         // ID
	HttpAddr    string `xorm:"default ':80' 'http_addr'" json:"httpAddr"`          // HTTP地址
	HttpsAddr   string `xorm:"default ':443' 'https_addr'" json:"httpsAddr"`       // HTTPS地址
	MaxBodySize int    `xorm:"default 5242880 'max_body_size'" json:"maxBodySize"` // 最大请求体大小
	CertPem     string `xorm:"text notnull 'cert_pem'" json:"certPem"`             // 证书PEM
	KeyPem      string `xorm:"text notnull 'key_pem'" json:"keyPem"`               // 私钥PEM
}

func (d *GlobalConfig) TableName() string {
	return "global_config"
}
