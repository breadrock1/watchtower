package telemetry

type OtlpConfig struct {
	Logger LoggerConfig `yaml:"logger"`
	Tracer TracerConfig `yaml:"tracer"`
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
