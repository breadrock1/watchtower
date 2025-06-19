package vectorizer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"watchtower/internal/application/dto"
	"watchtower/internal/application/utils"
)

const EmbeddingsURL = "/embed"

type VectorizerClient struct {
	config *Config
}

func New(config *Config) *VectorizerClient {
	return &VectorizerClient{
		config: config,
	}
}

func (vc *VectorizerClient) Load(ctx context.Context, inputText string) (*dto.ComputedTokens, error) {
	if vc.config.ChunkBySelf {
		return vc.LoadByOwnChunked(ctx, inputText)
	}

	textVectors := &dto.ComputeTokensForm{
		Inputs:    inputText,
		Truncate:  false,
		Normalize: true,
	}

	jsonData, err := json.Marshal(textVectors)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tokens: %w", err)
	}

	reqBody := bytes.NewBuffer(jsonData)
	mimeType := echo.MIMEApplicationJSON
	timeoutReq := time.Duration(300) * time.Second
	targetURL := utils.BuildTargetURL(vc.config.EnableSSL, vc.config.Address, EmbeddingsURL)
	respData, err := utils.POST(ctx, reqBody, targetURL, mimeType, timeoutReq)
	if err != nil {
		return nil, fmt.Errorf("failed to load embeddings: %w", err)
	}

	tokensResult := &dto.ComputedTokens{}
	err = json.Unmarshal(respData, &tokensResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal embeddings: %w", err)
	}

	if tokensResult.ChunksCount < 1 {
		return nil, errors.New("returned empty tokens from tokenizer service")
	}

	return tokensResult, nil
}

func (vc *VectorizerClient) LoadByOwnChunked(ctx context.Context, inputText string) (*dto.ComputedTokens, error) {
	contentData := strings.ReplaceAll(inputText, "\n", " ")
	chunkedText := vc.splitContent(contentData, vc.config.ChunkSize)

	tokensResult := &dto.ComputedTokens{}
	for _, textData := range chunkedText {
		result, err := vc.Load(ctx, textData)
		if err != nil {
			log.Printf("failed to load tokens: %v", err)
			continue
		}

		tokensResult.ChunksCount++
		tokensResult.Vectors = append(tokensResult.Vectors, result.Vectors[0])
		tokensResult.ChunkedText = append(tokensResult.ChunkedText, result.ChunkedText[0])
	}

	return tokensResult, nil
}

func (vc *VectorizerClient) splitContent(content string, chunkSize int) []string {
	strLength := len(content)
	splitLength := int(math.Ceil(float64(strLength) / float64(chunkSize)))
	splitString := make([]string, splitLength)
	var start, stop int
	for charIndex := 0; charIndex < splitLength; charIndex++ {
		start = charIndex * chunkSize
		stop = start + chunkSize
		if stop > strLength {
			stop = strLength
		}

		splitString[charIndex] = content[start:stop]
	}

	return splitString
}
