package doc_searcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
)

const DocumentJsonMime = echo.MIMEApplicationJSON

type DocSearcherClient struct {
	config *Config
}

func New(config *Config) *DocSearcherClient {
	return &DocSearcherClient{
		config: config,
	}
}

func (dsc *DocSearcherClient) Store(folder string, doc *dto.StorageDocument) error {
	if err := dsc.storeDocument(folder, doc); err != nil {
		return fmt.Errorf("failed to store document: %v", err)
	}

	if err := dsc.storeTokens(folder, doc); err != nil {
		log.Printf("failed to store tokens: %v", err)
	}

	return nil
}

func (dsc *DocSearcherClient) storeDocument(folder string, doc *dto.StorageDocument) error {
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed while marshaling doc: %v", err)
	}

	buildURL := strings.Builder{}
	buildURL.WriteString(utils.GetHttpSchema(dsc.config.EnableSSL))
	buildURL.WriteString("://")
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString("/storage/folders/")
	buildURL.WriteString(folder)
	buildURL.WriteString("/documents/create")
	targetURL := buildURL.String()

	log.Printf("storing document %s to index %s", doc.ID, folder)

	reqBody := bytes.NewBuffer(jsonData)
	timeoutReq := time.Duration(300) * time.Second
	_, err = utils.PUT(reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		return err
	}

	return nil
}

func (dsc *DocSearcherClient) storeTokens(folder string, doc *dto.StorageDocument) error {
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed while marshaling doc: %v", err)
	}

	folderID := fmt.Sprintf("%s-vector", folder)

	buildURL := strings.Builder{}
	buildURL.WriteString(utils.GetHttpSchema(dsc.config.EnableSSL))
	buildURL.WriteString("://")
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString("/storage/folders/")
	buildURL.WriteString(folderID)
	buildURL.WriteString("/documents/create")
	buildURL.WriteString("?folder_type=vectors")
	targetURL := buildURL.String()

	log.Printf("storing document %s to index %s", doc.ID, folder)

	reqBody := bytes.NewBuffer(jsonData)
	timeoutReq := time.Duration(300) * time.Second
	_, err = utils.PUT(reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		return err
	}

	return nil
}
