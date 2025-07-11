package doc_searcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
)

const DocumentJsonMime = "application/json"

type DocSearcherClient struct {
	config *Config
}

func New(config *Config) *DocSearcherClient {
	return &DocSearcherClient{
		config: config,
	}
}

func (dsc *DocSearcherClient) StoreDocument(
	ctx context.Context,
	folder string,
	doc *dto.StorageDocument,
) (string, error) {
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed while marshaling doc: %w", err)
	}

	buildURL := strings.Builder{}
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s/create", folder))
	targetURL := buildURL.String()

	log.Printf("storing document to index %s", folder)

	reqBody := bytes.NewBuffer(jsonData)
	timeoutReq := time.Duration(300) * time.Second
	_, err = utils.PUT(ctx, reqBody, targetURL, DocumentJsonMime, timeoutReq)
	if err != nil {
		return "", fmt.Errorf("failed to store document to storage: %w", err)
	}

	status := &StoreDocumentResult{}
	err = json.Unmarshal(reqBody.Bytes(), status)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return status.Message, nil
}

func (dsc *DocSearcherClient) UpdateDocument(ctx context.Context, folder string, document *dto.StorageDocument) error {
	return nil
}

func (dsc *DocSearcherClient) DeleteDocument(ctx context.Context, folder, id string) error {
	buildURL := strings.Builder{}
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s/%s", folder, id))
	targetURL := buildURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, targetURL, bytes.NewReader([]byte{}))
	if err != nil {
		return fmt.Errorf("failed while creating new request: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(100) * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed while sending request: %w", err)
	}

	if response.StatusCode/100 > 2 {
		return fmt.Errorf("bad response error status %s", response.Status)
	}

	return nil
}

func (dsc *DocSearcherClient) CreateIndex(ctx context.Context, folder string) error {
	form := &CreateIndexForm{
		folder,
		folder,
		"./",
	}

	data, err := json.Marshal(form)
	if err != nil {
		return fmt.Errorf("failed while marshaling form: %w", err)
	}

	buildURL := strings.Builder{}
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s", folder))
	targetURL := buildURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed while creating new request: %w", err)
	}

	req.Header.Set("Content-Type", DocumentJsonMime)
	client := &http.Client{Timeout: time.Duration(100) * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed while sending request: %w", err)
	}

	if response.StatusCode/100 > 2 {
		return fmt.Errorf("bad response error status: %s", response.Status)
	}

	return nil
}

func (dsc *DocSearcherClient) DeleteIndex(ctx context.Context, folder string) error {
	buildURL := strings.Builder{}
	buildURL.WriteString(dsc.config.Address)
	buildURL.WriteString(fmt.Sprintf("/storage/%s", folder))
	targetURL := buildURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, targetURL, bytes.NewReader([]byte{}))
	if err != nil {
		return fmt.Errorf("failed while creating new request: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(100) * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed while sending request: %w", err)
	}

	if response.StatusCode/100 > 2 {
		return fmt.Errorf("bad response error status %s", response.Status)
	}

	return nil
}
