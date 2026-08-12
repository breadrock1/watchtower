package utils

import (
	"bytes"
	"context"
	otlp_go "github.com/breadrock1/otlp-go/otlp"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testBaseURL           = "http://host"
	testPath              = "/path"
	testTargetURL         = "http://host/path"
	testContentType       = "text/plain"
	testResponseBody      = "ok"
	testPostBody          = "post"
	testPutBody           = "put"
	testRequestBody       = "body"
	testErrorResponseBody = "failed"
)

func init() {
	otlp_go.GlobalTracer = otel.Tracer("sender-test")
}
func TestHTTPSender(t *testing.T) {
	t.Run("Build target URL", func(t *testing.T) {
		targetURL := BuildTargetURL(testBaseURL, testPath)

		assert.Equal(t, testTargetURL, targetURL, "unexpected target url")
	})

	t.Run("Send POST request", func(t *testing.T) {
		var gotMethod string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method

			assert.Equal(t, testContentType, r.Header.Get("Content-Type"), "unexpected content type")

			_, _ = w.Write([]byte(testResponseBody))
		}))
		defer server.Close()

		resp, err := POST(
			context.Background(),
			bytes.NewBufferString(testPostBody), 
			server.URL, 
			testContentType, 
			time.Second,
		)

		assert.NoError(t, err, "failed to send post request")
		assert.Equal(t, testResponseBody, string(resp), "unexpected post response")
		assert.Equal(t, http.MethodPost, gotMethod, "unexpected request methods")
	})

	t.Run("Send PUT request", func(t *testing.T) {
		var gotMethod string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method

			assert.Equal(t, testContentType, r.Header.Get("Content-Type"), "unexpected content type")

			_, _ = w.Write([]byte(testResponseBody))
		}))
		defer server.Close()

		resp, err := PUT(
			context.Background(), 
			bytes.NewBufferString(testPutBody), 
			server.URL, 
			testContentType, 
			time.Second,
		)

		assert.NoError(t, err, "failed to send put request")
		assert.Equal(t, testResponseBody, string(resp), "unexpected put response")
		assert.Equal(t, http.MethodPut, gotMethod, "unexpected request methods")
	})

	t.Run("Return error on non success status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, testErrorResponseBody, http.StatusBadGateway)
		}))
		defer server.Close()

		_, err := POST(
			context.Background(), 
			bytes.NewBufferString(testRequestBody), 
			server.URL, 
			testContentType, 
			time.Second,
		)

		assert.Error(t, err, "expected error for non-success response")
	})
}
