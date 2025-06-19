package httpserver

type Config struct {
	Address string       `mapstructure:"address"`
	Logger  LoggerConfig `mapstructure:"logger"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Address    string `mapstructure:"address"`
	EnableLoki bool   `mapstructure:"enable_loki"`
}
