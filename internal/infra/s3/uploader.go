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

// s3の依存をまとめる部分。
type Config struct {
	/*S3
		└── bucket
		    └── key
			Bucket = prod-task-events
			s3://prod-task-events/raw/task/year=2026/month=02/day=14/hour=15/aaa.jsonl
	これにするとDWH / Athena / BigQueryが読みやすい。

	*/

	Bucket          string //保存場所
	Endpoint        string // MinIO: "http://localhost:9000", AWS: ""。AWSデプロイ環境はSDKがエンドポイント自動解決。
	Region          string //S3が置いてあるリージョン。例えばap-northeast-1
	AccessKeyID     string
	SecretAccessKey string
	//デプロイ環境は直接記述しない。IAM Role、EC2 / ECS / EKS のメタデータ経由、IRSA。SDKが自動解決。
	ForcePathStyle bool // MinIO は true 必須
	/*S3には2種類のURL方式がある。
	Virtual Hosted Style（AWS標準） https://bucket.s3.amazonaws.com/key
	Path Style https://s3.amazonaws.com/bucket/key
	MinIOはPath Style 必須。AWSはfalse*/
}

type Uploader struct {
	client *s3.Client
	bucket string
}

// s3クライアントのファクトリ。
func NewUploader(ctx context.Context, cfg Config) (*Uploader, error) {
	var opts []func(*awsconfig.LoadOptions) error
	//awsconfig.LoadOptions を受け取って error を返す関数のさらにスライス
	opts = append(opts, awsconfig.WithRegion(cfg.Region))
	//これはregionを設定する関数を作ってる。まだ設定しない。そして awsconfig.LoadDefaultConfigですべての関数を実行する。
	//これすることで重ね書きできる。
	/* opts = append(opts, awsconfig.WithRegion("A"))
		opts = append(opts, awsconfig.WithRegion("B"))
	の場合Bになる
	*/
	//regionは awsconfig（AWS共通設定ロード）側の設定で、endpoint は s3（S3クライアント）側の設定。
	//両方重ね書きするけどawsconfig系はawsconfig.LoadDefaultConfigで落とす。
	//S3クライアント系はs3.NewFromConfigで落とす
	/*awsconfig系は用意されたヘルパーがある項目もあるが、
	S3の endpoint は状況依存が強いから 無名関数で直接Optionsを書き換える形がよく使われる（同じパターン）。*/

	//鍵が与えられたなら(普通はSDKが自動でとるから与えられない)、それを使う
	//credentials.NewStaticCredentialsProvider。key系を返す関数。標準記法。重ね書き。
	//第三引数はsession token。普通はない
	if cfg.AccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...) //これAWSconfig系落とすというより重ね書きか
	//いや違う。デフォルトの探索をする+与えられた option 関数を実行する。読み込み->実行やから重ね書き
	if err != nil {
		return nil, fmt.Errorf("s3 load config: %w", err)
	}
	//無名関数。これもSDKが自動解決するからもし与えられてた場合(ローカル、MinIO、S3互換))。
	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		//awsconfig系では自動で定義してくれてたヘルパー関数を自分で定義するイメージ
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint) //SDKのフィールド型が*stringやかたaws.String使う。
			o.UsePathStyle = cfg.ForcePathStyle       //これもエンドポイント系設定。これはSDKのフィールド型bool。
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...) //clientにまとめる。これはs3系実行+まとめる。
	return &Uploader{client: client, bucket: cfg.Bucket}, nil
}

// S3にファイル保存する部分
func (u *Uploader) Upload(ctx context.Context, key string, body io.Reader) error {
	//さっき作ったclient使う
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{ //PutObjectはpost
		//一つ目の返り値はETagとかVersionId。今回は使わない
		/*Bucket *string
		Key    string*/
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil { //infra系はwrapして返す
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
		var notFound *types.NotFound //types.NotFoundの記法
		if isNotFoundError(err, notFound) {
			return false, nil
		} //notfoundは異常ではないから分類してerr=nilで返す。
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
	}) //tokenとか面倒なやつをNewListObjectsV2Paginatorが操作してくれる。
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3 list prefix=%s: %w", prefix, err)
		}
		for _, obj := range page.Contents {
			//page.Contentsページに含まれるオブジェクト一覧
			keys = append(keys, *obj.Key) //keyだけ格納
			//keyが分かったらget使える。
		}
	}
	return keys, nil
}

// Get は指定キーのオブジェクトを読み取り用に返す。呼び出し側が Close する。
func (u *Uploader) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := u.client.GetObject(ctx, &s3.GetObjectInput{ //get
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get object key=%s: %w", key, err)
	}
	return out.Body, nil
	//返り値はio.ReadCloser。普通いつもはデータ返して関数側で閉じてるけど今回は読み取りは受け取り側に任せてる。
	/*type ReadCloser interface {
	    Read(p []byte) (int, error)
	    Close() error
	}
	だからclose必須


	*/
	//S3はデータ大きい。全部返したら詰む。受け取り側がstreamで使う。
}
