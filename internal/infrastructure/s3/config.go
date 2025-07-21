package s3

type Config struct {
	Address   string `mapstructure:"address"`
	AccessID  string `mapstructure:"access_id"`
	SecretKey string `mapstructure:"secret_key"`
	Token     string `mapstructure:"token"`
	EnableSSL bool   `mapstructure:"enable_ssl"`
}
