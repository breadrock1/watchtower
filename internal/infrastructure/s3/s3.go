package s3

import (
	"bytes"
	"context"
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
	prefix       = ""
	eventsFilter = []string{
		"s3:ObjectCreated:Post",
		"s3:ObjectCreated:Put",
		"s3:ObjectCreated:Copy",
		"s3:ObjectRemoved:Delete",
		"s3:BucketCreated:*",
		"s3:BucketRemoved:*",
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

	s3Client, err := minio.New(config.Address, opts)
	if err != nil {
		return nil, fmt.Errorf("failed while connecting to s3: %w", err)
	}

	client := &S3Client{
		mc:          s3Client,
		bindBuckets: &sync.Map{},
		eventsCh:    make(chan dto.TaskEvent),
	}

	return client, nil
}

func (s *S3Client) GetEventsChannel() chan dto.TaskEvent {
	return s.eventsCh
}

func (s *S3Client) GetWatchedDirs(_ context.Context) ([]dto.Directory, error) {
	var watchedDirs []dto.Directory
	s.bindBuckets.Range(func(k, v interface{}) bool {
		dir := v.(dto.Directory)
		watchedDirs = append(watchedDirs, dir)
		return true
	})

	return watchedDirs, nil
}

func (s *S3Client) AttachWatchedDir(_ context.Context, dir dto.Directory) error {
	s.bindBuckets.Store(dir.Bucket, dir)
	return nil
}

func (s *S3Client) DetachWatchedDir(_ context.Context, path string) error {
	s.bindBuckets.Delete(path)
	return nil
}

func (s *S3Client) LaunchWatcher(ctx context.Context, dirs []dto.Directory) error {
	for _, dir := range dirs {
		err := s.AttachWatchedDir(ctx, dir)
		if err != nil {
			log.Printf("failed to launch watcher for directory %s: %v", dir.Bucket, err)
		}
	}

	go func() {
		if err := s.startBucketListener(ctx); err != nil {
			log.Printf("failed to start bucket listener: %v", err)
		}

		<-ctx.Done()
		if err := s.TerminateWatcher(ctx); err != nil {
			log.Printf("failed to terminate s3 watchers: %v", err)
		}
	}()

	return nil
}

func (s *S3Client) TerminateWatcher(_ context.Context) error {
	return nil
}

func (s *S3Client) startBucketListener(ctx context.Context) error {
	for event := range s.mc.ListenNotification(ctx, prefix, suffix, eventsFilter) {
		if event.Err != nil {
			log.Printf("caughet error event: %v", event.Err)
			continue
		}

		var eventType dto.EventType
		for _, record := range event.Records {
			s3Object := record.S3
			bucketName := s3Object.Bucket.Name

			_, ok := s.bindBuckets.Load(bucketName)

			switch record.EventName {
			case "s3:BucketCreated:*":
				eventType = dto.CreateBucket

			case "s3:BucketRemoved:*":
				eventType = dto.DeleteBucket

			case "s3:ObjectCreated:Put":
				if !ok {
					continue
				}
				eventType = dto.CreateFile

			case "s3:ObjectCreated:Post":
				if !ok {
					continue
				}
				eventType = dto.CreateFile

			case "s3:ObjectCreated:Copy":
				if !ok {
					continue
				}
				eventType = dto.CopyFile

			case "s3:ObjectRemoved:Delete":
				if !ok {
					continue
				}
				eventType = dto.DeleteFile

			default:
				log.Printf("unknown event: %v", record.EventName)
				continue
			}

			log.Printf("[%s]: s3 event type: %s", bucketName, record.EventName)

			s.eventsCh <- dto.TaskEvent{
				Id:         uuid.New(),
				Bucket:     bucketName,
				FilePath:   s3Object.Object.Key,
				FileSize:   s3Object.Object.Size,
				CreatedAt:  time.Now(),
				ModifiedAt: time.Now(),
				Status:     dto.Pending,
				EventType:  eventType,
			}
		}

	}

	return nil
}

func (s *S3Client) UploadFile(ctx context.Context, bucket, filePath string, data *bytes.Buffer) error {
	opts := minio.PutObjectOptions{}
	dataLen := int64(data.Len())
	_, err := s.mc.PutObject(ctx, bucket, filePath, data, dataLen, opts)
	if err != nil {
		return fmt.Errorf("failed to upload file to s3: %w", err)
	}
	return nil
}

func (s *S3Client) DownloadFile(ctx context.Context, bucket, filePath string) (bytes.Buffer, error) {
	var objBody bytes.Buffer

	opts := minio.GetObjectOptions{}
	obj, err := s.mc.GetObject(ctx, bucket, filePath, opts)
	if err != nil {
		return objBody, fmt.Errorf("failed to get object from s3: %w", err)
	}

	_, err = objBody.ReadFrom(obj)
	if err != nil {
		return objBody, fmt.Errorf("failed to read loaded object from s3: %w", err)
	}

	return objBody, nil
}

func (s *S3Client) CreateBucket(ctx context.Context, bucket string) error {
	opts := minio.MakeBucketOptions{}
	return s.mc.MakeBucket(ctx, bucket, opts)
}
