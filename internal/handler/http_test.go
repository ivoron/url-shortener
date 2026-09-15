package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/model"
)

type mockService struct {
	shortenURLFn     func(ctx context.Context, originalURL string) (*model.URL, error)
	getOriginalURLFn func(ctx context.Context, shortCode string) (*model.URL, error)
}

func (m *mockService) Shorten(ctx context.Context, originalURL string) (*model.URL, error) {
	if m.shortenURLFn != nil {
		return m.shortenURLFn(ctx, originalURL)
	}
	return nil, nil
}

func (m *mockService) Resolve(ctx context.Context, shortCode string) (*model.URL, error) {
	if m.getOriginalURLFn != nil {
		return m.getOriginalURLFn(ctx, shortCode)
	}
	return nil, nil
}

func TestHandler_ShortenURL(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSvc        *mockService
		expectedStatus int
	}{
		{
			name:        "Success - 201 Created",
			requestBody: `{"url": "https://golang.org"}`,
			mockSvc: &mockService{
				shortenURLFn: func(ctx context.Context, originalURL string) (*model.URL, error) {
					return &model.URL{
						OriginalUrl: "https://golang.org",
						ShortUrl:   "aB3k9X",
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Error - 400 Bad JSON",
			requestBody:    `{"invalid_json":}`,
			mockSvc:        &mockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Error - 500 Internal Error",
			requestBody: `{"url": "https://example.com"}`,
			mockSvc: &mockService{
				shortenURLFn: func(ctx context.Context, originalURL string) (*model.URL, error) {
					return nil, errors.New("db error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(tt.mockSvc)
			router := h.InitRoutes()

			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	tests := []struct {
		name             string
		shortCode        string
		mockSvc          *mockService
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:      "Success - 301 Redirect",
			shortCode: "aB3k9X",
			mockSvc: &mockService{
				getOriginalURLFn: func(ctx context.Context, shortCode string) (*model.URL, error) {
					return &model.URL{
						OriginalUrl: "https://github.com",
						ShortUrl:   "aB3k9X",
					}, nil
				},
			},
			expectedStatus:   http.StatusMovedPermanently,
			expectedLocation: "https://github.com",
		},
		{
			name:      "Error - 404 Not Found",
			shortCode: "unknown",
			mockSvc: &mockService{
				getOriginalURLFn: func(ctx context.Context, shortCode string) (*model.URL, error) {
					return nil, errors.New("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHTTPHandler(tt.mockSvc)
			router := h.InitRoutes()

			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortCode, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedLocation != "" {
				location := rec.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("expected Location header %s, got %s", tt.expectedLocation, location)
				}
			}
		})
	}
}