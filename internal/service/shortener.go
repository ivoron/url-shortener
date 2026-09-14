package service

import (
	"context"
	"errors"
	"math/rand"
	"strings"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

const (
	symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortURLLength = 6
)

type ShortenerService interface {
	Shorten(ctx context.Context, originalURL string) (*model.URL, error)
	Resolve(ctx context.Context, shortURL string) (*model.URL, error)
}

type shortenerService struct {
	repo repository.URLRepository
}

func NewShortenerService(repo repository.URLRepository) ShortenerService {
	return &shortenerService{
		repo: repo,
	}
}

func (s *shortenerService) Shorten(ctx context.Context, originalUrl string) (*model.URL, error) {
	if originalUrl == "" {
		return nil, errors.New("URL is empty")
	}

	shortUrl := generateShortURL()

	url := &model.URL{
		OriginalUrl: originalUrl,
		ShortUrl:    shortUrl,
	}

	err := s.repo.Save(ctx, url)

	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *shortenerService) Resolve(ctx context.Context, shortURL string) (*model.URL, error) {
	url, err := s.repo.FindByShortURL(ctx, shortURL)

	if err != nil {
		return nil, err
	}

	if url == nil {
		return nil, errors.New("URL not found")
	}

	go func() {
		bgCtx := context.Background()
		
		_ = s.repo.IncrementVisits(bgCtx, url.ShortUrl)
	}()

	return url, nil
}

func generateShortURL() string {
	var sb strings.Builder

	sb.Grow(shortURLLength)

	for i := 0; i < shortURLLength; i++ {
		sb.WriteByte(symbols[rand.Intn(len(symbols))])
	}

	return sb.String()
}