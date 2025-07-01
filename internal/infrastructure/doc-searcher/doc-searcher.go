package doc_searcher

import (
	"bytes"
	"context"
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

func (dsc *DocSearcherClient) StoreDocument(ctx context.Context, folder string, doc *dto.StorageDocument) error {
	if err := dsc.storeDocument(ctx, folder, doc); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	if err := dsc.storeTokens(ctx, folder, doc); err != nil {
		log.Printf("failed to store tokens: %v", err)
	}

	return nil
}

func (dsc *DocSearcherClient) storeDocument(ctx context.Context, folder string, doc *dto.StorageDocument) error {
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed while marshaling doc: %w", err)
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
	_, err = utils.PUT(ctx, reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		return fmt.Errorf("failed to store document to storage: %w", err)
	}

	return nil
}

func (dsc *DocSearcherClient) storeTokens(ctx context.Context, folder string, doc *dto.StorageDocument) error {
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed while marshaling doc: %w", err)
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
	_, err = utils.PUT(ctx, reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		return fmt.Errorf("failed to store vectors to storage: %w", err)
	}

	return nil
}

func (dsc *DocSearcherClient) UpdateDocument(ctx context.Context, folder string, document *dto.StorageDocument) error {
	return nil
}

func (dsc *DocSearcherClient) DeleteDocument(ctx context.Context, folder string, id string) error {
	return nil
}

func (dsc *DocSearcherClient) CreateIndex(ctx context.Context, folder string) error {
	return nil
}

func (dsc *DocSearcherClient) DeleteIndex(ctx context.Context, folder string) error {
	return nil
}
