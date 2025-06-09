package doc_searcher

type Config struct {
	Address   string `mapstructure:"address"`
	EnableSSL bool   `mapstructure:"enable_ssl"`
}
