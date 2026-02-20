package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Bucket          string
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string

	ForcePathStyle bool
}

type Uploader struct {
	client *s3.Client
	bucket string
}

func NewUploader(ctx context.Context, cfg Config) (*Uploader, error) {
	var opts []func(*awsconfig.LoadOptions) error

	opts = append(opts, awsconfig.WithRegion(cfg.Region))

	if cfg.AccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)

	if err != nil {
		return nil, fmt.Errorf("s3 load config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {

		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.ForcePathStyle
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	return &Uploader{client: client, bucket: cfg.Bucket}, nil
}

func (u *Uploader) Upload(ctx context.Context, key string, body io.Reader) error {

	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{

		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("s3 put object key=%s: %w", key, err)
	}
	return nil
}

func (u *Uploader) Exists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		//headobject:=データは取らずに、存在とメタ情報だけ確認するAPI
		//httpでいうところのHEAD。今回はGet使わない。重い。
		//存在確認->HeadObject
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if isNotFoundError(err, notFound) {
			return false, nil
		}
		return false, fmt.Errorf("s3 head object key=%s: %w", key, err)
	}
	return true, nil
}

// List は指定プレフィックス配下の全"キー"だけを返す。すべては返さない。APIの責務を分けてる。
// データが重いからS3では責務を分けて無駄なデータを取らないってのにフォーカスしてる。
// S3のList APIは 1回で全件返さない->Paginator
func (u *Uploader) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(u.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(u.bucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3 list prefix=%s: %w", prefix, err)
		}
		for _, obj := range page.Contents {

			keys = append(keys, *obj.Key)

		}
	}
	return keys, nil
}

func (u *Uploader) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := u.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get object key=%s: %w", key, err)
	}
	return out.Body, nil

}
