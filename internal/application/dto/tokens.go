package dto

type ComputedTokens struct {
	ChunksCount int         `json:"chunks"`
	ChunkedText []string    `json:"chunked_text"`
	Vectors     [][]float64 `json:"vectors"`
}

type ComputeTokensForm struct {
	Inputs    string `json:"inputs"`
	Truncate  bool   `json:"truncate"`
	Normalize bool   `json:"normalize"`
}
