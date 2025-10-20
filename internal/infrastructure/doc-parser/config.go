package doc_parser

import "time"

type Config struct {
	Address string        `mapstructure:"address"`
	Timeout time.Duration `mapstructure:"timeout"`
}
