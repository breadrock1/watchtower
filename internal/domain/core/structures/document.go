package domain

import "time"

type Document struct {
	Content    string    `json:"content"`
	SSDEEP     string    `json:"ssdeep"`
	ID         string    `json:"id"`
	Class      string    `json:"class"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int       `json:"file_size"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	Tokens     Tokens    `json:"embeddings"`
}

func (d *Document) SetClass(class string) {
	if len(class) > 0 {
		d.Class = class
	} else {
		d.Class = "unknown"
	}
}
