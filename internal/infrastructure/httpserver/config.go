package httpserver

type Config struct {
	Address string       `mapstructure:"address"`
	Logger  LoggerConfig `mapstructure:"logger"`
	Tracer  TracerConfig `mapstructure:"tracer"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Address    string `mapstructure:"address"`
	EnableLoki bool   `mapstructure:"enable_loki"`
}

type TracerConfig struct {
	Address      string `mapstructure:"address"`
	EnableJaeger bool   `mapstructure:"enable_jaeger"`
}
