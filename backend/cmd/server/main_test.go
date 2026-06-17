package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleHealth はヘルスチェックエンドポイントが正常に動作することを確認します。
func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handleHealth(w, req)

	if got := w.Code; got != http.StatusOK {
		t.Errorf("ステータスコードが期待値と異なります: got %d, want %d", got, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Typeが期待値と異なります: got %s, want application/json", contentType)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("レスポンスのdataフィールドの型が期待値と異なります")
	}

	if status, ok := data["status"]; !ok || status != "ok" {
		t.Errorf("statusフィールドが期待値と異なります: got %v, want ok", status)
	}
}

// TestLoggingMiddleware はロギングミドルウェアが次のハンドラを正常に呼び出すことを確認します。
func TestLoggingMiddleware(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := loggingMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("ミドルウェアが次のハンドラを呼び出しませんでした")
	}
}

// TestResponseWriterStatus はresponseWriterがステータスコードを正しく記録することを確認します。
func TestResponseWriterStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)

	if rw.status != http.StatusNotFound {
		t.Errorf("statusが期待値と異なります: got %d, want %d", rw.status, http.StatusNotFound)
	}
}
