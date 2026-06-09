package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/boxingoctopus/snackmates/api/internal/config"
)

type Client struct {
	s3              *s3.Client
	clientBucket    string
	staticBucket    string
	endpoint        string
	publicBaseURL   string
	usePathStyle    bool
	presignPrivate  bool
	presignExpiry   time.Duration
}

func New(cfg config.Config) (*Client, error) {
	endpoint := strings.TrimSpace(cfg.S3Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("s3 endpoint is required")
	}
	if strings.TrimSpace(cfg.S3Region) == "" {
		return nil, fmt.Errorf("s3 region is required")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = cfg.S3UsePathStyle
		o.Region = cfg.S3Region
	})

	expiry := time.Duration(cfg.S3PresignExpirySeconds) * time.Second
	if expiry <= 0 {
		expiry = time.Hour
	}

	return &Client{
		s3:             s3Client,
		clientBucket:   cfg.S3ClientBucket,
		staticBucket:   cfg.S3StaticBucket,
		endpoint:       strings.TrimRight(endpoint, "/"),
		publicBaseURL:  strings.TrimRight(strings.TrimSpace(cfg.S3PublicBaseURL), "/"),
		usePathStyle:   cfg.S3UsePathStyle,
		presignPrivate: cfg.S3PresignPrivateObjects,
		presignExpiry:  expiry,
	}, nil
}

func (c *Client) UploadAvatar(ctx context.Context, key string, body io.Reader, contentType string) error {
	return c.putObject(ctx, c.clientBucket, key, body, contentType)
}

func (c *Client) UploadBanner(ctx context.Context, key string, body io.Reader, contentType string) error {
	return c.putObject(ctx, c.clientBucket, key, body, contentType)
}

// ResolveObjectURL returns a browser-accessible URL for a stored key or external URL.
func (c *Client) ResolveObjectURL(ctx context.Context, storedURL, key *string) (string, error) {
	return c.ResolveAvatarURL(ctx, storedURL, key)
}

func (c *Client) UploadStatic(ctx context.Context, key string, body io.Reader, contentType string) error {
	return c.putObject(ctx, c.staticBucket, key, body, contentType)
}

func (c *Client) putObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

// ObjectURL builds a public object URL for S3-compatible providers (Cloudflare R2, MinIO, etc.).
func (c *Client) ObjectURL(bucket, key string) string {
	return buildObjectURL(c.urlBase(), bucket, key, c.usePathStyle)
}

func (c *Client) AvatarURL(key string) string {
	return c.ObjectURL(c.clientBucket, key)
}

func (c *Client) StaticURL(key string) string {
	return c.ObjectURL(c.staticBucket, key)
}

func (c *Client) urlBase() string {
	if c.publicBaseURL != "" {
		return c.publicBaseURL
	}
	return c.endpoint
}

// ResolveAvatarURL returns a browser-accessible avatar URL.
// Discord and other external URLs pass through; S3 keys are presigned when configured.
func (c *Client) ResolveAvatarURL(ctx context.Context, storedURL, key *string) (string, error) {
	if key != nil && strings.TrimSpace(*key) != "" {
		if c.presignPrivate {
			return c.PresignObject(ctx, c.clientBucket, *key, c.presignExpiry)
		}
		return c.AvatarURL(*key), nil
	}
	if storedURL != nil {
		return strings.TrimSpace(*storedURL), nil
	}
	return "", nil
}

func (c *Client) PresignObject(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = c.presignExpiry
	}
	presigner := s3.NewPresignClient(c.s3)
	out, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) PresignAvatar(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return c.PresignObject(ctx, c.clientBucket, key, expiry)
}

func buildObjectURL(base, bucket, key string, pathStyle bool) string {
	base = strings.TrimRight(base, "/")
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return base
	}
	if pathStyle {
		return fmt.Sprintf("%s/%s/%s", base, bucket, key)
	}
	return virtualHostedObjectURL(base, bucket, key)
}

func virtualHostedObjectURL(endpoint, bucket, key string) string {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(endpoint, "/"), bucket, key)
	}
	parsed.Host = bucket + "." + parsed.Host
	parsed.Path = path.Join("/", key)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}
