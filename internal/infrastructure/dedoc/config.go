package dedoc

import "time"

type Config struct {
	Address   string        `mapstructure:"address"`
	EnableSSL bool          `mapstructure:"enable_ssl"`
	Timeout   time.Duration `mapstructure:"timeout"`
}
