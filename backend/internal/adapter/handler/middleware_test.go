package handler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	// Capture log output.
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	handler := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	logged := buf.String()
	assert.Contains(t, logged, "/test-path", "log output must contain the request path")
}

func TestCORSMiddleware_SetsAccessControlHeader(t *testing.T) {
	handler := makeCORSMiddleware([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, origin, "Access-Control-Allow-Origin header must be set")
}

func TestRecoveryMiddleware_CatchesPanicAndReturns500(t *testing.T) {
	// Suppress log output during this test to avoid test noise.
	log.SetOutput(&bytes.Buffer{})
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	// Must not propagate the panic to the caller.
	assert.NotPanics(t, func() {
		handler.ServeHTTP(w, req)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), "Internal Server Error") ||
		strings.Contains(w.Body.String(), "500"), "response body must indicate server error")
}
