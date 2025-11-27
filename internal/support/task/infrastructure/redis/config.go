package redis

import "time"

type Config struct {
	Address  string        `mapstructure:"address"`
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
	Expired  time.Duration `mapstructure:"expired"`
}
