package storage

import "time"

type Type string

const (
	TypeLocal Type = "local"
	TypeS3    Type = "s3"
)

type Storage struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      Type      `json:"type"`
	Endpoint  string    `json:"endpoint"`
	AccessKey string    `json:"accessKey"`
	SecretKey string    `json:"secretKey"`
	Bucket    string    `json:"bucket"`
	Region    string    `json:"region"`
	LocalPath string    `json:"localPath"`
	Quota     int64     `json:"quota"`
	Status    int       `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
