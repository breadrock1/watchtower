package doc_searcher

type CreateIndexForm struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type StoreDocumentResult struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
