package model

import "time"

type URL struct {
	ID          string		`gorm:"primaryKey"`
	OriginalUrl	string		`gorm:"type:text;not null"`
	ShortUrl		string		`gorm:"type:varchar(20);uniqueIndex;not null"`
	CreatedAt		time.Time	`gorm:"autoCreateTime"`
	Visits			uint			`gorm:"default:0"`
}