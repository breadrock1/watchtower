package doc_storage

type Config struct {
	Address   string `mapstructure:"address"`
	EnableSSL bool   `mapstructure:"enable_ssl"`
}
