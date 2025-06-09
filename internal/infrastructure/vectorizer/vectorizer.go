package vectorizer

import (
	"bytes"
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

const EmbeddingsAssistantURL = "/embed"

type VectorizerClient struct {
	config *Config
}

func New(config *Config) *VectorizerClient {
	return &VectorizerClient{
		config: config,
	}
}

func (vc *VectorizerClient) Load(inputText string) (*dto.Tokens, error) {
	if vc.config.ChunkBySelf {
		return vc.LoadByOwnChunked(inputText)
	}

	textVectors := &dto.TokensInputForm{
		Inputs:    inputText,
		Truncate:  false,
		Normalize: true,
	}

	jsonData, err := json.Marshal(textVectors)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tokens: %v", err)
	}

	reqBody := bytes.NewBuffer(jsonData)
	mimeType := echo.MIMEApplicationJSON
	timeoutReq := time.Duration(300) * time.Second
	targetURL := utils.BuildTargetURL(vc.config.EnableSSL, vc.config.Address, EmbeddingsAssistantURL)
	respData, err := utils.POST(reqBody, targetURL, mimeType, timeoutReq)
	if err != nil {
		return nil, fmt.Errorf("failed to load embeddings: %v", err)
	}

	tokensResult := &dto.Tokens{}
	err = json.Unmarshal(respData, &tokensResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal embeddings: %v", err)
	}

	if tokensResult.Chunks < 1 {
		return nil, errors.New("returned empty tokens from tokenizer service")
	}

	return tokensResult, nil
}

func (vc *VectorizerClient) LoadByOwnChunked(inputText string) (*dto.Tokens, error) {
	contentData := strings.ReplaceAll(inputText, "\n", " ")
	chunkedText := vc.splitContent(contentData, vc.config.ChunkSize)

	tokensResult := &dto.Tokens{}
	for _, textData := range chunkedText {

		result, err := vc.Load(textData)
		if err != nil {
			log.Printf("failed to load tokens: %v", err)
			continue
		}

		tokensResult.Chunks++
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
	for i := 0; i < splitLength; i++ {
		start = i * chunkSize
		stop = start + chunkSize
		if stop > strLength {
			stop = strLength
		}

		splitString[i] = content[start:stop]
	}

	return splitString
}
