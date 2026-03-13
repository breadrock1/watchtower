package domain

import (
	"net/url"

	"watchtower/internal/shared/kernel"
)

// ICloudStorage defines the complete interface for cloud storage operations.
// It combines bucket management, object operations, and sharing capabilities
// into a unified API.
type ICloudStorage interface {
	IBucketManager
	IObjectManager
	IObjectWalker
	IShareManager
}

// IBucketManager defines operations for managing storage buckets/containers.
// Buckets are the top-level organizational units in cloud storage.
type IBucketManager interface {
	// GetAllBuckets retrieves a list of all buckets available in the storage system.
	// Returns a slice of Bucket objects or an error if the operation fails.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//
	// Returns:
	//   - []Bucket: List of all buckets
	//   - error: ErrUnauthorized if credentials are invalid,
	//            ErrServiceUnavailable if cloud provider is unreachable,
	//            or other provider-specific errors
	//
	// Example:
	//   buckets, err := storage.GetAllBuckets(ctx)
	//   for _, bucket := range buckets {
	//       fmt.Printf("Bucket: %s, Created: %s\n", bucket.ID, bucket.CreatedAt)
	//   }
	GetAllBuckets(ctx kernel.Ctx) ([]Bucket, error)

	// IsBucketExist checks if a bucket with the given ID exists.
	// This is useful for validation before performing bucket operations.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket to check
	//
	// Returns:
	//   - bool: true if bucket exists, false otherwise
	//   - error: ErrUnauthorized if credentials are invalid, or other provider-specific errors
	//
	// Example:
	//   exists, err := storage.IsBucketExist(ctx, "my-app-data")
	//   if !exists {
	//       // Create bucket or handle missing bucket
	//   }
	IsBucketExist(ctx kernel.Ctx, bucketID kernel.BucketID) (bool, error)

	// CreateBucket creates a new bucket with the specified ID.
	// Bucket IDs must be globally unique across the storage system.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: Unique identifier for the new bucket
	//
	// Returns:
	//   - error: ErrBucketAlreadyExists if bucket ID is taken,
	//            ErrInvalidBucketID if ID format is invalid,
	//            ErrQuotaExceeded if account limit reached,
	//            or other provider-specific errors
	//
	// Example:
	//   err := storage.CreateBucket(ctx, "user-uploads-2024")
	//   if errors.Is(err, ErrBucketAlreadyExists) {
	//       // Handle existing bucket
	//   }
	CreateBucket(ctx kernel.Ctx, bucketID kernel.BucketID) error

	// DeleteBucket removes an existing bucket and all its contents.
	// This operation is irreversible and should be used with caution.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket to delete
	//
	// Returns:
	//   - error: ErrBucketNotFound if bucket doesn't exist,
	//            ErrBucketNotEmpty if bucket still contains objects,
	//            or other provider-specific errors
	//
	// Note: Some providers require the bucket to be empty before deletion.
	DeleteBucket(ctx kernel.Ctx, bucketID kernel.BucketID) error
}

