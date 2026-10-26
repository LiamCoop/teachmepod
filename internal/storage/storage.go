// Package storage talks to the audio bucket. The bucket is private; listeners
// get time-limited presigned URLs, never the bucket itself.
//
// Configuration comes from the standard AWS variables (AWS_ENDPOINT_URL_S3,
// AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY), which Railway fills
// in from the bucket, plus AUDIO_BUCKET for the bucket name.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Store struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
}

// FromEnv builds a Store from the environment.
func FromEnv(ctx context.Context) (*Store, error) {
	bucket := os.Getenv("AUDIO_BUCKET")
	if bucket == "" {
		return nil, errors.New("AUDIO_BUCKET is not set")
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &Store{bucket: bucket, client: client, presign: s3.NewPresignClient(client)}, nil
}

// Put uploads one object.
func (s *Store) Put(ctx context.Context, key, contentType string, body io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put %s: %w", key, err)
	}
	return nil
}

// Delete removes one object. Deleting a missing key is not an error.
func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}

// PresignGet returns a URL that downloads key directly from the bucket until ttl passes.
func (s *Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign %s: %w", key, err)
	}
	return req.URL, nil
}
