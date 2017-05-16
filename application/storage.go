package application

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/thoas/gostorages"
)

// Storage holds the information for connecting to a storage backend
type Storage struct {
	Bucket   string `yaml:"bucket"`
	URL      string `yaml:"url"`
	Key      string `yaml:"key"`
	Secret   string `yaml:"secret"`
	Region   string `yaml:"region"`
	Root     string `yaml:"root"`
	CacheAge int    `yaml:"cache_age"`
	s3       gostorages.Storage
}

// S3 connects to an S3 storage backend
func (s *Storage) S3() gostorages.Storage {
	// TODO: only create a new one if one doesn't exist
	storage := gostorages.NewS3Storage(s.Key, s.Secret, s.Bucket, "", aws.Regions[s.Region], s3.PublicReadWrite, s.URL)

	return storage
}