// IObjectManager defines operations for managing individual objects/files
// within buckets. This includes CRUD operations and data access.
type IObjectManager interface {
	// GetObjectInfo retrieves metadata about an object without downloading its content.
	// Useful for checking object properties, size, or last modified time.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket containing the object
	//   - objID: ID or path of the object to inspect
	//
	// Returns:
	//   - Object: Complete object metadata (without data)
	//   - error: ErrObjectNotFound if object doesn't exist,
	//            ErrBucketNotFound if bucket doesn't exist,
	//            or other provider-specific errors
	//
	// Example:
	//   info, err := storage.GetObjectInfo(ctx, "documents", "reports/annual.pdf")
	//   if err == nil {
	//       fmt.Printf("Size: %d bytes, Modified: %s\n", info.Size, info.LastModified)
	//   }
	GetObjectInfo(ctx kernel.Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) (Object, error)

	// GetObjectData retrieves both the object metadata and its content.
	// The object data is returned as a buffer that can be read or streamed.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket containing the object
	//   - objID: ID or path of the object to download
	//
	// Returns:
	//   - ObjectData: Buffer containing the object's content
	//   - error: ErrObjectNotFound if object doesn't exist,
	//            ErrBucketNotFound if bucket doesn't exist,
	//            or other provider-specific errors
	//
	// Example:
	//   data, err := storage.GetObjectData(ctx, "images", "profile.jpg")
	//   if err == nil {
	//       imgBytes := data.Bytes()
	//       // Process image data...
	//   }
	GetObjectData(ctx kernel.Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) (ObjectData, error)

	// StoreObject uploads a new object or replaces an existing one.
	// If an object already exists at the specified path, it will be overwritten.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket to store the object in
	//   - params: Upload parameters including file path, data, and options
	//
	// Returns:
	//   - ObjectID: Unique identifier for the stored object
	//   - error: ErrBucketNotFound if bucket doesn't exist,
	//            ErrInvalidFilePath if path format is invalid,
	//            ErrQuotaExceeded if bucket or account limit reached,
	//            or other provider-specific errors
	//
	// Example:
	//   data := bytes.NewBuffer([]byte("file content"))
	//   params := &UploadObjectParams{
	//       FilePath: "uploads/notes.txt",
	//       FileData: data,
	//       ContentType: "text/plain",
	//       Metadata: map[string]string{"author": "john"},
	//   }
	//   objID, err := storage.StoreObject(ctx, "my-bucket", params)
	StoreObject(ctx kernel.Ctx, bucketID kernel.BucketID, params *UploadObjectParams) (kernel.ObjectID, error)

	// CopyObject duplicates an object from one location to another within the same bucket.
	// This operation is often more efficient than download+upload for large files.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket containing both source and destination
	//   - params: Copy parameters with source and destination paths
	//
	// Returns:
	//   - error: ErrObjectNotFound if source doesn't exist,
	//            ErrInvalidPath if destination path is invalid,
	//            or other provider-specific errors
	//
	// Example:
	//   params := &CopyObjectParams{
	//       SourcePath: "originals/document.pdf",
	//       DestinationPath: "backups/document-2024.pdf",
	//   }
	//   err := storage.CopyObject(ctx, "documents", params)
	CopyObject(ctx kernel.Ctx, bucketID kernel.BucketID, params *CopyObjectParams) error

	// DeleteObject permanently removes an object from storage.
	// This operation cannot be undone.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket containing the object
	//   - objID: ID or path of the object to delete
	//
	// Returns:
	//   - error: ErrObjectNotFound if object doesn't exist (idempotent),
	//            ErrBucketNotFound if bucket doesn't exist,
	//            or other provider-specific errors
	//
	// Example:
	//   err := storage.DeleteObject(ctx, "temp-files", "cache/session-123.tmp")
	//   if err != nil && !errors.Is(err, ErrObjectNotFound) {
	//       // Handle error, but ignore "not found" as it's already gone
	//   }
	DeleteObject(ctx kernel.Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) error
}

// IObjectWalker defines operations for listing and iterating through objects in a bucket.
// This is useful for directory listings, backups, and batch operations.
type IObjectWalker interface {
	// GetBucketObjects retrieves a list of objects in a bucket, optionally filtered by prefix.
	// This implements directory-like listing functionality.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket to list objects from
	//   - params: Filtering and pagination parameters
	//
	// Returns:
	//   - []Object: Slice of object metadata matching the criteria
	//   - error: ErrBucketNotFound if bucket doesn't exist, or other provider-specific errors
	//
	// Example:
	//   params := &GetObjectsParams{
	//       PrefixPath: "images/2024/",
	//   }
	//   objects, err := storage.GetBucketObjects(ctx, "media", params)
	//   for _, obj := range objects {
	//       fmt.Printf("Found: %s (%d bytes)\n", obj.Path, obj.Size)
	//   }
	GetBucketObjects(ctx kernel.Ctx, bucketID kernel.BucketID, params *GetObjectsParams) ([]Object, error)
}

// IShareManager defines operations for generating temporary access URLs to objects.
// This enables secure sharing of private objects without making them public.
type IShareManager interface {
	// GenShareURL generates a time-limited URL that provides access to a private object.
	// The URL includes authentication tokens and expires after the specified duration.
	//
	// Parameters:
	//   - kernel.Ctx: Context for cancellation and timeout
	//   - bucketID: ID of the bucket containing the object
	//   - params: Sharing parameters including file path and expiration
	//
	// Returns:
	//   - *url.URL: Pre-signed URL that grants temporary access
	//   - error: ErrObjectNotFound if object doesn't exist,
	//            ErrInvalidExpiration if duration is invalid,
	//            or other provider-specific errors
	//
	// Example:
	//   params := &ShareObjectParams{
	//       FilePath: "shared/report.pdf",
	//       Expired: 24 * time.Hour,
	//   }
	//   shareURL, err := storage.GenShareURL(ctx, "documents", params)
	//   if err == nil {
	//       fmt.Printf("Shareable link (valid for 24h): %s\n", shareURL.String())
	//   }
	GenShareURL(ctx kernel.Ctx, bucketID kernel.BucketID, params *ShareObjectParams) (*url.URL, error)
}
