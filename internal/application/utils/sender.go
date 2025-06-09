package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func PUT(body *bytes.Buffer, url, mime string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(echo.HeaderContentType, mime)
	client := &http.Client{Timeout: timeout}
	return SendRequest(client, req)
}

func POST(body *bytes.Buffer, url, mime string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(echo.HeaderContentType, mime)
	client := &http.Client{Timeout: timeout}
	return SendRequest(client, req)
}

func SendRequest(client *http.Client, req *http.Request) ([]byte, error) {
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	respData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode > 200 {
		return nil, fmt.Errorf("non success response %s: %s", response.Status, string(respData))
	}

	return respData, nil
}

func BuildTargetURL(enableSSL bool, host, path string) string {
	httpSchema := GetHttpSchema(enableSSL)
	targetURL := fmt.Sprintf("%s://%s%s", httpSchema, host, path)
	return targetURL
}

func GetHttpSchema(enableSSL bool) string {
	if enableSSL {
		return "https"
	} else {
		return "http"
	}
}
