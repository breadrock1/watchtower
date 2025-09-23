package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"watchtower/internal/application/utils/telemetry"
)

func PUT(ctx context.Context, body *bytes.Buffer, url, mime string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(echo.HeaderContentType, mime)
	client := &http.Client{Timeout: timeout}
	return SendRequest(ctx, client, req)
}

func POST(ctx context.Context, body *bytes.Buffer, url, mime string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(echo.HeaderContentType, mime)
	client := &http.Client{Timeout: timeout}
	return SendRequest(ctx, client, req)
}

func SendRequest(ctx context.Context, client *http.Client, req *http.Request) ([]byte, error) {
	ctx, span := telemetry.GlobalTracer.Start(ctx, "http-request")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.method", req.Method),
		attribute.String("request.uri", req.RequestURI),
	)

	injectSpanContext(ctx, req)
	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to send request: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	ctx = extractSpanContext(ctx, response)
	respData, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	if response.StatusCode/100 > 2 {
		err = fmt.Errorf("non success response %s: %s", response.Status, string(respData))
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return respData, nil
}

func extractSpanContext(ctx context.Context, resp *http.Response) context.Context {
	propagator := telemetry.TracePropagator
	carrier := propagation.HeaderCarrier(resp.Header)
	return propagator.Extract(ctx, carrier)
}

func injectSpanContext(ctx context.Context, req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.HeaderCarrier(req.Header)
	propagator.Inject(ctx, carrier)
}

func BuildTargetURL(host, path string) string {
	targetURL := fmt.Sprintf("%s%s", host, path)
	return targetURL
}
