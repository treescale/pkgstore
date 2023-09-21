package storage

import (
	"encoding/json"
	"github.com/alin-io/pkgproxy/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
)

type S3Backend struct {
	BaseStorageBackend
	s3Session *session.Session
	s3        *s3.S3

	Bucket string
}

func NewS3Backend() *S3Backend {
	s, err := session.NewSession(&aws.Config{
		Endpoint:         &config.Get().Storage.S3.ApiHost,
		Region:           &config.Get().Storage.S3.Region,
		Credentials:      credentials.NewStaticCredentials(config.Get().Storage.S3.ApiKey, config.Get().Storage.S3.ApiSecret, ""),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	return &S3Backend{
		s3Session: s,
		s3:        s3.New(s),

		Bucket: config.Get().Storage.S3.Bucket,
	}
}

func (s *S3Backend) WriteFile(key string, fileMeta interface{}, r io.ReadSeeker) error {
	metadata := make(map[string]*string)
	if fileMeta != nil {
		metadataBuffer, err := json.Marshal(fileMeta)
		if err != nil {
			return err
		}
		err = json.Unmarshal(metadataBuffer, &metadata)
		if err != nil {
			return err
		}
	}
	putObjectInput := &s3.PutObjectInput{
		Bucket:   aws.String(s.Bucket),
		Key:      aws.String(key),
		Body:     r,
		Metadata: metadata,
	}
	_, err := s.s3.PutObject(putObjectInput)
	return err
}

func (s *S3Backend) GetFile(key string) (io.ReadCloser, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}
	obj, err := s.s3.GetObject(getObjectInput)
	if err != nil {
		return nil, err
	}
	return obj.Body, nil
}

func (s *S3Backend) DeleteFile(key string) error {
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}
	_, err := s.s3.DeleteObject(deleteObjectInput)
	return err
}

func (s *S3Backend) GetMetadata(key string, value interface{}) error {
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}
	obj, err := s.s3.GetObject(getObjectInput)
	if err != nil {
		return err
	}
	defer obj.Body.Close()
	metadataBuffer, err := json.Marshal(obj.Metadata)
	if err != nil {
		return err
	}
	err = json.Unmarshal(metadataBuffer, value)
	return err
}
