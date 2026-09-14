package repository

import (
	"context"
	"errors"
	"url-shortener/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("url not found")

type URLRepository interface {
	Save(ctx context.Context, url *model.URL) error
	FindByShortURL(ctx context.Context, shortURL string) (*model.URL, error)
	IncrementVisits(ctx context.Context, shortURL string) error
}

type urlRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Save(ctx context.Context, url *model.URL) error {
	if url.ID == "" {
		url.ID = uuid.NewString()
	}

	return r.db.WithContext(ctx).Create(url).Error
}

func (r *urlRepository) FindByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	var url model.URL

	if err := r.db.WithContext(ctx).Where("short_url = ?", shortURL).First(&url).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &url, nil
}

func (r *urlRepository) IncrementVisits(ctx context.Context, shortURL string) error {
	res := r.db.WithContext(ctx).
		Model(&model.URL{}).
		Where("short_url = ?", shortURL).
		UpdateColumn("visits", gorm.Expr("visits + ?", 1))

	return res.Error
}