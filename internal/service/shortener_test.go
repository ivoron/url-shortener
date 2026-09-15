package service

import (
	"context"
	"errors"
	"testing"

	"url-shortener/internal/model"
)

type mockRepository struct {
	saveFn            func(ctx context.Context, url *model.URL) error
	findByShortUrlFn func(ctx context.Context, shortCode string) (*model.URL, error)
	incrementVisitsFn func(ctx context.Context, shortCode string) error
}

func (m *mockRepository) Save(ctx context.Context, url *model.URL) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, url)
	}
	return nil
}

func (m *mockRepository) FindByShortURL(ctx context.Context, shortCode string) (*model.URL, error) {
	if m.findByShortUrlFn != nil {
		return m.findByShortUrlFn(ctx, shortCode)
	}
	return nil, nil
}

func (m *mockRepository) IncrementVisits(ctx context.Context, shortCode string) error {
	if m.incrementVisitsFn != nil {
		return m.incrementVisitsFn(ctx, shortCode)
	}
	return nil
}

// TestShortenURL проверяет логику создания короткой ссылки
func TestShortenURL(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		mockRepo    *mockRepository
		wantErr     bool
	}{
		{
			name:        "Success - valid URL",
			originalURL: "https://golang.org",
			mockRepo: &mockRepository{
				saveFn: func(ctx context.Context, url *model.URL) error {
					// Проверяем, что сервисный слой передает корректные данные в БД
					if url.OriginalUrl != "https://golang.org" {
						t.Errorf("expected URL 'https://golang.org', got %s", url.OriginalUrl)
					}
					if len(url.ShortUrl) != shortURLLength {
						t.Errorf("expected code length %d, got %d", shortURLLength, len(url.ShortUrl))
					}
					return nil
				},
			},
			wantErr: false,
		},
		{
			name:        "Error - empty URL",
			originalURL: "",
			mockRepo:    &mockRepository{},
			wantErr:     true,
		},
		{
			name:        "Error - repository failure",
			originalURL: "https://example.com",
			mockRepo: &mockRepository{
				saveFn: func(ctx context.Context, url *model.URL) error {
					return errors.New("db connection error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewShortenerService(tt.mockRepo)

			result, err := svc.Shorten(context.Background(), tt.originalURL)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result == nil {
				t.Fatal("expected non-nil result on success")
			}
		})
	}
}

// TestGetOriginalURL проверяет поиск ссылки по коду
func TestGetOriginalURL(t *testing.T) {
	tests := []struct {
		name      string
		shortCode string
		mockRepo  *mockRepository
		wantErr   bool
		wantURL   string
	}{
		{
			name:      "Success - existing code",
			shortCode: "aB3k9X",
			mockRepo: &mockRepository{
				findByShortUrlFn: func(ctx context.Context, shortCode string) (*model.URL, error) {
					return &model.URL{
						OriginalUrl: "https://github.com",
						ShortUrl:   "aB3k9X",
					}, nil
				},
			},
			wantErr: false,
			wantURL: "https://github.com",
		},
		{
			name:      "Error - empty short code",
			shortCode: "",
			mockRepo:  &mockRepository{},
			wantErr:   true,
		},
		{
			name:      "Error - URL not found in DB",
			shortCode: "unknown",
			mockRepo: &mockRepository{
				findByShortUrlFn: func(ctx context.Context, shortCode string) (*model.URL, error) {
					return nil, errors.New("record not found")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewShortenerService(tt.mockRepo)

			result, err := svc.Resolve(context.Background(), tt.shortCode)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result.OriginalUrl != tt.wantURL {
				t.Errorf("Resolve() got = %v, want %v", result.OriginalUrl, tt.wantURL)
			}
		})
	}
}