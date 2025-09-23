package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

	injectTracingToHeader(ctx, req)
	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to send request: %w", err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

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

func injectTracingToHeader(ctx context.Context, req *http.Request) {
	span := trace.SpanFromContext(ctx)
	sCtx := span.SpanContext()

	headers := req.Header
	headers.Add("trace-id", sCtx.TraceID().String())
	headers.Add("span-id", sCtx.SpanID().String())
	headers.Add("trace-flags", sCtx.TraceFlags().String())
	headers.Add("trace-state", sCtx.TraceState().String())
}

func extractTracingFromHeader(ctx context.Context, resp *http.Response) {

}

func BuildTargetURL(host, path string) string {
	targetURL := fmt.Sprintf("%s%s", host, path)
	return targetURL
}
