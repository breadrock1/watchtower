package dto

type Tokens struct {
	Chunks      int         `json:"chunks"`
	ChunkedText []string    `json:"chunked_text"`
	Vectors     [][]float64 `json:"vectors"`
}

type TokensInputForm struct {
	Inputs    string `json:"inputs"`
	Truncate  bool   `json:"truncate"`
	Normalize bool   `json:"normalize"`
}
