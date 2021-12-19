package client

import (
	"backupagent/config"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	s3p "github.com/aws/aws-sdk-go/service/s3"
)

type S3client struct {
	uploader   *s3manager.Uploader
	client	   *s3p.S3
	bucket     string
}

func NewS3client(cfg *config.Config) (*S3client, error) {

	creds := credentials.NewStaticCredentials(cfg.S3.Key, cfg.S3.Secret, "")
	sess, err := session.NewSession(aws.NewConfig().
		WithCredentials(creds).
		WithEndpoint(cfg.S3.Endpoint).
		WithRegion(cfg.S3.Region))
	if err != nil {
		return nil, err
	}
	buf := int64(cfg.S3.PartSizeMB * 1024 * 1024)
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = buf
	})

	client := s3p.New(sess)
	return &S3client{uploader: uploader, client: client, bucket: cfg.S3.Bucket}, nil
}

func (s3 *S3client) Upload(backupName string, r *io.Reader) error {

	_, err := s3.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.bucket),
		Key:    aws.String(backupName),
		Body:   *r,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s3 *S3client) Download(backupName string) (*s3p.GetObjectOutput, error) {

	res,err := s3.client.GetObject(&s3p.GetObjectInput{
		Bucket: aws.String(s3.bucket),
		Key:    aws.String(backupName),
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
