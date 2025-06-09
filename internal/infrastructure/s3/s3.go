package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"watchtower/internal/application/dto"
)

var (
	suffix       = ""
	eventsFilter = []string{
		"s3:ObjectCreated:*",
		"s3:ObjectRemoved:*",
	}
)

type S3Client struct {
	eventsCh    chan dto.TaskEvent
	mc          *minio.Client
	bindBuckets *sync.Map
}

func New(config *Config) (*S3Client, error) {
	creds := credentials.NewStaticV4(config.AccessID, config.SecretKey, config.Token)
	opts := &minio.Options{
		Creds:  creds,
		Secure: config.EnableSSL,
	}

	mc, err := minio.New(config.Address, opts)
	if err != nil {
		return nil, fmt.Errorf("failed while connecting to s3: %v", err)
	}

	client := &S3Client{
		mc:          mc,
		bindBuckets: &sync.Map{},
		eventsCh:    make(chan dto.TaskEvent),
	}

	return client, nil
}

func (s *S3Client) GetEventsChannel() chan dto.TaskEvent {
	return s.eventsCh
}

func (s *S3Client) GetWatchedDirs(ctx context.Context) ([]dto.Directory, error) {
	buckets, err := s.mc.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	directories := make([]dto.Directory, len(buckets))
	for index, info := range buckets {
		directories[index] = dto.Directory{
			Bucket:    info.Name,
			Path:      "/",
			CreatedAt: info.CreationDate,
		}
	}

	return directories, nil
}

func (s *S3Client) AttachWatchedDir(_ context.Context, dir dto.Directory) error {
	_, ok := s.bindBuckets.Load(dir)
	if ok {
		return fmt.Errorf("directory already attached: %s", dir.Path)
	}

	go func() {
		err := s.startBucketListener(dir)
		if err != nil {
			log.Printf("failed to start bucket listener: %v", err)
		}
	}()

	return nil
}

func (s *S3Client) DetachWatchedDir(_ context.Context, path string) error {
	ch, ok := s.bindBuckets.Load(path)
	if !ok {
		return errors.New("there is no such bucket to detach")
	}

	cancel := ch.(context.CancelFunc)
	cancel()

	return nil
}

func (s *S3Client) LaunchWatcher(ctx context.Context, dirs []dto.Directory) error {
	var err error
	for _, dir := range dirs {
		go func() {
			if err = s.startBucketListener(dir); err != nil {
				log.Printf("failed to start bucket listener for %s: %v", dir.Bucket, err)
			}

			<-ctx.Done()
			if err = s.TerminateWatcher(ctx); err != nil {
				log.Printf("failed to terminate s3 watchers: %v", err)
			}
		}()
	}

	return nil
}

func (s *S3Client) TerminateWatcher(_ context.Context) error {
	s.bindBuckets.Range(func(key, value interface{}) bool {
		value.(context.CancelFunc)()
		return true
	})
	return nil
}

func (s *S3Client) startBucketListener(dir dto.Directory) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.bindBuckets.Store(dir, cancel)
	defer func() {
		s.bindBuckets.Delete(dir)
	}()

	for event := range s.mc.ListenBucketNotification(ctx, dir.Bucket, dir.Path, suffix, eventsFilter) {
		if event.Err == nil {
			for _, record := range event.Records {
				s3Object := record.S3
				bucketName := s3Object.Bucket.Name
				s.eventsCh <- dto.TaskEvent{
					Id:         uuid.New(),
					Bucket:     bucketName,
					FilePath:   s3Object.Object.Key,
					FileSize:   s3Object.Object.Size,
					CreatedAt:  time.Now(),
					ModifiedAt: time.Now(),
					Status:     dto.Received,
				}
			}

		}
	}

	return nil
}

func (s *S3Client) UploadFile(ctx context.Context, bucket, filePath string, data *bytes.Buffer) error {
	opts := minio.PutObjectOptions{}
	dataLen := int64(data.Len())
	_, err := s.mc.PutObject(ctx, bucket, filePath, data, dataLen, opts)
	return err
}

func (s *S3Client) DownloadFile(ctx context.Context, bucket, filePath string) (bytes.Buffer, error) {
	var objBody bytes.Buffer

	opts := minio.GetObjectOptions{}
	obj, err := s.mc.GetObject(ctx, bucket, filePath, opts)
	if err != nil {
		return objBody, err
	}

	_, err = objBody.ReadFrom(obj)
	if err != nil {
		return objBody, err
	}

	return objBody, nil
}

func (s *S3Client) CreateBucket(ctx context.Context, bucket string) error {
	opts := minio.MakeBucketOptions{}
	return s.mc.MakeBucket(ctx, bucket, opts)
}
