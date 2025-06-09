package dto

import (
	"time"

	"watchtower/internal/application/utils"
)

type Document struct {
	FolderID          string    `json:"folder_id"`
	FolderPath        string    `json:"folder_path"`
	Content           string    `json:"content"`
	DocumentID        string    `json:"document_id"`
	DocumentSSDEEP    string    `json:"document_ssdeep"`
	DocumentName      string    `json:"document_name"`
	DocumentPath      string    `json:"document_path"`
	DocumentClass     string    `json:"document_class"`
	DocumentSize      int64     `json:"document_size"`
	DocumentExtension string    `json:"document_extension"`
	Tokens            Tokens    `json:"embeddings"`
	DocumentCreated   time.Time `json:"document_created"`
	DocumentModified  time.Time `json:"document_modified"`
}

func (d *Document) SetDocumentClass(class string) {
	if len(class) > 0 {
		d.DocumentClass = class
	} else {
		d.DocumentClass = "unknown"
	}
}

func (d *Document) ComputeMd5Hash() {
	data := []byte(d.Content)
	res := utils.ComputeMd5(data)
	d.DocumentID = res
}

func (d *Document) ComputeSsdeepHash() {
	data := []byte(d.Content)
	res, _ := utils.ComputeSSDEEP(data)
	d.DocumentSSDEEP = res
}
