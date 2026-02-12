package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/jus1d/kypidbot/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	bucket string
}

func New(c *config.S3) (*Client, error) {
	url := fmt.Sprintf("%s:%s", c.Host, c.Port)
	client, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKeyID, c.SecretAccessKey, ""),
		Secure: c.UseSSL,
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		bucket: c.Bucket,
	}, nil
}

func (c *Client) GetPhoto(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
