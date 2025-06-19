package vectorizer

type Config struct {
	Address      string `mapstructure:"address"`
	EnableSSL    bool   `mapstructure:"enable_ssl"`
	ChunkSize    int    `mapstructure:"chunk_size"`
	ChunkOverlap int    `mapstructure:"chunk_overlap"`
	ReturnChunks bool   `mapstructure:"return_chunks"`
	ChunkBySelf  bool   `mapstructure:"chunk_by_self"`
}
