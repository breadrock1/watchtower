package doc_storage

type CreateIndexForm struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type StoreDocumentForm struct {
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	FileSize   int    `json:"file_size"`
	Content    string `json:"content"`
	CreatedAt  int64  `json:"created_at"`
	ModifiedAt int64  `json:"modified_at"`
}

type StoreDocumentResult struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
